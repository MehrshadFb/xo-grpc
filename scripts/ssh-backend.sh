#!/usr/bin/env bash
set -euo pipefail
export PATH="/opt/homebrew/bin:$PATH"
cd "$(dirname "$0")/.."

command -v railway >/dev/null 2>&1 || { echo "Railway CLI not found. Install: brew install railway" >&2; exit 1; }

choice="${1:-}"
if [ -z "$choice" ]; then
  echo "What do you want to do?"
  echo "  1) Shell into backend container"
  echo "  2) Stream backend logs"
  echo "  3) Connect to Postgres (psql)"
  printf "Enter choice [1-3]: "
  read -r n
  case "$n" in
    1) choice=shell ;;
    2) choice=logs ;;
    3) choice=db ;;
    *) echo "Invalid choice."; exit 1 ;;
  esac
fi

case "$choice" in
  shell) exec railway ssh --service backend ;;
  logs)  exec railway logs --service backend ;;
  db)    exec railway connect Postgres ;;
  *)     echo "Unknown option: '$choice' (use: shell | logs | db)"; exit 1 ;;
esac
