use jsonwebtoken::{Algorithm, DecodingKey, Validation, decode};
use serde::{Deserialize, Serialize};

use crate::configs::jwt_public_key;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub uuid: String,
    pub email: Option<String>,
    pub sid: Option<String>,
    pub ver: Option<i64>,
    pub jti: Option<String>,
    pub iat: Option<i64>,
    pub exp: usize,
    pub purpose: Option<String>,
}

pub fn algorithm_from_name(name: &str) -> Algorithm {
    match name {
        "EdDSA" => Algorithm::EdDSA,
        other => panic!("unsupported JWT_ALGORITHM: {other}"),
    }
}

pub fn validate_token(token: &str, algorithm: Algorithm) -> Result<Claims, String> {
    let public_key = jwt_public_key()?;
    let key = DecodingKey::from_ed_pem(public_key.as_bytes()).map_err(|e| e.to_string())?;
    let data = decode::<Claims>(token, &key, &Validation::new(algorithm)).map_err(|e| e.to_string())?;
    if data.claims.uuid.is_empty() {
        return Err("invalid token: empty uuid".to_owned());
    }
    if data.claims.purpose.as_deref() != Some("access") {
        return Err("invalid token purpose".to_owned());
    }
    Ok(data.claims)
}