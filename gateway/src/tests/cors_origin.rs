use crate::configs::cors::is_allowed_origin;

#[test]
fn any_localhost_port_is_allowed() {
    for origin in ["http://localhost", "http://localhost:3000", "https://localhost:9999", "http://127.0.0.1:5173", "http://0.0.0.0:1"] {
        assert!(is_allowed_origin(origin), "{origin}");
    }
}

#[test]
fn listed_origins_are_allowed() {
    assert!(is_allowed_origin("http://localhost:5786"));
    assert!(is_allowed_origin("https://localhost:4173"));
}

#[test]
fn project_domains_and_subdomains_are_allowed() {
    for origin in [
        "https://analogwakatime.com",
        "https://app.analogwakatime.com",
        "https://a.b.analogwakatime.com",
        "https://testingmyproject.space",
        "http://192.168.0.15",
        "http://192.168.0.15:8080",
    ] {
        assert!(is_allowed_origin(origin), "{origin}");
    }
}

#[test]
fn lookalike_domains_are_rejected() {
    for origin in [
        "https://evilanalogwakatime.com",
        "https://analogwakatime.com.evil.com",
        "https://analogwakatime.co",
        "https://notlocalhost.example",
        "http://192.168.0.150",
        "https://evil.com",
    ] {
        assert!(!is_allowed_origin(origin), "{origin}");
    }
}

#[test]
fn origin_is_normalized_before_the_check() {
    assert!(is_allowed_origin("HTTP://LOCALHOST:3000"));
    assert!(is_allowed_origin("http://localhost:80"));
    assert!(is_allowed_origin("https://App.AnalogWakaTime.com/"));
    assert!(is_allowed_origin("  https://analogwakatime.com  "));
}

#[test]
fn garbage_is_rejected() {
    for origin in ["", "null", "not an origin", "://", "javascript:alert(1)", "ftp://evil.com"] {
        assert!(!is_allowed_origin(origin), "{origin:?}");
    }
}
