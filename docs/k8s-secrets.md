# Kubernetes: configuration and secrets

**English** | [Русский](k8s-secrets.ru.md)

The frontend is **not** part of the cluster (it is hosted on Vercel). The cluster runs `gateway`, `auth`, `backend` and PostgreSQL.

## Layout

```
k8s/
├── kustomization.yaml      # entry point: kubectl apply -k k8s/
├── namespace.yaml          # low-code-bot-platform
├── configmap.yaml          # non-secret settings (TTLs, algorithm, Argon2, ...)
├── serviceaccounts.yaml    # one per workload, API token not mounted
├── networkpolicies.yaml    # default deny + gateway -> auth/backend -> postgres
├── postgres/ auth/ backend/ gateway/
└── secrets/
    ├── secrets.example.yaml  # template with placeholders (safe to commit)
    ├── create-secrets.sh     # builds real secrets from an env file
    └── sealed/               # optional: SealedSecrets (safe to commit)
```

## What is a secret and who gets it

| Secret              | Keys                                                  | Consumers                  |
|---------------------|-------------------------------------------------------|----------------------------|
| `postgres-credentials` | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | postgres                   |
| `database-url`      | `DATABASE_URL`                                        | auth, backend              |
| `jwt-private-key`   | `JWT_SECRET_KEY`                                      | **auth only**              |
| `jwt-public-key`    | `JWT_PUBLIC_KEY`                                      | auth, backend, gateway     |

Everything else (`JWT_ALGORITHM`, TTLs, Argon2 parameters, ...) is not secret and lives in the `app-config` ConfigMap.
The private key is split into its own secret on purpose: only `auth` mounts it.

## Deploying

1. **Regenerate the JWT keys** — see [jwt-keys.md](jwt-keys.md). Put real values in a local env file (e.g. `.env`, never committed with real secrets).
2. Create the secrets:

   ```bash
   k8s/secrets/create-secrets.sh --env-file .env
   ```

   The script refuses to run if `JWT_SECRET_KEY` equals the public mock key from `.env.example`
   (`--allow-mock` overrides this, only for throwaway clusters).
3. Set your image names in `k8s/kustomization.yaml` (`DOCKERHUB_USERNAME`) and apply:

   ```bash
   kubectl apply -k k8s/
   ```

## Production: don't keep secrets in plain `kubectl apply`

Plain Kubernetes Secrets are only base64 and are readable by anyone with `get secret` rights. Recommended options:

- **Sealed Secrets** (works with the script): `k8s/secrets/create-secrets.sh --env-file .env --seal`
  writes encrypted manifests to `k8s/secrets/sealed/`. They can be committed and applied with `kubectl apply -f k8s/secrets/sealed/`.
  Requires the Sealed Secrets controller and the `kubeseal` CLI.
- **External Secrets Operator** + Vault / cloud secret manager, if you already have one.
- Enable **encryption at rest** for Secrets in the cluster and restrict RBAC on `secrets`.

## Rotating JWT keys

1. Generate a new pair ([jwt-keys.md](jwt-keys.md)) and update your env file.
2. Re-run `create-secrets.sh`.
3. Restart the workloads: `kubectl -n low-code-bot-platform rollout restart deploy/auth deploy/backend deploy/gateway`.
4. All issued tokens become invalid; users must log in again.

## Notes and assumptions

- The gateway `routes.yaml` points to `Low-Code-Bot-Platform-*` service hostnames; the Services are named in lowercase
  (`low-code-bot-platform-auth-service`, `...-users-backend`) — DNS is case-insensitive, so the gateway resolves them.
- `bot` service referenced by the gateway has no manifests yet (no source in the repo).
- The gateway is exposed as a `NodePort` (`30080`) on every node, e.g. `http://192.168.0.31:30080/main/api/v1/`. No Ingress is defined.
- `networkpolicies.yaml` is enforced: the cluster runs Calico (VXLAN, pod CIDR `10.244.0.0/16`). Ingress is default-deny; the gateway accepts traffic from anywhere (needed for NodePort), `auth`/`backend` only from the gateway, Postgres only from `auth`/`backend`.
- Replicas are set to 1 and memory limits are low because the workers have ~4 GB RAM. Postgres uses `local-path` storage, so its data is tied to one node.
- `postgresql.conf` from `Infra/` is not mounted; the stock PostgreSQL config is used.
- The CI workflow currently builds only `backend` (image `test-k8s-api`); `auth` and `gateway` images need to be built and pushed too.
