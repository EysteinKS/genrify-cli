# Copilot instructions (genrify)

## Verification requirements

When you change Go code:

- Run `make lint` (or `go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4 run --timeout=5m`).
- Run `go test ./...` (prefer `make test`, which runs lint first).
- If you touch CLI/config behavior, ensure related tests are updated/added.

## CI parity

- Keep golangci-lint version in sync with `.github/workflows/ci.yml` and `Makefile` (`v1.63.4` currently).
