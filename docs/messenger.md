# Messenger (MVP)

End-to-end encrypted chat. The server is a "dumb courier": it stores public keys and opaque encrypted envelopes and never sees plaintext.

## Pieces

| Part | Where | Role |
|------|-------|------|
| Courier service | `messenger/` (Go) | public key directory, encrypted key backup blob, envelope delivery, conversation list, unread counters |
| Crypto | `frontend/src/lib/messenger/crypto/` (TypeScript) | the scheme below, in the browser |
| Reference | `mm/crypto_tui.py` (Python) | prototype the TypeScript code is tested against |
| UI | `frontend/src/pages/messenger/` | chats, key setup, dictionaries |

## Scheme (as in the prototype)

1. Every user has a long-term X25519 key pair. The public key (with `key_id = sha256(public)[:32]`) is registered on the server.
2. For each message a random salt (32 B) and nonce (16 B) are drawn. `HKDF-SHA256(X25519(private, peer_public), salt, info)` gives 64 bytes: a mask key and a MAC key. `info` binds the envelope type and version, algorithm, sender and recipient key ids and the hashes of both dictionaries.
3. Each character becomes a group of 4-7 tokens: one real token (the character index shifted by a value from an HMAC-SHA256 keystream, taken from dictionary 1 or 2) hidden among noise tokens. Dictionaries are open; the keyed shift is what protects the text.
4. The envelope carries `type, version, algorithm, alphabet_hash, alphabet_hash2, sender_key_id, recipient_key_id, salt, nonce, ciphertext, auth_tag`. `auth_tag` is HMAC-SHA256 over the metadata and ciphertext (encrypt-then-MAC), checked before any decoding.

Because the key schedule is symmetric, the sender can also re-read their own sent messages (`role: 'sender'`).

## Courier API (`/main/api/v1/messenger/...`, all require a JWT)

| Method | Path | Purpose |
|--------|------|---------|
| PUT | `put/my/key` | register or rotate the public key document |
| GET | `get/my/key` | own key document and the encrypted backup blob |
| PUT | `put/my/key/backup` | store an opaque backup (PBKDF2 + AES-GCM, made in the browser) |
| GET | `get/user/key/:uuid` | public key of another user |
| GET | `search/users?q=` | find users by username prefix |
| POST | `send/message` | store an envelope for a recipient (shape and key ids are checked, content is not readable) |
| GET | `get/my/messages?peer=&after=&limit=` | a thread; `after` is an id cursor for polling |
| GET | `get/my/conversations` | chats with unread counters |
| PATCH | `mark/my/messages/read` | mark a peer's messages as read |
| DELETE | `delete/my/message/:id` | the sender deletes a message for everyone |

## Known limits of the MVP

- The private key is kept in the browser's `localStorage`. XSS, extensions or another person with access to the profile can read it. Native apps should use the OS key store.
- No forward secrecy (static keys), no replay protection, no safety-number verification yet.
- Messages grow about 40x because of the noise layer; the envelope limit is 1 MB.
- Only characters present in the dictionaries can be sent.
- Delivery is by polling (3 s in a thread, 5 s for the list). A WebSocket hub is planned together with the todo module.
- The scheme is custom and has not been reviewed by a cryptographer.

## Tests

- `cd messenger && go test ./test/...` (needs Postgres; skips otherwise)
- `cd frontend && npm test` (Node 22+, no dependencies): byte-for-byte checks against vectors made by the Python prototype
- `python frontend/test/messenger/make-vectors.py` regenerates `vectors.json` (test keys only, derived from fixed bytes)
