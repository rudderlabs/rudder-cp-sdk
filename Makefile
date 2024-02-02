GO=go
LDFLAGS?=-s -w
TESTFILE=_testok

.PHONY: default
default: lint

.PHONY: help
help: ## Show the available commands
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: install-tools
	$(GO) generate ./...

.PHONY: test
test: install-tools test-run test-teardown ## Run all tests

test-teardown:
	@if [ -f "$(TESTFILE)" ]; then \
    	echo "Tests passed, tearing down..." ;\
		rm -f $(TESTFILE) ;\
		echo "mode: atomic" > coverage.txt ;\
		find . -name "profile.out" | while read file; do grep -v 'mode: atomic' $${file} >> coverage.txt; rm -f $${file}; done ;\
	else \
    	rm -f coverage.txt coverage.html ; find . -name "profile.out" | xargs rm -f ;\
		echo "Tests failed :-(" ;\
		exit 1 ;\
	fi

.PHONY: test-run
test-run:
	$(eval TEST_CMD = go test)
	$(eval TEST_OPTIONS = -v -count 1 -race -failfast -shuffle=on -coverprofile=profile.out -covermode=atomic -coverpkg=./... -vet=all --timeout=15m)
ifdef package
	$(TEST_CMD) $(TEST_OPTIONS) ./$(package)/... && touch $(TESTFILE) || true
else
	$(TEST_CMD) $(TEST_OPTIONS) ./... && touch $(TESTFILE) || true
endif

.PHONY: coverage
coverage: test-run
	go tool cover -html=coverage.txt -o coverage.html

.PHONY: test-with-coverage
test-with-coverage: test coverage

.PHONY: install-tools
install-tools:
	go install github.com/golang/mock/mockgen@v1.6.0
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/tools/cmd/goimports@latest
	bash ./scripts/install-golangci-lint.sh v1.55.0

.PHONY: lint
lint: fmt ## Run linters on all go files
	golangci-lint run -v --timeout 5m

.PHONY: fmt
fmt: install-tools ## Formats all go files
	gofumpt -l -w -extra  .

.PHONY: bench
bench:
	go test -bench=. -benchmem ./benchmark
