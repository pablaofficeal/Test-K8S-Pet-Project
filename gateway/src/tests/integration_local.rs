use super::support::{TestGateway, json_of};

#[tokio::test]
async fn hello_healthz_and_readyz() {
    let gateway = TestGateway::start().await;
    let hello = gateway.client.get(gateway.url("/")).send().await.expect("response");
    assert_eq!(hello.status(), 200);
    assert_eq!(json_of(&hello.text().await.expect("body"))["message"], "API Gateway is working");

    let health = json_of(&gateway.client.get(gateway.url("/healthz")).send().await.expect("response").text().await.expect("body"));
    assert_eq!(health["status"], "ok");
    assert_eq!(health["service"], "gateway");

    let ready = json_of(&gateway.client.get(gateway.url("/readyz")).send().await.expect("response").text().await.expect("body"));
    assert_eq!(ready["gateway"], "ok");
}

#[tokio::test]
async fn metrics_endpoint_reports_the_counters() {
    let gateway = TestGateway::start().await;
    gateway.client.get(gateway.url("/pub/echo")).send().await.expect("response");
    let metrics = json_of(&gateway.client.get(gateway.url("/metrics")).send().await.expect("response").text().await.expect("body"));
    assert_eq!(metrics["proxied_requests"], 1);
    assert_eq!(metrics["total_requests"], 2);
    for key in ["auth_failures", "upstream_errors", "rate_limited"] {
        assert!(metrics.get(key).is_some(), "{key}");
    }
}

#[tokio::test]
async fn docs_and_openapi_are_served() {
    let gateway = TestGateway::start().await;
    let docs = gateway.client.get(gateway.url("/docs")).send().await.expect("response");
    assert_eq!(docs.status(), 200);
    assert!(docs.headers()["content-type"].to_str().expect("type").starts_with("text/html"));
    assert!(docs.text().await.expect("body").to_lowercase().contains("<html"));

    let spec = gateway.client.get(gateway.url("/openapi/openapi.yml")).send().await.expect("response");
    assert_eq!(spec.status(), 200);
    assert!(spec.headers()["content-type"].to_str().expect("type").starts_with("application/yaml"));
    let text = spec.text().await.expect("body");
    assert!(text.contains("openapi:"));
    assert!(text.contains("/main/api/v1/start/log/in/my/account"));
}

#[tokio::test]
async fn preflight_from_an_allowed_origin_is_answered() {
    let gateway = TestGateway::start().await;
    let response = gateway
        .client
        .request(reqwest::Method::OPTIONS, gateway.url("/priv/echo"))
        .header("origin", "http://localhost:5173")
        .header("access-control-request-method", "PATCH")
        .header("access-control-request-headers", "content-type,x-requested-with")
        .send()
        .await
        .expect("response");
    assert!(response.status().is_success());
    let headers = response.headers();
    assert_eq!(headers["access-control-allow-origin"], "http://localhost:5173");
    assert_eq!(headers["access-control-allow-credentials"], "true");
    assert!(headers["access-control-allow-methods"].to_str().expect("methods").contains("PATCH"));
    assert!(headers["access-control-allow-headers"].to_str().expect("headers").to_lowercase().contains("x-requested-with"));
}

#[tokio::test]
async fn preflight_needs_no_authentication() {
    let gateway = TestGateway::start().await;
    let response = gateway
        .client
        .request(reqwest::Method::OPTIONS, gateway.url("/priv/echo"))
        .header("origin", "https://app.analogwakatime.com")
        .header("access-control-request-method", "GET")
        .send()
        .await
        .expect("response");
    assert!(response.status().is_success());
}

#[tokio::test]
async fn preflight_from_a_foreign_origin_gets_no_cors_headers() {
    let gateway = TestGateway::start().await;
    let response = gateway
        .client
        .request(reqwest::Method::OPTIONS, gateway.url("/pub/echo"))
        .header("origin", "https://evil.example")
        .header("access-control-request-method", "GET")
        .send()
        .await
        .expect("response");
    assert!(response.headers().get("access-control-allow-origin").is_none());
}

#[tokio::test]
async fn simple_requests_get_cors_headers_only_for_allowed_origins() {
    let gateway = TestGateway::start().await;
    let allowed = gateway.client.get(gateway.url("/pub/echo")).header("origin", "http://127.0.0.1:3000").send().await.expect("response");
    assert_eq!(allowed.headers()["access-control-allow-origin"], "http://127.0.0.1:3000");
    assert_eq!(allowed.headers()["access-control-allow-credentials"], "true");

    let foreign = gateway.client.get(gateway.url("/pub/echo")).header("origin", "https://evil.example").send().await.expect("response");
    assert!(foreign.headers().get("access-control-allow-origin").is_none());
}

#[tokio::test]
async fn real_config_serves_local_routes_and_protects_private_ones() {
    super::support::init_env();
    let cfg = crate::configs::routes::load("configs/routers/routes.yaml").expect("real config");
    let gateway = TestGateway::from_routes(&cfg.routes).await;

    assert_eq!(gateway.client.get(gateway.url("/")).send().await.expect("response").status(), 200);
    assert_eq!(gateway.client.get(gateway.url("/docs")).send().await.expect("response").status(), 200);

    for route in cfg.routes.iter().filter(|r| r.auth == crate::configs::routes::RouteAuth::Private) {
        let path = route.path.replace(":id", "x").replace(":key", "x").replace("/main/api/v1", "");
        let method = if route.method == "ANY" { "GET" } else { route.method.as_str() };
        let response = gateway.client.request(method.parse().expect("method"), gateway.url(&path)).send().await.expect("response");
        assert_eq!(response.status(), 401, "{method} {}", route.path);
    }
}
