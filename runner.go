package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
)

func queryMap(res []byte) drm.MapData {
	var resultMap drm.MapData
	json.Unmarshal(res, &resultMap)

	return resultMap
}

func RunCommands(command string, args map[string]string, db *sql.DB) {
	switch command {
	case "add":
		// query map
		mapToQueue := queryMap(drm.RequestDRM("query", args["id"]))

		// grab user
		user := args["username"]

		// grab number of user requests
		userNumRequests := database.GetUserRequests(user, db)

		// put map through filters
		mapToQueue, err := FilterMap(mapToQueue, user, userNumRequests, db)

		if err != nil {
			// print a message
			fmt.Printf("{\"message\": \"%s\"}", err)
		} else {
			// add to queue
			addKeyArgs := mapToQueue.BsrKey + "?user=" + user
			drm.RequestDRM("addKey", addKeyArgs)

			// increment user requests
			userNumRequests++
			database.SetUserRequests(user, userNumRequests, db)

			// ...grab the request limit
			requestLimit := viper.GetInt("bsr.request-limit")

			// and then, print the message
			fmt.Printf("\"message\": \"Added to queue! You have %d requests left.\"", requestLimit-userNumRequests)
		}

	case "mtt":
		// grab user
		user := args["username"]
		// find the user's last request
		positions := drm.WhereDRM(user)
		pos := fmt.Sprintf("%d", positions[len(positions)-1])

		// move the request to top
		drm.RequestDRM("queue", "move/"+pos+"/1")

		// print confirmation message
		fmt.Printf("\"message\": \"Moved request to the top of the queue.\"")

	case "wip":
		// grab WIP ID and username
		id := args["id"]
		user := args["username"]

		// add the WIP
		drm.RequestDRM("addWip", id+"?user"+user)

		// and then, print the message
		fmt.Printf("\"message\": \"Added WIP %s.\"", id)

	case "getqueue":
		// grab username
		user := args["username"]

		// grab positions of requests
		positions := drm.WhereDRM(user)

		// create long string
		message := "Your map"
		if len(positions) > 1 {
			stringBuilder := ""

			for idx, position := range positions {
				if idx < len(positions) {
					stringBuilder += fmt.Sprintf("%d, ", position)
				} else {
					stringBuilder += fmt.Sprintf("and %d.", position)
				}
			}

			message += "s are in positions " + stringBuilder
		} else if len(positions) == 1 {
			message += fmt.Sprintf(" is in position %d.", positions[0])
		} else {
			message = "You have no maps in queue."
		}

		// print message
		fmt.Printf("\"message\": \"%s\"", message)

	case "ban":
		// query map
		mapToBan := queryMap(drm.RequestDRM("query", args["id"]))

		// ban map
		database.BanMap(mapToBan.BsrKey, mapToBan.Hash, db)

		// print message
		fmt.Printf("\"message\": \"%s is now banned from being requested.\"", mapToBan.BsrKey)

	case "unban":
		// query map
		mapToUnban := queryMap(drm.RequestDRM("query", args["id"]))

		// unban map
		database.BanMap(mapToUnban.BsrKey, mapToUnban.Hash, db)

		// print message
		fmt.Printf("\"message\": \"%s is now banned from being requested.\"", mapToUnban.BsrKey)
	}
}
