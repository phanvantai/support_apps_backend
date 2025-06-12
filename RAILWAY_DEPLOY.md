# Railway Deployment Guide

## Quick Deploy to Railway

### 1. Prerequisites

- GitHub repository with your code pushed
- Railway account (sign up at [railway.app](https://railway.app))

### 2. Deploy Steps

#### Option A: One-Click Deploy (Recommended)

1. Go to [Railway](https://railway.app)
2. Click "Start a New Project"
3. Select "Deploy from GitHub repo"
4. Choose this repository
5. Railway will automatically detect the Dockerfile and start building

#### Option B: Railway CLI

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Initialize project in your app directory
cd /path/to/support_apps_backend
railway init

# Deploy
railway up
```

### 3. Add PostgreSQL Database

1. In Railway dashboard, click "Add Service"
2. Select "PostgreSQL"
3. Railway will automatically create `DATABASE_URL` environment variable
4. **Important**: Your app will automatically connect using this DATABASE_URL

### 4. Set Environment Variables

#### Required Variables (Set in Railway Dashboard > Variables)

```bash
# CRITICAL: Generate a secure JWT secret (required for production)
JWT_SECRET=your-64-character-secret-here

# Production environment settings
ENVIRONMENT=production
GIN_MODE=release
```

**Note:** When using Railway's PostgreSQL service via `DATABASE_URL`, the app automatically bypasses individual database credential validation since Railway generates secure credentials automatically.

#### Generate Secure JWT Secret

```bash
# Generate a secure 64-character JWT secret
openssl rand -hex 32

# Example output: 3f2504e04f8911...
# Copy this value to Railway's JWT_SECRET variable
```

#### Optional Variables (defaults work for most cases)

```bash
# Railway automatically sets PORT - usually not needed
PORT=8080

# Rate limiting (optional)
RATE_LIMIT=10.0
RATE_BURST=20
```

### 5. Database Setup

#### Automatic Database Configuration

- Railway PostgreSQL service automatically provides `DATABASE_URL`
- Your app will parse this URL and configure the database connection
- SSL is automatically enabled for Railway PostgreSQL

#### Database Migrations

Railway will automatically run migrations during deployment via the Dockerfile.

**Manual Migration (if needed):**

```bash
# Using Railway CLI
railway run ./scripts/railway_migrate.sh

# Or connect to Railway PostgreSQL directly
railway connect postgresql
```

### 6. Health Check & Monitoring

- Health check endpoint: `https://your-app.railway.app/health`
- Railway provides built-in monitoring for CPU, memory, and network
- View logs in Railway dashboard or via CLI: `railway logs`

### 7. Domain Setup

- **Railway Domain**: Automatically provided at `https://your-app.railway.app`
- **Custom Domain**: Add in Railway Dashboard > Settings > Domains

## Environment Variables Reference

| Variable | Required | Railway Handling | Description | Example |
|----------|----------|------------------|-------------|---------|
| `DATABASE_URL` | ✅ | Auto-generated | PostgreSQL connection URL | `postgresql://user:pass@host:5432/db` |
| `JWT_SECRET` | ✅ | Manual setup | 64-character secret for JWT | Use `openssl rand -hex 32` |
| `ENVIRONMENT` | ✅ | Manual setup | Set to `production` | `production` |
| `GIN_MODE` | ✅ | Manual setup | Set to `release` | `release` |
| `PORT` | ❌ | Auto-generated | Server port (Railway managed) | `8080` |
| `RATE_LIMIT` | ❌ | Optional | API rate limit | `10.0` |
| `RATE_BURST` | ❌ | Optional | Rate limit burst | `20` |

## Railway-Specific Features

### Automatic Deployments

- **Git Integration**: Every push to main branch triggers deployment
- **Zero Downtime**: Rolling deployments with health checks
- **Rollback**: Instant rollback to previous versions

### Database Features

- **Automatic Backups**: Daily backups included
- **Connection Pooling**: Built into Railway PostgreSQL
- **SSL/TLS**: Always enabled for security

### Scaling & Performance

- **Vertical Scaling**: Upgrade CPU/RAM in dashboard
- **Multiple Regions**: Deploy closer to users
- **CDN**: Static assets served via Railway's edge network

## Cost & Billing

### Free Tier

- $5/month in credits (enough for hobby projects)
- Automatic sleep after inactivity
- 500MB storage included

### Paid Plans

- **Developer**: $20/month for active development
- **Team**: $99/month for production workloads
- **Pro**: Custom pricing for enterprise

### Cost Optimization Tips

```bash
# Monitor resource usage
railway status

# Check current usage
railway usage

# Scale down during low traffic
# Use Railway's auto-scaling features
```

## Troubleshooting

### Common Issues

#### Build Failures

```bash
# Check build logs
railway logs --deployment

# Common fixes:
# 1. Ensure Dockerfile is in root directory
# 2. Check go.mod dependencies
# 3. Verify all source files are committed
```

#### Database Connection Issues

```bash
# Verify DATABASE_URL is set
railway variables

# Test database connection
railway run go run cmd/main.go

# Check PostgreSQL service status
railway status postgresql
```

#### Environment Configuration

```bash
# List all environment variables
railway variables

# Add missing variables
railway variables set JWT_SECRET=your-secret-here
railway variables set ENVIRONMENT=production
railway variables set GIN_MODE=release
```

#### SSL/TLS Issues

```bash
# Railway PostgreSQL always uses SSL
# Ensure your app handles SSL correctly
# DATABASE_URL automatically includes sslmode=require
```

### Debug Commands

```bash
# View real-time logs
railway logs -f

# Connect to your app's environment
railway shell

# Run one-off commands
railway run <command>

# Check service health
curl https://your-app.railway.app/health
```

### Performance Optimization

- **Database Indexing**: Ensure proper indexes for queries
- **Connection Pooling**: Already configured in the app
- **Resource Monitoring**: Use Railway dashboard metrics
- **Caching**: Consider adding Redis for session storage

## Security Best Practices

### Environment Variables

- **Never commit secrets** to version control
- **Use Railway's variable management** for sensitive data
- **Rotate JWT secrets** regularly in production

### Database Security

- **SSL enforced** automatically by Railway
- **Network isolation** between services
- **Regular backups** with point-in-time recovery

### Application Security

- **Rate limiting** configured for API endpoints
- **Input validation** with Gin binding
- **CORS properly configured** for web clients

## Support & Resources

### Railway Documentation

- [Railway Docs](https://docs.railway.app/)
- [Go Deployment Guide](https://docs.railway.app/deploy/go)
- [PostgreSQL Guide](https://docs.railway.app/databases/postgresql)

### Project Resources

- **Health Check**: `GET /health`
- **API Documentation**: `GET /swagger/index.html`
- **Repository**: Link to your GitHub repo

### Getting Help

- **Railway Discord**: Active community support
- **GitHub Issues**: For app-specific issues
- **Railway Support**: For platform issues
