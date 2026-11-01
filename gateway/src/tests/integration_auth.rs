use jsonwebtoken::{Algorithm, EncodingKey, Header, encode};
use serde_json::json;

use super::keys::{PRIVATE_A, PRIVATE_B, PUBLIC_A};
use super::support::{TestGateway, access_token, claims, json_of, now, sign};

const COOKIE: &str = "Cadence-Cloud-Auth-Token";

async fn get(gateway: &TestGateway, headers: &[(&str, String)]) -> reqwest::Response {
    let mut request = gateway.client.get(gateway.url("/priv/echo"));
    for (name, value) in headers {
        request = request.header(*name, value);
    }
    request.send().await.expect("response")
}

#[tokio::test]
async fn missing_token_is_401_and_counted() {
    let gateway = TestGateway::start().await;
    let response = get(&gateway, &[]).await;
    assert_eq!(response.status(), 401);
    assert_eq!(json_of(&response.text().await.expect("body"))["error"], "authorization required");
    assert_eq!(gateway.state.metrics.snapshot()["auth_failures"], 1);
}

#[tokio::test]
async fn valid_bearer_token_reaches_the_upstream_unchanged() {
    let gateway = TestGateway::start().await;
    let token = access_token("user-1", "sid-1");
    let response = get(&gateway, &[("authorization", format!("Bearer {token}"))]).await;
    assert_eq!(response.status(), 200);
    let body = json_of(&response.text().await.expect("body"));
    assert_eq!(body["headers"]["authorization"], format!("Bearer {token}"));
    assert_eq!(body["path"], "/echo");
}

#[tokio::test]
async fn bearer_scheme_is_case_insensitive() {
    let gateway = TestGateway::start().await;
    let token = access_token("user-1", "sid-1");
    for scheme in ["bearer", "BEARER", "Bearer"] {
        assert_eq!(get(&gateway, &[("authorization", format!("{scheme} {token}"))]).await.status(), 200, "{scheme}");
    }
}

#[tokio::test]
async fn cookie_token_is_converted_to_a_bearer_header_and_the_cookie_is_hidden() {
    let gateway = TestGateway::start().await;
    let token = access_token("user-1", "sid-1");
    let response = get(&gateway, &[("cookie", format!("theme=dark; {COOKIE}={token}; other=1"))]).await;
    assert_eq!(response.status(), 200);
    let body = json_of(&response.text().await.expect("body"));
    assert_eq!(body["headers"]["authorization"], format!("Bearer {token}"));
    assert!(body["headers"].get("cookie").is_none(), "cookie leaked to the upstream");
}

#[tokio::test]
async fn bearer_wins_over_the_cookie() {
    let gateway = TestGateway::start().await;
    let bearer = access_token("user-1", "sid-1");
    let cookie = access_token("user-2", "sid-2");
    let response = get(&gateway, &[("authorization", format!("Bearer {bearer}")), ("cookie", format!("{COOKIE}={cookie}"))]).await;
    let body = json_of(&response.text().await.expect("body"));
    assert_eq!(body["headers"]["authorization"], format!("Bearer {bearer}"));
}

#[tokio::test]
async fn cookie_auth_needs_the_csrf_header_on_unsafe_methods() {
    let gateway = TestGateway::start().await;
    let cookie = format!("{COOKIE}={}", access_token("user-1", "sid-1"));
    for method in ["POST", "PUT", "PATCH", "DELETE"] {
        let method: reqwest::Method = method.parse().expect("method");
        let blocked = gateway.client.request(method.clone(), gateway.url("/priv/echo")).header("cookie", &cookie).send().await.expect("response");
        assert_eq!(blocked.status(), 403, "{method}");
        assert_eq!(json_of(&blocked.text().await.expect("body"))["error"], "csrf header required");
        let allowed = gateway.client.request(method.clone(), gateway.url("/priv/echo")).header("cookie", &cookie).header("x-requested-with", "XMLHttpRequest").send().await.expect("response");
        assert_eq!(allowed.status(), 200, "{method}");
    }
    assert_eq!(gateway.state.metrics.snapshot()["auth_failures"], 4);
}

#[tokio::test]
async fn bearer_auth_does_not_need_the_csrf_header() {
    let gateway = TestGateway::start().await;
    let response = gateway
        .client
        .post(gateway.url("/priv/echo"))
        .header("authorization", format!("Bearer {}", access_token("user-1", "sid-1")))
        .send()
        .await
        .expect("response");
    assert_eq!(response.status(), 200);
}

#[tokio::test]
async fn safe_methods_with_a_cookie_do_not_need_the_csrf_header() {
    let gateway = TestGateway::start().await;
    let response = get(&gateway, &[("cookie", format!("{COOKIE}={}", access_token("user-1", "sid-1")))]).await;
    assert_eq!(response.status(), 200);
}

#[tokio::test]
async fn empty_or_foreign_cookies_do_not_authenticate() {
    let gateway = TestGateway::start().await;
    assert_eq!(get(&gateway, &[("cookie", format!("{COOKIE}="))]).await.status(), 401);
    assert_eq!(get(&gateway, &[("cookie", format!("Other-{COOKIE}={}", access_token("u", "s")))]).await.status(), 401);
}

#[tokio::test]
async fn broken_authorization_headers_are_401() {
    let gateway = TestGateway::start().await;
    let token = access_token("user-1", "sid-1");
    for header in ["Bearer", "Bearer ", &format!("Basic {token}"), &format!("Token {token}"), &token] {
        assert_eq!(get(&gateway, &[("authorization", header.to_owned())]).await.status(), 401, "{header:?}");
    }
}

#[tokio::test]
async fn bad_tokens_are_401() {
    let gateway = TestGateway::start().await;
    let mut expired = claims(Some("access"), "u", "s");
    expired["exp"] = json!(now() - 3600);
    let hmac = encode(&Header::new(Algorithm::HS256), &claims(Some("access"), "u", "s"), &EncodingKey::from_secret(PUBLIC_A.as_bytes())).expect("hs256");
    let tampered = {
        let token = access_token("u", "s");
        let mut parts: Vec<String> = token.split('.').map(str::to_owned).collect();
        parts[1] = parts[1].chars().rev().collect();
        parts.join(".")
    };
    let tokens = [
        "garbage".to_owned(),
        sign(&expired, PRIVATE_A),
        sign(&claims(Some("refresh"), "u", "s"), PRIVATE_A),
        sign(&claims(None, "u", "s"), PRIVATE_A),
        sign(&claims(Some("access"), "", "s"), PRIVATE_A),
        sign(&claims(Some("access"), "u", "s"), PRIVATE_B),
        hmac,
        tampered,
    ];
    for token in tokens {
        let response = get(&gateway, &[("authorization", format!("Bearer {token}"))]).await;
        assert_eq!(response.status(), 401, "{token}");
        assert_eq!(json_of(&response.text().await.expect("body"))["error"], "invalid or expired token");
    }
}

#[tokio::test]
async fn revoked_session_is_rejected_but_other_sessions_work() {
    let gateway = TestGateway::start().await;
    gateway.state.store.revoke_session("dead");
    let dead = get(&gateway, &[("authorization", format!("Bearer {}", access_token("u", "dead")))]).await;
    assert_eq!(dead.status(), 401);
    assert_eq!(json_of(&dead.text().await.expect("body"))["error"], "session revoked");
    assert_eq!(get(&gateway, &[("authorization", format!("Bearer {}", access_token("u", "alive")))]).await.status(), 200);
}

#[tokio::test]
async fn tokens_issued_before_the_rejection_epoch_are_refused() {
    let gateway = TestGateway::start().await;
    gateway.state.store.reject_issued_before("user-1", now() + 3600);
    assert_eq!(get(&gateway, &[("authorization", format!("Bearer {}", access_token("user-1", "s")))]).await.status(), 401);
    assert_eq!(get(&gateway, &[("authorization", format!("Bearer {}", access_token("user-2", "s")))]).await.status(), 200);
}

#[tokio::test]
async fn private_route_with_a_path_parameter() {
    let gateway = TestGateway::start().await;
    let response = gateway
        .client
        .get(gateway.url("/priv/item/42"))
        .header("authorization", format!("Bearer {}", access_token("u", "s")))
        .send()
        .await
        .expect("response");
    assert_eq!(json_of(&response.text().await.expect("body"))["path"], "/echo/item/42");
    assert_eq!(gateway.client.get(gateway.url("/priv/item/42")).send().await.expect("response").status(), 401);
}

#[tokio::test]
async fn public_routes_ignore_a_bad_token() {
    let gateway = TestGateway::start().await;
    let response = gateway.client.get(gateway.url("/pub/echo")).header("authorization", "Bearer garbage").send().await.expect("response");
    assert_eq!(response.status(), 200);
}
