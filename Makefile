PROTO_DIR := proto/v1

.PHONY: deps lint build-proto generate-go generate-python generate verify tidy-go build-python clean

deps:
	cd $(PROTO_DIR) && buf dep update

lint:
	cd $(PROTO_DIR) && buf lint

build-proto:
	cd $(PROTO_DIR) && buf build

generate-go: build-proto
	./scripts/generate-go.sh

generate-python: build-proto
	./scripts/generate-python.sh

generate: generate-go generate-python

sync-from-main:
	./scripts/sync-from-main.sh
tidy-go:
	go mod tidy

build-python:
	cd sdk/python && python3 -m pip install --upgrade build && python3 -m build

verify: lint build-proto generate tidy-go

clean:
	rm -rf gen/go gen/python docs/openapi
	mkdir -p gen/go gen/python docs/openapi