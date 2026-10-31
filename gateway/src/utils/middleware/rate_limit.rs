use std::{
    collections::HashMap,
    net::SocketAddr,
    sync::{Arc, Mutex},
    time::{SystemTime, UNIX_EPOCH},
};

use axum::{
    body::Body,
    extract::ConnectInfo,
    http::{Request, StatusCode},
    response::{IntoResponse, Response},
};

use crate::{configs::routes::Route, state::AppState};

#[derive(Debug, Clone)]
pub struct RateLimiter {
    windows: Arc<Mutex<HashMap<String, (u64, u32)>>>,
}

impl RateLimiter {
    pub fn new() -> Self {
        Self {
            windows: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn check(&self, key: &str, limit: u32, window_secs: u64) -> bool {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or(0);
        let window_secs = window_secs.max(1);
        let slot = now / window_secs;
        let mut windows = self.windows.lock().expect("rate limit lock poisoned");
        match windows.get_mut(key) {
            Some((s, n)) if *s == slot => {
                *n += 1;
                *n <= limit
            }
            _ => {
                windows.insert(key.to_owned(), (slot, 1));
                1 <= limit
            }
        }
    }
}

pub fn client_ip(req: &Request<Body>) -> String {
    if let Some(forwarded) = req
        .headers()
        .get("x-forwarded-for")
        .and_then(|v| v.to_str().ok())
        && let Some(first) = forwarded.split(',').next().map(str::trim)
        && !first.is_empty()
    {
        return first.to_owned();
    }
    if let Some(real) = req
        .headers()
        .get("x-real-ip")
        .and_then(|v| v.to_str().ok())
        .map(str::trim)
        && !real.is_empty()
    {
        return real.to_owned();
    }
    req.extensions()
        .get::<ConnectInfo<SocketAddr>>()
        .map(|c| c.0.ip().to_string())
        .unwrap_or_else(|| "-".to_owned())
}

pub fn check_request(state: &AppState, req: &Request<Body>, route: &Route) -> Option<Response> {
    let rate = route.rate.as_ref()?;
    if state.limiter.check(
        &format!("{}:{}", client_ip(req), route.path),
        rate.limit,
        rate.window_secs,
    ) {
        return None;
    }
    state
        .metrics
        .rate_limited
        .fetch_add(1, std::sync::atomic::Ordering::Relaxed);
    Some(
        (
            StatusCode::TOO_MANY_REQUESTS,
            axum::Json(serde_json::json!({"error": "TOO MANY REQUESTS"})),
        )
            .into_response(),
    )
}