use std::sync::atomic::Ordering;

use axum::{
    body::Body,
    extract::State,
    http::{HeaderMap, HeaderValue, Method, Request, StatusCode, header},
    middleware::Next,
    response::{IntoResponse, Response},
};

use crate::{
    db::session_allowed,
    state::AppState,
    utils::jwt::{algorithm_from_name, validate_token},
};

const ACCESS_COOKIE: &str = "Cadence-Cloud-Auth-Token";
const CSRF_HEADER: &str = "x-requested-with";

fn bearer_token(headers: &HeaderMap) -> Option<String> {
    let value = headers.get(header::AUTHORIZATION)?.to_str().ok()?;
    let (scheme, token) = value.split_once(' ')?;
    (scheme.eq_ignore_ascii_case("bearer") && !token.is_empty()).then(|| token.to_owned())
}

fn cookie_token(headers: &HeaderMap) -> Option<String> {
    headers
        .get_all(header::COOKIE)
        .iter()
        .filter_map(|v| v.to_str().ok())
        .flat_map(|v| v.split(';'))
        .filter_map(|pair| pair.trim().split_once('='))
        .find(|(name, value)| *name == ACCESS_COOKIE && !value.is_empty())
        .map(|(_, value)| value.to_owned())
}

pub async fn auth_middleware(
    State(state): State<AppState>,
    req: Request<Body>,
    next: Next,
) -> Response {
    let fail = |msg: &str| {
        state.metrics.auth_failures.fetch_add(1, Ordering::Relaxed);
        unauthorized(msg)
    };

    let (token, from_cookie) = match bearer_token(req.headers()) {
        Some(token) => (token, false),
        None => match cookie_token(req.headers()) {
            Some(token) => (token, true),
            None => return fail("authorization required"),
        },
    };

    let safe_method = matches!(*req.method(), Method::GET | Method::HEAD | Method::OPTIONS);
    if from_cookie && !safe_method && !req.headers().contains_key(CSRF_HEADER) {
        state.metrics.auth_failures.fetch_add(1, Ordering::Relaxed);
        return (
            StatusCode::FORBIDDEN,
            axum::Json(serde_json::json!({"error": "csrf header required"})),
        )
            .into_response();
    }

    let algorithm = algorithm_from_name(&state.jwt_algorithm);
    match validate_token(&token, algorithm) {
        Ok(claims) => {
            if !session_allowed(&state.store, &claims) {
                return fail("session revoked");
            }
            let mut req = req;
            if from_cookie {
                // upstream-сервисы ждут Bearer, а куку им отдавать незачем
                if let Ok(value) = HeaderValue::from_str(&format!("Bearer {token}")) {
                    req.headers_mut().insert(header::AUTHORIZATION, value);
                }
                req.headers_mut().remove(header::COOKIE);
            }
            req.extensions_mut().insert(claims);
            next.run(req).await
        }
        Err(_) => fail("invalid or expired token"),
    }
}

fn unauthorized(msg: &str) -> Response {
    (
        StatusCode::UNAUTHORIZED,
        axum::Json(serde_json::json!({"error": msg})),
    )
        .into_response()
}