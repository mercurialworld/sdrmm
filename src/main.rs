use clap::Parser;

use crate::{commands::SDRMM, config::SDRMMConfig, database::Database, drm::DRM};

mod commands;
#[expect(unused)]
mod config;
#[expect(unused)]
mod database;
#[expect(unused)]
mod drm;
#[expect(unused)]
mod filter;

fn main() {
    let sdrmm_config = SDRMMConfig::new("config.yaml").unwrap();

    let db = Database::from_file("database.db").unwrap();
    let _ = db.init_db();

    let drm = DRM::new(sdrmm_config.drm.url, sdrmm_config.drm.port);

    let args = SDRMM::parse();

    match args.command {
        commands::Commands::New => todo!(),
        commands::Commands::Request {
            id,
            user,
            service,
            modadd,
        } => todo!(),
        commands::Commands::Wip { wip, user } => todo!(),
        commands::Commands::Queue { command } => todo!(),
        commands::Commands::GetQueue { user } => todo!(),
        commands::Commands::Clear => todo!(),
        commands::Commands::Top { user } => todo!(),
        commands::Commands::Oops { user } => todo!(),
        commands::Commands::Refund { user } => todo!(),
    }
}
