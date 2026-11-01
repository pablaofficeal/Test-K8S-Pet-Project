from datetime import datetime, timedelta, timezone

import pytest

from tests.integration.conftest import CSRF
from utils.cookies.auth_cookies import ACCESS_COOKIE, REFRESH_COOKIE, REFRESH_COOKIE_PATH
from utils.jwt.jwt import generate_refresh_token, hash_refresh_token, validate_token

URL = "/main/api/v1/start/refresh/my/token"


@pytest.fixture()
def session_client(client, register, login):
    register()
    login()
    return client


def refresh(client, **kwargs):
    return client.post(URL, headers=CSRF, **kwargs)


def test_refresh_rotates_the_token_pair(session_client, database):
    old_refresh = session_client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value
    response = refresh(session_client)
    assert response.status_code == 200
    assert response.get_json()["token_type"] == "Bearer"
    new_refresh = session_client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value
    assert new_refresh != old_refresh
    assert validate_token(session_client.get_cookie(ACCESS_COOKIE).value)["uuid"]
    (session,) = database.sessions()
    assert session.token_hash == hash_refresh_token(new_refresh)


def test_refresh_keeps_the_session_id(session_client, database):
    (before,) = database.sessions()
    refresh(session_client)
    (after,) = database.sessions()
    assert after.sid == before.sid
    assert after.expires_at >= before.expires_at


def test_csrf_header_is_required(session_client):
    assert session_client.post(URL).status_code == 403


def test_refresh_cookie_is_required(client):
    response = refresh(client)
    assert response.status_code == 400
    assert "cookie" in response.get_json()["error"]


def test_garbage_refresh_token(client):
    client.set_cookie(REFRESH_COOKIE, "garbage", path=REFRESH_COOKIE_PATH)
    assert refresh(client).status_code == 401


def test_access_token_is_not_accepted_as_refresh(session_client):
    access = session_client.get_cookie(ACCESS_COOKIE).value
    session_client.set_cookie(REFRESH_COOKIE, access, path=REFRESH_COOKIE_PATH)
    assert refresh(session_client).status_code == 401


def test_token_for_an_unknown_session(session_client):
    token, _ = generate_refresh_token("u", "e@x.com", "no-such-sid")
    session_client.set_cookie(REFRESH_COOKIE, token, path=REFRESH_COOKIE_PATH)
    assert refresh(session_client).status_code == 401


def test_reusing_a_rotated_token_revokes_every_session(session_client, login, database):
    stolen = session_client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value
    refresh(session_client)
    login()
    assert len(database.sessions()) == 2

    session_client.set_cookie(REFRESH_COOKIE, stolen, path=REFRESH_COOKIE_PATH)
    assert refresh(session_client).status_code == 401
    assert all(session.revoked and session.revoked_at for session in database.sessions())

    session_client.set_cookie(REFRESH_COOKIE, session_client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value, path=REFRESH_COOKIE_PATH)
    login()
    fresh = database.sessions()[-1]
    assert fresh.revoked is False


def test_revoked_session_cannot_refresh(session_client, database):
    (session,) = database.sessions()
    database.update_session(session.sid, revoked=True, revoked_at=datetime.now(timezone.utc))
    assert refresh(session_client).status_code == 401


def test_expired_session_is_revoked(session_client, database):
    (session,) = database.sessions()
    database.update_session(session.sid, expires_at=datetime.now(timezone.utc) - timedelta(minutes=1))
    assert refresh(session_client).status_code == 401
    (after,) = database.sessions()
    assert after.revoked is True and after.revoked_at is not None


def test_deleted_user_cannot_refresh(session_client, database):
    database.delete_user("alice@example.com")
    assert refresh(session_client).status_code == 401
    (session,) = database.sessions()
    assert session.revoked is True


def test_rotation_chain_works_repeatedly(session_client):
    for _ in range(3):
        assert refresh(session_client).status_code == 200
