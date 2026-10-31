# Kubernetes: configuration and secrets

**English** | [Русский](k8s-secrets.ru.md)

The frontend is **not** part of the cluster (it is hosted on Vercel). The cluster runs `gateway`, `auth` and `backend`. **PostgreSQL runs outside the cluster** (docker compose on the host, `192.168.0.15:56799`).

## Layout

```
k8s/
├── kustomization.yaml      # entry point: kubectl apply -k k8s/
├── namespace.yaml          # low-code-bot-platform
├── configmap.yaml          # non-secret settings (TTLs, algorithm, Argon2, ...)
├── serviceaccounts.yaml    # one per workload, API token not mounted
├── networkpolicies.yaml    # default deny + gateway -> auth/backend
├── database/               # Service + EndpointSlice -> external PostgreSQL on the host
├── auth/ backend/ gateway/
├── migrations/             # Job: `migrations up` (AutoMigrate + seed), runs as an Argo CD Sync hook
└── secrets/
    ├── secrets.example.yaml  # template with placeholders (safe to commit)
    ├── create-secrets.sh     # builds real secrets from an env file
    └── sealed/               # optional: SealedSecrets (safe to commit)
```

## What is a secret and who gets it

| Secret              | Keys                                                  | Consumers                  |
|---------------------|-------------------------------------------------------|----------------------------|
| `database-url`      | `DATABASE_URL`                                        | auth, backend              |
| `migration-db-url`  | `MIGRATION_DB_URL`                                    | migrations Job             |
| `jwt-private-key`   | `JWT_SECRET_KEY`                                      | **auth only**              |
| `jwt-public-key`    | `JWT_PUBLIC_KEY`                                      | auth, backend, gateway     |

Everything else (`JWT_ALGORITHM`, TTLs, Argon2 parameters, ...) is not secret and lives in the `app-config` ConfigMap.
The private key is split into its own secret on purpose: only `auth` mounts it.

## Deploying

1. **Regenerate the JWT keys** — see [jwt-keys.md](jwt-keys.md). Put real values in a local, uncommitted env file (`.env.prod`).
2. Create the secrets:

   ```bash
   k8s/secrets/create-secrets.sh --env-file .env.prod
   ```

   For a throwaway/learning cluster you can use the committed mock `.env` as is (`--env-file .env`). The script does not check for mock keys — never do that on anything real.
3. Set your image names in `k8s/kustomization.yaml` (`DOCKERHUB_USERNAME`) and apply:

   ```bash
   kubectl apply -k k8s/
   ```

## Production: don't keep secrets in plain `kubectl apply`

Plain Kubernetes Secrets are only base64 and are readable by anyone with `get secret` rights. Recommended options:

- **Sealed Secrets** (works with the script): `k8s/secrets/create-secrets.sh --env-file .env.prod --seal`
  writes encrypted manifests to `k8s/secrets/sealed/`. They can be committed and applied with `kubectl apply -f k8s/secrets/sealed/`.
  Requires the Sealed Secrets controller and the `kubeseal` CLI.
- **External Secrets Operator** + Vault / cloud secret manager, if you already have one.
- Enable **encryption at rest** for Secrets in the cluster and restrict RBAC on `secrets`.

## Rotating JWT keys

1. Generate a new pair ([jwt-keys.md](jwt-keys.md)) and update your env file.
2. Re-run `create-secrets.sh`.
3. Restart the workloads: `kubectl -n low-code-bot-platform rollout restart deploy/auth deploy/backend deploy/gateway`.
4. All issued tokens become invalid; users must log in again.

## Argo CD

Argo CD excludes `Endpoints`/`EndpointSlice` by default, which silently drops `k8s/database/service.yaml`.
Apply `argocd/resource-exclusions-patch.yaml` once (the command is in the file), then create the app with `kubectl apply -f argocd/application.yaml`.

## Ingress

Ingress is handled by the official `ingress-nginx` Helm chart, installed as a separate Argo CD app:

```bash
kubectl apply -f argocd/ingress-nginx.yaml
```

It listens on NodePort `30080` (HTTP) and `30444` (HTTPS) on every node. The settings from `Infra/nginx/nginx.conf` are ported into the controller ConfigMap in that file
(`proxy-read-timeout` 365s, `proxy-buffering` off, 50M body, keep-alive, gzip, JSON access log with `request_id`, `server-tokens` off). `k8s/gateway/ingress.yaml` routes `/` to the gateway.
WebSocket upgrade, `X-Real-IP`, `X-Forwarded-For` and `X-Request-ID` are handled by ingress-nginx by default. The `limit_req` zones and `proxy_cache_path` in `nginx.conf` were declared but never used in a `location`, so they were not ported (the gateway has its own rate limiting).

Apply the Ingress app **after** the main app has synced the gateway Service as `ClusterIP`, because both want port `30080`.

> ⚠️ `kubernetes/ingress-nginx` was archived in March 2026 and no longer receives fixes. Fine for a test cluster; plan a move to another controller for anything real.

## Migrations

`auth` and `backend` wait for the DB tables to exist; the tables are created by the `migrations` Job (`migrations up`).
It runs on every sync (idempotent). Run other commands (`status`, `down`, `seed`) manually with `kubectl run` and the migrations image.

## Notes and assumptions

- The gateway `routes.yaml` points to `Low-Code-Bot-Platform-*` service hostnames; the Services are named in lowercase
  (`low-code-bot-platform-auth-service`, `...-users-backend`) — DNS is case-insensitive, so the gateway resolves them.
- `bot` service referenced by the gateway has no manifests yet (no source in the repo).
- The gateway Service is `ClusterIP`; it is reached through ingress-nginx (see Ingress), e.g. `http://192.168.0.31:30080/main/api/v1/`.
- `networkpolicies.yaml` is enforced: the cluster runs Calico (VXLAN, pod CIDR `10.244.0.0/16`). Ingress is default-deny; the gateway accepts traffic from anywhere (needed for NodePort), `auth`/`backend` only from the gateway.
- Replicas are set to 1 and memory limits are low because the workers have ~4 GB RAM.
- `DATABASE_URL` keeps the compose-style host `low-code-bot-platform-postgresql:5432`; inside the cluster it resolves via `k8s/database/service.yaml` to the host DB. If the host IP or `POSTGRES_PORT` changes, edit that file. The DB container must be running and publish its port on the LAN.
- CI (`.github/workflows/docker-publish.yml`) builds and pushes `backend` (`test-k8s-api`), `auth` and `gateway` to Docker Hub. Needs the `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` repo secrets.
