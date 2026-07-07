#!/usr/bin/env bash
set -euo pipefail
export PATH="/opt/homebrew/bin:$PATH"
cd "$(dirname "$0")/.."   # repo root

command -v railway >/dev/null 2>&1 || { echo "Railway CLI not found. Install: brew install railway" >&2; exit 1; }

echo "==> Redeploying backend to Railway..."
railway up --service backend
echo "==> Done."
