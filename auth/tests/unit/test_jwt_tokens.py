import time
from datetime import datetime, timedelta, timezone

import jwt
import pytest

from tests.conftest import make_key_pair
from utils.jwt.jwt import (
    ACCESS_PURPOSE,
    REFRESH_PURPOSE,
    _parse_ttl,
    generate_access_token,
    generate_refresh_token,
    generate_token,
    generate_token_pair,
    hash_refresh_token,
    validate_token,
)


@pytest.mark.parametrize(
    "value, expected",
    [
        ("30s", timedelta(seconds=30)),
        ("15m", timedelta(minutes=15)),
        ("2h", timedelta(hours=2)),
        ("7d", timedelta(days=7)),
        ("60d", timedelta(days=60)),
        ("10M", timedelta(minutes=10)),
        (" 5s ", timedelta(seconds=5)),
    ],
)
def test_parse_ttl(value, expected):
    assert _parse_ttl(value) == expected


@pytest.mark.parametrize("value", ["", None, "abc", "m", "1.5h"])
def test_parse_ttl_rejects_garbage(value):
    with pytest.raises((ValueError, TypeError, AttributeError)):
        _parse_ttl(value)


def test_access_token_claims(keys):
    token, expires_at = generate_access_token("user-1", "a@b.com", "sid-1")
    claims = validate_token(token)
    assert claims["uuid"] == "user-1"
    assert claims["email"] == "a@b.com"
    assert claims["sid"] == "sid-1"
    assert claims["purpose"] == ACCESS_PURPOSE
    assert claims["ver"] == 0
    assert claims["jti"] == "sid-1"
    assert claims["exp"] == int(expires_at.timestamp())
    assert 14 * 60 < claims["exp"] - claims["iat"] <= 15 * 60


def test_refresh_token_claims_and_unique_jti(keys):
    first, first_exp = generate_refresh_token("user-1", "a@b.com", "sid-1")
    second, _ = generate_refresh_token("user-1", "a@b.com", "sid-1")
    claims = validate_token(first, expected_purpose=REFRESH_PURPOSE)
    assert claims["purpose"] == REFRESH_PURPOSE
    assert claims["sid"] == "sid-1"
    assert (first_exp - datetime.now(timezone.utc)) > timedelta(days=6)
    assert first != second
    assert claims["jti"] != validate_token(second, expected_purpose=REFRESH_PURPOSE)["jti"]


def test_token_pair_shares_session_id(keys):
    access, access_exp, refresh, refresh_exp, sid = generate_token_pair("u", "e@x.com")
    assert validate_token(access)["sid"] == sid
    assert validate_token(refresh, expected_purpose=REFRESH_PURPOSE)["sid"] == sid
    assert refresh_exp > access_exp


def test_token_pair_reuses_given_session_id(keys):
    *_, sid = generate_token_pair("u", "e@x.com", session_id="fixed-sid")
    assert sid == "fixed-sid"


def test_generate_token_wraps_access_token(keys):
    token, _ = generate_token("u", "e@x.com", "sid")
    assert validate_token(token)["uuid"] == "u"


def test_purpose_is_enforced(keys):
    access, _ = generate_access_token("u", "e@x.com", "s")
    refresh, _ = generate_refresh_token("u", "e@x.com", "s")
    with pytest.raises(ValueError):
        validate_token(access, expected_purpose=REFRESH_PURPOSE)
    with pytest.raises(ValueError):
        validate_token(refresh)


def test_expired_token_is_rejected(keys, monkeypatch):
    from configs.app_config import Config

    monkeypatch.setattr(Config, "JWT_ACCESS_TTL", "1s")
    token, _ = generate_access_token("u", "e@x.com", "s")
    time.sleep(1.2)
    with pytest.raises(jwt.ExpiredSignatureError):
        validate_token(token)


def test_tampered_payload_is_rejected(keys):
    token, _ = generate_access_token("u", "e@x.com", "s")
    header, payload, signature = token.split(".")
    forged = ".".join([header, payload[:-2] + ("AA" if payload[-2:] != "AA" else "BB"), signature])
    with pytest.raises(jwt.PyJWTError):
        validate_token(forged)


def test_token_signed_with_another_key_is_rejected(keys):
    foreign_private, _ = make_key_pair()
    payload = {"uuid": "u", "purpose": "access", "exp": datetime.now(timezone.utc) + timedelta(minutes=5)}
    forged = jwt.encode(payload, foreign_private, algorithm="EdDSA")
    with pytest.raises(jwt.InvalidSignatureError):
        validate_token(forged)


def test_algorithm_confusion_is_rejected(keys):
    # атакующий подписывает HS256 публичным ключом как секретом; PyJWT сам такой токен не соберёт
    import base64
    import hashlib
    import hmac
    import json

    _, public_pem = keys
    encode = lambda data: base64.urlsafe_b64encode(data).rstrip(b"=").decode()
    header = encode(json.dumps({"alg": "HS256", "typ": "JWT"}).encode())
    payload = encode(json.dumps({"uuid": "u", "purpose": "access", "exp": int(time.time()) + 300}).encode())
    signature = encode(hmac.new(public_pem.encode(), f"{header}.{payload}".encode(), hashlib.sha256).digest())
    with pytest.raises(jwt.PyJWTError):
        validate_token(f"{header}.{payload}.{signature}")


def test_unsigned_token_is_rejected(keys):
    payload = {"uuid": "u", "purpose": "access", "exp": datetime.now(timezone.utc) + timedelta(minutes=5)}
    unsigned = jwt.encode(payload, None, algorithm="none")
    with pytest.raises(jwt.PyJWTError):
        validate_token(unsigned)


def test_empty_uuid_is_rejected(keys):
    private_pem, _ = keys
    payload = {"uuid": "", "purpose": "access", "exp": datetime.now(timezone.utc) + timedelta(minutes=5)}
    token = jwt.encode(payload, private_pem, algorithm="EdDSA")
    with pytest.raises(ValueError):
        validate_token(token)


@pytest.mark.parametrize("garbage", ["", "abc", "a.b.c", "Bearer x"])
def test_garbage_is_rejected(keys, garbage):
    with pytest.raises(jwt.PyJWTError):
        validate_token(garbage)


def test_refresh_token_hash_is_stable_sha256():
    digest = hash_refresh_token("token")
    assert digest == hash_refresh_token("token")
    assert digest != hash_refresh_token("token2")
    assert len(digest) == 64
    assert digest == "3c469e9d6c5875d37a43f353d4f88e61fcf812c66eee3457465a40b0da4153e0"


def test_escaped_newlines_in_pem_are_accepted(keys, monkeypatch):
    from configs.app_config import Config

    private_pem, public_pem = keys
    monkeypatch.setattr(Config, "JWT_SECRET_KEY", private_pem.replace("\n", "\\n"))
    monkeypatch.setattr(Config, "JWT_PUBLIC_KEY", public_pem.replace("\n", "\\n"))
    token, _ = generate_access_token("u", "e@x.com", "s")
    assert validate_token(token)["uuid"] == "u"
