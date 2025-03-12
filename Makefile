help: ## show this help
	@echo "\nspecify a command. choices are:\n"
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-12s\033[m %s\n", $$1, $$2}'
	@echo ""
.PHONY: help

build: ## build ccc binary
	go build -ldflags="-s -w" -o ccc main.go 
.PHONY: build

test: ## run tests
	go test -failfast github.com/rcastellotti/ccc/...
.PHONY: test

test-cov: ## run tests with coverage
	go test github.com/rcastellotti/ccc/cmd -covermode=count -coverpkg=./... -coverprofile cover.out
	go tool cover -html cover.out -o cover.html
.PHONY: test-coverage

clean: ## cleanup
	rm -f cover.out cover.html
	rm -f ccc
.PHONY: clean

format: ## format files using gofumpt
	go run mvdan.cc/gofumpt@latest -w .
	gofmt -s -w .
.PHONY: format


lint: ## lint files using golangci-lint
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6 run
.PHONY: lint

debug: ## debug main.go using delve
	go run github.com/go-delve/delve/cmd/dlv@latest debug
.PHONY: debug