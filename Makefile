.PHONY: help dev-setup dev-up dev-down dev-restart test test-coverage migrate-up migrate-down migrate-status logs clean railway-prepare swagger-docs

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
	@echo "📖 Documentation:"
	@echo "  swagger-docs  - Generate Swagger API documentation"
	@echo ""
	@echo "🚂 Railway Deployment:"
	@echo "  railway-prepare - Prepare for Railway deployment"
	@echo "  railway-verify  - Verify Railway deployment readiness"
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
	@echo ""
	@echo "📋 Pre-deployment checklist:"
	@echo "  ✅ Dockerfile optimized for Railway"
	@echo "  ✅ DATABASE_URL support implemented"
	@echo "  ✅ Health check endpoint available at /health"
	@echo "  ✅ Non-root user configuration for security"
	@echo "  ✅ Production-ready build optimizations"
	@echo ""
	@echo "🔧 Required Railway setup:"
	@echo "  1. Create Railway project from GitHub repository"
	@echo "  2. Add PostgreSQL service (generates DATABASE_URL automatically)"
	@echo "  3. Set required environment variables:"
	@echo "     - JWT_SECRET=\$$(openssl rand -hex 32)"
	@echo "     - ENVIRONMENT=production"
	@echo "     - GIN_MODE=release"
	@echo ""
	@echo "📚 Railway will automatically:"
	@echo "  - Build using Dockerfile"
	@echo "  - Set PORT environment variable"
	@echo "  - Provide DATABASE_URL from PostgreSQL service"
	@echo "  - Enable SSL for database connections"
	@echo "  - Run health checks on /health endpoint"
	@echo ""
	@echo "🔍 Next steps:"
	@echo "  1. Read RAILWAY_DEPLOY.md for detailed instructions"
	@echo "  2. Generate JWT secret: openssl rand -hex 32"
	@echo "  3. Push to GitHub and deploy to Railway"
	@echo ""
	@go mod tidy
	@echo "✅ Dependencies cleaned up and ready for Railway deployment"

# Verify Railway deployment readiness
railway-verify:
	@echo "🔍 Verifying Railway deployment readiness..."
	@./scripts/railway_verify.sh

# === DOCUMENTATION COMMANDS ===

# Generate Swagger API documentation
swagger-docs:
	@echo "📖 Generating Swagger API documentation..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag CLI tool..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Swagger documentation generated in docs/ directory"
	@echo "💡 Start the server and visit http://localhost:8080/swagger/index.html"

# === CLEANUP COMMANDS ===

# Clean up build artifacts and containers
clean:
	@echo "🧹 Cleaning up..."
	@docker-compose down --rmi all --volumes --remove-orphans 2>/dev/null || true
	@docker system prune -f
	@rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete"
