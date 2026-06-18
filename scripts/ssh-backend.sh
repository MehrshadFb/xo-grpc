#!/usr/bin/env bash
set -euo pipefail

XO_BACKEND_HOST="${XO_BACKEND_HOST:-40.233.102.214}"
XO_BACKEND_USER="${XO_BACKEND_USER:-opc}"
XO_BACKEND_KEY="${XO_BACKEND_KEY:-$HOME/.ssh/id_ed25519}"

if [[ ! -f "$XO_BACKEND_KEY" ]]; then
  echo "SSH key not found: $XO_BACKEND_KEY" >&2
  exit 1
fi

exec ssh \
  -i "$XO_BACKEND_KEY" \
  -o IdentitiesOnly=yes \
  "$XO_BACKEND_USER@$XO_BACKEND_HOST" \
  "$@"
