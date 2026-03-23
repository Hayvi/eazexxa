.PHONY: build test clean run-api run-settlement run-presence

# Build all binaries
build:
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o bin/api cmd/api/main.go
	go build -o bin/settlement cmd/settlement/main.go
	go build -o bin/presence cmd/presence/main.go
	go build -o bin/nginxtraffic cmd/nginxtraffic/main.go
	@echo "Build complete!"

# Run tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run tests (skip integration)
test-short:
	go test -short ./... -v

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run API server
run-api:
	go run cmd/api/main.go

# Run settlement worker
run-settlement:
	go run cmd/settlement/main.go

# Run presence worker
run-presence:
	go run cmd/presence/main.go

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run all checks
check: fmt lint test-short

# Development mode (with auto-reload)
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air
