#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

(
  cd "${ROOT_DIR}/proto"
  buf lint
  buf build
)

"${ROOT_DIR}/scripts/generate-go.sh"
"${ROOT_DIR}/scripts/generate-python.sh"

(
  cd "${ROOT_DIR}"
  go mod tidy
)

echo "Verification completed."
