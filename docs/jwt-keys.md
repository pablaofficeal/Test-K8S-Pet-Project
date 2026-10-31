# JWT Keys

**English** | [Русский](jwt-keys.ru.md)

> ⚠️ **You MUST regenerate the JWT keys before using this project.**
>
> The keys in the repository's `.env` are **mock values**. They are public and exist only so the project
> runs out of the box in a dev environment. Anyone can mint a valid token for your service if you leave
> them as they are. **Never use them in production or on any publicly reachable environment.**

## How it works

Tokens are signed with **EdDSA (Ed25519)**:

| Variable         | What it is                      | Used by                                           |
|------------------|---------------------------------|---------------------------------------------------|
| `JWT_SECRET_KEY` | Private key (PKCS#8, PEM)       | `auth` — signs access/refresh tokens              |
| `JWT_PUBLIC_KEY` | Public key (PEM)                | `gateway`, `backend`, `auth` — verify signatures  |
| `JWT_ALGORITHM`  | Always `EdDSA`                  | all services                                      |

Only the `auth` service needs the private key. The others only need the public one.

## Regenerating

Requires `openssl` 1.1.1+.

```bash
# 1. Private key
openssl genpkey -algorithm ed25519 -out priv.pem

# 2. Public key derived from the private one
openssl pkey -in priv.pem -pubout -out pub.pem
```

`priv.pem` and `pub.pem` are already in `.gitignore` — do not commit them, and delete them once copied into your env file.

## Writing to the env file

In the env file the PEM is stored **on a single line**, with newlines replaced by `\n`, wrapped in double quotes:

```bash
echo "JWT_SECRET_KEY=\"$(awk 'NF {printf "%s\\n", $0}' priv.pem)\""
echo "JWT_PUBLIC_KEY=\"$(awk 'NF {printf "%s\\n", $0}' pub.pem)\""
```

Copy the output into your **own, uncommitted** env file (e.g. `.env.prod`), replacing the old values:

```env
JWT_SECRET_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
JWT_ALGORITHM=EdDSA
```

Then remove the temporary files:

```bash
rm priv.pem pub.pem
```

## Verifying the pair

```bash
echo test > msg
openssl pkeyutl -sign -inkey priv.pem -rawin -in msg -out sig
openssl pkeyutl -verify -pubin -inkey pub.pem -rawin -in msg -sigfile sig
# Signature Verified Successfully
```

## After rotating keys

- Restart all services (`auth`, `gateway`, `backend`) so they pick up the new keys.
- All previously issued access and refresh tokens become invalid; users will have to log in again.
- In production keep `JWT_SECRET_KEY` in a secret store (Kubernetes Secret, Vault, etc.), not in the repository.
  The values in the committed `.env` are for local development only.
- If the private key leaks, regenerate the pair immediately.

## About `.env` in the repository

`.env` is intentionally committed: it contains only mock values for local startup. Keep real secrets in a separate file such as `.env.prod` (already in `.gitignore`).
Never commit real secrets to it.
