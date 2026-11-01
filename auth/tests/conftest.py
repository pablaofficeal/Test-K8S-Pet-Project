import os
from urllib.parse import urlsplit, urlunsplit

import pytest
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey

from configs.app_config import Config


def make_key_pair():
    private = Ed25519PrivateKey.generate()
    private_pem = private.private_bytes(
        serialization.Encoding.PEM,
        serialization.PrivateFormat.PKCS8,
        serialization.NoEncryption(),
    ).decode()
    public_pem = private.public_key().public_bytes(
        serialization.Encoding.PEM,
        serialization.PublicFormat.SubjectPublicKeyInfo,
    ).decode()
    return private_pem, public_pem


@pytest.fixture()
def keys(monkeypatch):
    # свежая пара ключей на тест, чтобы не зависеть от моков из .env
    private_pem, public_pem = make_key_pair()
    monkeypatch.setattr(Config, "JWT_SECRET_KEY", private_pem)
    monkeypatch.setattr(Config, "JWT_PUBLIC_KEY", public_pem)
    monkeypatch.setattr(Config, "JWT_ALGORITHM", "EdDSA")
    monkeypatch.setattr(Config, "JWT_ACCESS_TTL", "15m")
    monkeypatch.setattr(Config, "JWT_REFRESH_TTL", "7d")
    return private_pem, public_pem


def database_url():
    explicit = os.getenv("TEST_DATABASE_URL")
    if explicit:
        return explicit
    parts = urlsplit(Config.SQLALCHEMY_DATABASE_URI or "")
    if not parts.hostname:
        return None
    credentials = parts.netloc.rpartition("@")[0]
    host = os.getenv("TEST_DATABASE_HOST", "localhost:56799")
    netloc = f"{credentials}@{host}" if credentials else host
    return urlunsplit((parts.scheme, netloc, parts.path, parts.query, parts.fragment))
