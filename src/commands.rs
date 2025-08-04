use clap::{command, ArgAction, Parser, Subcommand};

#[derive(Debug, Parser)]
pub struct SDRMM {
    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Debug, Subcommand)]
pub enum Commands {
    /// Clears session request tracker, clears queue, and closes queue
    New,
    /// Sends a request to DRM's addKey endpoint after going through map filters
    #[command(arg_required_else_help = true)]
    Request {
        /// The 4-5 digit code of the map on BeatSaver
        id: String,
        /// The user who requested the map
        user: String,
        /// The service the user chatted from.
        #[arg(short, long)]
        service: Option<String>,
        /// Whether a mod added this map or not
        #[arg(long)]
        #[clap(action=ArgAction::SetTrue)]
        modadd: Option<bool>,
    },
    /// Sends a request to DRM's addWIP endpoint
    #[command(arg_required_else_help = true)]
    Wip {
        /// A link to the WIP file, or the code from wipbot.com
        wip: String,
        /// The user who requested the WIP
        user: String,
    },
    /// Shows/changes the status of the queue
    #[command(arg_required_else_help = true)]
    Queue {
        /// The subcommand
        command: String,
    },
    /// Gets length of the queue, and optionally where the user's requests are in it
    #[command(arg_required_else_help = true, name = "getqueue")]
    GetQueue {
        /// The user who invoked the command
        #[arg(short, long)]
        user: String,
    },
    /// Clears queue
    Clear,
    /// Moves a user's most recent request to the top of the queue
    #[command(arg_required_else_help = true)]
    Top {
        /// The user who invoked the command
        user: String,
    },
    /// Refunds a request if it's skipped or banned, if you have session_max set
    #[command(arg_required_else_help = true)]
    Refund {
        /// The user whose request was refunded
        user: String,
    },
    /// Gets a formatted message with current map information
    Link,
}
