#!/bin/bash

# Generate secure JWT secret
# Usage: ./scripts/generate_jwt_secret.sh

set -e

echo "Generating secure JWT secret..."

# Generate a 64-character random string using openssl
JWT_SECRET=$(openssl rand -base64 48 | tr -d "=+/" | cut -c1-64)

echo ""
echo "Generated JWT Secret:"
echo "JWT_SECRET=$JWT_SECRET"
echo ""
echo "Please update your .env file with this secure secret."
echo "For production, make sure to use a different secret than development."
echo ""
echo "Security Notes:"
echo "- This secret should be at least 32 characters long"
echo "- Keep it confidential and never commit it to version control"
echo "- Use different secrets for development and production"
echo "- Rotate secrets periodically for enhanced security"
