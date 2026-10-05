# Firebird API Contracts & Client SDKs

Tenant-facing gRPC/protobuf contracts for the Firebird API, together with the Go and
Python client SDKs and OpenAPI generated from them. Tenants depend on **tagged releases**
of this repository for a stable, reviewable API surface, decoupled from the server's
release cadence.

The `.proto` files here are the **approved public subset** of the Firebird API —
operator/internal services are deliberately excluded. They are vendored from the upstream
`firebird-api` service via `scripts/sync-from-main.sh`. Everything under `gen/` and
`docs/openapi/` is **generated**; do not hand-edit it.

## Layout

```text
proto/v1/                        # approved public contracts — source of truth for codegen
  buf.yaml                       # buf module + lint/breaking config
  buf.gen.yaml                   # Go + gRPC + grpc-gateway + OpenAPI (local plugins)
  buf.gen.python.remote.yaml     # Python + gRPC (buf remote plugins)
  *.proto                        # audit, bm, bm_dhcp, common, filesystem, image,
                                 # machine_type, maintenance_event, operation,
                                 # project, serial_logs, user_management, vpc
gen/go/                          # generated Go client (package v1)
gen/python/                      # generated Python client (*_pb2.py, *_pb2_grpc.py)
docs/openapi/                    # generated merged OpenAPI (firebird.swagger.json)
examples/{go,python}/            # runnable auth + call samples
scripts/                         # sync + codegen helpers
```

## Regenerating

Prerequisites: [`buf`](https://buf.build), a Go toolchain, and Python 3.10+. Install the Go
protoc plugins once:

```bash
./scripts/install-go-tools.sh    # protoc-gen-go, -go-grpc, -grpc-gateway, -openapiv2
```

Then:

```bash
make sync-from-main   # refresh proto/v1 from firebird-api (override source: SOURCE_REPO=/path/to/firebird-api)
make generate         # generate-go + generate-python (+ merged OpenAPI)
make verify           # lint + build + generate + go mod tidy
```

Go codegen uses the locally installed plugins above. Python codegen uses **buf remote
plugins** — no local `protoc`/plugins needed, but network access to `buf.build` is required.
Regenerate and commit `proto/`, `gen/`, and `docs/openapi/` together so every tag is
internally consistent.

## Consuming

**Go** — module `github.com/firebird-dc/firebird-proto-contracts`:

```go
import v1 "github.com/firebird-dc/firebird-proto-contracts/gen/go"

client := v1.NewUserManagementServiceClient(conn)
resp, err := client.Token(ctx, &v1.TokenRequest{GrantType: "client_credentials", /* … */})
```

**Python** — install a tagged release with pip (pulls a compatible protobuf/grpcio runtime):

```bash
python -m pip install "firebird-api-client @ git+https://github.com/firebird-dc/firebird-proto-contracts@v0.5.1"
```

or, from a checkout, install the runtime and put the generated modules on `PYTHONPATH`:

```bash
python -m pip install "protobuf>=7.34.0,<8" "grpcio>=1.78.0,<2" "googleapis-common-protos>=1.63"
export PYTHONPATH="$(pwd)/gen/python"
```
```python
import user_management_pb2, user_management_pb2_grpc
```

## Examples

`examples/go/main.go` and `examples/python/example.py` demonstrate the tenant auth flow:
obtain a bearer token via `UserManagementService.Token` (`client_credentials` grant), then
call `GetUserInfo` with it.

```bash
# Go
go run ./examples/go/main.go     --api-url https://<host> --client-id <id> --client-secret <secret>

# Python
PYTHONPATH="$(pwd)/gen/python" python examples/python/example.py \
  --api-url https://<host> --client-id <id> --client-secret <secret>
```

Both accept `--insecure` for plaintext gRPC against a local server.

## Releases

Consume from **tagged releases** (`vX.Y.Z`) — Go modules require the `vX.Y.Z` tag form; the
Python package version tracks `pyproject.toml`. A release is the committed `proto/` + `gen/`
+ `docs/openapi/` at that tag, guaranteed regenerated from the approved contracts.