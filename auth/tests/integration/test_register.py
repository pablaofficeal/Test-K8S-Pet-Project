import pytest

from utils.password.check import check_password

URL = "/main/api/v1/start/create/my/account"


def test_registers_user(client, register, database, published):
    response = register()
    assert response.status_code == 201
    assert response.get_json() == {"username": "alice", "email": "alice@example.com"}
    user = database.user("alice@example.com")
    assert user is not None and user.username == "alice"


def test_password_is_stored_as_argon2_hash(client, register, database):
    register(password="secret-pass")
    stored = database.user("alice@example.com").password
    assert stored.startswith("$argon2id$")
    assert "secret-pass" not in stored
    assert check_password("secret-pass", stored)


def test_response_never_contains_the_password(client, register):
    body = register().get_data(as_text=True)
    assert "secret-pass" not in body and "password" not in body


def test_verification_request_is_published(client, register, database, published):
    register()
    assert len(published) == 1
    user_uuid, trace_id = published[0]
    assert user_uuid == str(database.user("alice@example.com").uuid)
    assert trace_id == "-"


def test_registration_survives_a_broken_queue(client, register, database, monkeypatch):
    monkeypatch.setattr("api.auth.registers.publish_verification_requested", lambda *args: False)
    assert register().status_code == 201
    assert database.user("alice@example.com") is not None


@pytest.mark.parametrize(
    "payload",
    [
        {},
        {"username": "a", "email": "a@x.com"},
        {"username": "a", "password": "secret-pass"},
        {"email": "a@x.com", "password": "secret-pass"},
        {"username": "", "email": "a@x.com", "password": "secret-pass"},
    ],
)
def test_missing_fields_are_rejected(client, payload):
    response = client.post(URL, json=payload)
    assert response.status_code == 400
    assert "required" in response.get_json()["error"]


@pytest.mark.parametrize("password", ["short", "a" * 65, "12345"])
def test_password_length_is_validated(client, register, password):
    response = register(password=password)
    assert response.status_code == 400
    assert "characters long" in response.get_json()["error"]


def test_non_string_password_is_rejected(client):
    response = client.post(URL, json={"username": "a", "email": "a@x.com", "password": 123456789})
    assert response.status_code == 400


def test_invalid_json_is_rejected(client):
    assert client.post(URL, data="not json", content_type="application/json").status_code == 400


def test_duplicate_email_is_rejected(client, register, database):
    register()
    response = register(username="other")
    assert response.status_code == 409
    assert response.get_json() == {"error": "user already exists"}


def test_duplicate_username_is_rejected(client, register):
    register()
    assert register(email="other@example.com").status_code == 409


def test_different_users_can_register(client, register):
    assert register().status_code == 201
    assert register(username="bob", email="bob@example.com").status_code == 201


def test_failed_registration_publishes_nothing(client, register, published):
    register(password="x")
    assert published == []
