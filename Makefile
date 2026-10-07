.PHONY: build clean test run install dev fmt vet lint e2e-test test-all build-tools test-first-layer test-beginner web web-pages test-web test-web-smoke

# Build the binary
build:
	mkdir -p dist
	go build -o dist/cube ./cmd/cube

# Build database tools
build-tools:
	mkdir -p dist/tools
	go build -o dist/tools/verify-algorithm ./tools/verify-algorithm
	go build -o dist/tools/verify-database ./tools/verify-database

# Build everything (main binary + tools)
build-all-local: build build-tools

# Clean build artifacts
clean:
	rm -rf dist/
	go clean

# Run tests
test:
	go test ./...

# Run the CLI
run:
	go run ./cmd/cube

# Static website; no JavaScript bundler or npm build step.
web:
	GOOS=js GOARCH=wasm go build -p 2 -trimpath -o web/cube.wasm ./cmd/cubewasm
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js

# Export only runtime assets; keep test packages out of GitHub Pages.
web-pages: web
	mkdir -p dist/web
	cp web/index.html web/style.css web/icon.svg web/app.js web/cube-view.js web/engine.js web/worker.js web/cube.wasm web/wasm_exec.js dist/web/

test-web: web
	node web/test/api.test.cjs

test-web-smoke: web
	node web/test/smoke.mjs

# Install dependencies
install:
	go mod download
	go mod tidy

# Development mode with hot reload (requires air)
dev:
	air -c .air.toml

# Format code
fmt:
	go fmt ./...
	@if [ "$$(uname)" = "Darwin" ]; then \
		find . -name "*.go" -exec sed -i '' 's/[[:space:]]*$$//' {} \; ; \
		find . -name "*.go" -exec sh -c 'if [ $$(tail -c1 "$$1" | wc -l) -eq 0 ]; then echo >> "$$1"; fi' _ {} \; ; \
	else \
		find . -name "*.go" -exec sed -i 's/[[:space:]]*$$//' {} \; ; \
		find . -name "*.go" -exec sh -c 'if [ $$(tail -c1 "$$1" | wc -l) -eq 0 ]; then echo >> "$$1"; fi' _ {} \; ; \
	fi

# Vet code
vet:
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install development tools
install-tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/cube-linux-amd64 ./cmd/cube
	GOOS=darwin GOARCH=amd64 go build -o dist/cube-darwin-amd64 ./cmd/cube
	GOOS=windows GOARCH=amd64 go build -o dist/cube-windows-amd64.exe ./cmd/cube

# Run end-to-end tests
e2e-test: build
	@echo "Running end-to-end tests..."
	@bash test/e2e_test.sh

# Run all tests (unit + e2e)
test-all: test e2e-test
	@echo "All tests completed!"

# Independent physical-cubie and interactive replay oracle (Python stdlib only)
test-first-layer: build
	python3 test/first_layer_oracle.py

# Independently replay full solutions, printed checkpoints and last-layer recovery
test-beginner: build
	python3 test/full_lesson_oracle.py
