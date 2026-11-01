use std::net::SocketAddr;
use std::sync::{Arc, Once};
use std::time::{SystemTime, UNIX_EPOCH};

use axum::{
    Json, Router,
    body::Body,
    extract::Request,
    http::{HeaderValue, StatusCode, header},
    response::{IntoResponse, Response},
    routing::{any, get},
};
use jsonwebtoken::{Algorithm, EncodingKey, Header, encode};
use serde_json::{Map, Value, json};

use super::keys::{PRIVATE_A, PUBLIC_A};
use crate::configs::routes::load;
use crate::db::RevocationStore;
use crate::metrics::Metrics;
use crate::state::AppState;
use crate::utils::middleware::rate_limit::RateLimiter;

static ENV: Once = Once::new();

pub fn init_env() {
    ENV.call_once(|| unsafe {
        std::env::set_var("JWT_PUBLIC_KEY", PUBLIC_A);
        std::env::set_var("JWT_ALGORITHM", "EdDSA");
    });
}

pub fn now() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("clock")
        .as_secs() as i64
}

pub fn claims(purpose: Option<&str>, uuid: &str, sid: &str) -> Value {
    let mut value = json!({
        "uuid": uuid,
        "email": "user@example.com",
        "sid": sid,
        "ver": 0,
        "jti": sid,
        "iat": now(),
        "exp": now() + 300,
    });
    if let Some(purpose) = purpose {
        value["purpose"] = json!(purpose);
    }
    value
}

pub fn sign(claims: &Value, private_pem: &str) -> String {
    let key = EncodingKey::from_ed_pem(private_pem.as_bytes()).expect("test key");
    encode(&Header::new(Algorithm::EdDSA), claims, &key).expect("sign token")
}

pub fn access_token(uuid: &str, sid: &str) -> String {
    sign(&claims(Some("access"), uuid, sid), PRIVATE_A)
}

async fn echo(req: Request<Body>) -> Json<Value> {
    let (parts, body) = req.into_parts();
    let mut headers = Map::new();
    for (name, value) in parts.headers.iter() {
        headers.insert(
            name.as_str().to_owned(),
            json!(value.to_str().unwrap_or("<binary>")),
        );
    }
    let bytes = axum::body::to_bytes(body, usize::MAX)
        .await
        .expect("read body");
    Json(json!({
        "method": parts.method.as_str(),
        "path": parts.uri.path(),
        "query": parts.uri.query().unwrap_or(""),
        "headers": headers,
        "body": String::from_utf8_lossy(&bytes),
    }))
}

async fn status(axum::extract::Path(code): axum::extract::Path<u16>) -> Response {
    (
        StatusCode::from_u16(code).expect("status"),
        Json(json!({"status": code})),
    )
        .into_response()
}

async fn slow() -> &'static str {
    tokio::time::sleep(std::time::Duration::from_secs(3)).await;
    "late"
}

async fn cookies() -> Response {
    let mut response = Json(json!({"ok": true})).into_response();
    response
        .headers_mut()
        .append(header::SET_COOKIE, HeaderValue::from_static("a=1; Path=/"));
    response
        .headers_mut()
        .append(header::SET_COOKIE, HeaderValue::from_static("b=2; Path=/x"));
    response
        .headers_mut()
        .insert("x-upstream", HeaderValue::from_static("yes"));
    response
}

pub async fn start_upstream() -> String {
    let app = Router::new()
        .route("/echo", any(echo))
        .route("/echo/*rest", any(echo))
        .route("/status/:code", get(status))
        .route("/slow", get(slow))
        .route("/cookies", get(cookies));
    let listener = tokio::net::TcpListener::bind("127.0.0.1:0")
        .await
        .expect("bind upstream");
    let addr = listener.local_addr().expect("addr");
    tokio::spawn(async move {
        axum::serve(listener, app).await.expect("serve upstream");
    });
    format!("http://{addr}")
}

pub fn routes_yaml(upstream: &str) -> String {
    format!(
        r#"
server:
  port: 0
  gateway_host: 127.0.0.1

microservices_hosts:
  svc:
    host: {upstream}
    timeout: 5s
  slow:
    host: {upstream}
    timeout: 1s
  down:
    host: http://127.0.0.1:1
    timeout: 2s

public_routes:
  - path: /main/api/v1/
    method: GET
    target: local
    handler: hello
  - path: /main/api/v1/healthz
    method: GET
    target: local
    handler: healthz
  - path: /main/api/v1/readyz
    method: GET
    target: local
    handler: readyz
  - path: /main/api/v1/metrics
    method: GET
    target: local
    handler: metrics
  - path: /main/api/v1/docs
    method: GET
    target: local
    handler: docs
  - path: /main/api/v1/openapi/openapi.yml
    method: GET
    target: local
    handler: openapi
  - path: /main/api/v1/pub/echo
    target: ${{microservices_hosts.svc}}
    timeout: ${{microservices_hosts.svc.timeout}}
    prefix: /echo
  - path: /main/api/v1/pub/status/:code
    method: GET
    target: ${{microservices_hosts.svc}}
    timeout: 5s
    prefix: /status/:code
  - path: /main/api/v1/pub/cookies
    method: GET
    target: ${{microservices_hosts.svc}}
    timeout: 5s
    prefix: /cookies
  - path: /main/api/v1/pub/slow
    method: GET
    target: ${{microservices_hosts.slow}}
    timeout: ${{microservices_hosts.slow.timeout}}
    prefix: /slow
  - path: /main/api/v1/pub/down
    method: GET
    target: ${{microservices_hosts.down}}
    timeout: ${{microservices_hosts.down.timeout}}
    prefix: /anything
  - path: /main/api/v1/pub/limited
    method: GET
    target: ${{microservices_hosts.svc}}
    timeout: 5s
    prefix: /echo
    rate_limit:
      - by: ip
        limit: 2
        window: 100000h
  - path: /main/api/v1/pub/limited-other
    method: GET
    target: ${{microservices_hosts.svc}}
    timeout: 5s
    prefix: /echo
    rate_limit:
      - by: ip
        limit: 1
        window: 100000h
  - path: /main/api/v1/pub/websocket
    method: GET
    target: ${{microservices_hosts.svc}}
    timeout: 5s
    prefix: /echo
    websocket: true

private_routes:
  - path: /main/api/v1/priv/echo
    target: ${{microservices_hosts.svc}}
    timeout: 5s
    prefix: /echo
  - path: /main/api/v1/priv/item/:id
    method: GET
    target: ${{microservices_hosts.svc}}
    timeout: 5s
    prefix: /echo/item/:id
"#
    )
}

pub struct TestGateway {
    pub base: String,
    pub client: reqwest::Client,
    pub state: AppState,
}

impl TestGateway {
    pub async fn start() -> Self {
        let upstream = start_upstream().await;
        let path = std::env::temp_dir().join(format!("gateway-test-{}.yaml", uuid::Uuid::new_v4()));
        std::fs::write(&path, routes_yaml(&upstream)).expect("write routes");
        let cfg = load(path.to_str().expect("path")).expect("load routes");
        let _ = std::fs::remove_file(&path);
        Self::from_routes(&cfg.routes).await
    }

    pub async fn from_routes(routes: &[crate::configs::routes::Route]) -> Self {
        init_env();
        let state = AppState {
            client: reqwest::Client::builder().build().expect("client"),
            store: RevocationStore::default(),
            limiter: RateLimiter::new(),
            metrics: Arc::new(Metrics::default()),
            jwt_algorithm: "EdDSA".to_owned(),
        };
        let app = crate::build_app(routes, state.clone());
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0")
            .await
            .expect("bind gateway");
        let addr = listener.local_addr().expect("addr");
        tokio::spawn(async move {
            axum::serve(
                listener,
                app.into_make_service_with_connect_info::<SocketAddr>(),
            )
            .await
            .expect("serve gateway");
        });
        Self {
            base: format!("http://{addr}"),
            client: reqwest::Client::builder()
                .redirect(reqwest::redirect::Policy::none())
                .build()
                .expect("client"),
            state,
        }
    }

    pub fn url(&self, path: &str) -> String {
        format!("{}/main/api/v1{}", self.base, path)
    }
}

pub fn json_of(text: &str) -> Value {
    serde_json::from_str(text).expect("json body")
}
