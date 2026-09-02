#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

mkdir -p "${ROOT_DIR}/gen/python"

(
  cd "${ROOT_DIR}/proto/v1"
  buf generate . --include-imports --template buf.gen.python.remote.yaml
)

touch "${ROOT_DIR}/gen/python/__init__.py"

echo "Python protobuf / gRPC generation completed."