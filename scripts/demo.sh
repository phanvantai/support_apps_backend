#!/bin/bash

# Comprehensive demonstration script for the Support App Backend

set -e

echo "🎯 Support App Backend - Complete Demonstration"
echo "==============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# Function to print colored headers
print_header() {
    echo ""
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
}

# Function to print step info
print_step() {
    echo -e "${CYAN}👉 $1${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to wait for user input
wait_for_user() {
    echo ""
    echo -e "${PURPLE}Press Enter to continue...${NC}"
    read -r
}

# Function to check if server is running
check_server() {
    print_step "Checking if server is running..."
    if curl -s "$BASE_URL/health" >/dev/null 2>&1; then
        print_success "Server is running and healthy"
        return 0
    else
        print_error "Server is not running"
        return 1
    fi
}

# Function to generate JWT token
generate_jwt_token() {
    print_step "Generating JWT token for admin access..."
    
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go not found. Cannot generate JWT token."
        return 1
    fi
    
    JWT_TOKEN=$(go run pkg/jwt_generator.go | grep -A1 "JWT Token for testing" | tail -n1)
    if [ -n "$JWT_TOKEN" ]; then
        print_success "JWT token generated successfully"
        echo "Token (first 50 chars): ${JWT_TOKEN:0:50}..."
        return 0
    else
        print_error "Failed to generate JWT token"
        return 1
    fi
}

# Function to demonstrate public API - Submit Support Request
demo_submit_support_request() {
    print_header "📝 DEMO: Submit Support Request (Public API)"
    
    print_step "Submitting a support request..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X POST "$API_URL/support-request" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "support",
            "user_email": "demo@example.com",
            "message": "I am having trouble logging into my account. The app keeps saying my credentials are invalid.",
            "platform": "iOS",
            "app_version": "2.1.0",
            "device_model": "iPhone 14 Pro"
        }')
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "201" ]; then
        print_success "Support request submitted successfully!"
        echo "Response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        
        # Extract ID for later use
        SUPPORT_ID=$(echo "$body" | jq -r '.data.id' 2>/dev/null || echo "1")
    else
        print_error "Failed to submit support request (HTTP $status)"
        echo "$body"
    fi
    
    wait_for_user
}

# Function to demonstrate public API - Submit Feedback
demo_submit_feedback() {
    print_header "💬 DEMO: Submit Feedback (Public API)"
    
    print_step "Submitting feedback without email..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X POST "$API_URL/support-request" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "feedback",
            "message": "Love the new dark mode feature! The app looks fantastic. Would love to see more customization options in the future.",
            "platform": "Android",
            "app_version": "2.0.8",
            "device_model": "Samsung Galaxy S23"
        }')
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "201" ]; then
        print_success "Feedback submitted successfully!"
        echo "Response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        print_error "Failed to submit feedback (HTTP $status)"
        echo "$body"
    fi
    
    wait_for_user
}

# Function to demonstrate rate limiting
demo_rate_limiting() {
    print_header "🚦 DEMO: Rate Limiting (Security Feature)"
    
    print_step "Testing rate limiting by sending multiple requests quickly..."
    
    success_count=0
    rate_limited_count=0
    
    for i in {1..15}; do
        status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/support-request" \
            -H "Content-Type: application/json" \
            -d '{
                "type": "support",
                "message": "Rate limit test request #'$i'",
                "platform": "iOS",
                "app_version": "1.0.0",
                "device_model": "iPhone"
            }')
        
        if [ "$status" = "201" ]; then
            ((success_count++))
            echo -n "✅"
        elif [ "$status" = "429" ]; then
            ((rate_limited_count++))
            echo -n "🚫"
        else
            echo -n "❓"
        fi
        
        # Small delay to avoid overwhelming
        sleep 0.1
    done
    
    echo ""
    print_success "Rate limiting test completed!"
    echo "Successful requests: $success_count"
    echo "Rate limited requests: $rate_limited_count"
    
    if [ "$rate_limited_count" -gt 0 ]; then
        print_success "Rate limiting is working correctly!"
    else
        print_warning "Rate limiting might need adjustment"
    fi
    
    wait_for_user
}

# Function to demonstrate validation
demo_validation() {
    print_header "🔍 DEMO: Input Validation (Security Feature)"
    
    print_step "Testing invalid request (missing required fields)..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X POST "$API_URL/support-request" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "support"
        }')
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "400" ]; then
        print_success "Validation working correctly - invalid request rejected!"
        echo "Error response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        print_warning "Validation might not be working as expected (HTTP $status)"
    fi
    
    print_step "Testing invalid enum value..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X POST "$API_URL/support-request" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "invalid_type",
            "message": "Test message",
            "platform": "Windows",
            "app_version": "1.0.0",
            "device_model": "PC"
        }')
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    
    if [ "$status" = "400" ]; then
        print_success "Enum validation working correctly!"
    else
        print_warning "Enum validation might need attention"
    fi
    
    wait_for_user
}

# Function to demonstrate admin API - Get All Requests
demo_admin_get_all() {
    print_header "👑 DEMO: Admin API - Get All Support Requests"
    
    if [ -z "$JWT_TOKEN" ]; then
        print_error "No JWT token available. Skipping admin demos."
        return 1
    fi
    
    print_step "Fetching all support requests with pagination..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X GET "$API_URL/support-requests?page=1&page_size=5" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "200" ]; then
        print_success "Successfully retrieved support requests!"
        echo "Response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        print_error "Failed to retrieve support requests (HTTP $status)"
        echo "$body"
    fi
    
    wait_for_user
}

# Function to demonstrate admin API - Get Single Request
demo_admin_get_single() {
    print_header "🔍 DEMO: Admin API - Get Single Support Request"
    
    if [ -z "$JWT_TOKEN" ]; then
        print_error "No JWT token available. Skipping admin demos."
        return 1
    fi
    
    local request_id=${SUPPORT_ID:-1}
    print_step "Fetching support request #$request_id..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X GET "$API_URL/support-requests/$request_id" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "200" ]; then
        print_success "Successfully retrieved support request!"
        echo "Response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    elif [ "$status" = "404" ]; then
        print_warning "Support request not found (this is expected if database is empty)"
    else
        print_error "Failed to retrieve support request (HTTP $status)"
        echo "$body"
    fi
    
    wait_for_user
}

# Function to demonstrate admin API - Update Request
demo_admin_update() {
    print_header "✏️ DEMO: Admin API - Update Support Request"
    
    if [ -z "$JWT_TOKEN" ]; then
        print_error "No JWT token available. Skipping admin demos."
        return 1
    fi
    
    local request_id=${SUPPORT_ID:-1}
    print_step "Updating support request #$request_id status and adding admin notes..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X PATCH "$API_URL/support-requests/$request_id" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "status": "in_progress",
            "admin_notes": "We have identified the issue and are working on a fix. Thank you for your patience!"
        }')
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "200" ]; then
        print_success "Successfully updated support request!"
        echo "Response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    elif [ "$status" = "404" ]; then
        print_warning "Support request not found (this is expected if database is empty)"
    else
        print_error "Failed to update support request (HTTP $status)"
        echo "$body"
    fi
    
    wait_for_user
}

# Function to demonstrate unauthorized access
demo_unauthorized_access() {
    print_header "🔒 DEMO: Security - Unauthorized Access Protection"
    
    print_step "Attempting to access admin endpoint without authentication..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X GET "$API_URL/support-requests")
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status" = "401" ]; then
        print_success "Unauthorized access properly blocked!"
        echo "Response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        print_error "Security issue: Unauthorized access not properly blocked (HTTP $status)"
    fi
    
    print_step "Attempting to access admin endpoint with invalid token..."
    
    response=$(curl -s -w "\nSTATUS:%{http_code}" -X GET "$API_URL/support-requests" \
        -H "Authorization: Bearer invalid-token-here")
    
    status=$(echo "$response" | grep "STATUS:" | cut -d: -f2)
    
    if [ "$status" = "401" ]; then
        print_success "Invalid token properly rejected!"
    else
        print_error "Security issue: Invalid token not properly rejected (HTTP $status)"
    fi
    
    wait_for_user
}

# Function to show system architecture
demo_architecture() {
    print_header "🏗️ DEMO: System Architecture Overview"
    
    echo "The Support App Backend follows Clean Architecture principles:"
    echo ""
    echo "┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐"
    echo "│   HTTP Layer    │    │  Business Logic │    │   Data Layer    │"
    echo "│   (Handlers)    │───▶│   (Services)    │───▶│ (Repositories)  │"
    echo "└─────────────────┘    └─────────────────┘    └─────────────────┘"
    echo "         │                       │                       │"
    echo "         ▼                       ▼                       ▼"
    echo "┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐"
    echo "│   Middleware    │    │     Models      │    │   PostgreSQL    │"
    echo "│ (Auth, Rate     │    │   (Domain)      │    │   Database      │"
    echo "│  Limiting)      │    │                 │    │                 │"
    echo "└─────────────────┘    └─────────────────┘    └─────────────────┘"
    echo ""
    echo "Key Features Demonstrated:"
    echo "• ✅ RESTful API Design"
    echo "• ✅ JWT Authentication for Admin Endpoints"
    echo "• ✅ Rate Limiting for Public Endpoints"
    echo "• ✅ Input Validation & Sanitization"
    echo "• ✅ Error Handling & Proper HTTP Status Codes"
    echo "• ✅ Database Abstraction with GORM"
    echo "• ✅ Comprehensive Test Coverage (TDD)"
    echo "• ✅ Docker Support for Easy Deployment"
    echo "• ✅ Environment-based Configuration"
    echo "• ✅ API Documentation"
    echo ""
    
    wait_for_user
}

# Function to show test results
demo_test_results() {
    print_header "🧪 DEMO: Test-Driven Development (TDD)"
    
    print_step "Running all unit tests..."
    
    if go test ./... -v; then
        print_success "All tests passed! The system is built with TDD principles."
    else
        print_error "Some tests failed. This should be investigated."
    fi
    
    print_step "Test coverage includes:"
    echo "• Models with validation logic"
    echo "• Services with business logic"
    echo "• Handlers with HTTP logic"
    echo "• Middleware with security logic"
    echo "• Repository interfaces (integration tests available)"
    
    wait_for_user
}

# Main demonstration flow
main() {
    echo "🚀 Welcome to the Support App Backend Demonstration!"
    echo ""
    echo "This demonstration will show you:"
    echo "1. Public API endpoints (support request submission)"
    echo "2. Security features (rate limiting, validation, auth)"
    echo "3. Admin API endpoints (CRUD operations)"
    echo "4. System architecture and testing"
    echo ""
    echo "Prerequisites:"
    echo "• The server should be running on $BASE_URL"
    echo "• PostgreSQL should be available for full functionality"
    echo "• Go should be installed for JWT token generation"
    echo ""
    
    wait_for_user
    
    # Check if server is running
    if ! check_server; then
        print_error "Please start the server first:"
        echo "  cd /path/to/support-app-backend"
        echo "  go run ./cmd"
        echo ""
        echo "Or with Docker:"
        echo "  docker-compose up"
        exit 1
    fi
    
    # Generate JWT token for admin demos
    generate_jwt_token
    
    # Run demonstrations
    demo_submit_support_request
    demo_submit_feedback
    demo_rate_limiting
    demo_validation
    demo_unauthorized_access
    
    if [ -n "$JWT_TOKEN" ]; then
        demo_admin_get_all
        demo_admin_get_single
        demo_admin_update
    fi
    
    demo_architecture
    demo_test_results
    
    print_header "🎉 DEMONSTRATION COMPLETE!"
    
    print_success "Thank you for exploring the Support App Backend!"
    echo ""
    echo "What you've seen:"
    echo "✅ Complete RESTful API for support ticket management"
    echo "✅ Security features (auth, rate limiting, validation)"
    echo "✅ Clean architecture with comprehensive testing"
    echo "✅ Production-ready features (Docker, monitoring, docs)"
    echo ""
    echo "Next steps:"
    echo "• Review the API documentation (API_DOCUMENTATION.md)"
    echo "• Explore the codebase structure"
    echo "• Run integration tests (scripts/integration_test.sh)"
    echo "• Deploy with Docker Compose"
    echo ""
    echo "For more information, see README.md"
}

# Run the demonstration
main "$@"
