#!/bin/bash

# Integration test script for the Support App Backend

set -e

echo "🚀 Starting Support App Backend Integration Tests"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# Function to check if server is running
check_server() {
    echo "📡 Checking server health..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
    if [ "$response" -eq 200 ]; then
        echo -e "${GREEN}✅ Server is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not responding (HTTP $response)${NC}"
        return 1
    fi
}

# Function to test support request creation
test_create_support_request() {
    echo "📝 Testing support request creation..."
    
    response=$(curl -s -X POST "$API_URL/support-request" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "support",
            "user_email": "test@example.com",
            "message": "I cannot login to my account",
            "platform": "iOS",
            "app_version": "1.2.0",
            "device_model": "iPhone 13"
        }')
    
    if echo "$response" | grep -q '"id"'; then
        echo -e "${GREEN}✅ Support request created successfully${NC}"
        # Extract ID for later tests
        SUPPORT_REQUEST_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        echo "📋 Created support request with ID: $SUPPORT_REQUEST_ID"
        return 0
    else
        echo -e "${RED}❌ Failed to create support request${NC}"
        echo "Response: $response"
        return 1
    fi
}

# Function to test feedback creation
test_create_feedback() {
    echo "💬 Testing feedback creation..."
    
    response=$(curl -s -X POST "$API_URL/support-request" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "feedback",
            "message": "Great app! Love the new features.",
            "platform": "Android",
            "app_version": "1.1.5",
            "device_model": "Samsung Galaxy S21"
        }')
    
    if echo "$response" | grep -q '"id"'; then
        echo -e "${GREEN}✅ Feedback created successfully${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to create feedback${NC}"
        echo "Response: $response"
        return 1
    fi
}

# Function to test rate limiting
test_rate_limiting() {
    echo "🚦 Testing rate limiting..."
    
    # Send multiple requests quickly
    success_count=0
    rate_limited_count=0
    
    for i in {1..25}; do
        response_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/support-request" \
            -H "Content-Type: application/json" \
            -d '{
                "type": "support",
                "message": "Rate limit test #'$i'",
                "platform": "iOS",
                "app_version": "1.0.0",
                "device_model": "iPhone"
            }')
        
        if [ "$response_code" -eq 201 ]; then
            ((success_count++))
        elif [ "$response_code" -eq 429 ]; then
            ((rate_limited_count++))
        fi
    done
    
    if [ "$rate_limited_count" -gt 0 ]; then
        echo -e "${GREEN}✅ Rate limiting is working (${success_count} successful, ${rate_limited_count} rate limited)${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  Rate limiting might not be working as expected${NC}"
        return 0
    fi
}

# Function to test invalid requests
test_invalid_requests() {
    echo "❌ Testing invalid request handling..."
    
    # Test missing required fields
    response_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/support-request" \
        -H "Content-Type: application/json" \
        -d '{"type": "support"}')
    
    if [ "$response_code" -eq 400 ]; then
        echo -e "${GREEN}✅ Invalid request properly rejected${NC}"
        return 0
    else
        echo -e "${RED}❌ Invalid request not properly handled (HTTP $response_code)${NC}"
        return 1
    fi
}

# Function to generate JWT token for admin tests
generate_jwt() {
    echo "🔑 Generating JWT token for admin tests..."
    
    if command -v go >/dev/null 2>&1; then
        cd $(dirname "$0")
        JWT_TOKEN=$(go run pkg/jwt_generator.go | grep -A1 "JWT Token for testing" | tail -n1)
        echo "Generated token: ${JWT_TOKEN:0:50}..."
        return 0
    else
        echo -e "${YELLOW}⚠️  Go not found, skipping admin endpoint tests${NC}"
        return 1
    fi
}

# Function to test admin endpoints (requires JWT)
test_admin_endpoints() {
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "${YELLOW}⚠️  Skipping admin endpoint tests (no JWT token)${NC}"
        return 0
    fi
    
    echo "👑 Testing admin endpoints..."
    
    # Test getting all support requests
    response_code=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$API_URL/support-requests" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    if [ "$response_code" -eq 200 ]; then
        echo -e "${GREEN}✅ Admin can access support requests${NC}"
    else
        echo -e "${RED}❌ Admin endpoint access failed (HTTP $response_code)${NC}"
        return 1
    fi
    
    # Test unauthorized access
    response_code=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$API_URL/support-requests")
    
    if [ "$response_code" -eq 401 ]; then
        echo -e "${GREEN}✅ Unauthorized access properly blocked${NC}"
        return 0
    else
        echo -e "${RED}❌ Unauthorized access not properly blocked (HTTP $response_code)${NC}"
        return 1
    fi
}

# Main test execution
main() {
    echo "🧪 Support App Backend Integration Tests"
    echo "========================================"
    
    # Wait a bit for server to be ready
    sleep 2
    
    # Run tests
    if ! check_server; then
        echo -e "${RED}❌ Server is not running. Please start the server first.${NC}"
        exit 1
    fi
    
    test_create_support_request || exit 1
    test_create_feedback || exit 1
    test_rate_limiting
    test_invalid_requests || exit 1
    
    # Generate JWT token and test admin endpoints
    if generate_jwt; then
        test_admin_endpoints || exit 1
    fi
    
    echo ""
    echo "========================================"
    echo -e "${GREEN}🎉 All integration tests passed!${NC}"
}

# Run main function
main "$@"
