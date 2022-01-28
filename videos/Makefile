lint:
	@echo "Running linter checks"
	golangci-lint run

test:
	@echo "Running tests"
	@go clean -testcache
	go test -cover -race -short ./... | { grep -v 'no test files'; true; }
