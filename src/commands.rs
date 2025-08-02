use clap::{Parser, Subcommand, command};

#[derive(Debug, Parser)]
struct SDRMM {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Debug, Subcommand)]
enum Commands {
    New,
    #[command(arg_required_else_help = true)]
    Request {
        id: String,
        user: String,
        service: Option<String>,
        modadd: bool,
    },
    #[command(arg_required_else_help = true)]
    Wip {
        wip: String,
        user: String,
    },
    #[command(arg_required_else_help = true)]
    Queue {
        command: String,
    },
    Clear,
    #[command(arg_required_else_help = true)]
    Top {
        user: String,
    },
    #[command(arg_required_else_help = true)]
    Oops {
        user: String,
    },
    #[command(arg_required_else_help = true)]
    Refund {
        user: String,
    },
}
