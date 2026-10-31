use std::sync::Arc;

use crate::{db::RevocationStore, metrics::Metrics, utils::middleware::rate_limit::RateLimiter};

#[derive(Clone)]
pub struct AppState {
    pub client: reqwest::Client,
    pub store: RevocationStore,
    pub limiter: RateLimiter,
    pub metrics: Arc<Metrics>,
    pub jwt_algorithm: String,
}