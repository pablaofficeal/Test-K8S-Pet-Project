PASSWORD_MIN_LENGTH = 6
PASSWORD_MAX_LENGTH = 64


def password_length_error(password):
    if not isinstance(password, str):
        return "password must be a string"
    if len(password) < PASSWORD_MIN_LENGTH or len(password) > PASSWORD_MAX_LENGTH:
        return f"password must be {PASSWORD_MIN_LENGTH}-{PASSWORD_MAX_LENGTH} characters long"
    return None


def password_too_long(password):
    # Argon2 дорогой, поэтому слишком длинный пароль отсекаем до хэширования
    return not isinstance(password, str) or len(password) > PASSWORD_MAX_LENGTH
