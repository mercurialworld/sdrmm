use chrono::{DateTime, NaiveDate, NaiveDateTime, Utc};
use config::{Config, ConfigError, File, builder};
use num::{Num, zero};
use serde::Deserialize;

#[derive(Debug, Deserialize, Default, Clone)]
pub struct DRMConfig {
    pub url: String,
    pub port: i32,
}

#[derive(Debug, Deserialize, Default)]
pub struct QueueConfig {
    pub session_max: i32,
    pub queue_max: i32,
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
    pub date: BSRDateConfig,
    pub length: IntRange,
    pub nps: FloatRange,
    pub njs: FloatRange,
}

#[derive(Debug, Deserialize, Default)]
pub struct SDRMMConfig {
    pub drm: DRMConfig,
    pub queue: QueueConfig,
    pub bsr: BSRConfig,
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

pub fn ignore_config<T: Num>(val: T) -> bool {
    val.is_zero()
}
