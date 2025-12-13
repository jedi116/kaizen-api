# Makefile
.PHONY: help install dev build migrate-diff migrate-apply-local migrate-apply-staging migrate-apply-prod migrate-status test clean

# Load environment variables
include .env
export

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod download
	go mod tidy

#dev: ## Run development server with hot reload
	#go run cmd/server/main.go

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: ## Run the application
	go run cmd/api/main.go

# Migration commands
migrate-diff: ## Create a new migration (usage: make migrate-diff)
	@read -p "Migration name: " name; \
	atlas migrate diff $$name --env local

migrate-apply-local: ## Apply migrations to local database
	atlas migrate apply --env local

migrate-apply-staging: ## Apply migrations to staging database
	atlas migrate apply --env staging

migrate-apply-prod: ## Apply migrations to production database
	@echo "⚠️  WARNING: Applying to PRODUCTION"
	@read -p "Type 'yes' to confirm: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		atlas migrate apply --env production; \
	else \
		echo "Cancelled"; \
	fi

migrate-status: ## Check migration status across all environments
	@echo "=== LOCAL ==="
	@atlas migrate status --env local || true
	@echo ""
	@echo "=== STAGING ==="
	@atlas migrate status --env staging || true
	@echo ""
	@echo "=== PRODUCTION ==="
	@atlas migrate status --env production || true

migrate-lint: ## Lint migrations for potential issues
	atlas migrate lint --env local --latest 1

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -cover ./...

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	go clean

# Docker (optional)
docker-build: ## Build Docker image
	docker build -t your-project:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env your-project:latest