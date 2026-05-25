BINARY_NAME   := api
BUILD_DIR     := bin
CMD_API       := ./cmd/api
CMD_SETUP     := ./cmd/setup
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html
DOCKER_COMPOSE := docker compose

.PHONY: help setup run dev build test test-cover swagger \
        docker-up docker-down docker-logs pgadmin-up \
        lint tidy clean

all: help

help: ## List all available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# ----------------------------------------------------------
# Development
# ----------------------------------------------------------

setup: ## Interactive setup - generates config.yaml from prompts
	go run $(CMD_SETUP)

run: ## Start the API server (requires config.yaml)
	go run $(CMD_API)

dev: ## Start the API with hot reload (requires: go install github.com/air-verse/air@latest)
	air

build: ## Compile binary to ./bin/api
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_API)

tidy: ## Run go mod tidy
	go mod tidy

clean: ## Remove build artifacts and coverage files
	@rm -rf $(BUILD_DIR) $(COVERAGE_FILE) $(COVERAGE_HTML)

# ----------------------------------------------------------
# Tests
# ----------------------------------------------------------

test: ## Run all tests with race detector
	go test ./... -v -race

test-cover: ## Run tests and open HTML coverage report
	go test ./... -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# ----------------------------------------------------------
# Code quality
# ----------------------------------------------------------

lint: ## Run golangci-lint (requires: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

swagger: ## Generate Swagger docs (requires: go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g $(CMD_API)/main.go -o docs

# ----------------------------------------------------------
# Docker
# ----------------------------------------------------------

docker-up: ## Start PostgreSQL container
	$(DOCKER_COMPOSE) up -d postgres

docker-down: ## Stop and remove containers and volumes
	$(DOCKER_COMPOSE) down -v

docker-logs: ## Tail PostgreSQL container logs
	$(DOCKER_COMPOSE) logs -f postgres

pgadmin-up: ## Start PostgreSQL + pgAdmin (UI at http://localhost:5050)
	$(DOCKER_COMPOSE) --profile pgadmin up -d
