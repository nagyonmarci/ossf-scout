# Contributing

## Reporting issues

Use [GitHub Issues](https://github.com/nagyonmarci/ossf-scout/issues) for bugs and feature requests.

## Submitting a pull request

1. Fork the repository and create a branch from `main`
2. Make your changes
3. Ensure the build passes: `make build` or `go build ./...`
4. Run tests: `go test ./...`
5. Open a pull request against `main`

## Development setup

Prerequisites: **Go 1.25+**, **Node 22+**

```bash
# Build everything (frontend + Go binary)
make build

# Development mode (watches frontend changes)
make dev

# Run tests
go test ./...

# Run the server locally
go run . -serve
```

## Code style

- Go: standard `gofmt` formatting, `go vet` must pass
- No new dependencies without discussion
- Keep changes focused — one concern per PR
