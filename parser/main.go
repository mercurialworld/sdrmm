package parser

import (
	"fmt"
	"os"
	"strconv"

	"github.com/alexflint/go-arg"
)

func Parse() (string, map[string]string, error) {
	var args struct {
		Request  *RequestCmd    `arg:"subcommand:request" help:"Put a map in the queue"`
		Mtt      *MttCmd        `arg:"subcommand:mtt" help:"Put a user's request to the top of the queue"`
		Wip      *WipCmd        `arg:"subcommand:wip" help:"Put a WIP in the queue"`
		GetQueue *GetQueueCmd   `arg:"subcommand:getqueue" help:"Get the queue"`
		Queue    *QueueCmd      `arg:"subcommand:queue" help:"Toggle/get queue open/closed status"`
		Clear    *ClearCmd      `arg:"subcommand:clear" help:"Clear the queue"`
		Ban      *BanCmd        `arg:"subcommand:ban" help:"Ban a map from being requested"`
		Unban    *UnbanCmd      `arg:"subcommand:unban" help:"Unban a map from being requested"`
		Oops     *OopsCmd       `arg:"subcommand:oops" help:"Undo a user's recent request"`
		New      *NewSessionCmd `arg:"subcommand:new" help:"Start a new session"`
	}

	arg.MustParse(&args)

	switch {
	case args.Request != nil:
		fmt.Printf("User %s requested map %s\n", args.Request.User, args.Request.Id)

		return "add", map[string]string{"id": args.Request.Id, "username": args.Request.User, "platform": args.Request.Platform}, nil

	case args.Mtt != nil:
		fmt.Printf("Putting last request of user %s to top of queue\n", args.Mtt.User)

		return "mtt", map[string]string{"username": args.Mtt.User}, nil

	case args.Wip != nil:
		fmt.Printf("User %s requested WIP %s\n", args.Wip.User, args.Wip.Id)

		return "wip", map[string]string{"id": args.Wip.Id, "username": args.Request.User}, nil

	case args.GetQueue != nil:
		fmt.Printf("Queue requested")

		return "getqueue", map[string]string{"username": args.GetQueue.User}, nil

	case args.Queue != nil:
		if args.Queue.Status {
			fmt.Println("Getting queue open/closed status")
			return "queuestatus", nil, nil
		} else {
			fmt.Println("Toggling queue open/close")
			return "togglequeue", nil, nil
		}

	case args.Clear != nil:
		fmt.Printf("Clearing queue (cache?: %t)\n", args.Clear.SaveQueue)

		return "clear", map[string]string{"save": strconv.FormatBool(args.Clear.SaveQueue)}, nil

	case args.Ban != nil:
		fmt.Printf("Banning map %s\n", args.Ban.Id)
		return "ban", map[string]string{"id": args.Ban.Id}, nil

	case args.Unban != nil:
		fmt.Printf("Unbanning map %s\n", args.Unban.Id)
		return "unban", map[string]string{"id": args.Unban.Id}, nil

	case args.Oops != nil:
		return "oops", map[string]string{"user": args.Oops.User}, nil

	case args.New != nil:
		return "new", nil, nil
	}

	return "", nil, fmt.Errorf("unable to parse arguments: %s", os.Args)
}
