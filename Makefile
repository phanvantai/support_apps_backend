.PHONY: help build run test test-coverage clean docker-build docker-run migrate-up migrate-down migrate-status env-setup env-generate-jwt

# Default target
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  migrate-up    - Apply database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  migrate-status- Check migration status"
	@echo "  env-setup     - Set up environment files"
	@echo "  env-generate-jwt - Generate secure JWT secret"

# Build the application
build:
	go build -o bin/main ./cmd

# Run the application
run:
	go run ./cmd

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	docker build -t support-app-backend .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	docker-compose down

# Apply database migrations
migrate-up:
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database="postgresql://postgres:password@localhost:5432/support_app?sslmode=disable" \
		up

# Rollback database migrations
migrate-down:
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database="postgresql://postgres:password@localhost:5432/support_app?sslmode=disable" \
		down 1

# Check migration status
migrate-status:
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database="postgresql://postgres:password@localhost:5432/support_app?sslmode=disable" \
		version

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Security check
security:
	gosec ./...

# Generate mocks (if needed)
mocks:
	go generate ./...

# Run integration tests
integration-test:
	./scripts/integration_test.sh

# Seed database with sample data
seed-db:
	./scripts/seed_database.sh

# Run comprehensive demo
demo:
	./scripts/demo.sh

# Generate JWT token for testing
jwt-token:
	go run pkg/jwt_generator.go

# Start development environment
dev-up:
	docker-compose up -d postgres
	sleep 3
	make migrate-up
	make seed-db

# Stop development environment  
dev-down:
	docker-compose down

# Environment setup targets
env-setup:
	@echo "Setting up environment files..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example"; \
		echo "Please edit .env with your configuration"; \
	else \
		echo ".env already exists"; \
	fi
	@echo "Run 'make env-generate-jwt' to generate a secure JWT secret"

# Generate secure JWT secret
env-generate-jwt:
	@./scripts/generate_jwt_secret.sh
