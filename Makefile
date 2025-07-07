# Makefile for Sentinel

.PHONY: build clean test run docker-build docker-run dev deps lint format help proto

# Variables
BINARY_NAME=sentinel
MAIN_PATH=./cmd/server
BUILD_DIR=./build
VERSION?=dev
LDFLAGS=-ldflags="-w -s -X main.version=${VERSION}"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: ## Run in development mode with auto-reload
	go run $(MAIN_PATH)

run: build ## Build and run the application
	./$(BUILD_DIR)/$(BINARY_NAME)

runtcpserver:
	go run ./cmd/tcpserver

rungrpcserver:
	go run ./cmd/grpcserver

# Build targets
build: deps ## Build the application
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)


build-linux: deps ## Build for Linux
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)

build-all: deps ## Build for all platforms
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Code quality
lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...
	goimports -w .

# Docker
docker-push: ## Push Docker image
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t sxwebdev/sentinel:latest .

docker-run: ## Run Docker container
	docker run -d \
		--name sentinel \
		-p 8080:8080 \
		-v $(PWD)/data:/root/data \
		-v $(PWD)/config.yaml:/root/config.yaml \
		sxwebdev/sentinel:latest

docker-stop: ## Stop Docker container
	docker stop sentinel || true
	docker rm sentinel || true

# Docker Compose
up: ## Start with docker-compose
	docker-compose up -d

down: ## Stop docker-compose
	docker-compose down

logs: ## Show docker-compose logs
	docker-compose logs -f

# Installation
install: build ## Install binary to system
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Cleanup
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	docker-compose down --volumes --remove-orphans || true

# Database
init-db: ## Initialize database directory
	mkdir -p data

# Configuration
init-config: ## Copy example configuration
	cp config.yaml.example config.yaml || echo "config.yaml already exists"
	cp .env.example .env || echo ".env already exists"

# Release
release: clean build-all test ## Create release build
	@echo "Release $(VERSION) built successfully"
	@ls -la $(BUILD_DIR)/

genswagger:
	rm -rf ./docs/*
	swag fmt -d ./internal/web/handlers
	swag init -o docs/docsv1 --dir ./internal/web/handlers -g handlers.go --parseDependency
