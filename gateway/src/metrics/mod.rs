use std::sync::atomic::{AtomicU64, Ordering};

use axum::{Json, body::Body, extract::State, http::Request, middleware::Next, response::Response};

use crate::state::AppState;

#[derive(Debug, Default)]
pub struct Metrics {
    pub total_requests: AtomicU64,
    pub proxied_requests: AtomicU64,
    pub auth_failures: AtomicU64,
    pub upstream_errors: AtomicU64,
    pub rate_limited: AtomicU64,
}

impl Metrics {
    pub fn snapshot(&self) -> serde_json::Value {
        serde_json::json!({
            "total_requests": self.total_requests.load(Ordering::Relaxed),
            "proxied_requests": self.proxied_requests.load(Ordering::Relaxed),
            "auth_failures": self.auth_failures.load(Ordering::Relaxed),
            "upstream_errors": self.upstream_errors.load(Ordering::Relaxed),
            "rate_limited": self.rate_limited.load(Ordering::Relaxed),
        })
    }
}

pub async fn metrics_middleware(
    State(state): State<AppState>,
    req: Request<Body>,
    next: Next,
) -> Response {
    state.metrics.total_requests.fetch_add(1, Ordering::Relaxed);
    next.run(req).await
}

pub async fn metrics_handler(State(state): State<AppState>) -> Json<serde_json::Value> {
    Json(state.metrics.snapshot())
}