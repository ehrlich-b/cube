.PHONY: build clean test run serve install dev fmt vet lint

# Build the binary
build:
	mkdir -p dist
	go build -o dist/cube ./cmd/cube

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

# Start the web server
serve:
	go run ./cmd/cube serve

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