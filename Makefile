.PHONY: help dev-setup dev-up dev-down dev-restart test test-coverage migrate-up migrate-down migrate-status logs clean railway-prepare

# Default target
help:
	@echo "🚀 Support App Backend - Development Commands:"
	@echo ""
	@echo "📦 Development (Docker-based):"
	@echo "  dev-setup     - Initial setup (env files + dependencies)"
	@echo "  dev-up        - Start development environment"
	@echo "  dev-down      - Stop development environment"
	@echo "  dev-restart   - Restart development environment"
	@echo "  logs          - View application logs"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  migrate-up    - Apply database migrations"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-status- Check migration status"
	@echo ""
	@echo "🚂 Railway Deployment:"
	@echo "  railway-prepare - Prepare for Railway deployment"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "  clean         - Clean up build artifacts and containers"

# === DEVELOPMENT COMMANDS ===

# Initial development setup
dev-setup:
	@echo "🔧 Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env 2>/dev/null || echo "DB_URL=postgresql://postgres:password@localhost:5432/support_app?sslmode=disable" > .env; \
		echo "JWT_SECRET=$$(openssl rand -hex 32)" >> .env; \
		echo "GIN_MODE=debug" >> .env; \
		echo "PORT=8080" >> .env; \
		echo "✅ Created .env file with default values"; \
	else \
		echo "ℹ️  .env file already exists"; \
	fi
	@go mod download
	@go mod tidy
	@echo "✅ Development environment setup complete!"
	@echo "💡 Run 'make dev-up' to start the application"

# Start development environment
dev-up:
	@echo "🚀 Starting development environment..."
	@docker-compose up -d postgres
	@echo "⏳ Waiting for database to be ready..."
	@sleep 5
	@$(MAKE) migrate-up
	@docker-compose up -d --build
	@echo "✅ Development environment is running!"
	@echo "🌐 API available at: http://localhost:8080"
	@echo "📊 Run 'make logs' to view application logs"

# Stop development environment
dev-down:
	@echo "⏹️  Stopping development environment..."
	@docker-compose down
	@echo "✅ Development environment stopped"

# Restart development environment
dev-restart: dev-down dev-up

# View application logs
logs:
	@docker-compose logs -f app

# === TESTING COMMANDS ===

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total:

# === DATABASE COMMANDS ===

# Apply database migrations
migrate-up:
	@echo "📈 Applying database migrations..."
	@set -a && source .env && set +a && \
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database="postgresql://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" \
		up
	@echo "✅ Migrations applied successfully"

# Rollback database migrations
migrate-down:
	@echo "📉 Rolling back last migration..."
	@set -a && source .env && set +a && \
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database="postgresql://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" \
		down 1
	@echo "✅ Migration rolled back successfully"

# Check migration status
migrate-status:
	@echo "📊 Checking migration status..."
	@set -a && source .env && set +a && \
	docker run --rm \
		-v $(PWD)/migrations:/migrations \
		--network host \
		migrate/migrate \
		-path=/migrations \
		-database="postgresql://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" \
		version

# === RAILWAY DEPLOYMENT ===

# Prepare for Railway deployment
railway-prepare:
	@echo "🚂 Preparing for Railway deployment..."
	@echo "📋 Pre-deployment checklist:"
	@echo "  ✅ Dockerfile is present"
	@echo "  ✅ Environment variables will be set in Railway dashboard:"
	@echo "     - DATABASE_URL (Railway PostgreSQL)"
	@echo "     - JWT_SECRET (generate with: openssl rand -hex 32)"
	@echo "     - GIN_MODE=release"
	@echo "     - PORT (Railway will set automatically)"
	@echo ""
	@echo "🔧 Railway deployment steps:"
	@echo "  1. Connect your GitHub repository to Railway"
	@echo "  2. Add PostgreSQL service in Railway"
	@echo "  3. Set environment variables in Railway dashboard"
	@echo "  4. Deploy will happen automatically on git push"
	@echo ""
	@echo "📚 Railway will use the Dockerfile to build and deploy"
	@go mod tidy
	@echo "✅ Dependencies cleaned up and ready for deployment"

# === CLEANUP COMMANDS ===

# Clean up build artifacts and containers
clean:
	@echo "🧹 Cleaning up..."
	@docker-compose down --rmi all --volumes --remove-orphans 2>/dev/null || true
	@docker system prune -f
	@rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete"
