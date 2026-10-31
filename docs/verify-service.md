# Email verification service

**English** | [Русский](verify-service.ru.md)

`verify/` is a Go microservice that confirms a user's email with a 6-digit code. Transport is RabbitMQ.

> ⚠️ **The test provider is active.** Until a real email provider exists, every account uses the fixed code
> **`893658`** (`VERIFY_PROVIDER=test`). Nothing is sent, and anyone who knows the code can "confirm" any address.
> Do not treat verified addresses as real until a real provider replaces it, and do not rely on verification for access control yet.

## Flow

```
register ─► auth ──(verify.events / verification.requested)──► RabbitMQ ──► verify ──► provider.Send()
                                                                              │
browser ──► gateway ──► verify  POST /verify/confirm {"code": "893 658"} ─────┘ sets users.email_verified
```

1. `auth` creates the user and publishes `{"event":"verification.requested","user_uuid":...,"trace_id":...}` to the `verify.events` exchange.
   A broker outage never fails the registration (the user can press "send again").
2. `verify` consumes `verify.requests`, creates a code (stored only as an HMAC hash), and hands it to the provider.
3. The user submits the code on `/verify`. After success `users.email_verified = true`.
4. Changing the email in settings resets the verification (`email_verified = false`); a code issued for the old address never verifies the new one.

## HTTP API (through the gateway, JWT required)

| Method | Path | Result |
|--------|------|--------|
| `GET`  | `/main/api/v1/verify/status`  | `{verified, email, resend_in, expires_in}` |
| `POST` | `/main/api/v1/verify/resend`  | `202` new code issued, `409` already verified, `429` cooldown / hourly limit |
| `POST` | `/main/api/v1/verify/confirm` | body `{"code":"893658"}` (spaces and dashes are ignored): `200`, `400` wrong code (`attempts_left`) or no active code, `429` too many attempts |

## Limits (all configurable)

| Variable | Default | Meaning |
|----------|---------|---------|
| `VERIFY_CODE_TTL` | `10m` | code lifetime |
| `VERIFY_RESEND_COOLDOWN` | `60s` | minimum time between codes |
| `VERIFY_MAX_ATTEMPTS` | `5` | wrong attempts per code, then the code is locked (even the right one is rejected) |
| `VERIFY_MAX_SENDS_PER_HOUR` | `5` | codes per user per hour |

The gateway adds an additional per-IP rate limit for `/verify/*`.

## Providers

`providers.Provider` is a small interface (`Name()` and `Send(ctx, Message)`). To add a real one (SMTP, an email API, ...):

1. Implement it in `verify/providers/`, register it in `providers/provider-new.go`.
2. Put its credentials into a Secret, **never into the repository** (the repository is public).
3. Set `VERIFY_PROVIDER` to its name. Codes become random 6 digits automatically, and `VERIFY_TEST_CODE` is ignored.

The service refuses to start with `VERIFY_PROVIDER=test` unless `VERIFY_ALLOW_TEST_CODE=true`, and logs a warning on every start.

## RabbitMQ roles

RabbitMQ runs on the host (`docker compose up -d rabbitmq`) and is declared in `Infra/rabbitmq/definitions.json`.
The vhost is `cadence`, exchanges and queues are pre-declared, so applications need no `configure` permission.

| User | configure | write | read | Used by |
|------|-----------|-------|------|---------|
| `admin` | `.*` | `.*` | `.*` | humans (management UI on `127.0.0.1:${RABBITMQ_MANAGEMENT_PORT}`) |
| `auth` | – | `^verify\.events$` | – | `auth` publishes only |
| `verify` | – | – | `^verify\.requests$` | `verify` consumes only |

`verify.requests` is a quorum queue with `x-delivery-limit: 5`; poison messages and messages that fail 5 times go to the
dead letter queue `verify.requests.dead` (via `verify.dead`).

The passwords in `.env` and `definitions.json` are public **mock** values. Generate real ones with
`Infra/rabbitmq/hash-password.sh 'password'` (prints a hash for `definitions.json`) and keep the real URLs in `.env.prod`.

## Running

- Local: `docker compose up -d postgres rabbitmq`, run migrations (`migrations up`), then `go run .` in `verify/` with
  `DATABASE_URL`, `JWT_PUBLIC_KEY`, `RABBITMQ_VERIFY_URL` and the `VERIFY_*` variables.
- Cluster: `k8s/verify/` (Deployment, Service, ConfigMap), `k8s/rabbitmq/` (Service + EndpointSlice to the host broker).
  Secrets `rabbitmq-auth-url`, `rabbitmq-verify-url` and `verify-pepper` are created by `k8s/secrets/create-secrets.sh`.

## Not done yet

- Nothing is blocked for unverified users. Enforcement would need an `email_verified` claim in the JWT checked by the gateway.
- Real email provider and email templates.

## Troubleshooting

| Symptom | Cause / fix |
|---------|-------------|
| Pod `CreateContainerConfigError`, `secret "rabbitmq-verify-url" not found` | Secrets are not in git. Run `k8s/secrets/create-secrets.sh --env-file .env`. |
| Log: `tables not ready` forever | The `migrations` Job has not run. It runs in sync wave `-1`, before the Deployments, on every Argo CD sync. |
| Log: `rabbitmq consumer stopped ... connection refused` | The AMQP port is not published on the host. `docker-compose.yml` must contain `"${RABBITMQ_PORT}:5672"` for `rabbitmq` (the management port alone is not enough). |
| Registration works but no code arrives, auth log says `RABBITMQ_AUTH_URL is empty` | `auth` started before the `rabbitmq-auth-url` secret existed (it is optional). `kubectl -n low-code-bot-platform rollout restart deploy/auth`. |

