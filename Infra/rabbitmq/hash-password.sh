#!/usr/bin/env bash
# Prints a RabbitMQ password hash (rabbit_password_hashing_sha256) for definitions.json
#   ./hash-password.sh 'new-password'
set -euo pipefail

password="${1:?usage: hash-password.sh PASSWORD}"
salt_hex="$(openssl rand -hex 4)"
hash_hex="$({ printf '%s' "$salt_hex" | xxd -r -p; printf '%s' "$password"; } | openssl dgst -sha256 -binary | xxd -p -c 64)"

printf '%s' "${salt_hex}${hash_hex}" | xxd -r -p | base64 -w0
echo
