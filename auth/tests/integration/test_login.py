import pytest

from tests.integration.conftest import CSRF
from utils.cookies.auth_cookies import ACCESS_COOKIE, REFRESH_COOKIE, REFRESH_COOKIE_PATH, USER_COOKIE
from utils.jwt.jwt import REFRESH_PURPOSE, hash_refresh_token, validate_token

URL = "/main/api/v1/start/log/in/my/account"


@pytest.fixture()
def account(client, register):
    register()


def test_login_sets_cookies_and_returns_expiry(client, account, login):
    response = login()
    assert response.status_code == 200
    body = response.get_json()
    assert body["token_type"] == "Bearer"
    assert 14 * 60 <= body["expires_in"] <= 15 * 60
    assert client.get_cookie(ACCESS_COOKIE) is not None
    assert client.get_cookie(USER_COOKIE) is not None
    assert client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH) is not None


def test_tokens_are_not_in_the_response_body(client, account, login):
    body = login().get_data(as_text=True)
    assert client.get_cookie(ACCESS_COOKIE).value not in body
    assert client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value not in body


def test_cookie_flags(client, account, login):
    login()
    access = client.get_cookie(ACCESS_COOKIE)
    refresh = client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH)
    user = client.get_cookie(USER_COOKIE)
    assert access.http_only and access.path == "/"
    assert refresh.http_only and refresh.path == REFRESH_COOKIE_PATH
    assert not user.http_only


def test_issued_tokens_belong_to_the_user(client, account, login, database):
    login()
    user = database.user("alice@example.com")
    access = validate_token(client.get_cookie(ACCESS_COOKIE).value)
    refresh = validate_token(client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value, expected_purpose=REFRESH_PURPOSE)
    assert access["uuid"] == str(user.uuid) == refresh["uuid"]
    assert access["email"] == "alice@example.com"
    assert access["sid"] == refresh["sid"]
    assert client.get_cookie(USER_COOKIE).value == str(user.uuid)


def test_refresh_session_is_stored_hashed(client, account, login, database):
    login()
    token = client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH).value
    (session,) = database.sessions()
    assert session.token_hash == hash_refresh_token(token)
    assert token not in session.token_hash
    assert session.revoked is False and session.revoked_at is None
    assert session.sid == validate_token(token, expected_purpose=REFRESH_PURPOSE)["sid"]


def test_every_login_creates_its_own_session(client, account, login, database):
    login()
    login()
    sessions = database.sessions()
    assert len(sessions) == 2 and sessions[0].sid != sessions[1].sid


def test_form_encoded_body_is_accepted(client, account):
    response = client.post(URL, data={"email": "alice@example.com", "password": "secret-pass"}, headers=CSRF)
    assert response.status_code == 200


def test_csrf_header_is_required(client, account):
    response = client.post(URL, json={"email": "alice@example.com", "password": "secret-pass"})
    assert response.status_code == 403
    assert client.get_cookie(ACCESS_COOKIE) is None


@pytest.mark.parametrize("payload", [{}, {"email": "alice@example.com"}, {"password": "secret-pass"}, {"email": "", "password": ""}])
def test_missing_credentials(client, account, payload):
    assert client.post(URL, json=payload, headers=CSRF).status_code == 400


def test_wrong_password(client, account, login, database):
    response = login(password="wrong-pass")
    assert response.status_code == 401
    assert response.get_json() == {"error": "invalid credentials"}
    assert client.get_cookie(ACCESS_COOKIE) is None
    assert database.sessions() == []


def test_unknown_user_looks_like_a_wrong_password(client, account, login):
    unknown = login(email="nobody@example.com")
    wrong = login(password="wrong-pass")
    assert unknown.status_code == wrong.status_code == 401
    assert unknown.get_json() == wrong.get_json()


def test_too_long_password_is_rejected_before_hashing(client, account, login):
    response = login(password="a" * 65)
    assert response.status_code == 401


def test_email_is_matched_exactly(client, account, login):
    assert login(email="ALICE@example.com").status_code == 401


def test_sql_injection_attempt_is_just_a_wrong_login(client, account, login):
    assert login(email="' OR '1'='1", password="x' OR '1'='1").status_code == 401
