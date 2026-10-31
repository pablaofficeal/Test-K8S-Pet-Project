use std::collections::{HashMap, HashSet};
use std::sync::{Arc, RwLock};

use crate::utils::jwt::Claims;

#[derive(Debug, Clone, Default)]
pub struct RevocationStore {
    revoked_sessions: Arc<RwLock<HashSet<String>>>,
    rejected_before: Arc<RwLock<HashMap<String, i64>>>,
}

impl RevocationStore {
    #[allow(dead_code)]
    pub fn revoke_session(&self, sid: &str) {
        self.revoked_sessions
            .write()
            .expect("revocation lock poisoned")
            .insert(sid.to_owned());
    }

    pub fn is_session_revoked(&self, sid: &str) -> bool {
        self.revoked_sessions
            .read()
            .expect("revocation lock poisoned")
            .contains(sid)
    }

    #[allow(dead_code)]
    pub fn reject_issued_before(&self, uuid: &str, epoch_secs: i64) {
        self.rejected_before
            .write()
            .expect("revocation lock poisoned")
            .insert(uuid.to_owned(), epoch_secs);
    }

    pub fn is_epoch_rejected(&self, uuid: &str, issued_at: i64) -> bool {
        self.rejected_before
            .read()
            .expect("revocation lock poisoned")
            .get(uuid)
            .is_some_and(|epoch| issued_at < *epoch)
    }
}

pub fn session_allowed(store: &RevocationStore, claims: &Claims) -> bool {
    if let Some(sid) = &claims.sid
        && store.is_session_revoked(sid)
    {
        return false;
    }
    if let Some(iat) = claims.iat
        && store.is_epoch_rejected(&claims.uuid, iat)
    {
        return false;
    }
    true
}