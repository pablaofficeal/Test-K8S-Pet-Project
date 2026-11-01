use super::support::{TestGateway, json_of};

async fn hit(gateway: &TestGateway, path: &str, ip: &str) -> reqwest::Response {
    gateway.client.get(gateway.url(path)).header("x-forwarded-for", ip).send().await.expect("response")
}

#[tokio::test]
async fn third_request_inside_the_window_is_429() {
    let gateway = TestGateway::start().await;
    assert_eq!(hit(&gateway, "/pub/limited", "10.0.0.1").await.status(), 200);
    assert_eq!(hit(&gateway, "/pub/limited", "10.0.0.1").await.status(), 200);
    let blocked = hit(&gateway, "/pub/limited", "10.0.0.1").await;
    assert_eq!(blocked.status(), 429);
    assert_eq!(json_of(&blocked.text().await.expect("body"))["error"], "TOO MANY REQUESTS");
    assert_eq!(gateway.state.metrics.snapshot()["rate_limited"], 1);
}

#[tokio::test]
async fn blocked_requests_never_reach_the_upstream() {
    let gateway = TestGateway::start().await;
    for _ in 0..5 {
        hit(&gateway, "/pub/limited", "10.0.0.2").await;
    }
    assert_eq!(gateway.state.metrics.snapshot()["proxied_requests"], 2);
    assert_eq!(gateway.state.metrics.snapshot()["rate_limited"], 3);
}

#[tokio::test]
async fn every_client_has_its_own_budget() {
    let gateway = TestGateway::start().await;
    for _ in 0..3 {
        hit(&gateway, "/pub/limited", "10.0.0.3").await;
    }
    assert_eq!(hit(&gateway, "/pub/limited", "10.0.0.3").await.status(), 429);
    assert_eq!(hit(&gateway, "/pub/limited", "10.0.0.4").await.status(), 200);
}

#[tokio::test]
async fn every_route_has_its_own_budget() {
    let gateway = TestGateway::start().await;
    assert_eq!(hit(&gateway, "/pub/limited-other", "10.0.0.5").await.status(), 200);
    assert_eq!(hit(&gateway, "/pub/limited-other", "10.0.0.5").await.status(), 429);
    assert_eq!(hit(&gateway, "/pub/limited", "10.0.0.5").await.status(), 200);
}

#[tokio::test]
async fn routes_without_a_limit_are_never_blocked() {
    let gateway = TestGateway::start().await;
    for _ in 0..30 {
        assert_eq!(hit(&gateway, "/pub/echo", "10.0.0.6").await.status(), 200);
    }
}

#[tokio::test]
async fn real_ip_header_is_used_when_there_is_no_forwarded_for() {
    let gateway = TestGateway::start().await;
    for _ in 0..3 {
        gateway.client.get(gateway.url("/pub/limited")).header("x-real-ip", "10.0.0.7").send().await.expect("response");
    }
    let blocked = gateway.client.get(gateway.url("/pub/limited")).header("x-real-ip", "10.0.0.7").send().await.expect("response");
    assert_eq!(blocked.status(), 429);
}
