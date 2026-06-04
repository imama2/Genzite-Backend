# ============================================
# Makefile for Portfolio Builder Backend
# ============================================

APP_NAME     := genzite-backend
CMD_DIR      := ./cmd/server
BUILD_DIR    := ./bin
GO           := go
GOFLAGS      :=
LDFLAGS      := -s -w
DOCKER       := docker compose
SERVER_URL   ?= http://localhost:8080
# TOKEN must be provided via environment variable

ifeq ($(OS),Windows_NT)
SHELL := cmd.exe
.SHELLFLAGS := /C
AIR := $(shell where air 2>NUL)
BUILD_DIR_WIN := $(subst /,\\,$(BUILD_DIR))
MKDIR := if not exist "$(BUILD_DIR_WIN)" mkdir "$(BUILD_DIR_WIN)"
RMDIR := if exist "$(BUILD_DIR_WIN)" rmdir /S /Q "$(BUILD_DIR_WIN)"
else
AIR := $(shell command -v air 2>/dev/null)
MKDIR := mkdir -p $(BUILD_DIR)
RMDIR := rm -rf $(BUILD_DIR)
endif

ifneq ($(strip $(AIR)),)
RUN_CMD := air
else
RUN_CMD := $(GO) run $(GOFLAGS) $(CMD_DIR)
endif

.PHONY: help
help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  run                 Run the server (air if available, otherwise go run)"
	@echo "  build               Build the binary to ./bin/$(APP_NAME)"
	@echo "  clean               Remove ./bin"
	@echo "  test                Run all unit tests"
	@echo "  test-coverage        Generate coverage.out and coverage.html"
	@echo "  docker-up           Start docker compose stack"
	@echo "  docker-down         Stop stack and remove volumes"
	@echo "  docker-rebuild      Rebuild and restart app container only"
	@echo "  docker-logs         Follow logs from app container"
	@echo "  migrate-cli         Run migrations + seed via CLI"
	@echo "  migrate-up          POST /api/v1/migrations/up (TOKEN required)"
	@echo "  migrate-down        POST /api/v1/migrations/down (TOKEN required)"
	@echo "  seed                POST /api/v1/migrations/seed-all (TOKEN required)"
	@echo "  fmt                 Run go fmt ./..."
	@echo "  tidy                Run go mod tidy"
	@echo "  lint                Run golangci-lint"
	@echo "  proto               Generate gRPC stubs for Go and Python"

.PHONY: run
run: ## Run the server with air (fallback to go run)
	$(RUN_CMD)

.PHONY: build
build: ## Build the binary
	@$(MKDIR)
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

.PHONY: clean
clean: ## Remove build artifacts
	@$(RMDIR)

.PHONY: test
test: ## Run all unit tests
	$(GO) test ./... -v -count=1

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: docker-up
docker-up: ## Start the full stack (app + postgres + rabbitmq)
	$(DOCKER) up -d

.PHONY: docker-down
docker-down: ## Stop the stack and remove volumes
	$(DOCKER) down -v

.PHONY: docker-rebuild
docker-rebuild: ## Rebuild and restart the app container only
	$(DOCKER) up -d --build app

.PHONY: docker-logs
docker-logs: ## Follow logs from the app container
	$(DOCKER) logs -f app

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations (TOKEN required)
	curl -X POST $(SERVER_URL)/api/v1/migrations/up \
		-H "Authorization: Bearer $(TOKEN)" \
		-H "Content-Type: application/json"

.PHONY: migrate-down
migrate-down: ## Rollback migrations (TOKEN required)
	curl -X POST $(SERVER_URL)/api/v1/migrations/down \
		-H "Authorization: Bearer $(TOKEN)" \
		-H "Content-Type: application/json"

.PHONY: seed
seed: ## Run all seeders (TOKEN required)
	curl -X POST $(SERVER_URL)/api/v1/migrations/seed-all \
		-H "Authorization: Bearer $(TOKEN)" \
		-H "Content-Type: application/json"

.PHONY: migrate-cli
migrate-cli: ## Run migrations + seed from CLI
	$(GO) run ./cmd/migrate

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...

.PHONY: tidy
tidy: ## Tidy go modules
	$(GO) mod tidy

.PHONY: lint
lint: ## Lint with golangci-lint
	golangci-lint run ./...

.PHONY: proto
proto: ## Generate gRPC stubs
	@echo "Generating Go gRPC stubs..."
	if not exist "gen" mkdir "gen"
	if not exist "gen\go" mkdir "gen\go"
	protoc --proto_path=proto \
		--go_out=gen/go --go_opt=paths=source_relative \
		--go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
		proto/agents/architect.proto

	@echo "Generating Python gRPC stubs..."
	if not exist "python\grpc_server\generated" mkdir "python\grpc_server\generated"
	python -m grpc_tools.protoc --proto_path=proto/agents \
		--python_out=python/grpc_server/generated \
		--grpc_python_out=python/grpc_server/generated \
		--pyi_out=python/grpc_server/generated \
		proto/agents/architect.proto
	python -c "from pathlib import Path; p=Path('python/grpc_server/generated/architect_pb2_grpc.py'); s=p.read_text(); p.write_text(s.replace('import architect_pb2 as architect__pb2', 'from . import architect_pb2 as architect__pb2'))"
	if not exist "python\grpc_server\generated\__init__.py" type nul > "python\grpc_server\generated\__init__.py"
	@echo "Done."
