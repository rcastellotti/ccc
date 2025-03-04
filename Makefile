help: ## show this help
	@echo "\nspecify a command. choices are:\n"
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-12s\033[m %s\n", $$1, $$2}'
	@echo ""
.PHONY: help

build: ## build
	go build -ldflags="-s -w" -o ccc main.go 
.PHONY: build

test: ## run tests
	go test -v -failfast github.com/rcastellotti/ccc/pkg/cmd
.PHONY: test

test-cov: ## run tests with coverage
	go test -covermode=count -coverpkg=./... -coverprofile cover.out
	go tool cover -html cover.out -o cover.html
.PHONY: test-coverage

clean: ## cleanup
	rm cover.out
	rm cover.html
	rm ccc
.PHONY: clean
