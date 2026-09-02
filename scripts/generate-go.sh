#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

mkdir -p "${ROOT_DIR}/gen/go" "${ROOT_DIR}/docs/openapi"

cd "${ROOT_DIR}/proto/v1"
buf generate . --template buf.gen.yaml

echo "Go / gateway / OpenAPI generation completed."