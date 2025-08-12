use chrono::{DateTime, Days, NaiveTime, TimeDelta, Utc};
use regex::Regex;

use crate::{
    config::SDRMMConfig,
    database::Database,
    drm::{DRM, schema::DRMMap},
    helpers::{
        ignore_config, ignore_or_geq, ignore_or_geq_vec, ignore_or_leq, ignore_or_leq_vec,
        ignore_or_lt, match_in_two_vecs,
    },
};

fn is_recent(map_date: DateTime<Utc>, min_date: DateTime<Utc>) -> bool {
    map_date.signed_duration_since(min_date) > TimeDelta::zero()
}

fn is_open(db: &Database) -> anyhow::Result<bool> {
    Ok(db.get_queue_status()?)
}

fn censor(map: &DRMMap) -> bool {
    map.censor_artist
        || map.censor_mapper
        || map.censor_sub_title
        || map.censor_title
        || map.metadata_has_spliced_censor
}

fn map_contains_tlds(map: &DRMMap) -> bool {
    let tld_regex = Regex::new(r"(\.(com|net|org))").unwrap();

    let to_check = vec![&map.title, &map.sub_title, &map.artist, &map.mapper];

    for field in to_check {
        if let Some(_) = tld_regex.captures(field) {
            return true;
        }
    }

    false
}

pub async fn filter_map(
    map: &DRMMap,
    drm: &DRM,
    config: &SDRMMConfig,
    db: &Database,
    user: &str,
    modadd: Option<bool>,
) -> Result<(), String> {
    // is the map already in queue?
    if !config.queue.repeat
        && let Ok(q) = drm.queue().await
        && let Some(_) = q.iter().find(|&map_data| map_data.bsr_key == map.bsr_key)
    {
        return Err("Map is already in queue!".into());
    }

    // modadd
    // putting this after the queue doubles check so we don't have a repeat of
    // the one time 3 of us modadded !bsr 48dd1 to the queue
    if let Some(m) = modadd
        && m
    {
        return Ok(());
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
        && !ignore_or_lt(config.queue.queue_max, in_queue.len().try_into().unwrap())
    {
        return Err(format!(
            "You have too many songs in queue! (max is {})",
            config.queue.queue_max
        ));
    }

    // did the user request enough maps this session?
    if let Ok(session_reqs) = db.get_user_requests(&user)
        && !ignore_or_lt(config.queue.session_max, session_reqs.try_into().unwrap())
    {
        return Err(format!(
            "You have no more requests this session! (max is {})",
            config.queue.session_max
        ));
    }

    // does the map have anything that's flagged for censoring?
    if config.bsr.censors.deny_censored && censor(&map) {
        // is the thing flagging the censors not a domain?
        // or, if the thing is a domain, does the user want urls auto-denied?
        if !map_contains_tlds(map) || config.bsr.censors.deny_urls {
            return Err("Map has terms that aren't allowed.".into());
        }
    }

    // vote status
    match map.vote_status {
        crate::drm::schema::VoteStatus::None => (),
        crate::drm::schema::VoteStatus::Liked => {
            if config.map_vote.allow_liked {
                return Ok(());
            }
        }
        crate::drm::schema::VoteStatus::Disliked => {
            if config.map_vote.deny_disliked {
                return Err("The streamer probably doesn't like the map.".into());
            }
        }
    }

    // is map banned?
    if map.blacklisted {
        return Err("Map is banned from being requested!".into());
    }

    // is map in an allowed playlist?
    if let Some(playlists) = &config.allowed_playlists
        && match_in_two_vecs(map.playlists.clone(), playlists.to_vec())
    {
        return Ok(());
    }

    // is map rating greater than a certain minimum?
    // needs a live beatsaver query
    if let Ok(live_map) = drm.query_nocache(&map.bsr_key).await
        && !ignore_or_geq(config.bsr.min_rating, live_map.rating * 100.0)
    {
        return Err(format!(
            "Map rating is less than {:.2}% (is {:.2}%)",
            config.bsr.min_rating,
            live_map.rating * 100.0
        ));
    }

    // is map ai-generated?
    if !config.bsr.allow_ai && map.automapped {
        return Err("Map is automapped!".into());
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
        && let Some(_) = Utc::now().checked_sub_days(Days::new(config.bsr.date.min_age as u64))
    {
        return Err(format!(
            "Map is less than {} days old (uploaded {})",
            config.bsr.date.min_age,
            map.upload_time.format("%b %e, %Y"),
        ));
    }

    // is the map too short?
    if !ignore_or_geq(config.bsr.length.min, map.duration) {
        return Err(format!(
            "Map is shorter than {} seconds (is {} seconds)",
            config.bsr.length.min, map.duration
        ));
    }

    // is the map too long?
    if !ignore_or_leq(config.bsr.length.max, map.duration) {
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
    let nps = map
        .diffs
        .iter()
        .map(|diff| diff.notes_per_second)
        .collect::<Vec<f32>>();
    let njs = map
        .diffs
        .iter()
        .map(|diff| diff.note_jump_speed)
        .collect::<Vec<f32>>();

    // NPS comparisons
    // greater than min
    if !ignore_or_geq_vec(&nps, config.bsr.nps.min) {
        return Err(format!(
            "Map does not have a difficulty with NPS higher than {}",
            config.bsr.nps.min
        ));
    }

    // less than max
    if !ignore_or_leq_vec(&nps, config.bsr.nps.max) {
        return Err(format!(
            "Map does not have a difficulty with NPS lower than {}",
            config.bsr.nps.max
        ));
    }

    // NJS check
    // greater than min
    if !ignore_or_geq_vec(&njs, config.bsr.njs.min) {
        return Err(format!(
            "Map does not have a difficulty with NJS higher than {}",
            config.bsr.njs.min
        ));
    }

    // less than max
    if !ignore_or_leq_vec(&njs, config.bsr.njs.max) {
        return Err(format!(
            "Map does not have a difficulty with NJS lower than {}",
            config.bsr.njs.max
        ));
    }

    Ok(())
}
