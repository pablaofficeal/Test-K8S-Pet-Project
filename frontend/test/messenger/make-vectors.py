"""Генерирует тестовые векторы эталонным Python-прототипом (mm/crypto_tui.py).

Ключи, соль и nonce фиксированные и только тестовые: по ним TypeScript-реализация обязана выдать
побайтно такие же конверты. Запуск из корня репозитория:
    python frontend/test/messenger/make-vectors.py
"""
import json
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "mm"))

import crypto_tui as c  # noqa: E402
from cryptography.hazmat.primitives.asymmetric.x25519 import X25519PrivateKey  # noqa: E402
from cryptography.hazmat.primitives import serialization  # noqa: E402

DICTIONARIES = ROOT / "frontend" / "src" / "lib" / "messenger" / "dictionaries"
ALPHABET1 = DICTIONARIES / "encryption.json"
ALPHABET2 = DICTIONARIES / "encryption2.json"


def key_docs(seed: bytes):
    private = X25519PrivateKey.from_private_bytes(seed)
    private_raw = private.private_bytes(serialization.Encoding.Raw, serialization.PrivateFormat.Raw, serialization.NoEncryption())
    public_raw = private.public_key().public_bytes(serialization.Encoding.Raw, serialization.PublicFormat.Raw)
    public_doc = {"version": 1, "algorithm": "X25519", "key_id": c._key_id(public_raw), "public_key": c._b64(public_raw)}
    return {**public_doc, "private_key": c._b64(private_raw)}, public_doc


alice_private, alice_public = key_docs(bytes(range(1, 33)))
bob_private, bob_public = key_docs(bytes(range(101, 133)))

TEXTS = [
    ("single character", "a"),
    ("ascii sentence", "hello world"),
    ("cyrillic and punctuation", "Привет, Cadence! Это тест 123."),
    ("multi-line with quotes and symbols", 'line one\nline "two" \\ back\n{}[]<>|~^%$#@&*()-_=+/:;,.?!\'`'),
    ("repeated characters", "a" * 40 + "я" * 40),
    ("long text", ("Съешь ещё этих мягких французских булок, да выпей чаю. The quick brown fox jumps over the lazy dog. " * 3)[:300]),
]

cases = []
with tempfile.TemporaryDirectory() as tmp:
    tmp = Path(tmp)
    for who, (private_doc, peer_doc) in {"alice_to_bob": (alice_private, bob_public), "bob_to_alice": (bob_private, alice_public)}.items():
        (tmp / "priv.json").write_text(json.dumps(private_doc))
        (tmp / "peer.json").write_text(json.dumps(peer_doc))
        for index, (name, text) in enumerate(TEXTS):
            salt = bytes((index * 7 + j) % 256 for j in range(32))
            nonce = bytes((index * 11 + 3 * j + 1) % 256 for j in range(16))
            queue = [salt, nonce]
            original = c.os.urandom
            c.os.urandom = lambda n, queue=queue: queue.pop(0)
            try:
                envelope = c.encrypt_text(text, ALPHABET1, tmp / "priv.json", tmp / "peer.json", ALPHABET2)
            finally:
                c.os.urandom = original
            cases.append({"name": f"{who}: {name}", "direction": who, "text": text, "salt": c._b64(salt), "nonce": c._b64(nonce), "envelope": envelope})

_, _, hash1 = c.load_alphabet(ALPHABET1)
_, _, hash2 = c.load_alphabet(ALPHABET2)
out = {
    "note": "TEST VECTORS ONLY. These keys are derived from fixed bytes and must never be used for real messages.",
    "alphabet_hash": hash1,
    "alphabet_hash2": hash2,
    "alice": {"private": alice_private, "public": alice_public},
    "bob": {"private": bob_private, "public": bob_public},
    "cases": cases,
}
target = Path(__file__).with_name("vectors.json")
target.write_text(json.dumps(out, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(f"{len(cases)} vectors written to {target}")
