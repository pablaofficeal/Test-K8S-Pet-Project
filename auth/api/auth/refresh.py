from datetime import datetime, timezone
from flask import Blueprint, request, jsonify
from db.import_all_models import *
from db.init_db import db
from utils.jwt.jwt import *
from utils.cookies.auth_cookies import *
import uuid

refresh_bpp = Blueprint("refresh_bpp", __name__)


def _extract_refresh_token():
    return request.cookies.get(REFRESH_COOKIE)


@refresh_bpp.route("/main/api/v1/start/refresh/my/token", methods=["POST"])
def refresh():
    csrf = csrf_error()
    if csrf:
        return csrf

    token = _extract_refresh_token()
    if not token:
        return jsonify({"error": "refresh token cookie is required"}), 400

    try:
        claims = validate_token(token, expected_purpose=REFRESH_PURPOSE)
    except Exception:
        return jsonify({"error": "invalid token"}), 401

    now = datetime.now(timezone.utc)
    session = RefreshSession.query.filter_by(sid=claims.get("sid")).first()

    if session is None:
        return jsonify({"error": "invalid token"}), 401

    if session.revoked or session.revoked_at is not None or session.token_hash != hash_refresh_token(token):
        RefreshSession.query.filter_by(user_uuid=session.user_uuid).update(
            {"revoked": True, "revoked_at": now, "updated_at": now}
        )
        db.session.commit()
        return jsonify({"error": "invalid token"}), 401

    exp = session.expires_at
    if exp.tzinfo is None:
        exp = exp.replace(tzinfo=timezone.utc)
    if exp <= now:
        session.revoked = True
        session.revoked_at = now
        session.updated_at = now
        db.session.commit()
        return jsonify({"error": "invalid token"}), 401

    user = Users.query.filter(Users.uuid == uuid.UUID(session.user_uuid)).first()
    if user is None:
        session.revoked = True
        session.revoked_at = now
        session.updated_at = now
        db.session.commit()
        return jsonify({"error": "invalid token"}), 401

    access_token, access_exp, new_refresh, refresh_exp, _ = generate_token_pair(
        user.uuid, user.email, session_id=session.sid
    )
    session.token_hash = hash_refresh_token(new_refresh)
    session.expires_at = refresh_exp
    session.updated_at = now
    db.session.commit()

    response = jsonify({
        "token_type": "Bearer",
        "expires_in": int((access_exp - datetime.now(timezone.utc)).total_seconds()),
    })
    return set_auth_cookies(response, access_token, access_exp, new_refresh, refresh_exp), 200