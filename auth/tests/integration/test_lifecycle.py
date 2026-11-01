from tests.integration.conftest import CSRF
from utils.cookies.auth_cookies import ACCESS_COOKIE, REFRESH_COOKIE, REFRESH_COOKIE_PATH
from utils.jwt.jwt import validate_token


def test_register_login_refresh_logout(client, register, login, database, published):
    assert register().status_code == 201
    assert len(published) == 1

    assert login().status_code == 200
    first_access = validate_token(client.get_cookie(ACCESS_COOKIE).value)

    assert client.post("/main/api/v1/start/refresh/my/token", headers=CSRF).status_code == 200
    second_access = validate_token(client.get_cookie(ACCESS_COOKIE).value)
    assert second_access["sid"] == first_access["sid"]
    assert second_access["uuid"] == first_access["uuid"]

    assert client.post("/main/api/v1/start/log/out/my/account", headers=CSRF).status_code == 200
    assert client.get_cookie(ACCESS_COOKIE) is None
    assert client.get_cookie(REFRESH_COOKIE, path=REFRESH_COOKIE_PATH) is None
    (session,) = database.sessions()
    assert session.revoked is True


def test_a_new_login_works_after_logout(client, register, login):
    register()
    login()
    client.post("/main/api/v1/start/log/out/my/account", headers=CSRF)
    assert login().status_code == 200
    assert client.post("/main/api/v1/start/refresh/my/token", headers=CSRF).status_code == 200


def test_two_users_do_not_share_sessions(client, register, login, database):
    register()
    register(username="bob", email="bob@example.com")
    login()
    login(email="bob@example.com")
    alice = database.user("alice@example.com")
    bob = database.user("bob@example.com")
    assert len(database.sessions(alice.uuid)) == 1
    assert len(database.sessions(bob.uuid)) == 1
