import hashlib
import uuid
import jwt
from datetime import datetime, timedelta, timezone
from configs.app_config import Config


ACCESS_PURPOSE = "access"
REFRESH_PURPOSE = "refresh"


def _parse_ttl(value: str) -> timedelta:
    if not value:
        raise ValueError("empty TTL")
    value = value.strip()
    unit = value[-1].lower()
    amount = int(value[:-1])
    if unit == "s":
        return timedelta(seconds=amount)
    if unit == "m":
        return timedelta(minutes=amount)
    if unit == "h":
        return timedelta(hours=amount)
    return timedelta(days=amount)


def _access_ttl() -> timedelta:
    return _parse_ttl(Config().JWT_ACCESS_TTL)


def _refresh_ttl() -> timedelta:
    return _parse_ttl(Config().JWT_REFRESH_TTL)


def _base_claims(user_uuid: str, email: str, sid: str, jti: str,
                 purpose: str, now: datetime, expires_at: datetime) -> dict:
    return {
        "uuid": str(user_uuid),
        "email": email,
        "sid": sid,
        "ver": 0,
        "jti": jti,
        "purpose": purpose,
        "iat": now,
        "exp": expires_at,
    }


def _pem(value: str) -> str:
    return value.replace("\\n", "\n")


def _encode(payload: dict) -> str:
    return jwt.encode(payload, _pem(Config().JWT_SECRET_KEY), algorithm=Config().JWT_ALGORITHM)


def generate_access_token(user_uuid: str, email: str, sid: str):
    now = datetime.now(timezone.utc)
    expires_at = now + _access_ttl()
    payload = _base_claims(user_uuid, email, sid, sid, ACCESS_PURPOSE, now, expires_at)
    return _encode(payload), expires_at


def generate_refresh_token(user_uuid: str, email: str, sid: str):
    now = datetime.now(timezone.utc)
    expires_at = now + _refresh_ttl()
    payload = _base_claims(user_uuid, email, sid, uuid.uuid4().hex, REFRESH_PURPOSE, now, expires_at)
    return _encode(payload), expires_at


def generate_token_pair(user_uuid: str, email: str, session_id: str | None = None):
    sid = session_id or uuid.uuid4().hex
    access_token, access_exp = generate_access_token(user_uuid, email, sid)
    refresh_token, refresh_exp = generate_refresh_token(user_uuid, email, sid)
    return access_token, access_exp, refresh_token, refresh_exp, sid


def hash_refresh_token(token: str) -> str:
    return hashlib.sha256(token.encode()).hexdigest()


def generate_token(user_uuid: str, email: str, session_id: str | None = None):
    sid = session_id or uuid.uuid4().hex
    return generate_access_token(str(user_uuid), email, sid)


def validate_token(token: str, expected_purpose: str = ACCESS_PURPOSE) -> dict:
    public_key = _pem(Config().JWT_PUBLIC_KEY)
    algorithm = Config().JWT_ALGORITHM
    claims = jwt.decode(token, public_key, algorithms=[algorithm])
    if not claims.get("uuid"):
        raise ValueError("invalid token: empty uuid")
    purpose = claims.get("purpose")
    if purpose != expected_purpose:
        raise ValueError("invalid token purpose")
    return claims