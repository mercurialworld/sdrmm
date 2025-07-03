package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
)

func queryMap(res []byte) drm.MapData {
	var resultMap drm.MapData
	json.Unmarshal(res, &resultMap)

	return resultMap
}

func queryMaps(res []byte) []drm.MapData {
	var resQueue []drm.MapData
	json.Unmarshal(res, &resQueue)

	return resQueue
}

func RunCommands(command string, args map[string]string, db *sql.DB) {
	switch command {
	case "new":
		// clear queue
		drm.RequestDRM("clear", "")

		// clear requests
		database.ClearRequestLimits(db)

		// close queue
		drm.RequestDRM("queue", "open/false")

		// get current date and time
		currentTime := time.Now()

		// add new entry in db
		database.NewSession(currentTime, db)

		// print confirmation message
		fmt.Printf("New session created! Queue has been cleared and closed.")

	case "add":
		// query map
		mapToQueue := queryMap(drm.RequestDRM("query", args["id"]))

		// grab user
		user := args["username"]

		// grab modadd
		modadd, _ := strconv.ParseBool(args["modadd"])

		// grab number of user requests
		userNumRequests := database.GetUserRequests(user, db)

		// put map through filters
		mapToQueue, err := FilterMap(mapToQueue, user, userNumRequests, modadd, db)

		if err != nil {
			// print a message
			fmt.Printf("%s", err)
		} else {
			// add to queue
			addKeyArgs := mapToQueue.BsrKey + "?user=" + user
			drm.RequestDRM("addKey", addKeyArgs)

			// ...grab the request limit
			requestLimit := viper.GetInt("bsr.request-limit")

			message := "Added to queue!"

			if requestLimit > 0 {
				// increment user requests
				userNumRequests++
				database.SetUserRequests(user, userNumRequests, db)

				message += fmt.Sprintf(" You have %d requests left.", requestLimit-userNumRequests)
			}

			// and then, print the message
			fmt.Printf("%s", message)
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
		fmt.Printf("Moved request to the top of the queue.")

	case "wip":
		// grab WIP ID and username
		id := args["id"]
		user := args["username"]

		// add the WIP
		drm.RequestDRM("addWip", id+"?user"+user)

		// and then, print the message
		fmt.Printf("Added WIP %s.", id)

	case "getqueue":
		// grab username
		user := args["username"]

		// grab positions of requests
		positions := drm.WhereDRM(user)

		// get entire queue
		queue := queryMaps(drm.GetDRMQueue())

		// create long string
		message := fmt.Sprintf("There are %d maps in queue!", len(queue))
		if user != "" {
			if len(positions) > 1 {
				stringBuilder := " Your maps are in positions "

				for idx, position := range positions {
					if idx < len(positions) {
						stringBuilder += fmt.Sprintf("%d, ", position)
					} else {
						stringBuilder += fmt.Sprintf("and %d.", position)
					}
				}

				message += stringBuilder
			} else if len(positions) == 1 {
				message += fmt.Sprintf(" Your map is in position %d.", positions[0])
			} else {
				message += " You have no maps in queue."
			}
		}

		// print message
		fmt.Printf("%s", message)

	case "ban":
		// query map
		mapToBan := queryMap(drm.RequestDRM("query", args["id"]))

		// ban map
		database.BanMap(mapToBan.BsrKey, mapToBan.Hash, db)

		// print message
		fmt.Printf("%s is now banned from being requested.", mapToBan.BsrKey)

	case "unban":
		// query map
		mapToUnban := queryMap(drm.RequestDRM("query", args["id"]))

		// unban map
		database.BanMap(mapToUnban.BsrKey, mapToUnban.Hash, db)

		// print message
		fmt.Printf("%s can now be requested again.", mapToUnban.BsrKey)

	case "queuestatus":
		// get queue status
		queueStatus := database.GetQueueStatus(db)

		statusString := "closed"

		if queueStatus {
			statusString = "open"
		}

		fmt.Printf("The queue is currently %s.", statusString)

	case "togglequeue":
		// get queue status
		newQueueStatus := !database.GetQueueStatus(db)
		if args != nil {
			newQueueStatus, _ = strconv.ParseBool(args["status"])
		}

		// set the queue to whatever we get
		database.SetQueueStatus(newQueueStatus, db)

		// format/print string
		statusString := "closed"

		if newQueueStatus {
			statusString = "open"
		}

		fmt.Printf("The queue is now %s.", statusString)

	case "oops":
		// get user
		user := args["username"]

		// get all the damn requests
		queue := queryMaps(drm.GetDRMQueue())

		// grab last request
		positions := drm.WhereDRM(user)
		lastRequest := queue[len(positions)-1]

		// clear the queue
		drm.RequestDRM("queue", "clear")

		// re-request map if it isn't lastRequest
		for _, mapData := range queue {
			if !(mapData.BsrKey == lastRequest.BsrKey && mapData.User == lastRequest.User) {
				drm.RequestDRM("addKey", mapData.BsrKey+"?user="+mapData.User)
			}
		}

		// print message
		fmt.Printf("Request %s removed from queue.", lastRequest.BsrKey)

	}
}
