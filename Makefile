.PHONY: help install-backend run-backend docker-up docker-down docker-logs test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-backend: ## Install backend dependencies
	cd backend && go mod download

run-backend: ## Run backend locally
	cd backend && go run main.go

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

docker-restart: ## Restart all services
	docker-compose restart

test-backend: ## Run backend tests
	cd backend && go test ./...

test-coverage: ## Run backend tests with coverage
	cd backend && go test -cover ./...

clean: ## Clean up build artifacts
	cd backend && rm -f url-shortener
	docker-compose down -v

db-reset: ## Reset database (removes all data)
	docker-compose down -v
	docker-compose up -d postgres redis
	@echo "Waiting for services to be ready..."
	@sleep 5

backend-build: ## Build backend binary
	cd backend && go build -o url-shortener main.go

backend-docker: ## Build backend Docker image
	cd backend && docker build -t url-shortener-backend .

setup: install-backend docker-up ## Complete setup (install + docker up)
	@echo "Setup complete! Backend running at http://localhost:8080"
	@echo "Check health: curl http://localhost:8080/health"
