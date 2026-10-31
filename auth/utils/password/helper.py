from dataclasses import dataclass
from configs.app_config import Config


@dataclass
class Params:
    memory: int
    iterations: int
    parallelism: int
    salt_length: int
    key_length: int

p = Params(
    memory=Config().ARGON2_MEMORY,
    iterations=Config().ARGON2_ITERATIONS,
    parallelism=Config().ARGON2_PARALLELISM,
    salt_length=Config().ARGON2_SALT_LENGTH,
    key_length=Config().ARGON2_KEY_LENGTH,
)