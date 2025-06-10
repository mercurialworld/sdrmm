package drm

import (
	"fmt"

	"github.com/alexflint/go-arg"
)

func Parse() (string, []byte, any) {
	var args struct {
		Request  *RequestCmd  `arg:"subcommand:request" help:"Put a map in the queue"`
		Mtt      *MttCmd      `arg:"subcommand:mtt" help:"Put a user's request to the top of the queue"`
		Wip      *WipCmd      `arg:"subcommand:wip" help:"Put a WIP in the queue"`
		GetQueue *GetQueueCmd `arg:"subcommand:getqueue" help:"Get the queue"`
		Queue    *QueueCmd    `arg:"subcommand:queue" help:"Toggle/get queue open/closed status"`
		Clear    *ClearCmd    `arg:"subcommand:clear" help:"Clear the queue"`
		Ban      *BanCmd      `arg:"subcommand:ban" help:"Ban a map from being requested"`
		Unban    *UnbanCmd    `arg:"subcommand:unban" help:"Unban a map from being requested"`
	}

	arg.MustParse(&args)

	switch {
	case args.Request != nil:
		requestArg := args.Request.Id

		if args.Request.User != "" {
			fmt.Printf("User %s r", args.Request.User)
			requestArg += "?user=" + args.Request.User
		} else {
			fmt.Printf("R")
		}
		fmt.Printf("equested map %s\n", args.Request.Id)

		return "add", requestDRM("addKey", requestArg), nil

	case args.Mtt != nil:
		fmt.Printf("Putting first request of user %s to top of queue\n", args.Mtt.User)
		pos := fmt.Sprintf("%d", whereDRM(args.Mtt.User)[0])

		return "mtt", requestDRM("queue", "move/"+pos+"/1"), nil

	case args.Wip != nil:
		fmt.Printf("User %s requested WIP %s\n", args.Wip.User, args.Wip.Id)
		return "wip", requestDRM("addWip", args.Wip.Id+"?user="+args.Wip.User), nil

	case args.GetQueue != nil:
		fmt.Printf("Queue requested")

		var pos []int

		if args.GetQueue.User != "" {
			fmt.Printf(" (also getting position for %s)\n", args.GetQueue.User)
			pos = whereDRM(args.GetQueue.User)
		}

		// this will work for sure
		return "getqueue", requestDRM("queue", ""), pos

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
			queue = requestDRM("queue", "")
		}

		return "clear", requestDRM("queue", "clear"), queue

	case args.Ban != nil:
		fmt.Printf("Map %s banned", args.Ban.Id)

		// TODO: add to database of banned map

	case args.Unban != nil:
		fmt.Printf("Map %s unbanned", args.Unban.Id)

		// TODO: remove from database of banned map
	}

	return "", nil, nil
}
