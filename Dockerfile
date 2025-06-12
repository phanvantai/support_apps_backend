# Build stage
FROM golang:1.23-alpine AS builder

# Install ca-certificates and git for dependencies
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations for production
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates and curl for health checks
RUN apk --no-cache add ca-certificates curl

# Create a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/main .

# Copy migrations with proper ownership
COPY --from=builder --chown=appuser:appgroup /app/migrations ./migrations

# Switch to non-root user
USER appuser

# Expose port (Railway will set the PORT environment variable)
EXPOSE 8080

# Health check for Railway
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${PORT:-8080}/health || exit 1

# Run the binary
CMD ["./main"]
