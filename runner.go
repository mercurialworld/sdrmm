package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"rustlang.pocha.moe/sdrmm/config"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
)

type Runner struct {
	config   config.BSRConfig
	database database.DRMDatabase
}

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

func (r Runner) refundRequest(user string, messageBuilder string) string {
	// grab number of user requests
	userNumRequests := r.database.GetUserRequests(user)
	// grab number of requests *in queue*
	userNumRequestsInQueue := r.database.GetUserRequestsInQueue(user)

	// grab the request limit
	requestLimit := r.config.RequestLimit
	// grab the limit of maps in queue
	queueRequestLimit := r.config.QueueRequestLimit

	if requestLimit > 0 {
		userNumRequests--
		r.database.SetUserRequests(user, userNumRequests)

		messageBuilder += fmt.Sprintf(" You have %d requests left.", requestLimit-userNumRequests)
	}

	if queueRequestLimit > 0 {
		userNumRequestsInQueue--
		r.database.SetUserRequestsInQueue(user, userNumRequestsInQueue)
	}

	return messageBuilder
}

func (r Runner) RunCommands(command string, args map[string]string) {
	filter := Filterer(r)

	switch command {
	case "new":
		// clear queue
		drm.RequestDRM("clear", "")

		// clear requests
		r.database.ClearRequestLimits()

		// close queue
		drm.RequestDRM("queue", "open/false")

		// get current date and time
		currentTime := time.Now()

		// add new entry in db
		r.database.NewSession(currentTime)

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
		userNumRequests := r.database.GetUserRequests(user)
		// grab number of requests *in queue*
		userNumRequestsInQueue := r.database.GetUserRequestsInQueue(user)

		// grab the request limit
		requestLimit := r.config.RequestLimit
		// grab the limit of maps in queue
		queueRequestLimit := r.config.QueueRequestLimit

		// put map through filters
		mapToQueue, err := filter.FilterMap(mapToQueue, user, modadd)

		if err != nil {
			// print a message
			fmt.Printf("%s", err)
		} else {
			// add to queue
			addKeyArgs := mapToQueue.BsrKey + "?user=" + user
			drm.RequestDRM("addKey", "/"+addKeyArgs)

			message := "Added to queue!"

			if requestLimit > 0 {
				// increment user requests
				userNumRequests++
				r.database.SetUserRequests(user, userNumRequests)

				message += fmt.Sprintf(" You have %d requests left.", requestLimit-userNumRequests)
			}

			if queueRequestLimit > 0 {
				// increment requests in queue
				userNumRequestsInQueue++
				r.database.SetUserRequestsInQueue(user, userNumRequestsInQueue)
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
		drm.RequestDRM("queue", "/move/"+pos+"/1")

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
		mapToBan := queryMap(drm.RequestDRM("query", "/"+args["id"]))

		// ban map
		r.database.BanMap(mapToBan.BsrKey, mapToBan.Hash)

		message := fmt.Sprintf("%s is now banned from being requested.", mapToBan.BsrKey)

		// if a user requested it, refund their requests
		if user, found := args["username"]; found {
			message = r.refundRequest(user, message)
		}

		// print message
		fmt.Print(message)

	case "unban":
		// query map
		mapToUnban := queryMap(drm.RequestDRM("query", "/"+args["id"]))

		// unban map
		r.database.BanMap(mapToUnban.BsrKey, mapToUnban.Hash)

		// print message
		fmt.Printf("%s can now be requested again.", mapToUnban.BsrKey)

	case "queuestatus":
		// get queue status
		queueStatus := r.database.GetQueueStatus()

		statusString := "closed"

		if queueStatus {
			statusString = "open"
		}

		fmt.Printf("The queue is currently %s.", statusString)

	case "togglequeue":
		// get queue status
		newQueueStatus := !r.database.GetQueueStatus()
		if args != nil {
			newQueueStatus, _ = strconv.ParseBool(args["status"])
		}

		// set the queue status to whatever we get in the database
		r.database.SetQueueStatus(newQueueStatus)

		// set it in game too
		drm.RequestDRM("queue", "open/"+strconv.FormatBool(newQueueStatus))

		// format string
		statusString := "closed"
		if newQueueStatus {
			statusString = "open"
		}

		// print string
		fmt.Printf("The queue is now %s.", statusString)

	case "oops":
		// get user
		user := args["username"]

		// get all the damn requests
		queue := queryMaps(drm.GetDRMQueue())

		// grab last request
		positions := drm.WhereDRM(user)
		lastRequest := queue[positions[len(positions)-1]]

		// clear the queue
		drm.RequestDRM("queue", "clear")

		// re-request map if it isn't lastRequest
		for _, mapData := range queue {
			if !(mapData.BsrKey == lastRequest.BsrKey && mapData.User == lastRequest.User) {
				drm.RequestDRM("addKey", "/"+mapData.BsrKey+"?user="+mapData.User)
			}
		}

		// print message
		fmt.Printf("Request %s removed from queue.", lastRequest.BsrKey)
	case "refund":
		user := args["username"]
		message := "Request skipped."

		message = r.refundRequest(user, message)

		fmt.Print(message)
	}
}
