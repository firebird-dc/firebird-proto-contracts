#!/usr/bin/env bash
set -euo pipefail

# Pin these in your real repository if you want deterministic regeneration.
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

echo "Installed Go protobuf plugins into $(go env GOPATH)/bin"
echo "Make sure that directory is on PATH before running buf generate."
