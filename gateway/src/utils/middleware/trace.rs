use axum::{
    body::Body,
    http::{HeaderName, HeaderValue, Request},
    middleware::Next,
    response::Response,
};

pub const REQUEST_ID_HEADER: &str = "x-request-id";

#[derive(Debug, Clone)]
pub struct TraceId(pub String);

impl TraceId {
    pub fn from_request(req: &Request<Body>) -> &str {
        req.extensions()
            .get::<TraceId>()
            .map(|t| t.0.as_str())
            .unwrap_or("-")
    }
}

pub async fn trace_middleware(mut req: Request<Body>, next: Next) -> Response {
    let trace_id = req
        .headers()
        .get(REQUEST_ID_HEADER)
        .and_then(|v| v.to_str().ok())
        .map(str::trim)
        .filter(|v| !v.is_empty())
        .map(str::to_owned)
        .unwrap_or_else(|| uuid::Uuid::new_v4().to_string());

    let header = HeaderName::from_static(REQUEST_ID_HEADER);
    if let Ok(value) = HeaderValue::from_str(&trace_id) {
        req.headers_mut().insert(header.clone(), value);
    }
    req.extensions_mut().insert(TraceId(trace_id.clone()));

    let mut response = next.run(req).await;
    if let Ok(value) = HeaderValue::from_str(&trace_id) {
        response.headers_mut().insert(header, value);
    }
    response
}