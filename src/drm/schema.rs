use chrono::{DateTime, Utc, serde::ts_seconds};
use serde::Deserialize;

#[derive(Debug, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DRMMapDiff {
    pub difficulty: String,
    pub characteristic: String,
    pub note_jump_speed: f32,
    pub notes_per_second: f32,
    pub map_mods: MapMods,
    #[serde(rename = "ScoreSaberStars")]
    pub scoresaber_stars: f32,
    #[serde(rename = "BeatLeaderStars")]
    pub beatleader_stars: f32,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct MapMods {
    pub chroma: bool,
    pub cinema: bool,
    pub mapping_extensions: bool,
    pub noodle_extensions: bool,
    pub vivify: bool,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DRMMap {
    pub bsr_key: String,
    pub hash: String,
    pub user: Option<String>,
    pub title: String,
    pub censor_title: bool,
    pub sub_title: String,
    pub censor_sub_title: bool,
    pub artist: String,
    pub censor_artist: bool,
    pub mapper: String,
    pub censor_mapper: bool,
    #[serde(rename = "MetaDataHasSplicedCensor")]
    pub metadata_has_spliced_censor: Option<bool>,
    pub duration: i32,
    pub votes: Vec<i32>,
    pub rating: f32,
    #[serde(with = "ts_seconds")]
    pub upload_time: DateTime<Utc>, // unix epoch
    #[serde(with = "ts_seconds")]
    pub last_updated: DateTime<Utc>, // unix epoch
    pub cover: String,
    pub automapped: bool,
    #[serde(rename = "ScoreSaberRanked")]
    pub scoresaber_ranked: bool,
    #[serde(rename = "BeatLeaderRanked")]
    pub beatleader_ranked: bool,
    pub curated: bool,
    pub curator_name: String,
    pub playlists: Vec<String>,
    pub vote_status: i32,
    pub uses_chroma: bool,
    pub uses_cinema: bool,
    pub uses_mapping_extensions: bool,
    pub uses_noodle_extensions: bool,
    pub uses_vivify: bool,
    pub has_played: bool,
    pub blacklisted: bool,
    pub diffs: Vec<DRMMapDiff>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DRMQueueItem {
    pub spot: i32,
    pub queue_item: DRMMap,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DRMHistoryItem {
    #[serde(with = "ts_seconds")]
    pub timestamp: DateTime<Utc>, // unix epoch
    pub history_item: DRMMap,
}

#[derive(Debug, Deserialize)]
pub struct DRMMessage {
    pub message: String,
}
