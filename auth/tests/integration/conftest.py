import pytest
from flask import Flask
from sqlalchemy import text

from api.routers.auth_routers import register_all_blueprints
from db.import_all_models import RefreshSession, Users
from db.init_db import db
from tests.conftest import database_url

SCHEMA = "auth_test"
CSRF = {"X-Requested-With": "XMLHttpRequest"}


@pytest.fixture(scope="session")
def app():
    url = database_url()
    if not url:
        pytest.skip("DATABASE_URL is not configured")
    app = Flask(__name__)
    app.config.update(
        SQLALCHEMY_DATABASE_URI=url,
        SQLALCHEMY_ENGINE_OPTIONS={"connect_args": {"options": f"-csearch_path={SCHEMA}"}},
        TESTING=True,
    )
    register_all_blueprints(app)
    db.init_app(app)
    try:
        with app.app_context():
            db.session.execute(text(f"CREATE SCHEMA IF NOT EXISTS {SCHEMA}"))
            db.session.commit()
            db.create_all()
    except Exception as err:
        pytest.skip(f"postgres is not available: {err}")
    yield app
    with app.app_context():
        db.session.remove()
        db.session.execute(text(f"DROP SCHEMA {SCHEMA} CASCADE"))
        db.session.commit()


@pytest.fixture()
def published(monkeypatch):
    # RabbitMQ в интеграционных тестах не нужен: запоминаем, что ушло бы в очередь
    calls = []

    def fake_publish(user_uuid, trace_id):
        calls.append((str(user_uuid), trace_id))
        return True

    monkeypatch.setattr("api.auth.registers.publish_verification_requested", fake_publish)
    return calls


@pytest.fixture()
def client(app, keys, published):
    with app.app_context():
        db.session.execute(text("TRUNCATE users, refresh_tokens"))
        db.session.commit()
    return app.test_client()


@pytest.fixture()
def database(app):
    # свежий контекст на каждое обращение, чтобы не держать устаревший снимок данных
    class Database:
        def user(self, email):
            with app.app_context():
                return Users.query.filter_by(email=email).first()

        def sessions(self, user_uuid=None):
            with app.app_context():
                query = RefreshSession.query
                if user_uuid:
                    query = query.filter_by(user_uuid=str(user_uuid))
                rows = query.order_by(RefreshSession.created_at).all()
                for row in rows:
                    db.session.expunge(row)
                return rows

        def update_session(self, sid, **values):
            with app.app_context():
                RefreshSession.query.filter_by(sid=sid).update(values)
                db.session.commit()

        def delete_user(self, email):
            with app.app_context():
                Users.query.filter_by(email=email).delete()
                db.session.commit()

    return Database()


@pytest.fixture()
def register(client):
    def do_register(username="alice", email="alice@example.com", password="secret-pass"):
        return client.post(
            "/main/api/v1/start/create/my/account",
            json={"username": username, "email": email, "password": password},
        )

    return do_register


@pytest.fixture()
def login(client):
    def do_login(email="alice@example.com", password="secret-pass", **kwargs):
        return client.post(
            "/main/api/v1/start/log/in/my/account",
            json={"email": email, "password": password},
            headers=CSRF,
            **kwargs,
        )

    return do_login
