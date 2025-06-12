package parser

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"rustlang.pocha.moe/sdrmm/drm"
	"rustlang.pocha.moe/sdrmm/utils"
)

func Parse() (string, []byte, any, error) {
	var args struct {
		Request  *drm.RequestCmd    `arg:"subcommand:request" help:"Put a map in the queue"`
		Mtt      *drm.MttCmd        `arg:"subcommand:mtt" help:"Put a user's request to the top of the queue"`
		Wip      *drm.WipCmd        `arg:"subcommand:wip" help:"Put a WIP in the queue"`
		GetQueue *drm.GetQueueCmd   `arg:"subcommand:getqueue" help:"Get the queue"`
		Queue    *drm.QueueCmd      `arg:"subcommand:queue" help:"Toggle/get queue open/closed status"`
		Clear    *drm.ClearCmd      `arg:"subcommand:clear" help:"Clear the queue"`
		Ban      *drm.BanCmd        `arg:"subcommand:ban" help:"Ban a map from being requested"`
		Unban    *drm.UnbanCmd      `arg:"subcommand:unban" help:"Unban a map from being requested"`
		Oops     *drm.OopsCmd       `arg:"subcommand:oops" help:"Undo a user's recent request"`
		New      *drm.NewSessionCmd `arg:"subcommand:new" help:"Start a new session"`
	}

	arg.MustParse(&args)

	switch {
	case args.Request != nil:
		if args.Request.User != "" {
			fmt.Printf("User %s r", args.Request.User)
		} else {
			fmt.Printf("R")
		}
		fmt.Printf("equested map %s\n", args.Request.Id)

		return "add", drm.RequestDRM("query", args.Request.Id), args.Request.User, nil

	case args.Mtt != nil:
		fmt.Printf("Putting first request of user %s to top of queue\n", args.Mtt.User)
		pos := fmt.Sprintf("%d", drm.WhereDRM(args.Mtt.User)[0])

		return "mtt", drm.RequestDRM("queue", "move/"+pos+"/1"), nil, nil

	case args.Wip != nil:
		fmt.Printf("User %s requested WIP %s\n", args.Wip.User, args.Wip.Id)
		return "wip", drm.RequestDRM("addWip", args.Wip.Id+"?user="+args.Wip.User), nil, nil

	case args.GetQueue != nil:
		fmt.Printf("Queue requested")

		var pos []int

		if args.GetQueue.User != "" {
			fmt.Printf(" (also getting position for %s)\n", args.GetQueue.User)
			pos = drm.WhereDRM(args.GetQueue.User)
		}

		// this will work for sure
		return "getqueue", drm.RequestDRM("queue", ""), pos, nil

	case args.Queue != nil:
		// TODO: queue state management
		if args.Queue.Status {
			fmt.Println("Getting queue open/closed status")
		} else {
			fmt.Println("Toggling queue open/close")
		}

	case args.Clear != nil:
		fmt.Printf("Clearing queue (cache?: %t)\n", args.Clear.SaveQueue)
		var queue []byte = nil

		if args.Clear.SaveQueue {
			queue = drm.GetDRMQueue()
			err := os.WriteFile("queue.json", queue, 0644)
			utils.PanicOnError(err)
		}

		return "clear", drm.RequestDRM("queue", "clear"), queue, nil

	case args.Ban != nil:
		fmt.Printf("Banning map %s\n", args.Ban.Id)
		return "ban", drm.RequestDRM("query", args.Ban.Id), nil, nil

	case args.Unban != nil:
		fmt.Printf("Unbanning map %s\n", args.Unban.Id)
		return "unban", drm.RequestDRM("query", args.Unban.Id), nil, nil

	case args.Oops != nil:
		// cache entire queue
		queue := drm.GetDRMQueue()

		var queueData []drm.MapData
		err := json.Unmarshal(queue, &queueData)
		utils.PanicOnError(err)

		// get index of last request
		userRequests := drm.WhereDRM(args.Oops.User)
		userLastRequestIndex := userRequests[len(userRequests)-1]
		var userLastRequest drm.MapData = drm.MapData{}

		// pop the person's most recent request out of it
		var newQueue []drm.MapData
		for idx, currentMap := range queueData {
			if idx != userLastRequestIndex {
				userLastRequest = currentMap
				newQueue = append(newQueue, currentMap)
			}
		}

		// if the user has any requests in queue
		if userLastRequest.BsrKey != "" {
			// clear queue
			drm.RequestDRM("queue", "clear")

			// then add the entire new queue one by one :^)
			for _, currentMap := range newQueue {
				drm.RequestDRM("addKey", currentMap.BsrKey+"?user="+currentMap.User)
			}

			lastRequest, _ := json.Marshal(userLastRequest)

			return "oops", lastRequest, nil, nil
		} else {
			return "oops", nil, nil, fmt.Errorf("user %s does not have any requests in the queue", args.Oops.User)
		}

	case args.New != nil:
		return "new", nil, nil, nil
	}

	return "", nil, nil, fmt.Errorf("error in argument parsing: %s", os.Args)
}
