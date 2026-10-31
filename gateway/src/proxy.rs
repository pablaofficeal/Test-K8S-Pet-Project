use std::sync::atomic::Ordering;
use std::time::Duration;

use axum::{
    Extension,
    body::Body,
    extract::Request,
    http::{HeaderMap, StatusCode},
    response::{IntoResponse, Response},
};
use http_body_util::BodyExt;

use crate::{configs::routes::Route, state::AppState};

#[derive(Clone)]
pub struct ProxyState {
    pub route: Route,
    pub app: AppState,
}

fn is_hop_by_hop(name: &str) -> bool {
    matches!(
        name,
        "connection"
            | "keep-alive"
            | "proxy-authenticate"
            | "proxy-authorization"
            | "te"
            | "trailer"
            | "transfer-encoding"
            | "upgrade"
            | "host"
            | "content-length"
    )
}

fn is_websocket_upgrade(headers: &HeaderMap) -> bool {
    let upgrade = headers
        .get("upgrade")
        .and_then(|v| v.to_str().ok())
        .unwrap_or_default();
    let connection = headers
        .get("connection")
        .and_then(|v| v.to_str().ok())
        .unwrap_or_default()
        .to_lowercase();
    upgrade.eq_ignore_ascii_case("websocket") && connection.contains("upgrade")
}

pub fn rewrite_path(route_path: &str, prefix: &str, req_path: &str) -> String {
    if !prefix.contains(':') {
        let stripped = req_path.strip_prefix(route_path).unwrap_or(req_path);
        return format!("{prefix}{stripped}");
    }
    let mut params: Vec<(&str, &str)> = Vec::new();
    for (r, q) in route_path.split('/').zip(req_path.split('/')) {
        if let Some(name) = r.strip_prefix(':') {
            params.push((name, q));
        }
    }
    let mut out = prefix.to_owned();
    for (name, value) in params {
        out = out.replace(&format!(":{name}"), value);
    }
    out
}

fn json_error(status: StatusCode, msg: &str) -> Response {
    (status, axum::Json(serde_json::json!({"error": msg}))).into_response()
}

fn count_upstream_error(state: &ProxyState) {
    state
        .app
        .metrics
        .upstream_errors
        .fetch_add(1, Ordering::Relaxed);
}

pub async fn proxy_handler(
    Extension(state): Extension<ProxyState>,
    req: Request<Body>,
) -> Response {
    let route = state.route.clone();

    if is_websocket_upgrade(req.headers()) {
        count_upstream_error(&state);
        if route.websocket {
            return json_error(
                StatusCode::NOT_IMPLEMENTED,
                "websocket not supported by upstream",
            );
        }
        return json_error(
            StatusCode::NOT_IMPLEMENTED,
            "websocket not enabled for route",
        );
    }

    state
        .app
        .metrics
        .proxied_requests
        .fetch_add(1, Ordering::Relaxed);

    let req_path = req.uri().path().to_owned();
    let query = req
        .uri()
        .query()
        .map(|q| format!("?{q}"))
        .unwrap_or_default();
    let upstream_path = rewrite_path(&route.path, &route.prefix, &req_path);
    let url = format!("{}{}{}", route.upstream, upstream_path, query);

    let method = req.method().clone();
    let original_host = req
        .headers()
        .get("host")
        .and_then(|v| v.to_str().ok())
        .unwrap_or_default()
        .to_owned();
    let headers_in = req.headers().clone();
    let body_bytes = match req.into_body().collect().await {
        Ok(collected) => collected.to_bytes(),
        Err(_) => {
            count_upstream_error(&state);
            return json_error(StatusCode::BAD_GATEWAY, "failed to read body");
        }
    };

    let mut outgoing = state.app.client.request(method, url.as_str());
    for (name, value) in headers_in.iter() {
        if is_hop_by_hop(name.as_str()) {
            continue;
        }
        outgoing = outgoing.header(name, value);
    }
    outgoing = outgoing
        .header("X-Forwarded-Host", original_host)
        .header("X-Origin-Host", route.upstream.clone())
        .header("X-Gateway-Forwarded", "true");
    if !body_bytes.is_empty() {
        outgoing = outgoing.body(body_bytes.to_vec());
    }

    let send_result = if route.timeout_secs > 0 {
        match tokio::time::timeout(Duration::from_secs(route.timeout_secs), outgoing.send()).await {
            Ok(r) => r,
            Err(_) => {
                tracing::warn!(upstream = route.upstream, "proxy upstream timeout");
                count_upstream_error(&state);
                return json_error(StatusCode::GATEWAY_TIMEOUT, "upstream timeout");
            }
        }
    } else {
        outgoing.send().await
    };
    let upstream_resp = match send_result {
        Ok(r) => r,
        Err(err) => {
            tracing::warn!(%err, upstream = route.upstream, "proxy upstream error");
            count_upstream_error(&state);
            return json_error(StatusCode::BAD_GATEWAY, "upstream unreachable");
        }
    };

    let status =
        StatusCode::from_u16(upstream_resp.status().as_u16()).unwrap_or(StatusCode::BAD_GATEWAY);
    let mut headers_out = HeaderMap::new();
    for (name, value) in upstream_resp.headers().iter() {
        if is_hop_by_hop(name.as_str()) {
            continue;
        }
        headers_out.append(name, value.clone());
    }
    match upstream_resp.bytes().await {
        Ok(body) => (status, headers_out, body.to_vec()).into_response(),
        Err(_) => {
            count_upstream_error(&state);
            json_error(StatusCode::BAD_GATEWAY, "failed to read upstream body")
        }
    }
}