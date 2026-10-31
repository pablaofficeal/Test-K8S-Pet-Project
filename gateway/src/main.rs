mod configs;
mod db;
mod metrics;
mod proxy;
mod state;
mod utils;

use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    Extension, Json, Router,
    body::Body,
    extract::{Request, State},
    handler::Handler,
    http::{StatusCode, header},
    middleware::{Next, from_fn, from_fn_with_state},
    response::{IntoResponse, Response},
    routing::{MethodRouter, any, delete, get, patch, post, put},
};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use configs::{
    cors::cors_layer,
    jwt_public_key, load_dotenv,
    routes::{LocalHandler, Route, RouteAuth, load as load_routes},
};
use db::RevocationStore;
use metrics::{Metrics, metrics_handler, metrics_middleware};
use proxy::{ProxyState, proxy_handler};
use state::AppState;
use utils::jwt::algorithm_from_name;
use utils::middleware::{
    auth::auth_middleware,
    rate_limit::{RateLimiter, check_request, client_ip},
    trace::{TraceId, trace_middleware},
};

async fn hello() -> Json<serde_json::Value> {
    Json(serde_json::json!({"message": "API Gateway is working"}))
}

async fn healthz() -> Json<serde_json::Value> {
    Json(serde_json::json!({"status": "ok", "service": "gateway"}))
}

async fn readyz() -> Json<serde_json::Value> {
    Json(serde_json::json!({"gateway": "ok"}))
}

const OPENAPI_SPEC: &str = include_str!("../configs/openapi/openapi.yml");
const OPENAPI_UI: &str = include_str!("../configs/openapi/html/index.html");

async fn openapi_spec() -> impl IntoResponse {
    (
        [(header::CONTENT_TYPE, "application/yaml; charset=utf-8")],
        OPENAPI_SPEC,
    )
}

async fn api_docs() -> impl IntoResponse {
    (
        [(header::CONTENT_TYPE, "text/html; charset=utf-8")],
        OPENAPI_UI,
    )
}

async fn not_found() -> Response {
    (
        StatusCode::NOT_FOUND,
        Json(serde_json::json!({"error": "route not found"})),
    )
        .into_response()
}

fn local_route<H, T>(method: &str, handler: H) -> MethodRouter<AppState>
where
    H: Handler<T, AppState>,
    T: 'static,
{
    match method {
        "GET" => get(handler),
        "POST" => post(handler),
        "PUT" => put(handler),
        "PATCH" => patch(handler),
        "DELETE" => delete(handler),
        _ => any(handler),
    }
}

async fn local_rate_limit(
    State((state, route)): State<(AppState, Route)>,
    req: Request<Body>,
    next: Next,
) -> Response {
    if let Some(limited) = check_request(&state, &req, &route) {
        return limited;
    }
    next.run(req).await
}

async fn gin_log_middleware(req: Request<Body>, next: Next) -> Response {
    let method = req.method().clone();
    let path = req.uri().path().to_owned();
    let ip = client_ip(&req);
    let trace_id = TraceId::from_request(&req).to_owned();
    let start = std::time::Instant::now();
    let response = next.run(req).await;
    tracing::info!(
        "[GATEWAY] [trace_id={trace_id}] {} | {:?} | {ip} | {method} \"{path}\"",
        response.status().as_u16(),
        start.elapsed()
    );
    response
}

fn register_route(app: Router<AppState>, route: &Route, state: &AppState) -> Router<AppState> {
    let handler = match route.handler {
        Some(LocalHandler::Hello) => local_route(&route.method, hello),
        Some(LocalHandler::Healthz) => local_route(&route.method, healthz),
        Some(LocalHandler::Readyz) => local_route(&route.method, readyz),
        Some(LocalHandler::Metrics) => local_route(&route.method, metrics_handler),
        Some(LocalHandler::Docs) => local_route(&route.method, api_docs),
        Some(LocalHandler::Openapi) => local_route(&route.method, openapi_spec),
        None => {
            let proxy_state = ProxyState {
                route: route.clone(),
                app: state.clone(),
            };
            match route.method.as_str() {
                "GET" => get(proxy_handler),
                "POST" => post(proxy_handler),
                "PUT" => put(proxy_handler),
                "PATCH" => patch(proxy_handler),
                "DELETE" => delete(proxy_handler),
                _ => any(proxy_handler),
            }
            .route_layer(Extension(proxy_state))
        }
    }
    .route_layer(from_fn_with_state(
        (state.clone(), route.clone()),
        local_rate_limit,
    ));

    let handler = if route.auth == RouteAuth::Private {
        handler.route_layer(from_fn_with_state(state.clone(), auth_middleware))
    } else {
        handler
    };
    app.route(&route.path, handler)
}

fn jwt_algorithm() -> String {
    std::env::var("JWT_ALGORITHM").expect("JWT_ALGORITHM")
}

async fn shutdown_signal() {
    let ctrl_c = async {
        tokio::signal::ctrl_c()
            .await
            .expect("install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
            .expect("install SIGTERM handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }
}

#[tokio::main]
async fn main() {
    load_dotenv();
    let cfg = load_routes("configs/routers/routes.yaml").expect("load configs/routers/routes.yaml");
    if let Err(e) = jwt_public_key() {
        panic!("{e}");
    }
    algorithm_from_name(&jwt_algorithm());

    std::fs::create_dir_all("logs").expect("create logs dir");
    let file_appender = tracing_appender::rolling::never("logs", "main-api-gateway.log");
    let (file_writer, file_guard) = tracing_appender::non_blocking(file_appender);

    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info".into()),
        )
        .with(
            tracing_subscriber::fmt::layer()
                .with_ansi(false)
                .with_writer(file_writer),
        )
        .init();
    let _file_guard = file_guard;

    let state = AppState {
        client: reqwest::Client::builder()
            .build()
            .expect("build reqwest client"),
        store: RevocationStore::default(),
        limiter: RateLimiter::new(),
        metrics: Arc::new(Metrics::default()),
        jwt_algorithm: jwt_algorithm(),
    };
    let listen_addr = cfg.listen_addr.clone();

    let mut app = Router::new();
    for route in &cfg.routes {
        app = register_route(app, route, &state);
    }

    let app = app
        .fallback(not_found)
        .layer(from_fn_with_state(state.clone(), metrics_middleware))
        .layer(from_fn(gin_log_middleware))
        .layer(cors_layer())
        .layer(from_fn(trace_middleware))
        .with_state(state);

    let listener = tokio::net::TcpListener::bind(&listen_addr)
        .await
        .expect("bind gateway listener");
    tracing::info!("gateway listening on {listen_addr}");

    axum::serve(
        listener,
        app.into_make_service_with_connect_info::<SocketAddr>(),
    )
    .with_graceful_shutdown(shutdown_signal())
    .await
    .expect("serve gateway");
}