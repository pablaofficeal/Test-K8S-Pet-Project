use crate::db::{RevocationStore, session_allowed};
use crate::utils::jwt::Claims;

fn claims(uuid: &str, sid: Option<&str>, iat: Option<i64>) -> Claims {
    Claims {
        uuid: uuid.to_owned(),
        email: None,
        sid: sid.map(str::to_owned),
        ver: None,
        jti: None,
        iat,
        exp: 0,
        purpose: Some("access".to_owned()),
    }
}

#[test]
fn fresh_store_allows_everything() {
    let store = RevocationStore::default();
    assert!(session_allowed(&store, &claims("u", Some("s"), Some(100))));
    assert!(session_allowed(&store, &claims("u", None, None)));
}

#[test]
fn revoked_session_is_rejected_others_are_not() {
    let store = RevocationStore::default();
    store.revoke_session("bad");
    assert!(!session_allowed(&store, &claims("u", Some("bad"), None)));
    assert!(session_allowed(&store, &claims("u", Some("good"), None)));
    assert!(store.is_session_revoked("bad"));
    assert!(!store.is_session_revoked("good"));
}

#[test]
fn tokens_issued_before_the_epoch_are_rejected() {
    let store = RevocationStore::default();
    store.reject_issued_before("u", 1_000);
    assert!(!session_allowed(&store, &claims("u", None, Some(999))));
    assert!(session_allowed(&store, &claims("u", None, Some(1_000))));
    assert!(session_allowed(&store, &claims("u", None, Some(1_001))));
}

#[test]
fn epoch_applies_only_to_its_user() {
    let store = RevocationStore::default();
    store.reject_issued_before("u", 1_000);
    assert!(session_allowed(&store, &claims("other", None, Some(1))));
}

#[test]
fn token_without_iat_is_not_affected_by_the_epoch() {
    let store = RevocationStore::default();
    store.reject_issued_before("u", i64::MAX);
    assert!(session_allowed(&store, &claims("u", None, None)));
}

#[test]
fn clones_share_the_same_state() {
    let store = RevocationStore::default();
    let clone = store.clone();
    store.revoke_session("s");
    assert!(clone.is_session_revoked("s"));
}

#[test]
fn later_epoch_replaces_the_earlier_one() {
    let store = RevocationStore::default();
    store.reject_issued_before("u", 100);
    store.reject_issued_before("u", 500);
    assert!(store.is_epoch_rejected("u", 400));
    assert!(!store.is_epoch_rejected("u", 500));
}
