.PHONY: build run clean test tidy docker-build docker-push

# Variables
IMAGE_NAME ?= k8s-manager
IMAGE_TAG ?= latest
REGISTRY ?=

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run with custom config
run-config:
	go run cmd/server/main.go -config=$(CONFIG)

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Tidy dependencies
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run
