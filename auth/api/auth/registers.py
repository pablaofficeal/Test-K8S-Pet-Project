from flask import Blueprint, request, jsonify, g
from db.import_all_models import *
from db.init_db import db
from utils.password.hash import *
from utils.password.validate import password_length_error
from utils.mq.verification_publisher import publish_verification_requested
import logging
import uuid

register_bpp = Blueprint("register_bpp", __name__)


@register_bpp.route("/main/api/v1/start/create/my/account", methods=["POST"])
def create_account():
    data = request.get_json(silent=True) or {}
    username = data.get("username")
    email = data.get("email")
    password = data.get("password")

    if not username or not email or not password:
        return jsonify({"error": "username, email and password are required"}), 400

    password_error = password_length_error(password)
    if password_error:
        return jsonify({"error": password_error}), 400

    existing_user = Users.query.filter((Users.username == username) | (Users.email == email)).first()
    if existing_user:
        return jsonify({"error": "user already exists"}), 409

    new_user = Users(
        uuid=uuid.uuid4(),
        username=username,
        email=email,
        password=hash_password(password),
    )
    db.session.add(new_user)
    db.session.commit()

    publish_verification_requested(new_user.uuid, getattr(g, "trace_id", "-"))

    return jsonify({"username": new_user.username, "email": new_user.email}), 201