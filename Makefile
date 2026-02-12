.PHONY: test lint fmt build build-cli

GOLANGCI_LINT_VERSION ?= v1.63.4

build:
	CGO_ENABLED=1 go build -o genrify ./cmd/genrify

build-cli:
	CGO_ENABLED=0 go build -tags nogui -o genrify ./cmd/genrify

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --timeout=5m

test: lint
	go test ./... -race -coverprofile=coverage.out -covermode=atomic

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')
