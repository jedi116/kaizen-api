# Makefile
.PHONY: help install build run dev swagger migrate-diff migrate-apply-local migrate-apply-staging migrate-apply-prod migrate-status migrate-lint test test-coverage clean docker-build docker-run

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

build: ## Build the application
	go build -o bin/server cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

dev: ## Run development server with hot reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

# Documentation
swagger: ## Generate swagger API documentation
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Migration commands (using Atlas OSS with GORM provider)
migrate-diff: ## Create a new migration (usage: make migrate-diff name=migration_name)
	@go run -mod=mod ariga.io/atlas-provider-gorm load --path ./internal/models --dialect postgres > /tmp/gorm_schema.sql
	atlas migrate diff $(name) \
		--dir "file://migrations" \
		--to "file:///tmp/gorm_schema.sql" \
		--dev-url "$(DATABASE_URL)"

migrate-apply-local: ## Apply migrations to local database
	atlas migrate apply \
		--dir "file://migrations" \
		--url "$(DATABASE_URL)" \
		--revisions-schema public \

migrate-apply-staging: ## Apply migrations to staging database
	atlas migrate apply \
		--dir "file://migrations" \
		--url "$(NEON_STAGING_DATABASE_URL)" \
		--revisions-schema public

migrate-apply-prod: ## Apply migrations to production database
	@echo "⚠️  WARNING: Applying to PRODUCTION"
	@read -p "Type 'yes' to confirm: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		atlas migrate apply \
			--dir "file://migrations" \
			--url "$(NEON_PROD_DATABASE_URL)" \
			--revisions-schema public
	else \
		echo "Cancelled"; \
	fi

migrate-status: ## Check migration status
	atlas migrate status \
		--dir "file://migrations" \
		--url "$(DATABASE_URL)"

migrate-lint: ## Lint migrations for potential issues
	atlas migrate lint \
		--dir "file://migrations" \
		--dev-url "$(DATABASE_URL)" \
		--latest 1

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