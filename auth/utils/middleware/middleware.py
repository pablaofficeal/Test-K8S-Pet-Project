from functools import wraps
from flask import request, jsonify, g
from utils.jwt.jwt import validate_token


def require_auth(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        header = request.headers.get("Authorization", "")
        parts = header.split()
        if len(parts) != 2 or parts[0].lower() != "bearer":
            return jsonify({"error": "missing token"}), 401
        try:
            claims = validate_token(parts[1])
        except Exception:
            return jsonify({"error": "invalid token"}), 401
        g.jwt_claims = claims
        g.user_uuid = claims.get("uuid")
        return f(*args, **kwargs)
    return decorated