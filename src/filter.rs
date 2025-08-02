use chrono::{DateTime, Days, NaiveTime, TimeDelta, Utc};
use num::Num;

use crate::{
    config::{SDRMMConfig, ignore_config},
    database::{Database, DatabaseError},
    drm::{DRM, schema::DRMMap},
};

fn is_older(map_date: DateTime<Utc>, min_date: DateTime<Utc>) -> bool {
    map_date.signed_duration_since(min_date) < TimeDelta::zero()
}

fn is_open(db: &Database) -> Result<bool, DatabaseError> {
    db.get_queue_status()
}

// Returns False if everything is fine, or the setting is 0
fn ignore_or_compare<T: Num + PartialOrd>(to_ignore: T, left: T, right: T) -> bool {
    !ignore_config(to_ignore) && left < right
}

// Returns False if everything is fine, or the setting is 0
fn ignore_or_equate<T: Num + PartialOrd + Clone>(option: T, map_val: T) -> bool {
    !ignore_config(option.clone()) && option == map_val
}

pub async fn filter_map(
    map: &DRMMap,
    drm: &DRM,
    config: &SDRMMConfig,
    db: &Database,
    user: &str,
    modadd: Option<bool>,
) -> Result<(), String> {
    // modadd
    if let Some(m) = modadd {
        if m { return Ok(()) }
    }

    // is queue closed?
    match is_open(db) {
        Ok(open) => {
            if !open {
                return Err("Queue is closed!".into());
            }
        }
        Err(_) => return Err("Queue is probably closed".into()),
    }

    // does the user already have enough stuff in queue?
    if let Ok(in_queue) = drm.queue_where(&user).await
        && ignore_or_equate(config.queue.queue_max, in_queue.len() as i32)
    {
        return Err(format!(
            "You have too many songs in queue! (max is {})",
            config.queue.queue_max
        ));
    }

    // did the user request enough maps this session?
    if let Ok(session_reqs) = db.get_user_requests(&user)
        && ignore_or_equate(config.queue.session_max, session_reqs)
    {
        return Err(format!(
            "You have no more requests this session! (max is {})",
            config.queue.session_max
        ));
    }

    // is map banned?
    match map.blacklisted {
        true => (),
        false => return Err("Map is banned from being requested!".into()),
    }

    // is map older than a certain date?
    match is_older(
        map.upload_time,
        config
            .bsr
            .date
            .earliest
            .and_time(NaiveTime::default())
            .and_utc(),
    ) {
        true => (),
        false => {
            return Err(format!(
                "Map is older than {} (uploaded {})",
                config.bsr.date.earliest.format("%b %e, %Y"),
                map.upload_time.format("%b %e, %Y"),
            ));
        }
    }

    // is map younger than a certain number of days?
    if !ignore_config(config.bsr.date.min_age)
        && let Some(new_map) =
            Utc::now().checked_sub_days(Days::new(config.bsr.date.min_age as u64))
    {
        return Err(format!(
            "Map is less than {} years old (uploaded {})",
            config.bsr.date.min_age,
            map.upload_time.format("%b %e, %Y"),
        ));
    }

    // is the map too short?
    if ignore_or_compare(config.bsr.length.min, map.duration, config.bsr.length.max) {
        return Err(format!(
            "Map is shorter than {} seconds (is {} seconds)",
            config.bsr.length.min, map.duration
        ));
    }

    // is the map too long?
    if ignore_or_compare(config.bsr.length.max, config.bsr.length.max, map.duration) {
        return Err(format!(
            "Map is longer than {} seconds (is {} seconds)",
            config.bsr.length.max, map.duration
        ));
    }

    // [TODO] NPS/NJS check
    // make a list of nps/njs of all difficulties
    // if one of them is in limits, then it should be allowed through

    Ok(())
}
