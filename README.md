# Test-K8s-Pet-Project

**English** | [Русский](README.ru.md)

> ⚠️ **Before you run this project, regenerate the JWT keys.**
> The keys in `.env` are public mock values and are not safe to use anywhere real.
> See [docs/jwt-keys.md](docs/jwt-keys.md).

## Services

| Directory   | Description                                   |
|-------------|-----------------------------------------------|
| `auth/`     | Auth microservice (Python), issues JWT tokens |
| `backend/`  | Backend service (Go)                          |
| `gateway/`  | API gateway (Rust): routing, middleware, metrics |
| `frontend/` | Web client (React + Vite + TypeScript)        |
| `verify/`   | Email verification service (Go, RabbitMQ)     |
| `messenger/` | Messenger relay: public keys and opaque encrypted envelopes (Go) |
| `todo/`     | Todo canvas: boards, nodes, tasks, markdown (Go, Redis) |
| `migrations/` | Database schema migrations (Go)             |
| `Infra/`    | Infrastructure configs (PostgreSQL)           |
| `k8s/`      | Kubernetes manifests                          |

## Configuration

`.env` is committed on purpose and contains **mock values only** for local development. Never put real secrets into it; keep them in an uncommitted file such as `.env.prod`.
Never put real secrets into it.

## Docs

- [JWT keys](docs/jwt-keys.md)
- [Kubernetes: configuration and secrets](docs/k8s-secrets.md)
- [Email verification service](docs/verify-service.md)
