use axum::http::{HeaderName, HeaderValue, Method};
use tower_http::cors::{AllowHeaders, AllowMethods, CorsLayer};

static ALLOWED_ORIGINS: &[&str] = &[
    "http://localhost:3000",
    "http://localhost:3001",
    "http://127.0.0.1:3000",
    "http://localhost:8080",
    "http://127.0.0.1:8080",
    "http://localhost:8090",
    "http://localhost:6644",
    "http://127.0.0.1:8090",
    "http://127.0.0.1:6644",
    "http://localhost:5173",
    "http://127.0.0.1:5173",
    "http://localhost:5786",
    "http://127.0.0.1:5786",
    "http://localhost:4173",
    "https://localhost:4173",
    "https://localhost:5173",
];

static ALLOWED_SUFFIXES: &[&str] = &["analogwakatime.com", "testingmyproject.space", "localhost", "192.168.0.15"];

fn normalize_origin(origin: &str) -> String {
    let mut o = origin.trim().to_owned();
    while o.ends_with('/') {
        o.pop();
    }
    let Some(rest) = o.split_once("://") else {
        return o.to_lowercase();
    };
    let (scheme, authority) = rest;
    let (host_port, _) = authority.split_once('/').unwrap_or((authority, ""));
    let (host, port) = match host_port.rsplit_once(':') {
        Some((h, p)) if !p.contains(']') => (h, p),
        _ => (host_port, ""),
    };
    let scheme = scheme.to_lowercase();
    let host = host.to_lowercase();
    let default_port = (scheme == "http" && port == "80") || (scheme == "https" && port == "443");
    if port.is_empty() || default_port {
        format!("{scheme}://{host}")
    } else {
        format!("{scheme}://{host}:{port}")
    }
}

fn host_of(normalized: &str) -> &str {
    normalized
        .split_once("://")
        .map_or("", |(_, rest)| rest.split(':').next().unwrap_or(""))
}

pub fn is_allowed_origin(origin: &str) -> bool {
    let clean = normalize_origin(origin);
    let host = host_of(&clean);
    if matches!(
        host,
        "localhost" | "127.0.0.1" | "::1" | "0.0.0.0" | "[::1]"
    ) {
        return true;
    }
    if ALLOWED_SUFFIXES
        .iter()
        .any(|s| host == *s || host.ends_with(&format!(".{s}")))
    {
        return true;
    }
    ALLOWED_ORIGINS.iter().any(|a| normalize_origin(a) == clean)
}

pub fn cors_layer() -> CorsLayer {
    CorsLayer::new()
        .allow_origin(tower_http::cors::AllowOrigin::predicate(
            |origin: &HeaderValue, _| origin.to_str().is_ok_and(is_allowed_origin),
        ))
        .allow_methods(AllowMethods::list([
            Method::GET,
            Method::POST,
            Method::PUT,
            Method::DELETE,
            Method::OPTIONS,
            Method::PATCH,
        ]))
        .allow_headers(AllowHeaders::list([
            HeaderName::from_static("origin"),
            HeaderName::from_static("content-type"),
            HeaderName::from_static("authorization"),
            HeaderName::from_static("accept"),
            HeaderName::from_static("ngrok-skip-browser-warning"),
            HeaderName::from_static("x-requested-with"),
            HeaderName::from_static("access-control-request-method"),
            HeaderName::from_static("access-control-request-headers"),
            HeaderName::from_static("id"),
            HeaderName::from_static("x-request-id"),
        ]))
        .expose_headers([
            HeaderName::from_static("content-length"),
            HeaderName::from_static("content-type"),
            HeaderName::from_static("content-disposition"),
            HeaderName::from_static("authorization"),
            HeaderName::from_static("x-request-id"),
        ])
        .allow_credentials(true)
        .max_age(std::time::Duration::from_secs(12 * 3600))
}