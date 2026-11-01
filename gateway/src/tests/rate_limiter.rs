use std::net::SocketAddr;

use axum::{
    body::Body,
    extract::ConnectInfo,
    http::{Request, StatusCode},
};

use crate::configs::routes::{RateLimit, Route, RouteAuth};
use crate::db::RevocationStore;
use crate::metrics::Metrics;
use crate::state::AppState;
use crate::utils::middleware::rate_limit::{RateLimiter, check_request, client_ip};
use std::sync::{Arc, atomic::Ordering};

fn state() -> AppState {
    AppState {
        client: reqwest::Client::new(),
        store: RevocationStore::default(),
        limiter: RateLimiter::new(),
        metrics: Arc::new(Metrics::default()),
        jwt_algorithm: "EdDSA".to_owned(),
    }
}

fn route(rate: Option<RateLimit>) -> Route {
    Route {
        method: "GET".to_owned(),
        path: "/main/api/v1/x".to_owned(),
        upstream: "http://127.0.0.1:1".to_owned(),
        prefix: "/x".to_owned(),
        timeout_secs: 0,
        auth: RouteAuth::Public,
        websocket: false,
        handler: None,
        rate,
    }
}

fn request(headers: &[(&str, &str)]) -> Request<Body> {
    let mut builder = Request::builder().uri("/main/api/v1/x");
    for (name, value) in headers {
        builder = builder.header(*name, *value);
    }
    builder.body(Body::empty()).expect("request")
}

#[test]
fn allows_up_to_the_limit_then_blocks() {
    let limiter = RateLimiter::new();
    assert!(limiter.check("k", 3, 100_000));
    assert!(limiter.check("k", 3, 100_000));
    assert!(limiter.check("k", 3, 100_000));
    assert!(!limiter.check("k", 3, 100_000));
    assert!(!limiter.check("k", 3, 100_000));
}

#[test]
fn keys_are_independent() {
    let limiter = RateLimiter::new();
    assert!(limiter.check("a", 1, 100_000));
    assert!(!limiter.check("a", 1, 100_000));
    assert!(limiter.check("b", 1, 100_000));
}

#[test]
fn zero_limit_blocks_everything() {
    assert!(!RateLimiter::new().check("k", 0, 100_000));
}

#[test]
fn window_resets_the_counter() {
    let limiter = RateLimiter::new();
    assert!(limiter.check("k", 1, 1));
    std::thread::sleep(std::time::Duration::from_millis(2100));
    assert!(limiter.check("k", 1, 1));
}

#[test]
fn cloned_limiter_shares_counters() {
    let limiter = RateLimiter::new();
    let clone = limiter.clone();
    assert!(limiter.check("k", 1, 100_000));
    assert!(!clone.check("k", 1, 100_000));
}

#[test]
fn zero_window_is_treated_as_one_second() {
    assert!(RateLimiter::new().check("k", 1, 0));
}

#[test]
fn client_ip_prefers_the_first_forwarded_address() {
    assert_eq!(client_ip(&request(&[("x-forwarded-for", "1.1.1.1, 2.2.2.2")])), "1.1.1.1");
    assert_eq!(client_ip(&request(&[("x-forwarded-for", "  3.3.3.3  ")])), "3.3.3.3");
}

#[test]
fn client_ip_falls_back_to_real_ip_then_the_socket() {
    assert_eq!(client_ip(&request(&[("x-real-ip", "4.4.4.4")])), "4.4.4.4");
    assert_eq!(client_ip(&request(&[("x-forwarded-for", ""), ("x-real-ip", "5.5.5.5")])), "5.5.5.5");

    let mut with_socket = request(&[]);
    let addr: SocketAddr = "6.6.6.6:1234".parse().expect("addr");
    with_socket.extensions_mut().insert(ConnectInfo(addr));
    assert_eq!(client_ip(&with_socket), "6.6.6.6");
}

#[test]
fn client_ip_is_a_dash_when_unknown() {
    assert_eq!(client_ip(&request(&[])), "-");
}

#[test]
fn route_without_a_rate_never_limits() {
    let state = state();
    for _ in 0..50 {
        assert!(check_request(&state, &request(&[]), &route(None)).is_none());
    }
}

#[test]
fn exceeding_the_route_limit_gives_429_and_counts_it() {
    let state = state();
    let route = route(Some(RateLimit { limit: 2, window_secs: 100_000 }));
    let req = request(&[("x-forwarded-for", "9.9.9.9")]);
    assert!(check_request(&state, &req, &route).is_none());
    assert!(check_request(&state, &req, &route).is_none());
    let blocked = check_request(&state, &req, &route).expect("third request is limited");
    assert_eq!(blocked.status(), StatusCode::TOO_MANY_REQUESTS);
    assert_eq!(state.metrics.rate_limited.load(Ordering::Relaxed), 1);
}

#[test]
fn limit_is_counted_per_client_and_per_route() {
    let state = state();
    let route = route(Some(RateLimit { limit: 1, window_secs: 100_000 }));
    assert!(check_request(&state, &request(&[("x-forwarded-for", "1.1.1.1")]), &route).is_none());
    assert!(check_request(&state, &request(&[("x-forwarded-for", "2.2.2.2")]), &route).is_none());
    assert!(check_request(&state, &request(&[("x-forwarded-for", "1.1.1.1")]), &route).is_some());
}
