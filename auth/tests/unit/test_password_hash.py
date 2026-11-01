import pytest

from utils.password.check import check_password
from utils.password.hash import hash_password


def test_hash_has_argon2id_format():
    encoded = hash_password("correct horse")
    parts = encoded.split("$")
    assert parts[1] == "argon2id"
    assert parts[2].startswith("v=")
    assert len(parts) == 6
    assert "m=" in parts[3] and "t=" in parts[3] and "p=" in parts[3]


def test_hash_does_not_contain_password():
    assert "correct horse" not in hash_password("correct horse")


def test_roundtrip():
    encoded = hash_password("correct horse")
    assert check_password("correct horse", encoded) is True


def test_wrong_password_is_rejected():
    encoded = hash_password("correct horse")
    assert check_password("correct horse ", encoded) is False
    assert check_password("Correct horse", encoded) is False
    assert check_password("", encoded) is False


def test_salt_makes_hashes_unique():
    first = hash_password("same password")
    second = hash_password("same password")
    assert first != second
    assert check_password("same password", first) and check_password("same password", second)


def test_unicode_password():
    encoded = hash_password("пароль-🔒")
    assert check_password("пароль-🔒", encoded) is True
    assert check_password("пароль-🔓", encoded) is False


def test_check_uses_parameters_stored_in_the_hash():
    # хэш с другими параметрами, чем в конфиге, обязан по-прежнему проверяться
    from argon2.low_level import Type, hash_secret_raw
    import base64

    salt = b"0123456789abcdef"
    raw = hash_secret_raw(b"legacy", salt, time_cost=1, memory_cost=8, parallelism=1, hash_len=16, type=Type.ID)
    encode = lambda data: base64.b64encode(data).rstrip(b"=").decode()
    encoded = f"$argon2id$v=19$m=8,t=1,p=1${encode(salt)}${encode(raw)}"
    assert check_password("legacy", encoded) is True
    assert check_password("other", encoded) is False


@pytest.mark.parametrize("broken", ["", "plain", "$argon2id$v=19$m=8,t=1,p=1$salt", "a$b$c$d$e"])
def test_malformed_hash_raises(broken):
    with pytest.raises(Exception):
        check_password("x", broken)
