# Security Guidelines

## Environment Variables Security

### Overview

This document outlines security best practices for managing environment variables and sensitive configuration in the Support App Backend.

### Environment Files Structure

```bash
.env.example      # Template (committed to git)
.env             # Development (never commit)
.env.production  # Production (never commit)
```

### Security Validation

The application performs automatic security validation on startup:

#### Development Environment

- Minimal validation for ease of development
- Warnings for weak configurations
- Allows default values for quick setup

#### Production Environment

- **Strict validation enforced**
- Application will refuse to start with insecure configuration
- Mandatory requirements:
  - JWT secret ≥ 32 characters
  - Database password ≥ 12 characters  
  - SSL enabled (`DB_SSLMODE=require`)
  - No default credentials

### Configuration Checklist

#### ✅ Development Setup

```bash
# 1. Copy environment template
cp .env.example .env

# 2. Generate secure JWT secret
make env-generate-jwt

# 3. Update .env with generated secret
# Edit .env and replace JWT_SECRET value

# 4. Customize other values as needed
```

#### ✅ Production Deployment

```bash
# 1. Create production environment file
cp .env.example .env.production

# 2. Generate production-grade secrets
./scripts/generate_jwt_secret.sh

# 3. Set secure database credentials
# - Use strong passwords (≥12 characters)
# - Create dedicated database user (not 'postgres')
# - Enable SSL connections

# 4. Validate configuration
ENVIRONMENT=production go run ./cmd
```

### Security Features

#### Environment Validation

- Automatic checks on application startup
- Fails fast with insecure configuration
- Different validation rules per environment

#### Sensitive Data Protection

- Environment variables for all secrets
- `.gitignore` prevents accidental commits
- Example files show structure without exposing values

#### Database Security

- SSL/TLS required in production
- No default credentials in production
- Connection string validation

#### JWT Security

- Minimum secret length enforcement
- No default secrets in production
- Configurable token expiration

### Common Security Issues

#### ❌ Avoid These Mistakes

```bash
# DON'T: Commit .env files
git add .env  # Never do this!

# DON'T: Use weak secrets in production
JWT_SECRET=123456  # Too short and weak

# DON'T: Use default credentials in production
DB_USER=postgres
DB_PASSWORD=password

# DON'T: Disable SSL in production
DB_SSLMODE=disable
```

#### ✅ Best Practices

```bash
# DO: Use .env.example for templates
cp .env.example .env

# DO: Generate strong secrets
JWT_SECRET=$(openssl rand -base64 48 | tr -d "=+/" | cut -c1-64)

# DO: Use dedicated database users
DB_USER=support_app_user
DB_PASSWORD=super_secure_random_password_2024

# DO: Enable SSL in production
DB_SSLMODE=require
```

### Environment Variable Reference

| Variable | Required | Production Requirements | Example |
|----------|----------|-------------------------|---------|
| `JWT_SECRET` | Yes | ≥32 chars, non-default | `abc123...` (64 chars) |
| `DB_PASSWORD` | Yes | ≥12 chars, non-default | `SecurePass123!` |
| `DB_USER` | Yes | Non-default in prod | `support_app_user` |
| `DB_SSLMODE` | Yes | `require` in prod | `require` |
| `ENVIRONMENT` | Yes | `production` | `production` |

### Troubleshooting

#### Configuration Validation Errors

```bash
# Error: JWT secret is insecure
# Solution: Generate a new secret
make env-generate-jwt

# Error: Database password is insecure for production
# Solution: Set a strong password (≥12 characters)
DB_PASSWORD=your_strong_password_here

# Error: SSL must be enabled for production
# Solution: Enable SSL mode
DB_SSLMODE=require

# Error: Default database user should not be used
# Solution: Create a dedicated user
DB_USER=support_app_user
```

#### Environment Loading Issues

```bash
# Check if .env file exists
ls -la .env

# Verify environment variables are loaded
go run -ldflags="-X main.version=dev" ./cmd

# Test configuration validation
ENVIRONMENT=production go run ./cmd
```

### Security Monitoring

#### What to Monitor

- Failed authentication attempts
- Rate limit violations
- Configuration validation failures
- Database connection errors

#### Logging Security Events

- Never log sensitive values (passwords, JWT secrets)
- Log security validation results
- Monitor for configuration drift

### Additional Resources

- [OWASP Environment Variable Security](https://owasp.org/www-community/vulnerabilities/Insecure_Configuration_Management)
- [12-Factor App Config](https://12factor.net/config)
- [JWT Security Best Practices](https://tools.ietf.org/html/rfc8725)
