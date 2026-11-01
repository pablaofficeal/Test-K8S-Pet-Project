use jsonwebtoken::{Algorithm, EncodingKey, Header, encode};
use serde_json::json;

use super::keys::{PRIVATE_A, PRIVATE_B, PUBLIC_A};
use super::support::{claims, init_env, now, sign};
use crate::utils::jwt::{algorithm_from_name, validate_token};

fn check(token: &str) -> Result<crate::utils::jwt::Claims, String> {
    init_env();
    validate_token(token, Algorithm::EdDSA)
}

#[test]
fn valid_access_token_is_accepted() {
    let parsed = check(&sign(&claims(Some("access"), "user-1", "sid-1"), PRIVATE_A)).expect("valid");
    assert_eq!(parsed.uuid, "user-1");
    assert_eq!(parsed.sid.as_deref(), Some("sid-1"));
    assert_eq!(parsed.email.as_deref(), Some("user@example.com"));
}

#[test]
fn refresh_token_is_not_an_access_token() {
    let error = check(&sign(&claims(Some("refresh"), "u", "s"), PRIVATE_A)).unwrap_err();
    assert!(error.contains("purpose"), "{error}");
}

#[test]
fn token_without_purpose_is_rejected() {
    assert!(check(&sign(&claims(None, "u", "s"), PRIVATE_A)).is_err());
}

#[test]
fn empty_uuid_is_rejected() {
    let error = check(&sign(&claims(Some("access"), "", "s"), PRIVATE_A)).unwrap_err();
    assert!(error.contains("uuid"), "{error}");
}

#[test]
fn expired_token_is_rejected() {
    let mut expired = claims(Some("access"), "u", "s");
    expired["exp"] = json!(now() - 3600);
    assert!(check(&sign(&expired, PRIVATE_A)).is_err());
}

#[test]
fn token_signed_with_another_key_is_rejected() {
    assert!(check(&sign(&claims(Some("access"), "u", "s"), PRIVATE_B)).is_err());
}

#[test]
fn tampered_payload_is_rejected() {
    let token = sign(&claims(Some("access"), "u", "s"), PRIVATE_A);
    let mut parts: Vec<String> = token.split('.').map(str::to_owned).collect();
    parts[1] = parts[1].chars().rev().collect();
    assert!(check(&parts.join(".")).is_err());
}

#[test]
fn hmac_token_signed_with_the_public_key_is_rejected() {
    let hmac = encode(
        &Header::new(Algorithm::HS256),
        &claims(Some("access"), "u", "s"),
        &EncodingKey::from_secret(PUBLIC_A.as_bytes()),
    )
    .expect("hs256");
    assert!(check(&hmac).is_err());
}

#[test]
fn unsigned_and_garbage_tokens_are_rejected() {
    for token in ["", "abc", "a.b.c", "eyJhbGciOiJub25lIn0.e30.", "Bearer x"] {
        assert!(check(token).is_err(), "{token:?}");
    }
}

#[test]
fn supported_algorithm_name() {
    assert_eq!(algorithm_from_name("EdDSA"), Algorithm::EdDSA);
}

#[test]
#[should_panic(expected = "unsupported JWT_ALGORITHM")]
fn unsupported_algorithm_name_panics() {
    algorithm_from_name("HS256");
}
