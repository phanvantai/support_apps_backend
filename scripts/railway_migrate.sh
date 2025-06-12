#!/bin/bash
# Railway Migration Script
# This script can be run as a Railway build command to apply migrations

set -e

echo "🚂 Railway Migration Script for Support App Backend"

# Check if DATABASE_URL is available
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL environment variable is required for Railway deployments"
    exit 1
fi

echo "✅ DATABASE_URL found, proceeding with migrations..."

# Check if migrate CLI tool is available
if ! command -v migrate &> /dev/null; then
    echo "📦 Installing migrate CLI tool..."
    
    # Download and install migrate for Railway's Ubuntu environment
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
    sudo mv migrate /usr/local/bin/
    chmod +x /usr/local/bin/migrate
    
    echo "✅ migrate CLI tool installed"
fi

# Apply migrations
echo "📊 Applying database migrations..."
migrate -path=./migrations -database="$DATABASE_URL" up

echo "✅ Database migrations completed successfully!"
echo "🌐 Application is ready for Railway deployment"
