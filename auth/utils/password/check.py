import base64
import hmac
from argon2.low_level import hash_secret_raw, Type

def check_password(password: str, encoded_hash: str) -> bool:
    parts = encoded_hash.split("$")
    if len(parts) != 6:
        raise ValueError("invalid hash format")

    params = {}
    for item in parts[3].split(","):
        key, value = item.split("=")
        params[key] = int(value)

    memory      = params["m"]
    iterations  = params["t"]
    parallelism = params["p"]

    def decode(s: str) -> bytes:
        padding = (4 - len(s) % 4) % 4
        return base64.b64decode(s + "=" * padding)

    salt        = decode(parts[4])
    decoded_hash = decode(parts[5])

    comparison_hash = hash_secret_raw(
        secret=password.encode(),
        salt=salt,
        time_cost=iterations,
        memory_cost=memory,
        parallelism=parallelism,
        hash_len=len(decoded_hash),
        type=Type.ID,
    )

    return hmac.compare_digest(decoded_hash, comparison_hash)