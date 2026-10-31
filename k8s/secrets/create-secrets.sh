#!/usr/bin/env bash
# Creates the cluster secrets from an env file.
#   ./create-secrets.sh [--env-file FILE] [--seal]
# Default: applies the secrets straight to the cluster with kubectl.
# --seal : writes SealedSecret manifests to ./sealed/ (safe to commit) instead.
set -euo pipefail

NAMESPACE="${NAMESPACE:-low-code-bot-platform}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE=".env"
SEAL=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env-file) ENV_FILE="$2"; shift 2 ;;
    --seal) SEAL=1; shift ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

[[ -f "$ENV_FILE" ]] || { echo "env file not found: $ENV_FILE" >&2; exit 1; }

get() {
  local line
  line="$(grep -E "^$1=" "$ENV_FILE" | tail -n1 || true)"
  [[ -n "$line" ]] || { echo "missing $1 in $ENV_FILE" >&2; exit 1; }
  local value="${line#*=}"
  value="${value%\"}"
  value="${value#\"}"
  printf '%s' "$value"
}

TMP="$(mktemp -d)"
umask 077
trap 'rm -rf "$TMP"' EXIT

put() { printf '%s' "$2" > "$TMP/$1"; }

publish() {
  local name="$1"; shift
  local manifest
  manifest="$(kubectl create secret generic "$name" -n "$NAMESPACE" --dry-run=client -o yaml "$@")"
  if [[ $SEAL -eq 1 ]]; then
    mkdir -p "$HERE/sealed"
    printf '%s\n' "$manifest" | kubeseal --format yaml -n "$NAMESPACE" > "$HERE/sealed/$name.yaml"
    echo "sealed: sealed/$name.yaml"
  else
    printf '%s\n' "$manifest" | kubectl apply -f -
  fi
}

if [[ $SEAL -eq 0 ]]; then
  kubectl get namespace "$NAMESPACE" >/dev/null 2>&1 || kubectl create namespace "$NAMESPACE"
fi

put DATABASE_URL "$(get DATABASE_URL)"
put JWT_SECRET_KEY "$(get JWT_SECRET_KEY)"
put JWT_PUBLIC_KEY "$(get JWT_PUBLIC_KEY)"

publish database-url --from-file=DATABASE_URL="$TMP/DATABASE_URL"
publish jwt-private-key --from-file=JWT_SECRET_KEY="$TMP/JWT_SECRET_KEY"
publish jwt-public-key --from-file=JWT_PUBLIC_KEY="$TMP/JWT_PUBLIC_KEY"
