import os
import base64
from argon2.low_level import hash_secret_raw, Type
from utils.password.helper import *
from configs.app_config import Config

ARGON2_VERSION = Config().ARGON2_VERSION

def hash_password(password: str) -> str:
    salt = os.urandom(p.salt_length)

    hash_bytes = hash_secret_raw(
        secret=password.encode(),
        salt=salt,
        time_cost=p.iterations,
        memory_cost=p.memory,
        parallelism=p.parallelism,
        hash_len=p.key_length,
        type=Type.ID,
    )

    b64_salt = base64.b64encode(salt).rstrip(b"=").decode()
    b64_hash = base64.b64encode(hash_bytes).rstrip(b"=").decode()

    return (
        f"$argon2id$v={ARGON2_VERSION}"
        f"$m={p.memory},t={p.iterations},p={p.parallelism}"
        f"${b64_salt}${b64_hash}"
    )