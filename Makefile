# NixGuard Build System
# Google-standard hermetic build with clear targets

SHELL := /bin/bash
.DEFAULT_GOAL := help

# ─── Variables ────────────────────────────────────────────────────
GO := go
GOFLAGS := -trimpath
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-X github.com/nixguard/nixguard/pkg/version.Version=$(VERSION) \
	-X github.com/nixguard/nixguard/pkg/version.BuildTime=$(BUILD_TIME)"

BIN_DIR := bin
CMD_DIR := cmd
PROTO_DIR := api/proto

# Binary names
SERVER_BIN := nixguard-server
AGENT_BIN := nixguard-agent
CLI_BIN := nixguard-cli
WORKER_BIN := nixguard-worker

# ─── Build ────────────────────────────────────────────────────────
.PHONY: build
build: build-server build-agent build-cli build-worker ## Build all binaries

.PHONY: build-server
build-server: ## Build API server
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(SERVER_BIN) ./$(CMD_DIR)/$(SERVER_BIN)

.PHONY: build-agent
build-agent: ## Build privileged agent
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(AGENT_BIN) ./$(CMD_DIR)/$(AGENT_BIN)

.PHONY: build-cli
build-cli: ## Build CLI tool
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(CLI_BIN) ./$(CMD_DIR)/$(CLI_BIN)

.PHONY: build-worker
build-worker: ## Build background worker
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(WORKER_BIN) ./$(CMD_DIR)/$(WORKER_BIN)

# ─── Proto ────────────────────────────────────────────────────────
.PHONY: proto
proto: ## Generate code from protobuf definitions
	@echo "Generating protobuf code..."
	buf generate $(PROTO_DIR)

.PHONY: proto-lint
proto-lint: ## Lint protobuf definitions
	buf lint $(PROTO_DIR)

# ─── Test ─────────────────────────────────────────────────────────
.PHONY: test
test: ## Run unit tests
	$(GO) test -race -count=1 ./internal/... ./pkg/...

.PHONY: test-integration
test-integration: ## Run integration tests
	$(GO) test -race -tags=integration -count=1 ./test/integration/...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	$(GO) test -race -tags=e2e -count=1 ./test/e2e/...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	$(GO) test -race -coverprofile=coverage.out ./internal/... ./pkg/...
	$(GO) tool cover -html=coverage.out -o coverage.html

# ─── Lint ─────────────────────────────────────────────────────────
.PHONY: lint
lint: ## Run linters
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...
	goimports -w .

# ─── Frontend ─────────────────────────────────────────────────────
.PHONY: web-install
web-install: ## Install frontend dependencies
	cd web && npm ci

.PHONY: web-build
web-build: ## Build frontend
	cd web && npm run build

.PHONY: web-dev
web-dev: ## Start frontend dev server
	cd web && npm run dev

.PHONY: web-test
web-test: ## Run frontend tests
	cd web && npm test

.PHONY: web-lint
web-lint: ## Lint frontend code
	cd web && npm run lint

# ─── Docker ───────────────────────────────────────────────────────
.PHONY: docker-build
docker-build: ## Build Docker images
	docker build -t nixguard-server:$(VERSION) -f deployments/docker/Dockerfile.server .
	docker build -t nixguard-agent:$(VERSION) -f deployments/docker/Dockerfile.agent .

.PHONY: docker-compose-up
docker-compose-up: ## Start with docker-compose
	docker-compose -f deployments/docker/docker-compose.yml up -d

# ─── Database ─────────────────────────────────────────────────────
.PHONY: db-migrate
db-migrate: ## Run database migrations
	$(GO) run ./scripts/migration/migrate.go up

.PHONY: db-rollback
db-rollback: ## Rollback last migration
	$(GO) run ./scripts/migration/migrate.go down 1

# ─── Install ──────────────────────────────────────────────────────
.PHONY: install
install: build ## Install binaries to system
	sudo install -m 755 $(BIN_DIR)/$(SERVER_BIN) /usr/local/bin/
	sudo install -m 755 $(BIN_DIR)/$(AGENT_BIN) /usr/local/bin/
	sudo install -m 755 $(BIN_DIR)/$(CLI_BIN) /usr/local/bin/
	sudo install -m 755 $(BIN_DIR)/$(WORKER_BIN) /usr/local/bin/

# ─── Clean ────────────────────────────────────────────────────────
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BIN_DIR) coverage.out coverage.html
	cd web && rm -rf dist node_modules/.cache

# ─── Dev ──────────────────────────────────────────────────────────
.PHONY: dev
dev: ## Start development environment (server + frontend)
	@echo "Starting NixGuard development environment..."
	@$(MAKE) -j2 dev-server dev-web

.PHONY: dev-server
dev-server:
	$(GO) run ./$(CMD_DIR)/$(SERVER_BIN) --config configs/defaults/server.yaml

.PHONY: dev-web
dev-web:
	cd web && npm run dev

# ─── Help ─────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
