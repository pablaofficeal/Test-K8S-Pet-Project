import pytest

from tests.integration.conftest import CSRF
from utils.cookies.auth_cookies import ACCESS_COOKIE, REFRESH_COOKIE, REFRESH_COOKIE_PATH, USER_COOKIE

LOGOUT = "/main/api/v1/start/log/out/my/account"
REFRESH = "/main/api/v1/start/refresh/my/token"


@pytest.fixture()
def session_client(client, register, login):
    register()
    login()
    return client


def test_logout_clears_cookies(session_client):
    response = session_client.post(LOGOUT, headers=CSRF)
    assert response.status_code == 200
    assert response.get_json() == {"ok": True}
    assert session_client.get_cookie(ACCESS_COOKIE) is None
    assert session_client.get_cookie(USER_COOKIE) is None
    assert session_client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH) is None


def test_logout_revokes_the_session(session_client, database):
    session_client.post(LOGOUT, headers=CSRF)
    (session,) = database.sessions()
    assert session.revoked is True and session.revoked_at is not None


def test_refresh_fails_after_logout(session_client):
    token = session_client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value
    session_client.post(LOGOUT, headers=CSRF)
    session_client.set_cookie(REFRESH_COOKIE, token, path=REFRESH_COOKIE_PATH)
    assert session_client.post(REFRESH, headers=CSRF).status_code == 401


def test_logout_with_only_the_access_cookie(session_client, database):
    session_client.delete_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH)
    session_client.post(LOGOUT, headers=CSRF)
    (session,) = database.sessions()
    assert session.revoked is True


def test_logout_without_a_session_is_harmless(client):
    assert client.post(LOGOUT, headers=CSRF).status_code == 200


def test_logout_with_garbage_cookies_is_harmless(client, database):
    client.set_cookie(ACCESS_COOKIE, "garbage")
    client.set_cookie(REFRESH_COOKIE, "garbage", path=REFRESH_COOKIE_PATH)
    assert client.post(LOGOUT, headers=CSRF).status_code == 200


def test_csrf_header_is_required(session_client, database):
    assert session_client.post(LOGOUT).status_code == 403
    (session,) = database.sessions()
    assert session.revoked is False


def test_logout_only_ends_its_own_session(client, register, login, database):
    register()
    login()
    other = client.application.test_client()
    other.post("/main/api/v1/start/log/in/my/account", json={"email": "alice@example.com", "password": "secret-pass"}, headers=CSRF)
    client.post(LOGOUT, headers=CSRF)
    sessions = database.sessions()
    assert len(sessions) == 2
    assert sorted(session.revoked for session in sessions) == [False, True]
