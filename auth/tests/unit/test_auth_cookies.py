from datetime import datetime, timedelta, timezone

import pytest
from flask import Flask, jsonify

from configs.app_config import Config
from utils.cookies.auth_cookies import (
    ACCESS_COOKIE,
    CSRF_HEADER,
    REFRESH_COOKIE,
    REFRESH_COOKIE_PATH,
    USER_COOKIE,
    clear_auth_cookies,
    csrf_error,
    set_auth_cookies,
)


@pytest.fixture()
def app():
    return Flask(__name__)


def cookies_of(response):
    parsed = {}
    for header in response.headers.getlist("Set-Cookie"):
        name, _, rest = header.partition("=")
        parsed[name] = rest
    return parsed


def build_response(app, secure):
    Config.COOKIE_SECURE = secure
    now = datetime.now(timezone.utc)
    with app.app_context():
        return set_auth_cookies(jsonify({}), "access", now + timedelta(minutes=15), "refresh", now + timedelta(days=7), "user-1")


def test_sets_three_cookies(app):
    cookies = cookies_of(build_response(app, False))
    assert set(cookies) == {ACCESS_COOKIE, REFRESH_COOKIE, USER_COOKIE}
    assert cookies[ACCESS_COOKIE].startswith("access;")
    assert cookies[REFRESH_COOKIE].startswith("refresh;")
    assert cookies[USER_COOKIE].startswith("user-1;")


def test_access_and_refresh_cookies_are_http_only_lax(app):
    cookies = cookies_of(build_response(app, False))
    for name in (ACCESS_COOKIE, REFRESH_COOKIE):
        assert "HttpOnly" in cookies[name]
        assert "SameSite=Lax" in cookies[name]


def test_user_cookie_is_readable_by_the_frontend(app):
    assert "HttpOnly" not in cookies_of(build_response(app, False))[USER_COOKIE]


def test_refresh_cookie_is_limited_to_auth_paths(app):
    cookies = cookies_of(build_response(app, False))
    assert f"Path={REFRESH_COOKIE_PATH}" in cookies[REFRESH_COOKIE]
    assert "Path=/;" in cookies[ACCESS_COOKIE] + ";"


def test_max_age_follows_expiry(app):
    cookies = cookies_of(build_response(app, False))
    access_age = int(cookies[ACCESS_COOKIE].split("Max-Age=")[1].split(";")[0])
    refresh_age = int(cookies[REFRESH_COOKIE].split("Max-Age=")[1].split(";")[0])
    assert 14 * 60 <= access_age <= 15 * 60
    assert 6 * 24 * 3600 < refresh_age <= 7 * 24 * 3600


def test_secure_flag_follows_config(app):
    assert "Secure" not in "".join(cookies_of(build_response(app, False)).values())
    secure = cookies_of(build_response(app, True))
    assert all("Secure" in value for value in secure.values())
    Config.COOKIE_SECURE = False


def test_naive_expiry_is_treated_as_utc(app):
    Config.COOKIE_SECURE = False
    naive = datetime.now(timezone.utc).replace(tzinfo=None) + timedelta(minutes=10)
    with app.app_context():
        response = set_auth_cookies(jsonify({}), "a", naive, "r", naive, "u")
    age = int(cookies_of(response)[ACCESS_COOKIE].split("Max-Age=")[1].split(";")[0])
    assert 9 * 60 <= age <= 10 * 60


def test_expired_tokens_get_zero_max_age(app):
    Config.COOKIE_SECURE = False
    past = datetime.now(timezone.utc) - timedelta(hours=1)
    with app.app_context():
        response = set_auth_cookies(jsonify({}), "a", past, "r", past, "u")
    assert "Max-Age=0" in cookies_of(response)[ACCESS_COOKIE]


def test_clear_cookies_expires_all_three(app):
    Config.COOKIE_SECURE = False
    with app.app_context():
        response = clear_auth_cookies(jsonify({}))
    cookies = cookies_of(response)
    assert set(cookies) == {ACCESS_COOKIE, REFRESH_COOKIE, USER_COOKIE}
    for value in cookies.values():
        assert "Max-Age=0" in value or "1970" in value
    assert f"Path={REFRESH_COOKIE_PATH}" in cookies[REFRESH_COOKIE]


def test_csrf_header_is_required(app):
    with app.test_request_context("/", method="POST"):
        body, status = csrf_error()
        assert status == 403
        assert body.get_json() == {"error": "csrf header required"}


def test_csrf_passes_with_header(app):
    with app.test_request_context("/", method="POST", headers={CSRF_HEADER: "XMLHttpRequest"}):
        assert csrf_error() is None


def test_empty_csrf_header_is_rejected(app):
    with app.test_request_context("/", method="POST", headers={CSRF_HEADER: ""}):
        assert csrf_error()[1] == 403
