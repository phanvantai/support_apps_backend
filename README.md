# Support App Backend

A modern, scalable backend system for handling support tickets and feedback from mobile applications.

## Features

- ✅ **RESTful API** for support ticket management
- ✅ **Rate Limiting** to prevent abuse
- ✅ **JWT Authentication** for admin endpoints
- ✅ **PostgreSQL Database** with proper indexing
- ✅ **Clean Architecture** with separation of concerns
- ✅ **Test-Driven Development** with comprehensive test coverage
- ✅ **Docker Support** for easy deployment
- ✅ **Environment Configuration** for different deployments

## Tech Stack

- **Backend**: Go 1.21+ with Gin framework
- **Database**: PostgreSQL 15+
- **Authentication**: JWT tokens
- **ORM**: GORM
- **Testing**: Testify
- **Containerization**: Docker & Docker Compose

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description | Rate Limited |
|--------|----------|-------------|--------------|
| `POST` | `/api/v1/support-request` | Submit a support ticket or feedback | ✅ |
| `GET` | `/health` | Health check endpoint | ❌ |

### Admin Endpoints (Authentication Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/support-requests` | List all support requests (paginated) |
| `GET` | `/api/v1/support-requests/{id}` | Get a specific support request |
| `PATCH` | `/api/v1/support-requests/{id}` | Update request status or add admin notes |
| `DELETE` | `/api/v1/support-requests/{id}` | Delete a support request |

## Data Schema

### Support Request Model

```json
{
  "id": 1,
  "type": "support",
  "user_email": "user@example.com",
  "message": "I need help with...",
  "platform": "iOS",
  "app_version": "1.2.0",
  "device_model": "iPhone 13",
  "status": "new",
  "admin_notes": "Admin response...",
  "created_at": "2025-06-12T10:30:00Z",
  "updated_at": "2025-06-12T10:30:00Z"
}
```

### Field Constraints

- `type`: Must be either `"support"` or `"feedback"`
- `platform`: Must be either `"iOS"` or `"Android"`
- `status`: Must be one of `"new"`, `"in_progress"`, `"resolved"`
- `message`: Required field
- `app_version`: Required field
- `device_model`: Required field
- `user_email`: Optional field

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd support-app-backend
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and secrets
   # For production, use .env.production with secure values
   ```

4. **Start PostgreSQL** (using Docker)

   ```bash
   docker run --name postgres-dev \
     -e POSTGRES_DB=support_app \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=password \
     -p 5432:5432 \
     -d postgres:15-alpine
   ```

5. **Run database migrations**

   ```bash
   make migrate-up
   ```

6. **Run the application**

   ```bash
   make run
   ```

7. **Run tests**

   ```bash
   make test
   ```

### Using Docker Compose

1. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start all services**

   ```bash
   docker-compose up -d
   ```

3. **View logs**

   ```bash
   docker-compose logs -f app
   ```

4. **Stop services**

   ```bash
   docker-compose down
   ```

## API Usage Examples

### Submit a Support Request

```bash
curl -X POST http://localhost:8080/api/v1/support-request \
  -H "Content-Type: application/json" \
  -d '{
    "type": "support",
    "user_email": "user@example.com",
    "message": "I cannot login to my account",
    "platform": "iOS",
    "app_version": "1.2.0",
    "device_model": "iPhone 13"
  }'
```

### Get All Support Requests (Admin)

```bash
curl -X GET http://localhost:8080/api/v1/support-requests \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -G \
  -d "page=1" \
  -d "page_size=20"
```

### Update Support Request Status (Admin)

```bash
curl -X PATCH http://localhost:8080/api/v1/support-requests/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_progress",
    "admin_notes": "Working on this issue"
  }'
```

## JWT Token Generation

For testing admin endpoints, you can generate a JWT token using the following Go code:

```go
package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func main() {
    claims := &JWTClaims{
        UserID:   1,
        Username: "admin",
        Role:     "admin",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("your-secret-key-change-in-production"))
    fmt.Println("JWT Token:", tokenString)
}
```

## Configuration

The application uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_NAME` | Database name | `support_app` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `RATE_LIMIT` | Requests per second limit | `10.0` |
| `RATE_BURST` | Rate limit burst | `20` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key-change-in-production` |

## Security & Environment Variables

### Environment Variable Security

This application uses environment variables to avoid hardcoding sensitive information. **Never commit sensitive data to version control.**

#### Quick Setup

```bash
# Set up environment files
make env-setup

# Generate a secure JWT secret
make env-generate-jwt
```

#### Security Requirements

**Development Environment:**

- Copy `.env.example` to `.env`
- Generate a secure JWT secret (minimum 32 characters)
- Use strong database passwords

**Production Environment:**

- Use `.env.production` or system environment variables
- **Mandatory security validations:**
  - JWT secret must be at least 32 characters
  - Database password must be at least 12 characters
  - SSL must be enabled (`DB_SSLMODE=require`)
  - Cannot use default credentials (`postgres`/`password`)

#### Environment Files

| File | Purpose | Commit to Git |
|------|---------|---------------|
| `.env.example` | Template with placeholder values | ✅ Yes |
| `.env` | Development configuration | ❌ No |
| `.env.production` | Production configuration | ❌ No |

#### Security Validation

The application automatically validates configuration on startup:

```bash
# This will fail with insecure configuration in production
ENVIRONMENT=production JWT_SECRET=weak go run ./cmd
# Error: JWT secret is insecure: must be at least 32 characters
```

#### Generating Secure Secrets

```bash
# Generate JWT secret
./scripts/generate_jwt_secret.sh

# Or manually with openssl
openssl rand -base64 48 | tr -d "=+/" | cut -c1-64
```

### Security Features

1. **Environment Validation**: Automatic security checks on startup
2. **Rate Limiting**: Prevents abuse of public endpoints  
3. **JWT Authentication**: Secures admin endpoints
4. **Input Validation**: Validates all incoming data
5. **SQL Injection Protection**: Uses parameterized queries via GORM
6. **CORS Support**: Configurable cross-origin requests
7. **SSL/TLS Support**: Required for production databases

## Testing

The project follows Test-Driven Development (TDD) principles:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test database interactions
- **Handler Tests**: Test HTTP endpoints

Run tests with coverage:

```bash
make test-coverage
```

## Database Migrations

Run migrations manually:

```bash
# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-status
```

## Monitoring and Health Checks

The application provides a health check endpoint:

```bash
curl http://localhost:8080/health
```

Response:

```json
{
  "status": "healthy",
  "service": "support-app-backend"
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for your changes
4. Implement the feature
5. Ensure all tests pass
6. Submit a pull request

## Production Deployment

1. **Environment Setup**
   - Use strong JWT secrets
   - Configure proper database credentials
   - Set appropriate rate limits
   - Enable SSL/TLS

2. **Database**
   - Use managed PostgreSQL service
   - Set up regular backups
   - Configure connection pooling

3. **Application**
   - Build optimized Docker image
   - Set up load balancer
   - Configure logging and monitoring
   - Implement graceful shutdown

## License

This project is licensed under the MIT License.
