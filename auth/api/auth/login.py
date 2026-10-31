from datetime import datetime, timezone
from flask import Blueprint, request, jsonify
from db.import_all_models import *
from db.init_db import db
from utils.password.check import *
from utils.password.validate import password_too_long
from utils.jwt.jwt import *
from utils.cookies.auth_cookies import *

login_bpp = Blueprint("login_bpp", __name__)


@login_bpp.route("/main/api/v1/start/log/in/my/account", methods=["POST"])
def login():
    csrf = csrf_error()
    if csrf:
        return csrf

    email = request.form.get("email") or (request.get_json(silent=True) or {}).get("email")
    password = request.form.get("password") or (request.get_json(silent=True) or {}).get("password")

    if not email or not password:
        return jsonify({"error": "email and password are required"}), 400

    if password_too_long(password):
        return jsonify({"error": "invalid credentials"}), 401

    user = Users.query.filter(Users.email == email).first()

    if user is None or not check_password(password, user.password):
        return jsonify({"error": "invalid credentials"}), 401

    access_token, access_exp, refresh_token, refresh_exp, sid = generate_token_pair(
        user.uuid, user.email
    )

    now = datetime.now(timezone.utc)
    db.session.add(RefreshSession(
        sid=sid,
        user_uuid=str(user.uuid),
        token_hash=hash_refresh_token(refresh_token),
        expires_at=refresh_exp,
        revoked=False,
        revoked_at=None,
        created_at=now,
        updated_at=now,
    ))
    db.session.commit()

    response = jsonify({
        "token_type": "Bearer",
        "expires_in": int((access_exp - datetime.now(timezone.utc)).total_seconds()),
    })
    return set_auth_cookies(response, access_token, access_exp, refresh_token, refresh_exp, user.uuid), 200