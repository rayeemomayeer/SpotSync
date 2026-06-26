# SpotSync — developer task runner
# Run `make` or `make help` to list available targets.

# Load variables from .env if present (so DATABASE_URL etc. are available to targets)
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

# ---- Configuration (override on the command line, e.g. `make run PORT=9090`) ----
APP_NAME       ?= spotsync
API_PKG        ?= ./cmd/api
MIGRATIONS_DIR ?= migrations
DATABASE_URL   ?= postgres://spotsync:spotsync@localhost:5432/spotsync?sslmode=disable
COMPOSE_FILE   ?= deploy/compose/docker-compose.yml
LOAD_SCRIPT    ?= test/load/reserve_stampede.js

.DEFAULT_GOAL := help

# ---- Help ----
.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ---- Run & build ----
.PHONY: run
run: ## Run the API locally
	go run $(API_PKG)

.PHONY: dev
dev: ## Run the API with hot reload (requires air)
	air

.PHONY: build
build: ## Compile the API binary into ./bin
	go build -o bin/$(APP_NAME) $(API_PKG)

# ---- Code quality ----
.PHONY: fmt
fmt: ## Format the codebase (gofumpt)
	gofumpt -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: tidy
tidy: ## Tidy and verify go modules
	go mod tidy
	go mod verify

# ---- Tests ----
.PHONY: test
test: ## Run unit tests
	go test ./...

.PHONY: test-race
test-race: ## Run tests with the race detector
	go test -race ./...

.PHONY: test-int
test-int: ## Run integration tests (requires Docker)
	go test -tags=integration ./test/integration/...

.PHONY: cover
cover: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# ---- Database migrations (requires the golang-migrate CLI) ----
.PHONY: migrate-up
migrate-up: ## Apply all up migrations
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: migrate-create
migrate-create: ## Create a new migration: make migrate-create name=add_table
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

.PHONY: seed
seed: ## Seed the database (e.g. an initial admin user)
	go run ./cmd/seed

# ---- Local environment ----
.PHONY: compose-up
compose-up: ## Start the full local stack via Docker Compose
	docker compose -f $(COMPOSE_FILE) up --build

.PHONY: compose-down
compose-down: ## Stop the local stack and remove volumes
	docker compose -f $(COMPOSE_FILE) down -v

# ---- Load testing (requires k6) — added in the observability phase ----
.PHONY: load-test
load-test: ## Run the k6 stampede load test
	k6 run $(LOAD_SCRIPT)

# ---- Kubernetes on kind — added in the deployment phase ----
.PHONY: kind-up
kind-up: ## Create a local kind cluster and deploy SpotSync
	bash deploy/scripts/kind-up.sh

.PHONY: kind-down
kind-down: ## Delete the local kind cluster
	bash deploy/scripts/kind-down.sh
