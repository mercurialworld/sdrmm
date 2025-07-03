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
		return "add", map[string]string{"id": args.Request.Id, "username": args.Request.User, "modadd": strconv.FormatBool(args.Request.ModAdd)}, nil

	case args.Mtt != nil:
		return "mtt", map[string]string{"username": args.Mtt.User}, nil

	case args.Wip != nil:
		return "wip", map[string]string{"id": args.Wip.Id, "username": args.Request.User}, nil

	case args.GetQueue != nil:
		return "getqueue", map[string]string{"username": args.GetQueue.User}, nil

	case args.Queue != nil:
		if args.Queue.Status {
			return "queuestatus", nil, nil
		} else {
			var extraArgs map[string]string = nil

			if !args.Queue.FromCommand {
				extraArgs = map[string]string{"status": strconv.FormatBool(args.Queue.SetStatus)}
			}

			return "togglequeue", extraArgs, nil
		}

	case args.Clear != nil:
		return "clear", map[string]string{"save": strconv.FormatBool(args.Clear.SaveQueue)}, nil

	case args.Ban != nil:
		return "ban", map[string]string{"id": args.Ban.Id}, nil

	case args.Unban != nil:
		return "unban", map[string]string{"id": args.Unban.Id}, nil

	case args.Oops != nil:
		return "oops", map[string]string{"username": args.Oops.User}, nil

	case args.New != nil:
		return "new", nil, nil
	}

	return "", nil, fmt.Errorf("unable to parse arguments: %s", os.Args)
}
