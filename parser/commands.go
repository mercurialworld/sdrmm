package parser

// BeatSaver request by a user
type RequestCmd struct {
	Id     string `arg:"positional,required" help:"The BeatSaver ID of the map to be requested"`
	User   string `arg:"-u,required" help:"The username of the requester"`
	ModAdd bool   `default:"false" help:"Whether to force the map in the queue"`
}

// Put a request to top of the queue
type MttCmd struct {
	User string `arg:"positional,required" help:"The username whose map you want to put on top of the queue"`
}

// WIP request by a user
type WipCmd struct {
	Id   string `arg:"positional,required" help:"The WIP code of the map to be requested"`
	User string `arg:"-u,required" help:"The username of the requester"`
}

// Get the user's requests in queue
type GetQueueCmd struct {
	User string `arg:"positional" help:"The username who invoked the command"`
}

// Blacklist/ban a map
type BanCmd struct {
	Id   string `arg:"positional,required" help:"The BeatSaver ID of the map to be banned"`
	User string `help:"The user to refund, if any"`
}

// Whitelist/unban a map
type UnbanCmd struct {
	Id string `arg:"positional,required" help:"The BeatSaver ID of the map to be unbanned"`
}

// Queue related commands
type QueueCmd struct {
	Get    *GetQueueStatusCmd    `arg:"subcommand:get" help:"Gets the status of queue"`
	Set    *SetQueueStatusCmd    `arg:"subcommand:set" help:"Sets the queue status"`
	Toggle *ToggleQueueStatusCmd `arg:"subcommand:toggle" help:"Toggles queue status"`
}

// Get queue status
type GetQueueStatusCmd struct{}

// Set queue status
type SetQueueStatusCmd struct {
	Open bool `arg:"positional,required" help:"Should the queue be open?"`
}

// Toggle queue status
type ToggleQueueStatusCmd struct{}

// Clear queue
type ClearCmd struct {
	SaveQueue bool `arg:"--save" default:"false" help:"Whether to save the queue as it was before deleting all of it, in case you want to refund points or something"`
}

// Undo last request
type OopsCmd struct {
	User string `arg:"positional" help:"The username who invoked the !oops command."`
}

// New session
type NewSessionCmd struct{}

// refund
type RefundCmd struct {
	User string `help:"The user to refund."`
}
