.PHONY: build run clean test tidy docker-build docker-push deploy undeploy

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

# Generate wire (if using wire for DI)
wire:
	wire ./...

# Docker build
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# Docker build with registry
docker-build-push: docker-build
ifdef REGISTRY
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
else
	@echo "REGISTRY is not set. Use: make docker-build-push REGISTRY=your-registry.com"
endif

# Deploy to Kubernetes
deploy:
	kubectl apply -k deploy/

# Undeploy from Kubernetes
undeploy:
	kubectl delete -k deploy/

# Show Kubernetes resources
status:
	kubectl get all -n k8s-manager
