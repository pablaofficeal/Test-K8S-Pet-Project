use serde_json::Value;

use super::support::{TestGateway, json_of};

async fn echo(gateway: &TestGateway, request: reqwest::RequestBuilder) -> (u16, Value) {
    let _ = gateway;
    let response = request.send().await.expect("response");
    let status = response.status().as_u16();
    (status, json_of(&response.text().await.expect("body")))
}

#[tokio::test]
async fn get_is_proxied_with_a_rewritten_path_and_the_query() {
    let gateway = TestGateway::start().await;
    let (status, body) = echo(&gateway, gateway.client.get(gateway.url("/pub/echo?x=1&y=two"))).await;
    assert_eq!(status, 200);
    assert_eq!(body["method"], "GET");
    assert_eq!(body["path"], "/echo");
    assert_eq!(body["query"], "x=1&y=two");
}

#[tokio::test]
async fn special_characters_in_the_query_survive() {
    let gateway = TestGateway::start().await;
    let (_, body) = echo(&gateway, gateway.client.get(gateway.url("/pub/echo?q=a%20b%26c&k=%D0%BF"))).await;
    assert_eq!(body["query"], "q=a%20b%26c&k=%D0%BF");
}

#[tokio::test]
async fn request_headers_are_forwarded_and_gateway_headers_added() {
    let gateway = TestGateway::start().await;
    let request = gateway
        .client
        .get(gateway.url("/pub/echo"))
        .header("x-custom", "value")
        .header("x-request-id", "trace-123");
    let (_, body) = echo(&gateway, request).await;
    let headers = &body["headers"];
    assert_eq!(headers["x-custom"], "value");
    assert_eq!(headers["x-request-id"], "trace-123");
    assert_eq!(headers["x-gateway-forwarded"], "true");
    assert_eq!(headers["x-forwarded-host"], gateway.base.trim_start_matches("http://"));
    assert!(headers["x-origin-host"].as_str().expect("origin").starts_with("http://127.0.0.1:"));
}

#[tokio::test]
async fn host_header_is_replaced_for_the_upstream() {
    let gateway = TestGateway::start().await;
    let (_, body) = echo(&gateway, gateway.client.get(gateway.url("/pub/echo"))).await;
    assert_ne!(body["headers"]["host"], gateway.base.trim_start_matches("http://"));
}

#[tokio::test]
async fn trace_id_is_echoed_or_generated() {
    let gateway = TestGateway::start().await;
    let given = gateway.client.get(gateway.url("/pub/echo")).header("x-request-id", "mine-1").send().await.expect("response");
    assert_eq!(given.headers()["x-request-id"], "mine-1");

    let generated = gateway.client.get(gateway.url("/pub/echo")).send().await.expect("response");
    let id = generated.headers()["x-request-id"].to_str().expect("id").to_owned();
    assert_eq!(id.len(), 36);
    let (_, body) = echo(&gateway, gateway.client.get(gateway.url("/pub/echo"))).await;
    assert_eq!(body["headers"]["x-request-id"].as_str().expect("forwarded").len(), 36);

    let blank = gateway.client.get(gateway.url("/pub/echo")).header("x-request-id", "   ").send().await.expect("response");
    assert_eq!(blank.headers()["x-request-id"].to_str().expect("id").len(), 36);
}

#[tokio::test]
async fn every_method_and_its_body_reach_the_upstream() {
    let gateway = TestGateway::start().await;
    for method in ["POST", "PUT", "PATCH", "DELETE"] {
        let request = gateway
            .client
            .request(method.parse().expect("method"), gateway.url("/pub/echo"))
            .header("content-type", "application/json")
            .body(r#"{"hello":"мир"}"#);
        let (status, body) = echo(&gateway, request).await;
        assert_eq!(status, 200, "{method}");
        assert_eq!(body["method"], method);
        assert_eq!(body["body"], r#"{"hello":"мир"}"#);
        assert_eq!(body["headers"]["content-type"], "application/json");
    }
}

#[tokio::test]
async fn large_body_goes_through_intact() {
    let gateway = TestGateway::start().await;
    let payload = "x".repeat(1024 * 1024);
    let (status, body) = echo(&gateway, gateway.client.post(gateway.url("/pub/echo")).body(payload.clone())).await;
    assert_eq!(status, 200);
    assert_eq!(body["body"].as_str().expect("body").len(), payload.len());
}

#[tokio::test]
async fn path_parameters_are_substituted_and_statuses_pass_through() {
    let gateway = TestGateway::start().await;
    for code in [200u16, 201, 204, 400, 404, 418, 500, 503] {
        let response = gateway.client.get(gateway.url(&format!("/pub/status/{code}"))).send().await.expect("response");
        assert_eq!(response.status().as_u16(), code);
    }
}

#[tokio::test]
async fn upstream_response_headers_and_repeated_cookies_survive() {
    let gateway = TestGateway::start().await;
    let response = gateway.client.get(gateway.url("/pub/cookies")).send().await.expect("response");
    assert_eq!(response.headers()["x-upstream"], "yes");
    let cookies: Vec<&str> = response.headers().get_all("set-cookie").iter().map(|v| v.to_str().expect("cookie")).collect();
    assert_eq!(cookies, ["a=1; Path=/", "b=2; Path=/x"]);
}

#[tokio::test]
async fn unreachable_upstream_gives_502_and_is_counted() {
    let gateway = TestGateway::start().await;
    let response = gateway.client.get(gateway.url("/pub/down")).send().await.expect("response");
    assert_eq!(response.status(), 502);
    assert_eq!(json_of(&response.text().await.expect("body"))["error"], "upstream unreachable");
    assert_eq!(gateway.state.metrics.snapshot()["upstream_errors"], 1);
}

#[tokio::test]
async fn slow_upstream_gives_504() {
    let gateway = TestGateway::start().await;
    let started = std::time::Instant::now();
    let response = gateway.client.get(gateway.url("/pub/slow")).send().await.expect("response");
    assert_eq!(response.status(), 504);
    assert_eq!(json_of(&response.text().await.expect("body"))["error"], "upstream timeout");
    assert!(started.elapsed() < std::time::Duration::from_secs(3));
    assert_eq!(gateway.state.metrics.snapshot()["upstream_errors"], 1);
}

#[tokio::test]
async fn websocket_upgrade_is_refused_with_501() {
    let gateway = TestGateway::start().await;
    for path in ["/pub/echo", "/pub/websocket"] {
        let response = gateway
            .client
            .get(gateway.url(path))
            .header("connection", "Upgrade")
            .header("upgrade", "websocket")
            .send()
            .await
            .expect("response");
        assert_eq!(response.status(), 501, "{path}");
    }
}

#[tokio::test]
async fn unknown_route_gives_a_json_404() {
    let gateway = TestGateway::start().await;
    let response = gateway.client.get(gateway.url("/no/such/route")).send().await.expect("response");
    assert_eq!(response.status(), 404);
    assert_eq!(json_of(&response.text().await.expect("body"))["error"], "route not found");
}

#[tokio::test]
async fn method_not_declared_for_a_route_is_405() {
    let gateway = TestGateway::start().await;
    let response = gateway.client.post(gateway.url("/pub/cookies")).send().await.expect("response");
    assert_eq!(response.status(), 405);
}

#[tokio::test]
async fn proxied_and_total_requests_are_counted() {
    let gateway = TestGateway::start().await;
    gateway.client.get(gateway.url("/pub/echo")).send().await.expect("response");
    gateway.client.get(gateway.url("/pub/echo")).send().await.expect("response");
    gateway.client.get(gateway.url("/nope")).send().await.expect("response");
    let snapshot = gateway.state.metrics.snapshot();
    assert_eq!(snapshot["proxied_requests"], 2);
    assert_eq!(snapshot["total_requests"], 3);
}
