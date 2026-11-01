from datetime import datetime, timezone
from flask import request, jsonify
from configs.app_config import Config

ACCESS_COOKIE = "Cadence-Cloud-Auth-Token"
REFRESH_COOKIE = "Cadence-Cloud-Refresh-Token"
USER_COOKIE = "Cadence-Cloud-User-Id"
REFRESH_COOKIE_PATH = "/main/api/v1/start"
CSRF_HEADER = "X-Requested-With"


def _max_age(expires_at: datetime) -> int:
    if expires_at.tzinfo is None:
        expires_at = expires_at.replace(tzinfo=timezone.utc)
    return max(0, int((expires_at - datetime.now(timezone.utc)).total_seconds()))


def set_auth_cookies(response, access_token, access_exp, refresh_token, refresh_exp, user_uuid):
    response.set_cookie(
        ACCESS_COOKIE, access_token, max_age=_max_age(access_exp), path="/",
        httponly=True, secure=Config.COOKIE_SECURE, samesite="Lax",
    )
    # refresh уходит только на auth-эндпоинты, а не на каждый запрос к API
    response.set_cookie(
        REFRESH_COOKIE, refresh_token, max_age=_max_age(refresh_exp), path=REFRESH_COOKIE_PATH,
        httponly=True, secure=Config.COOKIE_SECURE, samesite="Lax",
    )
    # Косметика для фронта (URL и отображение). Не HttpOnly, доверять ему нельзя:
    # права проверяет шлюз по access-токену
    response.set_cookie(
        USER_COOKIE, str(user_uuid), max_age=_max_age(refresh_exp), path="/",
        httponly=False, secure=Config.COOKIE_SECURE, samesite="Lax",
    )
    return response


def clear_auth_cookies(response):
    response.delete_cookie(ACCESS_COOKIE, path="/", httponly=True, secure=Config.COOKIE_SECURE, samesite="Lax")
    response.delete_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH, httponly=True, secure=Config.COOKIE_SECURE, samesite="Lax")
    response.delete_cookie(USER_COOKIE, path="/", secure=Config.COOKIE_SECURE, samesite="Lax")
    return response


def csrf_error():
    if request.headers.get(CSRF_HEADER):
        return None
    return jsonify({"error": "csrf header required"}), 403