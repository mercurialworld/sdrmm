use chrono::Utc;
use clap::Parser;
use tokio::main;

use crate::{
    commands::SDRMM, config::SDRMMConfig, database::Database, drm::DRM, filter::filter_map,
};

mod commands;
#[expect(unused)]
mod config;
#[expect(unused)]
mod database;
#[expect(unused)]
mod drm;
#[expect(unused)]
mod filter;

fn format_time(duration: i32) -> String {
    let seconds = duration % 60;
    let minutes = (duration / 60) % 60;
    let hours = (duration / 60) / 60;

    if hours > 0 {
        format!("{}:{:0>2}:{:0>2}", hours, minutes, seconds)
    } else {
        format!("{}:{:0>2}", minutes, seconds)
    }
}

async fn new(drm: &DRM, db: &Database) {
    match drm.queue_control("clear").await {
        Ok(_) => println!("Queue cleared from in-game!"),
        Err(_) => println!("Unable to clear queue from in-game."),
    };

    match db.clear_user_requests() {
        Ok(_) => println!("Cleared requests from database."),
        Err(_) => println!("Something happened."),
    };

    match db.new_session(Utc::now(), true) {
        Ok(_) => println!("Created new session."),
        Err(_) => println!("Something happened."),
    };
}

async fn get_queue(user: Option<String>, drm: &DRM) {
    match drm.queue().await {
        Ok(q) => {
            let sum: i32 = q
                .iter()
                .map(|map| map.duration)
                .reduce(|s, m| s + m)
                .unwrap();

            print!(
                "There are {} maps in queue (length {}).",
                q.len(),
                format_time(sum)
            );

            if let Some(u) = user {
                let user_maps = drm.queue_where(&u).await.unwrap();

                match user_maps.len() {
                    0 => (),
                    1 => print!(" Your map is in position {}.", user_maps[0].spot),
                    _ => print!(
                        " Your maps are in positions {}.",
                        user_maps
                            .iter()
                            .map(|map| map.spot.to_string())
                            .collect::<Vec<String>>()
                            .join(", ")
                    ),
                }
            }

            println!();
        }
        Err(_) => println!("Unable to get queue."),
    }
}

async fn set_queue(open: bool, drm: &DRM, db: &Database) {
    match drm
        .queue_control(&format!("open/{}", open.to_string()))
        .await
    {
        Ok(_) => (),
        Err(_) => println!("Something happened."),
    }

    match db.set_queue_status(true) {
        Ok(_) => (),
        Err(_) => println!("Something happened."),
    }
}

async fn queue(command: String, drm: &DRM, db: &Database) {
    match command.as_str() {
        "open" => set_queue(true, drm, db).await,
        "close" => set_queue(false, drm, db).await,
        "toggle" => match db.get_queue_status() {
            Ok(s) => set_queue(s, drm, db).await,
            Err(_) => (),
        },
        &_ => println!("Possible commands are open/close/toggle."),
    }
}

async fn clear_queue(drm: &DRM, db: &Database) {
    match drm.queue_control("clear").await {
        Ok(_) => (),
        Err(_) => println!("Something happened."),
    }

    match db.clear_user_requests() {
        Ok(_) => (),
        Err(_) => println!("Something happened."),
    }
}

async fn request(
    bsr: String,
    user: String,
    service: Option<String>,
    modadd: Option<bool>,
    drm: &DRM,
    db: &Database,
    config: SDRMMConfig,
) {
    let map = drm.query(&bsr).await.unwrap();

    match filter_map(&map, drm, &config, db, &user, modadd).await {
        Ok(_) => {
            let _ = match service {
                Some(s) => drm.add_with_service(&bsr, &user, &s).await,
                None => drm.add(&bsr, &user).await,
            };
        }
        Err(e) => println!("{}", e),
    }
}

async fn add_wip(wip: String, user: String, drm: &DRM) {
    match drm.wip(&wip, &user).await {
        Ok(_) => (),
        Err(e) => println!("{}", e),
    };
}

#[main]
async fn main() {
    let sdrmm_config = SDRMMConfig::new("config.yaml").unwrap();

    let db = Database::from_file("database.db").unwrap();
    let _ = db.init_db();

    let drm = DRM::new(sdrmm_config.drm.url.clone(), sdrmm_config.drm.port);

    let args = SDRMM::parse();

    match args.command {
        commands::Commands::New => new(&drm, &db).await,
        commands::Commands::Request {
            id,
            user,
            service,
            modadd,
        } => request(id, user, service, modadd, &drm, &db, sdrmm_config).await,
        commands::Commands::Wip { wip, user } => add_wip(wip, user, &drm).await,
        commands::Commands::Queue { command } => queue(command, &drm, &db).await,
        commands::Commands::GetQueue { user } => get_queue(Some(user), &drm).await,
        commands::Commands::Clear => clear_queue(&drm, &db).await,
        commands::Commands::Top { user: _ } => todo!(),
        commands::Commands::Oops { user: _ } => todo!(),
        commands::Commands::Refund { user: _ } => todo!(),
    }
}
