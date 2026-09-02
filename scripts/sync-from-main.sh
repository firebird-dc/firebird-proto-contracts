#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Priority:
# 1. first script argument
# 2. SOURCE_REPO env var
# 3. default path on mac
SOURCE_REPO="${1:-${SOURCE_REPO:-${HOME}/Developer}}"

DEST_DIR="${ROOT_DIR}/proto/v1"
SRC_DIR="${SOURCE_REPO}/proto/v1"

mkdir -p "${DEST_DIR}"

FILES=(
  common.proto
  audit.proto
  bm.proto
  bm_dhcp.proto
  filesystem.proto
  image.proto
  machine_type.proto
  maintenance_event.proto
  network_acl.proto
  operation.proto
  project.proto
  serial_logs.proto
  user_management.proto
  vpc.proto
)

GO_PACKAGE="github.com/firebird-dc/firebird-proto-contracts/gen/go;v1"

for file in "${FILES[@]}"; do
  cp "${SRC_DIR}/${file}" "${DEST_DIR}/${file}"
  # The upstream files point go_package at the server module; retarget it at this module.
  sed -i.bak -E "s#^option go_package = \"[^\"]+\";#option go_package = \"${GO_PACKAGE}\";#" "${DEST_DIR}/${file}"
  rm -f "${DEST_DIR}/${file}.bak"
done

echo "Synced approved proto files from ${SOURCE_REPO}."