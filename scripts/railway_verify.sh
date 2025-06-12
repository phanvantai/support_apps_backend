#!/bin/bash
# Railway Deployment Verification Script
# This script verifies that your app is ready for Railway deployment

set -e

echo "🚂 Railway Deployment Verification"
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check functions
check_file() {
    if [ -f "$1" ]; then
        echo -e "✅ ${GREEN}$1 exists${NC}"
        return 0
    else
        echo -e "❌ ${RED}$1 missing${NC}"
        return 1
    fi
}

check_directory() {
    if [ -d "$1" ]; then
        echo -e "✅ ${GREEN}$1/ directory exists${NC}"
        return 0
    else
        echo -e "❌ ${RED}$1/ directory missing${NC}"
        return 1
    fi
}

# Verification steps
echo "🔍 Checking required files..."
check_file "Dockerfile"
check_file "go.mod"
check_file "go.sum"
check_file "railway.toml"
check_file "RAILWAY_DEPLOY.md"
check_directory "migrations"
check_directory "cmd"

echo ""
echo "🧪 Running tests..."
if go test ./... > /dev/null 2>&1; then
    echo -e "✅ ${GREEN}All tests passing${NC}"
else
    echo -e "❌ ${RED}Tests failing${NC}"
    echo "Run 'make test' to see details"
    exit 1
fi

echo ""
echo "🔧 Testing DATABASE_URL parsing..."
export JWT_SECRET="test-jwt-secret-that-is-long-enough-32-chars"
export DATABASE_URL="postgresql://testuser:testpass@testhost:5432/testdb?sslmode=require"

# Test config loading
if go run -ldflags="-X main.version=test" ./cmd/main.go 2>&1 | grep -q "testuser"; then
    echo -e "✅ ${GREEN}DATABASE_URL parsing works${NC}"
else
    echo -e "❌ ${RED}DATABASE_URL parsing failed${NC}"
    exit 1
fi

echo ""
echo "🐳 Testing Docker build..."
if docker build -t railway-test . > /dev/null 2>&1; then
    echo -e "✅ ${GREEN}Docker build successful${NC}"
    # Cleanup
    docker rmi railway-test > /dev/null 2>&1 || true
else
    echo -e "❌ ${RED}Docker build failed${NC}"
    exit 1
fi

echo ""
echo "🔐 Testing JWT secret validation..."
if JWT_SECRET="short" go run ./cmd/main.go 2>&1 | grep -q "JWT secret is insecure"; then
    echo -e "✅ ${GREEN}JWT validation working${NC}"
else
    echo -e "❌ ${RED}JWT validation not working${NC}"
    exit 1
fi

echo ""
echo "📊 Checking health endpoint..."
# This would require the server to be running, so we'll check the code instead
if grep -q "HealthCheck" internal/handlers/support_request_handler.go; then
    echo -e "✅ ${GREEN}Health check endpoint implemented${NC}"
else
    echo -e "❌ ${RED}Health check endpoint missing${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Railway Deployment Ready!${NC}"
echo ""
echo "📋 Next Steps:"
echo "1. Push your code to GitHub"
echo "2. Create Railway project from GitHub repo"
echo "3. Add PostgreSQL service in Railway"
echo "4. Set environment variables:"
echo "   - JWT_SECRET=\$(openssl rand -hex 32)"
echo "   - ENVIRONMENT=production"
echo "   - GIN_MODE=release"
echo ""
echo "📚 For detailed instructions, see: RAILWAY_DEPLOY.md"
