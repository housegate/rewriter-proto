BUF_VERSION                := v1.71.0
PROTOC_GEN_GO_VERSION      := v1.36.11
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1
BREAKING_BASELINE          ?= main
VERIFY_BREAKING_BASELINE   ?=

.PHONY: tools proto lint breaking test verify tidy

tools:
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

proto:
	buf generate

lint:
	buf lint

breaking:
	@if ! git rev-parse --verify --quiet "$(BREAKING_BASELINE)^{commit}" >/dev/null; then \
		echo "Breaking baseline does not resolve: $(BREAKING_BASELINE)" >&2; \
		exit 1; \
	fi
	@if git ls-tree -r --name-only "$(BREAKING_BASELINE)" -- proto | grep -Eq '\.proto$$'; then \
		buf breaking --against ".git#ref=$(BREAKING_BASELINE)"; \
	else \
		echo "No protobuf baseline on $(BREAKING_BASELINE); skipping initial breaking check"; \
	fi

test:
	scripts/next-version_test.sh
	go build ./...
	go vet ./...
	go test ./...

verify:
	$(MAKE) lint
	@if [ -n "$(VERIFY_BREAKING_BASELINE)" ]; then \
		$(MAKE) breaking BREAKING_BASELINE="$(VERIFY_BREAKING_BASELINE)"; \
	else \
		echo "No breaking baseline supplied; skipping breaking check"; \
	fi
	$(MAKE) proto
	git diff --exit-code -- gen/
	test -z "$$(git ls-files --others --exclude-standard -- gen/)"
	$(MAKE) test

tidy:
	go mod tidy
