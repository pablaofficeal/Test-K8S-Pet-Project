#!/usr/bin/env python3
"""Small TUI prototype for alphabet-aware end-to-end encryption."""
import argparse
import base64
import curses
import hashlib
import hmac
import json
import os
import sys
from pathlib import Path
from typing import Any
from cryptography.hazmat.primitives import *
from cryptography.hazmat.primitives.asymmetric.x25519 import *
from cryptography.hazmat.primitives.kdf.hkdf import *


KEY_VERSION = 1
ENVELOPE_VERSION = 2
ENVELOPE_TYPE = "my-masengers/message"
VERSION = KEY_VERSION
ALGORITHM = "X25519-HKDF-SHA256+DualAlphabetNoise-FullRandom-v2"
ALGORITHM_FULL_V1 = "X25519-HKDF-SHA256+DualAlphabetNoise-FullRandom"
ALGORITHM_DUAL_V1 = "X25519-HKDF-SHA256+DualAlphabetNoise-HMAC-SHA256"
LEGACY_ALGORITHM = "X25519-HKDF-SHA256+AlphabetNoise-HMAC-SHA256"
SUPPORTED_ALGORITHMS = (ALGORITHM,)
LEGACY_ENVELOPE_VERSION = 1
MAX_TEXT_BYTES = 1_048_576
ENVELOPE_REQUIRED_KEYS = frozenset(
    {
        "type",
        "version",
        "algorithm",
        "alphabet_hash",
        "alphabet_hash2",
        "sender_key_id",
        "recipient_key_id",
        "salt",
        "nonce",
        "ciphertext",
        "auth_tag",
    }
)


def _b64(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode("ascii")


def _unb64(value: str, name: str) -> bytes:
    try:
        padding = "=" * (-len(value) % 4)
        return base64.b64decode(value + padding, altchars=b"-_", validate=True)
    except (ValueError, TypeError) as error:
        raise ValueError(f"Некорректное поле {name}") from error


def _key_id(public_key: bytes) -> str:
    return hashlib.sha256(public_key).hexdigest()[:32]


def _read_json(path: Path) -> dict[str, Any]:
    if path.stat().st_size > MAX_TEXT_BYTES * 20:
        raise ValueError(f"Файл слишком большой: {path}")
    with path.open("r", encoding="utf-8") as file:
        raw = file.read()

    def _no_dupes(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
        seen: set[str] = set()
        result: dict[str, Any] = {}
        for key, val in pairs:
            if key in seen:
                raise ValueError(f"В {path} дублируется ключ {key!r}")
            seen.add(key)
            result[key] = val
        return result

    try:
        value = json.loads(raw, object_pairs_hook=_no_dupes)
    except json.JSONDecodeError as error:
        raise ValueError(f"В {path} битый JSON: {error}") from error
    if not isinstance(value, dict):
        raise ValueError(f"В {path} ожидается JSON-объект")
    return value


def _write_json(
    path: Path, value: dict[str, Any], mode: int, overwrite: bool = False
) -> None:
    open_mode = "w" if overwrite else "x"
    with path.open(open_mode, encoding="utf-8") as file:
        json.dump(value, file, ensure_ascii=False, indent=2)
        file.write("\n")
    path.chmod(mode)


def load_alphabet(path: Path) -> tuple[dict[str, str], dict[str, str], str]:
    """Load encryption.json and create reversible character/token maps."""
    data = _read_json(path)
    forward: dict[str, str] = {}
    reverse: dict[str, str] = {}

    for section_name in ("US", "RU", "symbols"):
        section = data.get(section_name)
        if not isinstance(section, dict) or not section:
            raise ValueError(f"В алфавите нет секции {section_name}")
        for character, token in section.items():
            if not isinstance(character, str) or len(character) != 1:
                raise ValueError(f"Ключ {character!r} должен быть одним символом")
            if not isinstance(token, str) or not token:
                raise ValueError(f"Пустой токен у символа {character!r}")
            if any(symbol.isspace() for symbol in token):
                raise ValueError(f"Токен {token!r} не должен содержать пробелы")
            if character in forward:
                raise ValueError(f"Символ {character!r} встречается дважды")
            if token in reverse:
                raise ValueError(f"Токен {token!r} встречается дважды")
            forward[character] = token
            reverse[token] = character

    canonical = json.dumps(
        data, ensure_ascii=False, sort_keys=True, separators=(",", ":")
    ).encode("utf-8")
    alphabet_hash = _b64(hashlib.sha256(canonical).digest())
    return forward, reverse, alphabet_hash


def generate_keys(
    private_path: Path, public_path: Path, overwrite: bool = False
) -> str:
    """Generate and save one X25519 key pair."""
    if private_path.resolve() == public_path.resolve():
        raise ValueError("Пути private.json и public.json должны различаться")
    if not overwrite and (private_path.exists() or public_path.exists()):
        raise FileExistsError("Файл ключа уже существует")

    private_key = X25519PrivateKey.generate()
    private_raw = private_key.private_bytes(
        serialization.Encoding.Raw,
        serialization.PrivateFormat.Raw,
        serialization.NoEncryption(),
    )
    public_raw = private_key.public_key().public_bytes(
        serialization.Encoding.Raw, serialization.PublicFormat.Raw
    )
    key_id = _key_id(public_raw)

    private_data = {
        "version": KEY_VERSION,
        "algorithm": "X25519",
        "key_id": key_id,
        "private_key": _b64(private_raw),
        "public_key": _b64(public_raw),
    }
    public_data = {
        "version": KEY_VERSION,
        "algorithm": "X25519",
        "key_id": key_id,
        "public_key": _b64(public_raw),
    }
    _write_json(private_path, private_data, 0o600, overwrite)
    _write_json(public_path, public_data, 0o644, overwrite)
    return key_id


def _load_private(path: Path) -> tuple[X25519PrivateKey, str]:
    data = _read_json(path)
    if set(data) != {"version", "algorithm", "key_id", "private_key", "public_key"}:
        raise ValueError("Неподдерживаемый формат приватного ключа")
    if data.get("version") != KEY_VERSION or data.get("algorithm") != "X25519":
        raise ValueError("Неподдерживаемый формат приватного ключа")
    private_raw = _unb64(str(data.get("private_key", "")), "private_key")
    stored_public = _unb64(str(data.get("public_key", "")), "public_key")
    if len(private_raw) != 32 or len(stored_public) != 32:
        raise ValueError("X25519-ключ должен содержать 32 байта")
    private_key = X25519PrivateKey.from_private_bytes(private_raw)
    actual_public = private_key.public_key().public_bytes(
        serialization.Encoding.Raw, serialization.PublicFormat.Raw
    )
    actual_id = _key_id(actual_public)
    if not hmac.compare_digest(stored_public, actual_public) or not hmac.compare_digest(
        str(data.get("key_id", "")), actual_id
    ):
        raise ValueError("Приватный ключ повреждён")
    return private_key, actual_id


def _load_public(path: Path) -> tuple[X25519PublicKey, str]:
    data = _read_json(path)
    if set(data) != {"version", "algorithm", "key_id", "public_key"}:
        raise ValueError("Неподдерживаемый формат публичного ключа")
    if data.get("version") != KEY_VERSION or data.get("algorithm") != "X25519":
        raise ValueError("Неподдерживаемый формат публичного ключа")
    public_raw = _unb64(str(data.get("public_key", "")), "public_key")
    if len(public_raw) != 32:
        raise ValueError("X25519-ключ должен содержать 32 байта")
    actual_id = _key_id(public_raw)
    if not hmac.compare_digest(str(data.get("key_id", "")), actual_id):
        raise ValueError("Публичный ключ повреждён")
    return X25519PublicKey.from_public_bytes(public_raw), actual_id


def _session_keys(
    private_key: X25519PrivateKey,
    public_key: X25519PublicKey,
    salt: bytes,
    sender_id: str,
    recipient_id: str,
    algorithm: str = ALGORITHM,
    alphabet_hash: str = "",
    alphabet_hash2: str = "",
) -> tuple[bytes, bytes]:
    shared_secret = private_key.exchange(public_key)
    key_material = HKDF(
        algorithm=hashes.SHA256(),
        length=64,
        salt=salt,
        info=(
            f"{ENVELOPE_TYPE}/v{ENVELOPE_VERSION}\0{algorithm}\0"
            f"{sender_id}\0{recipient_id}\0{alphabet_hash}\0{alphabet_hash2}"
        ).encode(),
    ).derive(shared_secret)
    return key_material[:32], key_material[32:]


def _mask_block(key: bytes, nonce: bytes, position: int) -> bytes:
    return hmac.digest(
        key,
        b"alphabet-mask-v1\0" + nonce + position.to_bytes(8, "big"),
        "sha256",
    )


def _mask_stream(key: bytes, nonce: bytes, position: int, length: int) -> bytes:
    out = bytearray()
    counter = 0
    while len(out) < length:
        out.extend(
            hmac.digest(
                key,
                b"alphabet-mask-v2\0"
                + nonce
                + position.to_bytes(8, "big")
                + counter.to_bytes(4, "big"),
                "sha256",
            )
        )
        counter += 1
    return bytes(out[:length])


def _alphabet_order(forward: dict[str, str]) -> tuple[list[str], dict[str, int]]:
    characters = sorted(forward)
    tokens = [forward[character] for character in characters]
    return tokens, {character: index for index, character in enumerate(characters)}


def _load_dual_alphabets(
    path1: Path, path2: Path
) -> tuple[dict[str, str], dict[str, str], str, dict[str, str], dict[str, str], str]:
    forward1, _, hash1 = load_alphabet(path1)
    forward2, _, hash2 = load_alphabet(path2)
    if set(forward1) != set(forward2):
        raise ValueError("Алфавиты должны содержать одинаковые символы")
    if hash1 == hash2:
        raise ValueError("Алфавиты совпадают, нужны два разных файла")
    tokens1 = set(forward1.values())
    tokens2 = set(forward2.values())
    if tokens1 & tokens2:
        raise ValueError("Токены алфавитов пересекаются, они должны быть разными")
    reverse1 = {token: char for char, token in forward1.items()}
    reverse2 = {token: char for char, token in forward2.items()}
    return forward1, reverse1, hash1, forward2, reverse2, hash2


def _add_dual_noise_v1(
    text: str,
    forward1: dict[str, str],
    forward2: dict[str, str],
    mask_key: bytes,
    nonce: bytes,
) -> str:
    tokens1, indexes1 = _alphabet_order(forward1)
    tokens2, indexes2 = _alphabet_order(forward2)
    if sorted(forward1) != sorted(forward2):
        raise ValueError("Алфавиты должны содержать одинаковые символы")
    size = len(tokens1)
    if size != len(tokens2) or size == 0:
        raise ValueError("Алфавиты пустые или разного размера")
    combined = tokens1 + tokens2
    combined_size = len(combined)
    output: list[str] = []
    for position, character in enumerate(text):
        if character not in indexes1 or character not in indexes2:
            raise ValueError(
                f"Символ {character!r} в позиции {position + 1} отсутствует в алфавите"
            )
        block = _mask_block(mask_key, nonce, position)
        noise_count = 2 + block[0] % 3
        group_size = noise_count + 1
        real_position = block[1] % group_size
        shift = int.from_bytes(block[2:6], "big") % size
        real_flag = block[31] % 2
        if real_flag == 0:
            real_token = tokens1[(indexes1[character] + shift) % size]
        else:
            real_token = tokens2[(indexes2[character] + shift) % size]
        group = [
            combined[
                int.from_bytes(block[6 + index * 4 : 10 + index * 4], "big")
                % combined_size
            ]
            for index in range(noise_count)
        ]
        group.insert(real_position, real_token)
        output.extend(group)
    return " ".join(output)


def _add_dual_noise(
    text: str,
    forward1: dict[str, str],
    forward2: dict[str, str],
    mask_key: bytes,
    nonce: bytes,
) -> str:
    tokens1, indexes1 = _alphabet_order(forward1)
    tokens2, indexes2 = _alphabet_order(forward2)
    if sorted(forward1) != sorted(forward2):
        raise ValueError("Алфавиты должны содержать одинаковые символы")
    size = len(tokens1)
    if size != len(tokens2) or size == 0:
        raise ValueError("Алфавиты пустые или разного размера")
    combined = tokens1 + tokens2
    combined_size = len(combined)
    output: list[str] = []
    for position, character in enumerate(text):
        if character not in indexes1 or character not in indexes2:
            raise ValueError(
                f"Символ {character!r} в позиции {position + 1} отсутствует в алфавите"
            )
        stream = _mask_stream(mask_key, nonce, position, 80)
        noise_count = 3 + (stream[0] % 4)
        group_size = noise_count + 1
        real_position = int.from_bytes(stream[1:3], "big") % group_size
        shift1 = int.from_bytes(stream[4:8], "big") % size
        shift2 = int.from_bytes(stream[8:12], "big") % size
        real_flag = (stream[12] ^ stream[13] ^ stream[14] ^ (position & 0xFF)) & 1
        if real_flag == 0:
            real_token = tokens1[(indexes1[character] + shift1) % size]
        else:
            real_token = tokens2[(indexes2[character] + shift2) % size]
        group: list[str] = []
        used: set[str] = {real_token}
        for index in range(noise_count):
            token_index = (
                int.from_bytes(stream[16 + index * 4 : 20 + index * 4], "big")
                % combined_size
            )
            token = combined[token_index]
            probe = 0
            while token in used and probe < combined_size:
                token_index = (token_index + 1) % combined_size
                token = combined[token_index]
                probe += 1
            used.add(token)
            group.append(token)
        group.insert(real_position, real_token)
        output.extend(group)
    return " ".join(output)


def _remove_dual_noise_v1(
    ciphertext: str,
    forward1: dict[str, str],
    forward2: dict[str, str],
    mask_key: bytes,
    nonce: bytes,
) -> str:
    tokens1, _ = _alphabet_order(forward1)
    tokens2, _ = _alphabet_order(forward2)
    chars = sorted(forward1)
    if chars != sorted(forward2):
        raise ValueError("Алфавиты должны содержать одинаковые символы")
    size = len(tokens1)
    token_to_index1 = {token: index for index, token in enumerate(tokens1)}
    token_to_index2 = {token: index for index, token in enumerate(tokens2)}
    token_alpha: dict[str, int] = {}
    for token in tokens1:
        token_alpha[token] = 1
    for token in tokens2:
        if token in token_alpha:
            raise ValueError("Токены алфавитов пересекаются")
        token_alpha[token] = 2
    encrypted_tokens = ciphertext.split()
    plaintext: list[str] = []
    offset = 0
    position = 0
    while offset < len(encrypted_tokens):
        block = _mask_block(mask_key, nonce, position)
        noise_count = 2 + block[0] % 3
        group_size = noise_count + 1
        if offset + group_size > len(encrypted_tokens):
            raise ValueError("Повреждена структура зашифрованного текста")
        real_position = block[1] % group_size
        shift = int.from_bytes(block[2:6], "big") % size
        real_flag = block[31] % 2
        real_token = encrypted_tokens[offset + real_position]
        if real_token not in token_alpha:
            raise ValueError(f"Токен {real_token!r} отсутствует в алфавитах")
        if token_alpha[real_token] != (1 if real_flag == 0 else 2):
            raise ValueError("Повреждена структура: алфавит токена не совпал")
        if real_flag == 0:
            plaintext_index = (token_to_index1[real_token] - shift) % size
        else:
            plaintext_index = (token_to_index2[real_token] - shift) % size
        plaintext.append(chars[plaintext_index])
        offset += group_size
        position += 1
    return "".join(plaintext)


def _remove_dual_noise(
    ciphertext: str,
    forward1: dict[str, str],
    forward2: dict[str, str],
    mask_key: bytes,
    nonce: bytes,
) -> str:
    tokens1, _ = _alphabet_order(forward1)
    tokens2, _ = _alphabet_order(forward2)
    chars = sorted(forward1)
    if chars != sorted(forward2):
        raise ValueError("Алфавиты должны содержать одинаковые символы")
    size = len(tokens1)
    token_to_index1 = {token: index for index, token in enumerate(tokens1)}
    token_to_index2 = {token: index for index, token in enumerate(tokens2)}
    token_alpha: dict[str, int] = {}
    for token in tokens1:
        token_alpha[token] = 1
    for token in tokens2:
        if token in token_alpha:
            raise ValueError("Токены алфавитов пересекаются")
        token_alpha[token] = 2
    encrypted_tokens = ciphertext.split()
    plaintext: list[str] = []
    offset = 0
    position = 0
    while offset < len(encrypted_tokens):
        stream = _mask_stream(mask_key, nonce, position, 80)
        noise_count = 3 + (stream[0] % 4)
        group_size = noise_count + 1
        if offset + group_size > len(encrypted_tokens):
            raise ValueError("Повреждена структура зашифрованного текста")
        real_position = int.from_bytes(stream[1:3], "big") % group_size
        shift1 = int.from_bytes(stream[4:8], "big") % size
        shift2 = int.from_bytes(stream[8:12], "big") % size
        real_flag = (stream[12] ^ stream[13] ^ stream[14] ^ (position & 0xFF)) & 1
        real_token = encrypted_tokens[offset + real_position]
        if real_token not in token_alpha:
            raise ValueError(f"Токен {real_token!r} отсутствует в алфавитах")
        if token_alpha[real_token] != (1 if real_flag == 0 else 2):
            raise ValueError("Повреждена структура: алфавит токена не совпал")
        if real_flag == 0:
            plaintext_index = (token_to_index1[real_token] - shift1) % size
        else:
            plaintext_index = (token_to_index2[real_token] - shift2) % size
        plaintext.append(chars[plaintext_index])
        offset += group_size
        position += 1
    return "".join(plaintext)


def _add_keyed_noise(
    text: str, forward: dict[str, str], mask_key: bytes, nonce: bytes
) -> str:
    alphabet_tokens, character_indexes = _alphabet_order(forward)
    size = len(alphabet_tokens)
    output: list[str] = []
    for position, character in enumerate(text):
        if character not in character_indexes:
            raise ValueError(
                f"Символ {character!r} в позиции {position + 1} отсутствует в алфавите"
            )
        block = _mask_block(mask_key, nonce, position)
        noise_count = 1 + block[0] % 3
        real_position = block[1] % (noise_count + 1)
        shift = int.from_bytes(block[2:6], "big") % size
        encrypted_index = (character_indexes[character] + shift) % size

        group = [
            alphabet_tokens[
                int.from_bytes(block[6 + index * 4 : 10 + index * 4], "big")
                % size
            ]
            for index in range(noise_count)
        ]
        group.insert(real_position, alphabet_tokens[encrypted_index])
        output.extend(group)
    return " ".join(output)


def _remove_keyed_noise(
    ciphertext: str, forward: dict[str, str], mask_key: bytes, nonce: bytes
) -> str:
    alphabet_tokens, _ = _alphabet_order(forward)
    token_indexes = {token: index for index, token in enumerate(alphabet_tokens)}
    encrypted_tokens = ciphertext.split()
    plaintext: list[str] = []
    offset = 0
    position = 0

    while offset < len(encrypted_tokens):
        block = _mask_block(mask_key, nonce, position)
        noise_count = 1 + block[0] % 3
        group_size = noise_count + 1
        real_position = block[1] % group_size
        if offset + group_size > len(encrypted_tokens):
            raise ValueError("Повреждена структура зашифрованного текста")
        real_token = encrypted_tokens[offset + real_position]
        if real_token not in token_indexes:
            raise ValueError(f"Токен {real_token!r} отсутствует в encryption.json")
        shift = int.from_bytes(block[2:6], "big") % len(alphabet_tokens)
        plaintext_index = (token_indexes[real_token] - shift) % len(alphabet_tokens)
        plaintext.append(sorted(forward)[plaintext_index])
        offset += group_size
        position += 1

    return "".join(plaintext)


def _metadata(envelope: dict[str, Any]) -> bytes:
    keys = (
        "type",
        "version",
        "algorithm",
        "alphabet_hash",
        "alphabet_hash2",
        "sender_key_id",
        "recipient_key_id",
        "salt",
        "nonce",
    )
    data = {key: envelope.get(key) for key in keys}
    return json.dumps(data, sort_keys=True, separators=(",", ":")).encode()


def _validate_envelope_descriptor(envelope: dict[str, Any]) -> None:
    if set(envelope) != set(ENVELOPE_REQUIRED_KEYS):
        raise ValueError("Нарушен дескриптор сообщения: лишний или недостающий ключ")
    if envelope.get("type") != ENVELOPE_TYPE:
        raise ValueError("Нарушен дескриптор сообщения: неверный type")
    if envelope.get("version") != ENVELOPE_VERSION:
        raise ValueError(
            "Устаревшая или неподдерживаемая версия сообщения, расшифровка запрещена"
        )
    if envelope.get("algorithm") not in SUPPORTED_ALGORITHMS:
        raise ValueError("Алгоритм сообщения запрещён политикой, возможен даунгрейд")
    for field in ("alphabet_hash", "alphabet_hash2", "salt", "nonce", "auth_tag"):
        value = envelope.get(field)
        if not isinstance(value, str) or not value:
            raise ValueError(f"Нарушен дескриптор сообщения: поле {field}")
        if any(symbol.isspace() for symbol in value):
            raise ValueError(f"Нарушен дескриптор сообщения: поле {field}")
    for field in ("sender_key_id", "recipient_key_id"):
        value = envelope.get(field)
        if not isinstance(value, str) or len(value) != 32:
            raise ValueError("Нарушен дескриптор сообщения: неверный key_id")
    ciphertext = envelope.get("ciphertext")
    if not isinstance(ciphertext, str) or not ciphertext:
        raise ValueError("Нарушен дескриптор сообщения: поле ciphertext")


def _metadata_legacy(envelope: dict[str, Any]) -> bytes:
    if "alphabet_hash2" in envelope:
        keys = (
            "version",
            "algorithm",
            "alphabet_hash",
            "alphabet_hash2",
            "sender_key_id",
            "recipient_key_id",
            "salt",
            "nonce",
        )
    else:
        keys = (
            "version",
            "algorithm",
            "alphabet_hash",
            "sender_key_id",
            "recipient_key_id",
            "salt",
            "nonce",
        )
    data = {key: envelope.get(key) for key in keys}
    return json.dumps(data, sort_keys=True, separators=(",", ":")).encode()


def _session_keys_legacy(
    private_key: X25519PrivateKey,
    public_key: X25519PublicKey,
    salt: bytes,
    sender_id: str,
    recipient_id: str,
) -> tuple[bytes, bytes]:
    shared_secret = private_key.exchange(public_key)
    key_material = HKDF(
        algorithm=hashes.SHA256(),
        length=64,
        salt=salt,
        info=f"my-masengers/v1\0{sender_id}\0{recipient_id}".encode(),
    ).derive(shared_secret)
    return key_material[:32], key_material[32:]


def encrypt_text(
    text: str,
    alphabet_path: Path,
    private_path: Path,
    peer_public_path: Path,
    alphabet2_path: Path | None = None,
) -> dict[str, Any]:
    """Encode text with two system alphabets and encrypt it for a peer."""
    if alphabet2_path is None:
        raise ValueError("Нужны два системных алфавита: encryption.json и encryption2.json")
    if len(text.encode("utf-8")) > MAX_TEXT_BYTES:
        raise ValueError("Текст слишком большой")
    if not text:
        raise ValueError("Пустой текст шифровать запрещено")
    forward1, _, hash1, forward2, _, hash2 = _load_dual_alphabets(
        alphabet_path, alphabet2_path
    )
    private_key, sender_id = _load_private(private_path)
    public_key, recipient_id = _load_public(peer_public_path)
    salt = os.urandom(32)
    nonce = os.urandom(16)
    envelope: dict[str, Any] = {
        "type": ENVELOPE_TYPE,
        "version": ENVELOPE_VERSION,
        "algorithm": ALGORITHM,
        "alphabet_hash": hash1,
        "alphabet_hash2": hash2,
        "sender_key_id": sender_id,
        "recipient_key_id": recipient_id,
        "salt": _b64(salt),
        "nonce": _b64(nonce),
    }
    mask_key, mac_key = _session_keys(
        private_key, public_key, salt, sender_id, recipient_id, ALGORITHM, hash1, hash2
    )
    ciphertext = _add_dual_noise(text, forward1, forward2, mask_key, nonce)
    envelope["ciphertext"] = ciphertext
    envelope["auth_tag"] = _b64(
        hmac.digest(mac_key, _metadata(envelope) + b"\0" + ciphertext.encode(), "sha256")
    )
    _validate_envelope_descriptor(envelope)
    return envelope


def _decrypt_legacy_text(
    envelope: dict[str, Any],
    alphabet_path: Path,
    private_path: Path,
    peer_public_path: Path,
    alphabet2_path: Path | None,
) -> str:
    algorithm = envelope.get("algorithm")
    if algorithm == LEGACY_ALGORITHM:
        forward, _, alphabet_hash = load_alphabet(alphabet_path)
        if not hmac.compare_digest(
            str(envelope.get("alphabet_hash", "")), alphabet_hash
        ):
            raise ValueError("Сообщение создано с другим encryption.json")
    elif algorithm in (ALGORITHM_DUAL_V1, ALGORITHM_FULL_V1):
        if alphabet2_path is None:
            raise ValueError("Для этого сообщения нужны два алфавита (--alphabet2)")
        forward1, _, hash1, forward2, _, hash2 = _load_dual_alphabets(
            alphabet_path, alphabet2_path
        )
        if not hmac.compare_digest(str(envelope.get("alphabet_hash", "")), hash1):
            raise ValueError("Сообщение создано с другим encryption.json")
        if not hmac.compare_digest(str(envelope.get("alphabet_hash2", "")), hash2):
            raise ValueError("Сообщение создано с другим encryption2.json")
    else:
        raise ValueError("Неподдерживаемый формат зашифрованного сообщения")
    private_key, recipient_id = _load_private(private_path)
    public_key, sender_id = _load_public(peer_public_path)
    if envelope.get("recipient_key_id") != recipient_id:
        raise ValueError("Сообщение предназначено для другого приватного ключа")
    if envelope.get("sender_key_id") != sender_id:
        raise ValueError("Указан неверный публичный ключ отправителя")
    salt = _unb64(str(envelope.get("salt", "")), "salt")
    nonce = _unb64(str(envelope.get("nonce", "")), "nonce")
    ciphertext = envelope.get("ciphertext")
    if not isinstance(ciphertext, str):
        raise ValueError("Поле ciphertext должно быть строкой")
    auth_tag = _unb64(str(envelope.get("auth_tag", "")), "auth_tag")
    if len(salt) != 32 or len(nonce) != 16 or len(auth_tag) != 32:
        raise ValueError("Некорректный salt или nonce")
    mask_key, mac_key = _session_keys_legacy(
        private_key, public_key, salt, sender_id, recipient_id
    )
    expected_tag = hmac.digest(
        mac_key, _metadata_legacy(envelope) + b"\0" + ciphertext.encode(), "sha256"
    )
    if not hmac.compare_digest(auth_tag, expected_tag):
        raise ValueError("Неверные ключи или сообщение было изменено")
    if algorithm == LEGACY_ALGORITHM:
        return _remove_keyed_noise(ciphertext, forward, mask_key, nonce)
    if algorithm == ALGORITHM_DUAL_V1:
        return _remove_dual_noise_v1(ciphertext, forward1, forward2, mask_key, nonce)
    return _remove_dual_noise(ciphertext, forward1, forward2, mask_key, nonce)


def decrypt_text(
    envelope_path: Path,
    alphabet_path: Path,
    private_path: Path,
    peer_public_path: Path,
    alphabet2_path: Path | None = None,
    allow_legacy: bool = False,
) -> str:
    """Authenticate, decrypt and decode a saved envelope. Any tampering fails."""
    envelope = _read_json(envelope_path)
    version = envelope.get("version")
    algorithm = envelope.get("algorithm")
    if version == LEGACY_ENVELOPE_VERSION or algorithm in (
        LEGACY_ALGORITHM,
        ALGORITHM_DUAL_V1,
        ALGORITHM_FULL_V1,
    ):
        if not allow_legacy:
            raise ValueError(
                "Устаревшая версия сообщения запрещена политикой, возможен даунгрейд"
            )
        return _decrypt_legacy_text(
            envelope, alphabet_path, private_path, peer_public_path, alphabet2_path
        )
    _validate_envelope_descriptor(envelope)
    if alphabet2_path is None:
        raise ValueError("Для этого сообщения нужны два алфавита (--alphabet2)")
    forward1, _, hash1, forward2, _, hash2 = _load_dual_alphabets(
        alphabet_path, alphabet2_path
    )
    if not hmac.compare_digest(str(envelope.get("alphabet_hash", "")), hash1):
        raise ValueError("Сообщение создано с другим encryption.json")
    if not hmac.compare_digest(str(envelope.get("alphabet_hash2", "")), hash2):
        raise ValueError("Сообщение создано с другим encryption2.json")
    private_key, recipient_id = _load_private(private_path)
    public_key, sender_id = _load_public(peer_public_path)
    if not hmac.compare_digest(str(envelope.get("recipient_key_id")), recipient_id):
        raise ValueError("Сообщение предназначено для другого приватного ключа")
    if not hmac.compare_digest(str(envelope.get("sender_key_id")), sender_id):
        raise ValueError("Указан неверный публичный ключ отправителя")
    try:
        salt = _unb64(str(envelope.get("salt", "")), "salt")
        nonce = _unb64(str(envelope.get("nonce", "")), "nonce")
        auth_tag = _unb64(str(envelope.get("auth_tag", "")), "auth_tag")
    except ValueError:
        raise ValueError("Неверные ключи или сообщение было изменено")
    ciphertext = envelope.get("ciphertext")
    if len(salt) != 32 or len(nonce) != 16 or len(auth_tag) != 32:
        raise ValueError("Неверные ключи или сообщение было изменено")
    if len(ciphertext.encode("utf-8")) > MAX_TEXT_BYTES * 20:
        raise ValueError("Неверные ключи или сообщение было изменено")
    mask_key, mac_key = _session_keys(
        private_key,
        public_key,
        salt,
        sender_id,
        recipient_id,
        ALGORITHM,
        hash1,
        hash2,
    )
    expected_tag = hmac.digest(
        mac_key, _metadata(envelope) + b"\0" + ciphertext.encode(), "sha256"
    )
    if not hmac.compare_digest(auth_tag, expected_tag):
        raise ValueError("Неверные ключи или сообщение было изменено")
    return _remove_dual_noise(ciphertext, forward1, forward2, mask_key, nonce)


def _set_cursor(visible: bool) -> None:
    try:
        curses.curs_set(1 if visible else 0)
    except curses.error:
        pass


def _prompt(screen: curses.window, label: str, default: str = "") -> str:
    value = list(default)
    _set_cursor(True)
    while True:
        height, width = screen.getmaxyx()
        screen.move(height - 2, 0)
        screen.clrtoeol()
        shown = f"{label}: "
        screen.addnstr(height - 2, 0, shown + "".join(value), width - 1)
        screen.refresh()
        key = screen.get_wch()
        if key in ("\n", "\r", curses.KEY_ENTER):
            _set_cursor(False)
            answer = "".join(value)
            return answer if answer else default
        if key == "\x1b":
            _set_cursor(False)
            raise KeyboardInterrupt
        if key in ("\b", "\x7f", curses.KEY_BACKSPACE):
            if value:
                value.pop()
        elif isinstance(key, str) and key.isprintable():
            value.append(key)


def _message(screen: curses.window, title: str, body: str) -> None:
    screen.clear()
    height, width = screen.getmaxyx()
    screen.addnstr(1, 2, title, width - 4, curses.A_BOLD)
    lines: list[str] = []
    for source_line in body.splitlines() or [""]:
        while len(source_line) > max(1, width - 4):
            lines.append(source_line[: width - 4])
            source_line = source_line[width - 4 :]
        lines.append(source_line)
    for row, line in enumerate(lines[: height - 5], start=3):
        screen.addnstr(row, 2, line, width - 4)
    screen.addnstr(height - 2, 2, "Нажми любую клавишу", width - 4, curses.A_DIM)
    screen.refresh()
    screen.get_wch()


def _confirm_overwrite(screen: curses.window, paths: tuple[Path, ...]) -> bool:
    existing = [str(path) for path in paths if path.exists()]
    if not existing:
        return False
    answer = _prompt(screen, f"Перезаписать {', '.join(existing)}? y/N", "N")
    if answer.lower() not in {"y", "yes", "д", "да"}:
        raise KeyboardInterrupt
    return True


def _generate_action(screen: curses.window) -> None:
    private_path = Path(_prompt(screen, "Приватный ключ", "private.json"))
    public_path = Path(_prompt(screen, "Публичный ключ", "public.json"))
    overwrite = _confirm_overwrite(screen, (private_path, public_path))
    key_id = generate_keys(private_path, public_path, overwrite)
    _message(screen, "Ключи созданы", f"ID: {key_id}\n{private_path}\n{public_path}")


def _encrypt_action(
    screen: curses.window, alphabet_path: Path, alphabet2_path: Path
) -> None:
    private_path = Path(_prompt(screen, "Твой приватный ключ", "private.json"))
    peer_path = Path(_prompt(screen, "Публичный ключ получателя", "public.json"))
    text = _prompt(screen, "Исходный текст")
    output_path = Path(_prompt(screen, "Файл результата", "message.json"))
    overwrite = _confirm_overwrite(screen, (output_path,))
    envelope = encrypt_text(text, alphabet_path, private_path, peer_path, alphabet2_path)
    _write_json(output_path, envelope, 0o600, overwrite)
    _message(screen, "Готово", f"Зашифрованное сообщение: {output_path}")


def _decrypt_action(
    screen: curses.window, alphabet_path: Path, alphabet2_path: Path, allow_legacy: bool
) -> None:
    private_path = Path(_prompt(screen, "Твой приватный ключ", "private.json"))
    peer_path = Path(_prompt(screen, "Публичный ключ отправителя", "public.json"))
    message_path = Path(_prompt(screen, "Зашифрованное сообщение", "message.json"))
    plaintext = decrypt_text(
        message_path,
        alphabet_path,
        private_path,
        peer_path,
        alphabet2_path,
        allow_legacy,
    )
    _message(screen, "Расшифрованный текст", plaintext)


def run_tui(
    screen: curses.window,
    alphabet_path: Path,
    alphabet2_path: Path,
    allow_legacy: bool = False,
) -> None:
    """Run the full-screen terminal menu."""
    _set_cursor(False)
    screen.keypad(True)
    cursor = 0
    items = (
        ("Создать пару ключей", _generate_action),
        (
            "Зашифровать текст [2 алфавита]",
            lambda window: _encrypt_action(window, alphabet_path, alphabet2_path),
        ),
        (
            "Расшифровать текст [2 алфавита]",
            lambda window: _decrypt_action(
                window, alphabet_path, alphabet2_path, allow_legacy
            ),
        ),
    )
    while True:
        screen.clear()
        height, width = screen.getmaxyx()
        screen.addnstr(1, 2, "MY MASENGERS · CRYPTO TUI", width - 4, curses.A_BOLD)
        screen.addnstr(
            3, 2, f"Алфавит 1: {alphabet_path} | Алфавит 2: {alphabet2_path}", width - 4, curses.A_DIM
        )
        for index, (title, _) in enumerate(items):
            marker = "▶ " if index == cursor else "  "
            style = curses.A_REVERSE if index == cursor else curses.A_NORMAL
            screen.addnstr(5 + index, 4, marker + title, width - 8, style)
        screen.addnstr(height - 2, 2, "↑/↓ — выбор · Enter — открыть · q — выход", width - 4)
        screen.refresh()
        key = screen.get_wch()
        if key in ("q", "Q", "й", "Й", "\x1b"):
            return
        if key in (curses.KEY_UP, "k", "w", "ц"):
            cursor = (cursor - 1) % len(items)
        elif key in (curses.KEY_DOWN, "j", "s", "ы"):
            cursor = (cursor + 1) % len(items)
        elif key in ("\n", "\r", curses.KEY_ENTER):
            try:
                items[cursor][1](screen)
            except KeyboardInterrupt:
                continue
            except (OSError, ValueError, json.JSONDecodeError) as error:
                _message(screen, "Ошибка", str(error))


def main() -> int:
    """Parse arguments and launch the TUI."""
    parser = argparse.ArgumentParser(description="TUI-прототип шифрования сообщений")
    parser.add_argument(
        "--alphabet", type=Path, default=Path("encryption.json"), help="путь к encryption.json"
    )
    parser.add_argument(
        "--alphabet2",
        type=Path,
        default=Path("encryption2.json"),
        help="путь ко второму системному алфавиту encryption2.json",
    )
    parser.add_argument(
        "--allow-legacy",
        action="store_true",
        help="разрешить расшифровку устаревших v1-сообщений (по умолчанию запрещено)",
    )
    parser.add_argument("--version", action="version", version="crypto-tui 3.0.0")
    args = parser.parse_args()
    if not sys.stdin.isatty() or not sys.stdout.isatty():
        parser.error("TUI нужно запускать в интерактивном терминале")
    try:
        _load_dual_alphabets(args.alphabet, args.alphabet2)
        curses.wrapper(run_tui, args.alphabet, args.alphabet2, args.allow_legacy)
    except KeyboardInterrupt:
        return 130
    except (OSError, ValueError, curses.error) as error:
        print(f"Ошибка: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
