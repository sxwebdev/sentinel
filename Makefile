# Makefile for Sentinel

.PHONY: build clean test run docker-build docker-run dev deps lint format help proto

# Variables
BINARY_NAME=sentinel
SENTINEL_PATH=./cmd/sentinel
BUILD_DIR=./build
VERSION?=dev
LDFLAGS=-ldflags="-w -s -X main.version=${VERSION}"
MIGRATIONS_DIR	 = ./sql/migrations/

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
hub: ## Run in development mode with auto-reload
	go run $(SENTINEL_PATH) hub start -c ./config.yaml

agent: ## Run in development mode with auto-reload
	go run $(SENTINEL_PATH) agent start -c ./config-agent.yaml

migrateup:
	go run $(SENTINEL_PATH) migrations up -db-path ./data/hub/sqlite/db.sqlite

migratedown:
	go run $(SENTINEL_PATH) migrations down -db-path ./data/hub/sqlite/db.sqlite

air:
	air -c .air.toml

run: build ## Build and run the application
	./$(BUILD_DIR)/$(BINARY_NAME)

runtestservers:
	go run ./cmd/testserver -http -grpc -tcp

front:
	cd frontend && pnpm dev

# Build targets
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SENTINEL_PATH)


build-linux: deps ## Build for Linux
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(SENTINEL_PATH)

build-all: deps ## Build for all platforms
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SENTINEL_PATH)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SENTINEL_PATH)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SENTINEL_PATH)

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

# Testing
test: ## Run tests
	go test -v ./...

test-api:
	@echo "Running Sentinel API Integration Tests..."
	go run ./cmd/testapi test

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Code quality
lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...
	goimports -w .

docker-push: ## Build and push Docker image
	docker buildx build --platform linux/amd64 --push \
		--build-arg VERSION=`git describe --tags --abbrev=0 || echo "0.0.0"` \
		--build-arg COMMIT=`git rev-parse --short HEAD` \
		--build-arg DATE=`date -u +'%Y-%m-%dT%H:%M:%SZ'` \
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

deocker-dev:
	docker compose -f docker-compose.local.yml up -d

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

# db-create-migration:
# 	migrate create -ext sql -format unix -dir "$(MIGRATIONS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

db-create-migration:
	go run ./cmd/sentinel migrations create -p ./sql/migrations -name $(filter-out $@,$(MAKECMDGOALS))

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
	swag fmt -d ./internal/web
	swag init -o docs/docsv1 --dir ./internal/web -g handlers.go --parseDependency

genenvs:
	go run ./cmd/sentinel config genenvs

gensql:
	pgxgen crud
	pgxgen sqlc generate

genproto: ## Generate protobuf code
	buf lint
	rm -rf ./internal/hub/hubserver/api/*
	rm -rf frontend/src/api/gen/*
	buf generate
	rm -rf frontend/src/api/gen/sentinel/hub

grpcui-hub:
	grpcui --plaintext localhost:8080

grpcui-server:
	grpcui --plaintext localhost:8080

%:
	@:
