use chrono::{DateTime, TimeDelta, Utc};
use clap::Parser;
use tokio::main;
use url::Url;

use crate::{
    commands::SDRMM,
    config::SDRMMConfig,
    database::Database,
    drm::{DRM, schema::DRMMap},
    filter::filter_map,
};

mod commands;
mod config;
mod database;
mod drm;
mod filter;
mod helpers;

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

// [TODO] PLEASE WRITE THIS BETTER
async fn new(drm: &DRM, db: &Database, config: &SDRMMConfig) {
    let now = Utc::now();

    // difference between last recorded session and now is less than (config value from DRM)
    if let Some(last_session_time) =
        DateTime::<Utc>::from_timestamp(db.get_latest_session().unwrap().timestamp.into(), 0)
        && now.signed_duration_since(last_session_time)
            > TimeDelta::minutes(config.drm.new_session_length.into())
    {
        // is history empty in-game? if not, then don't bother
        if let Ok(hist) = drm.history().await
            && hist.len() == 0
        {
            match drm.queue_control("clear").await {
                Ok(_) => println!("Queue cleared from in-game!"),
                Err(_) => println!("unable to clear queue from in-game."),
            };

            match db.clear_user_requests() {
                Ok(_) => println!("Cleared requests from database."),
                Err(_) => println!("unable to clear requests from database."),
            };

            match db.new_session(Utc::now(), true) {
                Ok(_) => println!("Created new session."),
                Err(_) => println!("unable to create new session."),
            };
        } else {
            println!("no need for new session, history already empty!");
        }
    } else {
        println!("no need for new session, not long enough!");
    }
}

async fn get_queue(user: Option<String>, drm: &DRM) {
    match drm.queue().await {
        Ok(q) => {
            if q.len() == 0 {
                println!("Queue is empty!");
                return;
            }

            let sum: i32 = q
                .iter()
                .map(|map| map.duration)
                .reduce(|s, m| s + m)
                .unwrap_or_default();

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
        Err(_) => println!("unable to get queue."),
    }
}

async fn set_queue(open: bool, drm: &DRM, db: &Database) {
    match drm
        .queue_control(&format!("open/{}", open.to_string()))
        .await
    {
        Ok(_) => (),
        Err(_) => println!("Unable to set queue status in-game."),
    }

    match db.set_queue_status(open) {
        Ok(_) => (),
        Err(e) => println!("{:?}", e),
        // Err(_) => println!("Unable to set queue status in database."),
    }
}

async fn queue(command: String, drm: &DRM, db: &Database) {
    match command.as_str() {
        "open" => set_queue(true, drm, db).await,
        "close" => set_queue(false, drm, db).await,
        "toggle" => match db.get_queue_status() {
            Ok(s) => set_queue(!s, drm, db).await,
            Err(_) => (),
        },
        &_ => println!("Possible commands are open/close/toggle."),
    }
}

async fn clear_queue(drm: &DRM, db: &Database) {
    match drm.queue_control("clear").await {
        Ok(_) => (),
        Err(_) => println!("Unable to clear queue in-game."),
    }

    match db.clear_user_requests() {
        Ok(_) => (),
        Err(_) => println!("Unable to clear requests in database."),
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
            let mut message_builder: String = format!("{} added to queue.", &bsr);

            let _ = match service {
                Some(s) => drm.add_with_service(&bsr, &user, &s).await,
                None => drm.add(&bsr, &user).await,
            };

            if config.queue.session_max > 0 {
                message_builder.push_str(" You have ");

                let _ = match db.get_user_requests(&user) {
                    Ok(r) => {
                        let _ = db.set_user_requests(&user, r + 1);
                        message_builder.push_str(&format!("{} requests left.", r + 1));
                    }
                    Err(_) => {
                        let _ = db.add_user_requests(&user);
                        message_builder
                            .push_str(&format!("{} requests left.", config.queue.session_max - 1));
                    }
                };
            }

            println!("{}", message_builder);
        }
        Err(e) => println!("{}", e),
    }
}

async fn add_wip(wip: String, user: String, drm: &DRM) {
    match drm.wip(&wip, &user).await {
        Ok(map) => {
            if let Ok(wip_domain) = Url::parse(&map.bsr_key) {
                println!("WIP from {} added to queue", wip_domain.host_str().unwrap());
            }
        },
        Err(e) => println!("{}", e),
    };
}

async fn get_link(drm: &DRM) {
    match drm.link().await {
        Ok(hist) => {
            let map = &hist.get(0).unwrap().history_item;

            println!(
                "{} - {} (mapped by {}) https://beatsaver.com/maps/{}",
                map.artist, map.title, map.mapper, map.bsr_key
            );
        }
        Err(_) => println!("No map available"),
    }
}

async fn move_to_top(user: &str, drm: &DRM) {
    if let Ok(q) = drm.queue_where(&user).await {
        let last_req = q.last().unwrap();

        match drm
            .queue_control(&*format!("move/{}/1", last_req.spot))
            .await
        {
            Ok(_) => println!("Map {} moved to top.", last_req.queue_item.bsr_key),
            Err(_) => println!("Unable to move your recent request to top."),
        };
    }
}

async fn oops(user: &str, drm: &DRM) -> bool {
    // check if there's a queue
    if let Ok(current_queue) = drm.queue().await
        && current_queue.len() > 0
    {
        // check if user has requests in queue
        let user_reqs: Vec<&DRMMap> = current_queue
            .iter()
            .filter(|&m| m.user.as_ref().unwrap() == user)
            .collect();

        if user_reqs.len() > 0 {
            // get last request
            let last_bsr = &user_reqs.last().unwrap().bsr_key;

            // create queue with request removed
            let new_queue: Vec<&DRMMap> = current_queue
                .iter()
                .filter(|&m| m.bsr_key != *last_bsr)
                .collect();

            // clear queue
            let _ = drm.queue_control("clear").await;

            // re-request everything!
            for map in new_queue.iter() {
                let _ = drm.add(&map.bsr_key, &map.user.as_ref().unwrap()).await;
            }

            println!("Request {} removed from queue.", last_bsr);

            return true;
        }
    }
    println!("Unable to remove recent request from queue.");

    false
}

async fn refund_request(user: &str, db: &Database, config: &SDRMMConfig) {
    if config.queue.session_max > 0 {
        if let Ok(r) = db.get_user_requests(&user) {
            match db.set_user_requests(&user, r - 1) {
                Ok(_) => println!("Request refunded."),
                Err(_) => println!("Unable to refund request"),
            };
        } else {
            println!("User not found.");
        }
    }
}

async fn ban(bsr: String, drm: &DRM) {
    match drm.blacklist(&bsr).await {
        Ok(_) => println!("{} is now banned from being requested.", bsr),
        Err(_) => todo!("Unable to access DRM."),
    }
}

async fn unban(bsr: String, drm: &DRM) {
    match drm.unblacklist(&bsr).await {
        Ok(_) => println!("{} can now be requested again.", bsr),
        Err(_) => todo!("Unable to access DRM."),
    }
}

#[main]
async fn main() {
    let sdrmm_config = SDRMMConfig::new("config.yaml").unwrap();

    let db = Database::from_file("database.db").unwrap();
    let _ = db.init_db();

    let drm = DRM::new(
        sdrmm_config.drm.url.clone(),
        sdrmm_config.drm.port.try_into().unwrap(),
    );

    let args = SDRMM::parse();

    match args.command {
        commands::Commands::New => new(&drm, &db, &sdrmm_config).await,
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
        commands::Commands::Top { user } => move_to_top(&user, &drm).await,
        commands::Commands::Refund { user } => refund_request(&user, &db, &sdrmm_config).await,
        commands::Commands::Link => get_link(&drm).await,
        commands::Commands::Oops { user } => {
            let res = oops(&user, &drm).await;

            if res {
                refund_request(&user, &db, &sdrmm_config).await;
            }
        },
        commands::Commands::Ban { id } => ban(id, &drm).await,
        commands::Commands::Unban { id } => unban(id, &drm).await,

    }
}
