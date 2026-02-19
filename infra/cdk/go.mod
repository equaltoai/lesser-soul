module github.com/equaltoai/lesser-soul/infra/cdk

// This nested module exists to prevent `go test ./...` (run from the repo root)
// from traversing this directory (and its `node_modules/`) as part of the main
// Go module.

go 1.26

toolchain go1.26.0

