from datetime import datetime, timezone
from flask import Blueprint, request, jsonify
from db.import_all_models import *
from db.init_db import db
from utils.jwt.jwt import *
from utils.cookies.auth_cookies import *

logout_bpp = Blueprint("logout_bpp", __name__)


def _revoke_session(sid):
    session = RefreshSession.query.filter_by(sid=sid).first()
    if session is None:
        return
    now = datetime.now(timezone.utc)
    session.revoked = True
    session.revoked_at = now
    session.updated_at = now
    db.session.commit()


def _session_id_from_cookies():
    for name, purpose in ((REFRESH_COOKIE, REFRESH_PURPOSE), (ACCESS_COOKIE, ACCESS_PURPOSE)):
        token = request.cookies.get(name)
        if not token:
            continue
        try:
            return validate_token(token, expected_purpose=purpose).get("sid")
        except Exception:
            continue
    return None


@logout_bpp.route("/main/api/v1/start/log/out/my/account", methods=["POST"])
def logout():
    csrf = csrf_error()
    if csrf:
        return csrf

    sid = _session_id_from_cookies()
    if sid:
        _revoke_session(sid)

    return clear_auth_cookies(jsonify({"ok": True})), 200