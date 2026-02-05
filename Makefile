.PHONY: help build dev prod test clean logs backup restore

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Environment file created. Please edit .env with your configuration."

dev-start: ## Start development databases only
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Development databases started (MongoDB: 27017, Redis: 6379)"

dev-stop: ## Stop development databases
	docker-compose -f docker-compose.dev.yml down
	@echo "Development databases stopped"

dev-logs: ## Show development logs
	docker-compose -f docker-compose.dev.yml logs -f

dev-reset: ## Reset development environment
	docker-compose -f docker-compose.dev.yml down -v
	@echo "Development environment reset (all data cleared)"

# Application targets
build: ## Build the application
	go mod tidy
	go build -o bin/wedding-api cmd/api/main.go

run: ## Run the application locally (requires databases to be running)
	go run cmd/api/main.go

run-docker: ## Run application with Docker (full stack)
	docker-compose up --build

run-detached: ## Run application in detached mode
	docker-compose up -d --build

# Production targets
prod-build: ## Build production image
	docker build -t wedding-api:latest .

prod-start: ## Start production environment
	docker-compose --profile production up -d --build

prod-stop: ## Stop production environment
	docker-compose --profile production down

prod-logs: ## Show production logs
	docker-compose --profile production logs -f

# Testing targets
test: ## Run unit tests
	go test ./...

test-integration: ## Run integration tests (requires databases)
	docker-compose -f docker-compose.yml up -d mongodb redis
	@sleep 10
	go test -tags=integration ./...
	docker-compose -f docker-compose.yml down

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Database targets
db-connect: ## Connect to MongoDB
	docker exec -it wedding-mongodb mongosh -u admin -p password123 --authenticationDatabase admin

db-backup: ## Backup database
	docker exec wedding-mongodb mongodump --uri="mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin" -o /backup
	docker cp wedding-mongodb:/backup ./backup-$(shell date +%Y%m%d_%H%M%S)

db-restore: ## Restore database (usage: make db-restore BACKUP_DIR=backup_20231201_120000)
	@if [ -z "$(BACKUP_DIR)" ]; then echo "Usage: make db-restore BACKUP_DIR=backup_directory"; exit 1; fi
	docker cp $(BACKUP_DIR) wedding-mongodb:/tmp/restore
	docker exec wedding-mongodb mongorestore --uri="mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin" --drop /tmp/restore/wedding_invitations

# Utility targets
clean: ## Clean up Docker resources
	docker-compose down -v --remove-orphans
	docker system prune -f
	@echo "Docker resources cleaned up"

logs: ## Show application logs
	docker-compose logs -f app

status: ## Show container status
	docker-compose ps

health: ## Check application health
	curl -f http://localhost:8080/health || echo "Application not responding"

# Deployment targets
deploy-staging: ## Deploy to staging
	@echo "Deploying to staging environment..."
	# Add staging deployment commands here

deploy-prod: ## Deploy to production
	@echo "Deploying to production environment..."
	# Add production deployment commands here

# Development workflow targets
setup: dev-setup build ## Complete development setup
	@echo "Development setup complete!"
	@echo "Run 'make dev-start' to start databases"
	@echo "Run 'make run' to start the application"

reset: dev-stop dev-reset ## Reset development environment
	@echo "Development environment reset!"

workflow-test: ## Run full test workflow
	@echo "Running test workflow..."
	make clean
	make dev-start
	@sleep 15
	make test-integration
	make dev-stop
	@echo "Test workflow completed!"

# Monitoring targets
monitor: ## Monitor resource usage
	docker stats

top: ## Show running processes
	docker-compose top

# Security targets
security-scan: ## Run security scan on dependencies
	go list -json -m all | nancy sleuth

lint: ## Run code linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

# Quick commands for local development
q: ## Quick start (dev databases + app)
	make dev-start && sleep 10 && make run

q-stop: ## Quick stop
	make dev-stop
	pkill -f "wedding-api" || true

# Generate documentation
docs: ## Generate API documentation
	swag init -g cmd/api/main.go -o docs/

# Version and info
version: ## Show version info
	@echo "Git commit: $(shell git rev-parse --short HEAD)"
	@echo "Git branch: $(shell git branch --show-current)"
	@echo "Build date: $(shell date)"
	@echo "Go version: $(shell go version)"

info: ## Show project information
	@echo "Wedding Invitation Backend"
	@echo "=========================="
	@echo "Ports:"
	@echo "  API:      http://localhost:8080"
	@echo "  MongoDB:  localhost:27017"
	@echo "  Redis:    localhost:6379"
	@echo "  Nginx:    http://localhost (production)"
	@echo ""
	@echo "Environment files:"
	@echo "  .env.example - Template configuration"
	@echo "  .env        - Local development"
	@echo "  .env.production - Production"
	@echo ""
	@echo "Use 'make help' to see all available commands"