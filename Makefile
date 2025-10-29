.PHONY: build test coverage install clean lint fmt help

# Build the binary
build:
	go build -o gcdiff ./cmd/gcdiff

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
test-race:
	go test -race -v ./...

# Install the binary
install:
	go install ./cmd/gcdiff

# Clean build artifacts
clean:
	rm -f gcdiff coverage.out coverage.html
	rm -rf dist/

# Run linter
lint:
	go vet ./...
	gofmt -l .

# Format code
fmt:
	go fmt ./...

# Run all checks
check: fmt lint test

# Display help
help:
	@echo "Available targets:"
	@echo "  build       - Build the gcdiff binary"
	@echo "  test        - Run all tests"
	@echo "  coverage    - Run tests with coverage report"
	@echo "  test-race   - Run tests with race detector"
	@echo "  install     - Install gcdiff to GOPATH/bin"
	@echo "  clean       - Remove build artifacts"
	@echo "  lint        - Run linters"
	@echo "  fmt         - Format code"
	@echo "  check       - Run all checks (fmt, lint, test)"
	@echo "  help        - Display this help message"
