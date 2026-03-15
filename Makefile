# Makefile for LiteClaw Go

# Binary name
BINARY_NAME=liteclaw
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) \
				  -X main.buildDate=$(BUILD_TIME) \
				  -X main.commit=$(GIT_COMMIT)"

# Main targets
.PHONY: all build clean test install run fmt lint

all: clean deps build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) .
	@echo "Build complete: bin/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Multi-platform build complete"

# Build with optimizations
build-prod:
	@echo "Building optimized binary..."
	$(GOBUILD) $(LDFLAGS) -ldflags "-s -w" -o bin/$(BINARY_NAME) .
	@echo "Optimized build complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Install binary to GOPATH
install:
	@echo "Installing to GOPATH..."
	$(GOCMD) install $(LDFLAGS) ./...

# Run the binary
run:
	$(GOBUILD) -o bin/$(BINARY_NAME) .
	./bin/$(BINARY_NAME)

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code
lint:
	@echo "Linting..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Security scan
security:
	@echo "Running security scan..."
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all > docs/api.txt

# Docker build
docker:
	@echo "Building Docker image..."
	docker build -t liteclaw:$(VERSION) .

# Run in Docker
docker-run:
	docker run -it --rm liteclaw:$(VERSION)

# Benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Check code quality
quality: fmt lint security test
	@echo "All quality checks passed!"

# Development mode with hot reload
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

# Binary size analysis
size:
	@echo "Binary size analysis..."
	$(GOBUILD) -o bin/$(BINARY_NAME) .
	ls -lh bin/$(BINARY_NAME)

# Profile CPU
profile-cpu:
	@echo "Profiling CPU..."
	$(GOBUILD) -o bin/$(BINARY_NAME)-profile .
	./bin/$(BINARY_NAME)-profile -cpuprofile=cpu.prof
	$(GOCMD) tool pprof cpu.prof

# Profile Memory
profile-mem:
	@echo "Profiling memory..."
	$(GOBUILD) -o bin/$(BINARY_NAME)-profile .
	./bin/$(BINARY_NAME)-profile -memprofile=mem.prof
	$(GOCMD) tool pprof mem.prof

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  build-prod     - Build optimized binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Install dependencies"
	@echo "  install        - Install to GOPATH"
	@echo "  run            - Build and run"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  security       - Security scan"
	@echo "  docker         - Build Docker image"
	@echo "  benchmark      - Run benchmarks"
	@echo "  quality        - Run all quality checks"
	@echo "  dev            - Development mode with hot reload"
