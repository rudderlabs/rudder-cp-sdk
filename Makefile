GO=go
LDFLAGS?=-s -w

.PHONY: default
default: lint

.PHONY: help
help: ## Show the available commands
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: install-tools
	$(GO) generate ./...

.PHONY: test
test: install-tools test-run ## Run all tests

.PHONY: test-run
test-run:
	$(eval TEST_CMD = go test)
	$(eval TEST_OPTIONS = -v -count 1 -race -coverprofile cover.out --timeout=15m)
ifdef package
	$(TEST_CMD) $(TEST_OPTIONS) $(package)
else
	$(TEST_CMD) $(TEST_OPTIONS) ./...
endif

.PHONY: coverage
coverage: test-run
	go tool cover -html=cover.out -o coverage.html

.PHONY: test-with-coverage
test-with-coverage: test coverage

.PHONY: install-tools
install-tools:
	go install github.com/golang/mock/mockgen@v1.6.0
	go install mvdan.cc/gofumpt@latest

.PHONY: lint
lint: fmt ## Run linters on all go files
	docker run --rm -v $(shell pwd):/app:ro -w /app golangci/golangci-lint:v1.51.1 bash -e -c \
		'golangci-lint run -v --timeout 5m'

.PHONY: fmt
fmt: install-tools ## Formats all go files
	gofumpt -l -w -extra  .
