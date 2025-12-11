use chrono::NaiveDate;
use config::{Config, ConfigError, File};
use serde::Deserialize;

#[derive(Debug, Deserialize, Default, Clone)]
pub struct DRMConfig {
    pub url: String,
    pub port: u32,
}

#[derive(Debug, Deserialize, Default)]
pub struct QueueConfig {
    pub session_max: u32,
    pub queue_max: u32,
    pub repeat: bool,
    pub replay: bool,
}

#[derive(Debug, Deserialize, Default)]
pub struct BSRDateConfig {
    pub earliest: NaiveDate,
    pub min_age: i32,
}

#[derive(Debug, Deserialize, Default)]
pub struct IntRange {
    pub min: i32,
    pub max: i32,
}

#[derive(Debug, Deserialize, Default)]
pub struct FloatRange {
    pub min: f32,
    pub max: f32,
}

#[derive(Debug, Deserialize, Default)]
pub struct BSRConfig {
    pub censors: BSRCensorConfig,
    pub allow_ai: bool,
    pub min_rating: f32,
    pub date: BSRDateConfig,
    pub length: IntRange,
    pub nps: FloatRange,
    pub njs: FloatRange,
}

#[derive(Debug, Deserialize, Default)]
pub struct BSRCensorConfig {
    pub deny_censored: bool,
    pub deny_urls: bool,
}

#[derive(Debug, Deserialize, Default)]
pub struct MapVoteConfig {
    pub allow_liked: bool,
    pub deny_disliked: bool,
}

#[derive(Debug, Deserialize, Default)]
pub struct SDRMMConfig {
    pub drm: DRMConfig,
    pub queue: QueueConfig,
    pub bsr: BSRConfig,
    pub map_vote: MapVoteConfig,
    pub allowed_playlists: Option<Vec<String>>,
    pub banned_mappers: Option<Vec<String>>,
}

impl SDRMMConfig {
    pub fn new(filename: &str) -> Result<Self, ConfigError> {
        Config::builder()
            .add_source(File::with_name(filename))
            .build()
            .unwrap()
            .try_deserialize::<SDRMMConfig>()
    }
}
