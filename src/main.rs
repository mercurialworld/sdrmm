
use clap::Parser;
use tokio::main;
use url::Url;

use crate::{
    commands::SDRMM,
    config::SDRMMConfig,
    database::Database,
    drm::{
        DRM,
        schema::{DRMMap},
    },
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

async fn new(drm: &DRM, db: &Database) {
    let mut create_new = false;

    // is history empty in-game?
    if let Ok(hist) = drm.history().await
        && hist.len() == 0
    {
        create_new = true;
    }

    if create_new {
        match drm.queue_control("clear").await {
            Ok(_) => println!("Queue cleared from in-game!"),
            Err(_) => println!("unable to clear queue from in-game."),
        };

        match db.clear_user_requests() {
            Ok(_) => println!("Cleared requests from database."),
            Err(_) => println!("unable to clear requests from database."),
        };
    } else {
        println!("no need for new session");
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
                    1 => print!(" Your map: {} ({}).", user_maps[0].spot, user_maps[0].queue_item.title),
                    _ => print!(
                        " Your maps: {}.",
                        user_maps
                            .iter()
                            .map(|m| format!("{} ({})", m.spot, m.queue_item.title))
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

async fn set_queue(open: bool, drm: &DRM) {
    match drm
        .queue_control(&format!("open/{}", open.to_string()))
        .await
    {
        Ok(_) => (),
        Err(_) => println!("Unable to set queue status in-game."),
    }
}

async fn queue(command: String, drm: &DRM) {
    match command.as_str() {
        "open" => set_queue(true, drm).await,
        "close" => set_queue(false, drm).await,
        "toggle" => match drm.queue_status().await {
            Ok(s) => set_queue(!s.queue_open, drm).await,
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
        }
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

async fn oops(user: &str, id: Option<String>, drm: &DRM) -> bool {
    match drm.queue().await {
        Ok(q) => {
            if q.len() == 0 {
                println!("Queue is empty!");
                return false;
            }

            let reqs: Vec<&DRMMap> = q
                .iter()
                .filter(|&m| m.user.clone().unwrap_or("".into()) == user)
                .collect();

            // if they have no requests in queue
            if reqs.len() == 0 {
                println!("You have no requests in queue!");
                return false;
            } else {
                let key_to_remove: String;

                if let Some(bsr) = id {
                    // if user specified a bsr, check if it's their request
                    let to_remove = reqs.iter().find(|&&m| m.bsr_key.clone() == bsr);

                    match to_remove {
                        Some(_) => key_to_remove = bsr,
                        None => {
                            println!("Map {} is either not in the queue or is not your request!", bsr);
                            return false;
                        }
                    }
                } else {
                    // if not, get their recent one
                    key_to_remove = drm
                        .queue_where(user)
                        .await
                        .unwrap()
                        .iter()
                        .last()
                        .unwrap()
                        .queue_item
                        .bsr_key
                        .clone();
                }

                match drm.remove(&key_to_remove).await {
                    Ok(map) => {
                        println!("Map {} removed from queue.", map.bsr_key);
                        return true;
                    }
                    Err(_) => {
                        println!("Unable to remove map from queue.");
                        return false;
                    }
                }
            }
        }
        Err(_) => {
            println!("Unable to get queue.");
            return false;
        }
    }
}

async fn remove(id: String, drm: &DRM) {
    match drm.remove(&id).await {
        Ok(m) => {
            println!("Map {} removed from queue.", m.bsr_key);
        },
        Err(_) => {
            println!("Unable to remove map from queue.")
        },
    }
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

async fn version(drm: &DRM) {
    match drm.version().await {
        Ok(v) => println!(
            "Beat Saber v{}, DumbRequestManager v{}",
            v.game_version,
            v.mod_version.split("+").collect::<Vec<&str>>()[0]
        ),
        Err(_) => println!("Unable to access DRM"),
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
        commands::Commands::New => new(&drm, &db).await,
        commands::Commands::Request {
            id,
            user,
            service,
            modadd,
        } => request(id, user, service, modadd, &drm, &db, sdrmm_config).await,
        commands::Commands::Wip { wip, user } => add_wip(wip, user, &drm).await,
        commands::Commands::Queue { command } => queue(command, &drm).await,
        commands::Commands::GetQueue { user } => get_queue(Some(user), &drm).await,
        commands::Commands::Clear => clear_queue(&drm, &db).await,
        commands::Commands::Top { user } => move_to_top(&user, &drm).await,
        commands::Commands::Refund { user } => refund_request(&user, &db, &sdrmm_config).await,
        commands::Commands::Link => get_link(&drm).await,
        commands::Commands::Oops { user, id } => {
            let res = oops(&user, id, &drm).await;

            if res {
                refund_request(&user, &db, &sdrmm_config).await;
            }
        },
        commands::Commands::Remove { id } => remove(id, &drm).await,
        commands::Commands::Ban { id } => ban(id, &drm).await,
        commands::Commands::Unban { id } => unban(id, &drm).await,
        commands::Commands::Version => version(&drm).await,
    }
}
