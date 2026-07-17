BUF_VERSION                := v1.71.0
PROTOC_GEN_GO_VERSION      := v1.36.11
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1
BREAKING_BASELINE          ?= main

.PHONY: tools proto lint breaking test tidy

tools:
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

proto:
	buf generate

lint:
	buf lint

breaking:
	@if git ls-tree -r --name-only "$(BREAKING_BASELINE)" -- proto | grep -Eq '\.proto$$'; then \
		buf breaking --against ".git#branch=$(BREAKING_BASELINE)"; \
	else \
		echo "No protobuf baseline on $(BREAKING_BASELINE); skipping initial breaking check"; \
	fi

test:
	go build ./...
	go vet ./...
	go test ./...

tidy:
	go mod tidy
