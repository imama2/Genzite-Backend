# ============================================
# Makefile for Portfolio Builder Backend
# ============================================

# Variables
APP_NAME     = portfolio-server
CMD_DIR      = ./cmd/server
BUILD_DIR    = ./bin
ENV_FILE     = .env
DOCKER_COMPOSE = docker-compose

# Go related
GO           = go
GOFLAGS      = -v
LDFLAGS      = -s -w

# Colors (optional, for prettier output)
GREEN  := \033[0;32m
YELLOW := \033[0;33m
NC     := \033[0m # No Color

# ============================================
# Default target
# ============================================
.PHONY: help
help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# ============================================
# Development (local without Docker)
# ============================================
.PHONY: run
run: ## Run the server with hot‑reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to go run..."; \
		$(GO) run $(GOFLAGS) $(CMD_DIR); \
	fi

.PHONY: build
build: ## Build the binary locally (no Docker)
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "Binary created at $(BUILD_DIR)/$(APP_NAME)"

.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf $(BUILD_DIR)

# ============================================
# Testing
# ============================================
.PHONY: test
test: ## Run all unit tests
	$(GO) test ./... -v -count=1

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# ============================================
# Docker operations
# ============================================
.PHONY: docker-build
docker-build: ## Build the Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME) .

.PHONY: docker-up
docker-up: ## Start the full stack (app + postgres + rabbitmq)
	$(DOCKER_COMPOSE) up -d

.PHONY: docker-down
docker-down: ## Stop the stack and remove volumes (caution: data loss)
	$(DOCKER_COMPOSE) down -v

.PHONY: docker-logs
docker-logs: ## Follow logs from the app container
	$(DOCKER_COMPOSE) logs -f app

.PHONY: docker-rebuild
docker-rebuild: ## Rebuild and restart the app container only
	$(DOCKER_COMPOSE) up -d --build app

# ============================================
# Database & Migrations via service endpoints
# (requires the app to be running)
# ============================================
.PHONY: migrate-up
migrate-up: ## Apply all pending migrations (needs admin token)
	@echo "Triggering migrations... (ensure SERVER_URL and TOKEN are set)"
	curl -X POST $(SERVER_URL)/api/v1/migrations/up \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json"

.PHONY: migrate-down
migrate-down: ## Rollback last migration batch (needs admin token)
	@echo "Rolling back migrations..."
	curl -X POST $(SERVER_URL)/api/v1/migrations/down \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json"

.PHONY: seed
seed: ## Run all seeders (needs admin token)
	@echo "Running seeders..."
	curl -X POST $(SERVER_URL)/api/v1/migrations/seed-all \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json"

# ============================================
# Utility
# ============================================
.PHONY: lint
lint: ## Lint the code (golangci-lint must be installed)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format the code
	$(GO) fmt ./...

.PHONY: tidy
tidy: ## Tidy go modules
	$(GO) mod tidy

# ============================================
# Dev environment helpers
# ============================================
.PHONY: dev-setup
dev-setup: ## Install required Go tools (air, golangci-lint, goose)
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

# ============================================
# Full workflow shortcuts
# ============================================
.PHONY: fresh-start
fresh-start: ## Pull images, rebuild app, apply migrations, seed
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d --build
	@sleep 5   # wait for DB/rabbitmq readiness
	@echo "Run migrations now with: make migrate-up SERVER_URL=http://localhost:8080 TOKEN=<your_admin_token>"