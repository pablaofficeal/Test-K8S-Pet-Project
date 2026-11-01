import pytest

from utils.password.validate import (
    PASSWORD_MAX_LENGTH,
    PASSWORD_MIN_LENGTH,
    password_length_error,
    password_too_long,
)


def test_accepts_boundary_lengths():
    assert password_length_error("a" * PASSWORD_MIN_LENGTH) is None
    assert password_length_error("a" * PASSWORD_MAX_LENGTH) is None


@pytest.mark.parametrize("length", [0, PASSWORD_MIN_LENGTH - 1, PASSWORD_MAX_LENGTH + 1, 500])
def test_rejects_out_of_range_lengths(length):
    assert "characters long" in password_length_error("a" * length)


@pytest.mark.parametrize("value", [None, 123456, ["password"], {"password": "x"}, b"secret1"])
def test_rejects_non_string(value):
    assert password_length_error(value) == "password must be a string"
    assert password_too_long(value) is True


def test_too_long_only_checks_upper_bound():
    assert password_too_long("a" * PASSWORD_MAX_LENGTH) is False
    assert password_too_long("a" * (PASSWORD_MAX_LENGTH + 1)) is True
    assert password_too_long("") is False


def test_length_counts_characters_not_bytes():
    assert password_length_error("ы" * PASSWORD_MAX_LENGTH) is None
