use reqwest::Client;
use serde::de::DeserializeOwned;
use thiserror::Error;
use url::Url;

use crate::drm::schema::{DRMHistoryItem, DRMMap, DRMMessage, DRMQueueItem};

pub(crate) mod schema;

#[derive(Debug, Error)]
pub enum ClientError {
    #[error("Reqwest error")]
    ReqwestError(#[from] reqwest::Error),
    #[error("Serde error")]
    SerdeError(#[from] serde_json::Error),
    #[error("URL error")]
    URLError(#[from] url::ParseError),
}

type DRMResult<T> = Result<T, ClientError>;

pub struct DRM {
    client: Client,
    url: String,
}

impl DRM {
    pub fn new(host: String, port: i32) -> Self {
        let client = Client::new();
        let url = format!("{}:{}", host, port);

        Self { client, url }
    }

    pub async fn get_endpoint<T>(&self, endpoint: &str) -> DRMResult<T>
    where
        T: DeserializeOwned,
    {
        let full_url = Url::parse(&self.url)
            .map_err(ClientError::URLError)?
            .join(endpoint);

        let res = self
            .client
            .get(full_url?.as_str())
            .send()
            .await
            .map_err(ClientError::ReqwestError)?
            .text()
            .await?;

        serde_json::from_str::<T>(&res).map_err(ClientError::SerdeError)
    }

    // query map info
    pub async fn query(&self, id: &str) -> DRMResult<DRMMap> {
        self.get_endpoint(&format!("query/{}", id)).await
    }

    // query map info from beatsaver
    pub async fn query_nocache(&self, id: &str) -> DRMResult<DRMMap> {
        self.get_endpoint(&format!("query/nocache/{}", id)).await
    }

    // the entire queue
    pub async fn queue(&self) -> DRMResult<Vec<DRMMap>> {
        self.get_endpoint("queue").await
    }

    // all of a user's queue stuff
    pub async fn queue_where(&self, user: &str) -> DRMResult<Vec<DRMQueueItem>> {
        self.get_endpoint(&format!("queue/where/{}", user)).await
    }

    // clear/open/move/shuffle
    pub async fn queue_control(&self, subcommands: &str) -> DRMResult<DRMMessage> {
        self.get_endpoint(&format!("queue/{}", subcommands)).await
    }

    // all history for the session
    #[expect(unused)]
    pub async fn history(&self) -> DRMResult<Vec<DRMHistoryItem>> {
        self.get_endpoint("history").await
    }

    // most recently played map
    pub async fn link(&self) -> DRMResult<Vec<DRMHistoryItem>> {
        self.get_endpoint("history?limit=1").await
    }

    // add map to queue
    pub async fn add(&self, id: &str, user: &str) -> DRMResult<DRMMap> {
        self.get_endpoint(&format!("addKey/{}?user={}", id, user))
            .await
    }

    // add map to queue, with optional service
    pub async fn add_with_service(&self, id: &str, user: &str, service: &str) -> DRMResult<DRMMap> {
        self.get_endpoint(&format!("addKey/{}?user={}&service={}", id, user, service))
            .await
    }

    // add wip to queue
    pub async fn wip(&self, wip: &str, user: &str) -> DRMResult<DRMMap> {
        let drm_url = Url::parse(&self.url)
            .map_err(ClientError::URLError)?
            .join(&format!("addWip?user={}", user));

        let url: String = wip.into();

        // if !wip.starts_with("https://") {
        //     // assume it's a wipbot code
        //     url = format!("https://wipbot.com/wips/{}.zip", wip);
        // }

        let res = self
            .client
            .post(drm_url?.as_str())
            .body(url)
            .send()
            .await
            .map_err(ClientError::ReqwestError)?
            .text()
            .await?;

        serde_json::from_str::<DRMMap>(&res).map_err(ClientError::SerdeError)
    }
}
