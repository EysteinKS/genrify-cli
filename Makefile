.PHONY: test lint fmt

GOLANGCI_LINT_VERSION ?= v1.63.4

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --timeout=5m

test: lint
	go test ./... -race -coverprofile=coverage.out -covermode=atomic

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')
