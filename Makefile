.PHONY: help build run test lint clean docker-up docker-down migrate

# Variables
BINARY_NAME=kapok
GATEWAY_BINARY=kapok-gateway
CLI_BINARY=kapok-cli
CONTROL_PLANE_BINARY=kapok-control-plane
VERSION?=0.1.0
BUILD_DIR=bin
GO=go

# Colors
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

help: ## Show this help
	@echo "Kapok - Backend-as-a-Service Multi-Tenant"
	@echo ""
	@echo "Usage:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

# =============================================================================
# Build
# =============================================================================

build: build-gateway build-cli build-control-plane ## Build all binaries

build-gateway: ## Build gateway binary
	@echo "$(YELLOW)Building gateway...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(GATEWAY_BINARY) ./cmd/gateway

build-cli: ## Build CLI binary
	@echo "$(YELLOW)Building CLI...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/cli

build-control-plane: ## Build control-plane binary
	@echo "$(YELLOW)Building control-plane...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(CONTROL_PLANE_BINARY) ./cmd/control-plane

# =============================================================================
# Run
# =============================================================================

run: ## Run gateway locally
	@echo "$(YELLOW)Running gateway...$(NC)"
	$(GO) run ./cmd/gateway

run-watch: ## Run gateway with hot reload (requires air)
	@echo "$(YELLOW)Running gateway with hot reload...$(NC)"
	air -c .air.toml

# =============================================================================
# Development
# =============================================================================

dev: docker-up ## Start development environment
	@echo "$(GREEN)Development environment started$(NC)"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis:      localhost:6379"
	@echo "  MinIO:      localhost:9000 (Console: localhost:9001)"
	@echo "  Mailhog:    localhost:8025"

stop: docker-down ## Stop development environment

# =============================================================================
# Docker
# =============================================================================

docker-up: ## Start Docker services
	@echo "$(YELLOW)Starting Docker services...$(NC)"
	docker compose up -d

docker-down: ## Stop Docker services
	@echo "$(YELLOW)Stopping Docker services...$(NC)"
	docker compose down

docker-logs: ## Show Docker logs
	docker compose logs -f

docker-clean: ## Remove Docker volumes
	docker compose down -v

# =============================================================================
# Database
# =============================================================================

migrate-up: ## Run migrations up
	@echo "$(YELLOW)Running migrations...$(NC)"
	$(GO) run ./cmd/cli migrate up

migrate-down: ## Run migrations down
	@echo "$(YELLOW)Rolling back migrations...$(NC)"
	$(GO) run ./cmd/cli migrate down

migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	@echo "$(YELLOW)Creating migration: $(name)$(NC)"
	@mkdir -p migrations/control-plane
	@touch migrations/control-plane/$$(date +%Y%m%d%H%M%S)_$(name).up.sql
	@touch migrations/control-plane/$$(date +%Y%m%d%H%M%S)_$(name).down.sql
	@echo "$(GREEN)Created migration files$(NC)"

db-reset: ## Reset database (drop and recreate)
	@echo "$(YELLOW)Resetting database...$(NC)"
	docker compose exec postgres psql -U kapok -c "DROP DATABASE IF EXISTS kapok_control;"
	docker compose exec postgres psql -U kapok -c "CREATE DATABASE kapok_control;"
	@echo "$(GREEN)Database reset$(NC)"

db-shell: ## Open PostgreSQL shell
	docker compose exec postgres psql -U kapok -d kapok_control

redis-shell: ## Open Redis shell
	docker compose exec redis redis-cli

# =============================================================================
# Testing
# =============================================================================

test: ## Run tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GO) test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	$(GO) test -v -tags=integration ./tests/integration/...

# =============================================================================
# Quality
# =============================================================================

lint: ## Run linter
	@echo "$(YELLOW)Running linter...$(NC)"
	golangci-lint run

fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GO) fmt ./...
	gofumpt -l -w .

vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	$(GO) vet ./...

# =============================================================================
# Dependencies
# =============================================================================

deps: ## Download dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	$(GO) mod download

deps-update: ## Update dependencies
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy

deps-tidy: ## Tidy dependencies
	$(GO) mod tidy

# =============================================================================
# Cleanup
# =============================================================================

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# =============================================================================
# Setup
# =============================================================================

setup: ## Initial project setup
	@echo "$(YELLOW)Setting up project...$(NC)"
	cp config.example.yaml config.yaml
	$(GO) mod download
	@echo "$(GREEN)Setup complete!$(NC)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit config.yaml with your settings"
	@echo "  2. Run 'make dev' to start development services"
	@echo "  3. Run 'make run' to start the gateway"

install-tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/cosmtrek/air@latest
	@echo "$(GREEN)Tools installed$(NC)"

# =============================================================================
# Web Console
# =============================================================================

web-install: ## Install web console dependencies
	@echo "$(YELLOW)Installing web console dependencies...$(NC)"
	cd web && npm install

web-dev: ## Start web console dev server
	@echo "$(YELLOW)Starting web console...$(NC)"
	cd web && npm run dev

web-build: ## Build web console for production
	@echo "$(YELLOW)Building web console...$(NC)"
	cd web && npm run build

web-lint: ## Lint web console
	@echo "$(YELLOW)Linting web console...$(NC)"
	cd web && npm run lint

web-typecheck: ## Type-check web console
	@echo "$(YELLOW)Type-checking web console...$(NC)"
	cd web && npm run typecheck
