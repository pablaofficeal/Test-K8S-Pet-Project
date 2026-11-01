import pytest
from flask import Flask, g, jsonify

from utils.jwt.jwt import generate_access_token, generate_refresh_token
from utils.middleware.middleware import require_auth


@pytest.fixture()
def client(keys):
    app = Flask(__name__)

    @app.get("/private")
    @require_auth
    def private():
        return jsonify({"uuid": g.user_uuid, "email": g.jwt_claims["email"]})

    return app.test_client()


def test_missing_header_is_rejected(client):
    response = client.get("/private")
    assert response.status_code == 401
    assert response.get_json() == {"error": "missing token"}


@pytest.mark.parametrize("header", ["abc", "Bearer", "Bearer a b", "Basic dXNlcjpwYXNz", "Token abc"])
def test_malformed_header_is_rejected(client, header):
    assert client.get("/private", headers={"Authorization": header}).status_code == 401


def test_garbage_token_is_rejected(client):
    response = client.get("/private", headers={"Authorization": "Bearer garbage"})
    assert response.status_code == 401
    assert response.get_json() == {"error": "invalid token"}


def test_refresh_token_cannot_be_used_as_access(client):
    token, _ = generate_refresh_token("u", "e@x.com", "s")
    assert client.get("/private", headers={"Authorization": f"Bearer {token}"}).status_code == 401


@pytest.mark.parametrize("scheme", ["Bearer", "bearer", "BEARER"])
def test_valid_token_sets_user_context(client, scheme):
    token, _ = generate_access_token("user-1", "a@b.com", "sid")
    response = client.get("/private", headers={"Authorization": f"{scheme} {token}"})
    assert response.status_code == 200
    assert response.get_json() == {"uuid": "user-1", "email": "a@b.com"}
