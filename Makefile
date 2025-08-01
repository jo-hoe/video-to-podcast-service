include help.mk

# Cross-platform build configuration

.DEFAULT_GOAL := start

# Build configuration
BUILD_DIR := bin
API_BINARY := api-service
UI_BINARY := ui-service

# Cross-platform build settings

# =============================================================================
# Binary Build Targets
# =============================================================================

.PHONY: build
build: build-api build-ui ## build both API and UI binaries

.PHONY: build-api
build-api: ## build API service binary
	go build -o $(API_BINARY) ./cmd/api

.PHONY: build-ui
build-ui: ## build UI service binary
	go build -o $(UI_BINARY) ./cmd/ui

.PHONY: start
start: ## build and start with docker compose
	docker compose up --build

# =============================================================================
# Testing and Quality Targets
# =============================================================================

.PHONY: test
test: ## run golang test (including integration tests)
	go test -timeout 0  ./...

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run ./...

.PHONY: update
update: ## update dependencies
	go mod tidy

# =============================================================================
# Docker and CI/CD Targets
# =============================================================================

.PHONY: docker-build
docker-build: ## build Docker images locally
	docker compose build

.PHONY: docker-test
docker-test: docker-build ## build and test Docker images
	docker compose up -d
	sleep 10
	curl -f http://localhost:8080/v1/health
	curl -f http://localhost:3000/health
	docker compose down

.PHONY: docker-clean
docker-clean: ## clean up Docker resources
	docker compose down -v
	docker system prune -f
