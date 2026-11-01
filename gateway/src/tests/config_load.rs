use std::path::PathBuf;

use crate::configs::routes::{GatewayConfig, RouteAuth, load};

fn write(files: &[(&str, &str)]) -> PathBuf {
    let dir = std::env::temp_dir().join(format!("gateway-config-{}", uuid::Uuid::new_v4()));
    std::fs::create_dir_all(&dir).expect("dir");
    for (name, content) in files {
        std::fs::write(dir.join(name), content).expect("write");
    }
    dir.join("routes.yaml")
}

fn parse(routes: &str) -> Result<GatewayConfig, String> {
    load(write(&[("routes.yaml", routes)]).to_str().expect("path"))
}

const BASE: &str = r#"
microservices_timeouts:
  auth: 40s
microservices_hosts:
  auth:
    host: http://auth-service:6637
    timeout: ${microservices_timeouts.auth}
  bare:
    host: bare-host:9000/
"#;

fn with(routes: &str) -> String {
    format!("{BASE}{routes}")
}

#[test]
fn listen_address_has_defaults() {
    let cfg = parse("public_routes: []\n").expect("config");
    assert_eq!(cfg.listen_addr, "0.0.0.0:8080");
    assert!(cfg.routes.is_empty());
}

#[test]
fn partial_server_section_keeps_the_other_defaults() {
    assert_eq!(parse("server:\n  port: 9100\n").expect("config").listen_addr, "0.0.0.0:9100");
    assert_eq!(parse("server:\n  gateway_host: 10.0.0.1\n").expect("config").listen_addr, "10.0.0.1:8080");
}

#[test]
fn listen_address_is_configurable() {
    let cfg = parse("server:\n  port: 9000\n  gateway_host: 127.0.0.1\n").expect("config");
    assert_eq!(cfg.listen_addr, "127.0.0.1:9000");
}

#[test]
fn route_is_resolved_against_hosts_and_timeouts() {
    let cfg = parse(&with(
        "public_routes:\n  - path: /main/api/v1/login\n    method: post\n    target: ${microservices_hosts.auth}\n    timeout: ${microservices_hosts.auth.timeout}\n    prefix: /login\n",
    ))
    .expect("config");
    let route = &cfg.routes[0];
    assert_eq!(route.method, "POST");
    assert_eq!(route.upstream, "http://auth-service:6637");
    assert_eq!(route.timeout_secs, 40);
    assert_eq!(route.prefix, "/login");
    assert_eq!(route.auth, RouteAuth::Public);
    assert!(route.handler.is_none() && !route.websocket && route.rate.is_none());
}

#[test]
fn private_routes_require_auth() {
    let cfg = parse(&with(
        "private_routes:\n  - path: /main/api/v1/me\n    target: ${microservices_hosts.auth}\n",
    ))
    .expect("config");
    assert_eq!(cfg.routes[0].auth, RouteAuth::Private);
}

#[test]
fn defaults_for_method_and_prefix() {
    let cfg = parse(&with("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.auth}\n")).expect("config");
    assert_eq!(cfg.routes[0].method, "ANY");
    assert_eq!(cfg.routes[0].prefix, "/main/api/v1/x");
    assert_eq!(cfg.routes[0].timeout_secs, 0);
}

#[test]
fn host_without_scheme_gets_http_and_loses_the_trailing_slash() {
    let cfg = parse(&with("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.bare}\n")).expect("config");
    assert_eq!(cfg.routes[0].upstream, "http://bare-host:9000");
}

#[test]
fn durations_understand_suffixes() {
    for (raw, secs) in [("20s", 20), ("2m", 120), ("1h", 3600), ("45", 45)] {
        let cfg = parse(&with(&format!("public_routes:\n  - path: /main/api/v1/x\n    target: ${{microservices_hosts.auth}}\n    timeout: {raw}\n"))).expect(raw);
        assert_eq!(cfg.routes[0].timeout_secs, secs, "{raw}");
    }
}

#[test]
fn invalid_duration_is_an_error() {
    let error = parse(&with("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.auth}\n    timeout: soon\n")).unwrap_err();
    assert!(error.contains("invalid duration"), "{error}");
}

#[test]
fn path_must_live_under_the_gateway_prefix() {
    let error = parse(&with("public_routes:\n  - path: /other/x\n    target: ${microservices_hosts.auth}\n")).unwrap_err();
    assert!(error.contains("must start with"), "{error}");
}

#[test]
fn local_target_needs_a_handler_and_vice_versa() {
    let error = parse("public_routes:\n  - path: /main/api/v1/x\n    target: local\n").unwrap_err();
    assert!(error.contains("requires a handler"), "{error}");
    let error = parse(&with("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.auth}\n    handler: hello\n")).unwrap_err();
    assert!(error.contains("requires target local"), "{error}");
}

#[test]
fn unknown_handler_is_rejected() {
    assert!(parse("public_routes:\n  - path: /main/api/v1/x\n    target: local\n    handler: shell\n").is_err());
}

#[test]
fn unresolved_placeholder_is_an_error() {
    let error = parse(&with("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.nope}\n")).unwrap_err();
    assert!(error.contains("unresolved placeholder"), "{error}");
}

#[test]
fn duplicate_method_and_path_is_an_error() {
    let error = parse(&with(
        "public_routes:\n  - path: /main/api/v1/x\n    method: GET\n    target: ${microservices_hosts.auth}\n  - path: /main/api/v1/x\n    method: GET\n    target: ${microservices_hosts.auth}\n",
    ))
    .unwrap_err();
    assert!(error.contains("duplicate route"), "{error}");
}

#[test]
fn same_path_with_different_methods_is_fine() {
    let cfg = parse(&with(
        "public_routes:\n  - path: /main/api/v1/x\n    method: GET\n    target: ${microservices_hosts.auth}\n  - path: /main/api/v1/x\n    method: POST\n    target: ${microservices_hosts.auth}\n",
    ))
    .expect("config");
    assert_eq!(cfg.routes.len(), 2);
}

#[test]
fn public_and_private_routes_cannot_collide() {
    assert!(parse(&with(
        "public_routes:\n  - path: /main/api/v1/x\n    method: GET\n    target: ${microservices_hosts.auth}\nprivate_routes:\n  - path: /main/api/v1/x\n    method: GET\n    target: ${microservices_hosts.auth}\n",
    ))
    .is_err());
}

#[test]
fn inline_rate_limit_prefers_the_ip_rule() {
    let cfg = parse(&with(
        "public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.auth}\n    rate_limit:\n      - by: user\n        limit: 1\n        window: 1m\n      - by: ip\n        limit: 7\n        window: 2m\n",
    ))
    .expect("config");
    let rate = cfg.routes[0].rate.as_ref().expect("rate");
    assert_eq!((rate.limit, rate.window_secs), (7, 120));
}

#[test]
fn rate_limit_rules_are_validated() {
    let error = parse(&with("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.auth}\n    rate_limit:\n      - by: ip\n        limit: 0\n        window: 1m\n")).unwrap_err();
    assert!(error.contains("limit must be > 0"), "{error}");
    let error = parse(&with("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.auth}\n    rate_limit:\n      - limit: 5\n        window: 1m\n")).unwrap_err();
    assert!(error.contains("by is required"), "{error}");
}

const LIMITS: &str = r#"
policies:
  strict:
    - by: ip
      limit: 10
      window: 1m
microservices:
  auth:
    - by: ip
      limit: 30
      window: 1m
routes:
  - path: /main/api/v1/special
    method: POST
    rate_limit: ${rate_limits.policies.strict}
"#;

fn with_limits(routes: &str) -> Result<GatewayConfig, String> {
    let text = format!("{BASE}rate_limits: ./limits.yaml\n{routes}");
    load(write(&[("routes.yaml", &text), ("limits.yaml", LIMITS)]).to_str().expect("path"))
}

#[test]
fn microservice_limit_applies_to_its_routes() {
    let cfg = with_limits("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.auth}\n").expect("config");
    let rate = cfg.routes[0].rate.as_ref().expect("rate");
    assert_eq!((rate.limit, rate.window_secs), (30, 60));
}

#[test]
fn services_without_a_limit_stay_unlimited() {
    let cfg = with_limits("public_routes:\n  - path: /main/api/v1/x\n    target: ${microservices_hosts.bare}\n").expect("config");
    assert!(cfg.routes[0].rate.is_none());
}

#[test]
fn per_route_limit_in_the_limits_file_wins() {
    let cfg = with_limits("public_routes:\n  - path: /main/api/v1/special\n    method: POST\n    target: ${microservices_hosts.auth}\n").expect("config");
    assert_eq!(cfg.routes[0].rate.as_ref().expect("rate").limit, 10);
}

#[test]
fn per_route_limit_matches_only_its_method() {
    let cfg = with_limits("public_routes:\n  - path: /main/api/v1/special\n    method: GET\n    target: ${microservices_hosts.auth}\n").expect("config");
    assert_eq!(cfg.routes[0].rate.as_ref().expect("rate").limit, 30);
}

#[test]
fn route_can_reference_a_policy_or_switch_the_limit_off() {
    let cfg = with_limits("public_routes:\n  - path: /main/api/v1/a\n    target: ${microservices_hosts.auth}\n    rate_limit: ${rate_limits.policies.strict}\n  - path: /main/api/v1/b\n    target: ${microservices_hosts.auth}\n    rate_limit: off\n").expect("config");
    assert_eq!(cfg.routes[0].rate.as_ref().expect("rate").limit, 10);
    assert!(cfg.routes[1].rate.is_none());
}

#[test]
fn unknown_policy_is_an_error() {
    let error = with_limits("public_routes:\n  - path: /main/api/v1/a\n    target: ${microservices_hosts.auth}\n    rate_limit: ${rate_limits.policies.nope}\n").unwrap_err();
    assert!(error.contains("unknown placeholder"), "{error}");
}

#[test]
fn missing_files_and_broken_yaml_are_errors() {
    assert!(load("/definitely/not/here.yaml").is_err());
    assert!(parse("public_routes: [unclosed").is_err());
    assert!(load(write(&[("routes.yaml", "rate_limits: ./gone.yaml\n")]).to_str().expect("path")).is_err());
}

#[test]
fn the_real_config_loads() {
    let cfg = load("configs/routers/routes.yaml").expect("configs/routers/routes.yaml");
    assert!(!cfg.routes.is_empty());
    assert!(cfg.routes.iter().all(|r| r.path.starts_with("/main/api/v1/")));
}

#[test]
fn the_real_config_protects_user_data() {
    let cfg = load("configs/routers/routes.yaml").expect("config");
    let public: Vec<&str> = cfg.routes.iter().filter(|r| r.auth == RouteAuth::Public).map(|r| r.path.as_str()).collect();
    for path in &public {
        assert!(
            !path.contains("/my/") || path.contains("/start/"),
            "{path} looks like user data but is public"
        );
    }
    for must_be_private in [
        "/main/api/v1/get/my/profile/info",
        "/main/api/v1/todo/get/my/boards",
        "/main/api/v1/todo/move/my/task/:id",
        "/main/api/v1/start/upload/images",
        "/main/api/v1/verify/confirm",
    ] {
        let route = cfg.routes.iter().find(|r| r.path == must_be_private).unwrap_or_else(|| panic!("{must_be_private} is not routed"));
        assert_eq!(route.auth, RouteAuth::Private, "{must_be_private}");
    }
}

#[test]
fn the_real_config_routes_auth_endpoints_publicly() {
    let cfg = load("configs/routers/routes.yaml").expect("config");
    for path in ["/main/api/v1/start/log/in/my/account", "/main/api/v1/start/create/my/account", "/main/api/v1/start/refresh/my/token", "/main/api/v1/start/log/out/my/account"] {
        let route = cfg.routes.iter().find(|r| r.path == path).unwrap_or_else(|| panic!("{path} is not routed"));
        assert_eq!(route.auth, RouteAuth::Public, "{path}");
        assert_eq!(route.method, "POST", "{path}");
    }
}
