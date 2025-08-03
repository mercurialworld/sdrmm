use chrono::{DateTime, Days, NaiveTime, TimeDelta, Utc};
use num::Num;

use crate::{
    config::{ignore_config, SDRMMConfig},
    database::{Database},
    drm::{schema::{DRMMap}, DRM},
};

fn is_recent(map_date: DateTime<Utc>, min_date: DateTime<Utc>) -> bool {
    map_date.signed_duration_since(min_date) > TimeDelta::zero()
}

fn is_open(db: &Database) -> anyhow::Result<bool> {
    Ok(db.get_queue_status()?)
}

// Returns False if everything is fine, or the setting is 0
fn gt_ignore<T: Num + PartialOrd + Clone>(to_compare: T, config_val: T) -> bool {
    !ignore_config(config_val.clone()) && to_compare > config_val
}

// Returns False if everything is fine, or the setting is 0
fn lt_ignore<T: Num + PartialOrd + Clone>(to_compare: T, config_val: T) -> bool {
    !ignore_config(config_val.clone()) && to_compare < config_val
}

// Returns False if everything is fine, or the setting is 0
fn eq_ignore<T: Num + PartialOrd + Clone>(config_val: T, to_compare: T) -> bool {
    !ignore_config(config_val.clone()) && config_val <= to_compare
}

// Returns true if there's at least one diff that meets requirements, or the setting is 0
fn gt_diffs_ignore<T: Num + PartialOrd + Clone>(to_compare: &Vec<T>, config_val: T) -> bool {
    if ignore_config(config_val.clone()) {
        return true;
    }

    let mut one_diff_meets_criteria = false;

    for diff_val in to_compare {
        if *diff_val > config_val {
            one_diff_meets_criteria = true;
        }
    }

    one_diff_meets_criteria
}

// Returns true if there's at least one diff that meets requirements, or the setting is 0
fn lt_diffs_ignore<T: Num + PartialOrd + Clone>(to_compare: &Vec<T>, config_val: T) -> bool {
    if ignore_config(config_val.clone()) {
        return true;
    }

    let mut one_diff_meets_criteria = false;

    for diff_val in to_compare {
        if *diff_val < config_val {
            one_diff_meets_criteria = true;
        }
    }

    one_diff_meets_criteria
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
        if m {
            return Ok(());
        }
    }

    // is queue closed?
    match is_open(db) {
        Ok(open) => {
            if !open {
                return Err("Queue is closed!".into());
            }
        }
        Err(e) => {
            let mut h: String = "".into();
            e.chain()
                .skip(1)
                .for_each(|cause| h.push_str(&format!("{:?}", cause)));
            return Err(h);
        }
    }

    // does the user already have enough stuff in queue?
    if let Ok(in_queue) = drm.queue_where(&user).await
        && eq_ignore(config.queue.queue_max, in_queue.len() as i32)
    {
        return Err(format!(
            "You have too many songs in queue! (max is {})",
            config.queue.queue_max
        ));
    }

    // did the user request enough maps this session?
    if let Ok(session_reqs) = db.get_user_requests(&user)
        && eq_ignore(config.queue.session_max, session_reqs)
    {
        return Err(format!(
            "You have no more requests this session! (max is {})",
            config.queue.session_max
        ));
    }

    // is map banned?
    if map.blacklisted {
        return Err("Map is banned from being requested!".into());
    }

    // is map older than a certain date?
    if !is_recent(
        map.upload_time,
        config
            .bsr
            .date
            .earliest
            .and_time(NaiveTime::default())
            .and_utc(),
    ) {
        return Err(format!(
            "Map is older than {} (uploaded {})",
            config.bsr.date.earliest.format("%b %e, %Y"),
            map.upload_time.format("%b %e, %Y"),
        ));
    }

    // is map younger than a certain number of days?
    if !ignore_config(config.bsr.date.min_age)
        && let Some(_) =
            Utc::now().checked_sub_days(Days::new(config.bsr.date.min_age as u64))
    {
        return Err(format!(
            "Map is less than {} days old (uploaded {})",
            config.bsr.date.min_age,
            map.upload_time.format("%b %e, %Y"),
        ));
    }

    // is the map too short?
    if lt_ignore(map.duration, config.bsr.length.min) {
        return Err(format!(
            "Map is shorter than {} seconds (is {} seconds)",
            config.bsr.length.min, map.duration
        ));
    }

    // is the map too long?
    if gt_ignore(map.duration, config.bsr.length.max) {
        return Err(format!(
            "Map is longer than {} seconds (is {} seconds)",
            config.bsr.length.max, map.duration
        ));
    }

    // has the map already been played?
    if !config.queue.replay && map.has_played {
        return Err("Map has already been played this session!".into());
    }

    // make a list of nps/njs of all difficulties
    let nps = map.diffs.iter().map(|diff| diff.notes_per_second).collect::<Vec<f32>>();
    let njs = map.diffs.iter().map(|diff| diff.note_jump_speed).collect::<Vec<f32>>();

    // NPS comparisons   
    if !gt_diffs_ignore(&nps, config.bsr.nps.min) {
        return Err(format!(
            "Map does not have a difficulty with NPS higher than {}",
            config.bsr.nps.min
        ))
    }

    if !lt_diffs_ignore(&nps, config.bsr.nps.max) {
        return Err(format!(
            "Map does not have a difficulty with NPS lower than {}",
            config.bsr.nps.max
        ))
    }

    // NJS comparisons
    if !gt_diffs_ignore(&njs, config.bsr.njs.min) {
        return Err(format!(
            "Map does not have a difficulty with NJS higher than {}",
            config.bsr.njs.min
        ))
    }

    if !lt_diffs_ignore(&njs, config.bsr.njs.max) {
        return Err(format!(
            "Map does not have a difficulty with NJS lower than {}",
            config.bsr.njs.max
        ))
    }


    Ok(())
}
