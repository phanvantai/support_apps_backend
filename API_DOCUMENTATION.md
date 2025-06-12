# Support App Backend API Documentation

## Base URL

```bash
http://localhost:8080
```

## Authentication

Admin endpoints require JWT authentication. Include the token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

Public endpoints are rate-limited to prevent abuse:

- **Rate**: 10 requests per second
- **Burst**: 20 requests maximum
- **Scope**: Per IP address

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:

- `400` - Bad Request (validation errors)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `429` - Too Many Requests (rate limited)
- `500` - Internal Server Error

---

## Endpoints

### Health Check

#### GET /health

Check if the API is running and healthy.

**Authentication**: None required

**Example Request:**

```bash
curl http://localhost:8080/health
```

**Example Response:**

```json
{
  "status": "healthy",
  "service": "support-app-backend"
}
```

---

### Submit Support Request

#### POST /api/v1/support-request

Submit a new support ticket or feedback request.

**Authentication**: None required (public endpoint)
**Rate Limited**: Yes

**Request Body:**

```json
{
  "type": "support|feedback",           // Required
  "user_email": "user@example.com",     // Optional
  "message": "Your message here",       // Required
  "platform": "iOS|Android",           // Required
  "app_version": "1.2.0",              // Required
  "device_model": "iPhone 13"          // Required
}
```

**Example Request:**

```bash
curl -X POST http://localhost:8080/api/v1/support-request \
  -H "Content-Type: application/json" \
  -d '{
    "type": "support",
    "user_email": "user@example.com",
    "message": "I cannot login to my account",
    "platform": "iOS",
    "app_version": "2.1.0",
    "device_model": "iPhone 14 Pro"
  }'
```

**Example Response:**

```json
{
  "data": {
    "id": 1,
    "type": "support",
    "user_email": "user@example.com",
    "message": "I cannot login to my account",
    "platform": "iOS",
    "app_version": "2.1.0",
    "device_model": "iPhone 14 Pro",
    "status": "new",
    "admin_notes": null,
    "created_at": "2025-06-12T10:30:00Z",
    "updated_at": "2025-06-12T10:30:00Z"
  }
}
```

**Validation Rules:**

- `type`: Must be either "support" or "feedback"
- `message`: Required, cannot be empty
- `platform`: Must be either "iOS" or "Android"
- `app_version`: Required, cannot be empty
- `device_model`: Required, cannot be empty
- `user_email`: Optional, must be valid email format if provided

---

### Get All Support Requests (Admin)

#### GET /api/v1/support-requests

Retrieve all support requests with pagination.

**Authentication**: Required (Admin only)

**Query Parameters:**

- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)

**Example Request:**

```bash
curl -X GET "http://localhost:8080/api/v1/support-requests?page=1&page_size=10" \
  -H "Authorization: Bearer <your-jwt-token>"
```

**Example Response:**

```json
{
  "data": [
    {
      "id": 1,
      "type": "support",
      "user_email": "user@example.com",
      "message": "I cannot login to my account",
      "platform": "iOS",
      "app_version": "2.1.0",
      "device_model": "iPhone 14 Pro",
      "status": "new",
      "admin_notes": null,
      "created_at": "2025-06-12T10:30:00Z",
      "updated_at": "2025-06-12T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### Get Single Support Request (Admin)

#### GET /api/v1/support-requests/{id}

Retrieve a specific support request by ID.

**Authentication**: Required (Admin only)

**Path Parameters:**

- `id`: Support request ID (integer)

**Example Request:**

```bash
curl -X GET http://localhost:8080/api/v1/support-requests/1 \
  -H "Authorization: Bearer <your-jwt-token>"
```

**Example Response:**

```json
{
  "data": {
    "id": 1,
    "type": "support",
    "user_email": "user@example.com",
    "message": "I cannot login to my account",
    "platform": "iOS",
    "app_version": "2.1.0",
    "device_model": "iPhone 14 Pro",
    "status": "new",
    "admin_notes": null,
    "created_at": "2025-06-12T10:30:00Z",
    "updated_at": "2025-06-12T10:30:00Z"
  }
}
```

---

### Update Support Request (Admin)

#### PATCH /api/v1/support-requests/{id}

Update the status or admin notes of a support request.

**Authentication**: Required (Admin only)

**Path Parameters:**

- `id`: Support request ID (integer)

**Request Body:**

```json
{
  "status": "new|in_progress|resolved",  // Optional
  "admin_notes": "Admin response"        // Optional
}
```

**Example Request:**

```bash
curl -X PATCH http://localhost:8080/api/v1/support-requests/1 \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_progress",
    "admin_notes": "We are investigating this issue and will get back to you soon."
  }'
```

**Example Response:**

```json
{
  "data": {
    "id": 1,
    "type": "support",
    "user_email": "user@example.com",
    "message": "I cannot login to my account",
    "platform": "iOS",
    "app_version": "2.1.0",
    "device_model": "iPhone 14 Pro",
    "status": "in_progress",
    "admin_notes": "We are investigating this issue and will get back to you soon.",
    "created_at": "2025-06-12T10:30:00Z",
    "updated_at": "2025-06-12T11:15:00Z"
  }
}
```

**Validation Rules:**

- `status`: Must be one of "new", "in_progress", or "resolved"
- At least one field (`status` or `admin_notes`) must be provided

---

### Delete Support Request (Admin)

#### DELETE /api/v1/support-requests/{id}

Soft delete a support request.

**Authentication**: Required (Admin only)

**Path Parameters:**

- `id`: Support request ID (integer)

**Example Request:**

```bash
curl -X DELETE http://localhost:8080/api/v1/support-requests/1 \
  -H "Authorization: Bearer <your-jwt-token>"
```

**Example Response:**

```bash
HTTP 204 No Content
```

---

## JWT Token Generation

For testing admin endpoints, you can generate a JWT token using the included utility:

```bash
go run pkg/jwt_generator.go
```

This will output a token that you can use for testing admin endpoints.

**Example Output:**

```bash
JWT Token for testing admin endpoints:
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Use this token in the Authorization header:
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Example curl command:
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." http://localhost:8080/api/v1/support-requests
```

---

## Data Types

### SupportRequestType

- `support` - Support ticket for issues or problems
- `feedback` - General feedback or suggestions

### Platform

- `iOS` - Apple iOS devices
- `Android` - Android devices

### Status

- `new` - Newly submitted request (default)
- `in_progress` - Currently being worked on by support team
- `resolved` - Issue has been resolved or feedback acknowledged

---

## Examples Collection

### Postman Collection

You can import this JSON into Postman for easy testing:

```json
{
  "info": {
    "name": "Support App Backend API",
    "description": "Collection for testing the Support App Backend API"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080"
    },
    {
      "key": "jwtToken",
      "value": "your-jwt-token-here"
    }
  ],
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "url": "{{baseUrl}}/health"
      }
    },
    {
      "name": "Submit Support Request",
      "request": {
        "method": "POST",
        "url": "{{baseUrl}}/api/v1/support-request",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"type\": \"support\",\n  \"user_email\": \"test@example.com\",\n  \"message\": \"Test support request\",\n  \"platform\": \"iOS\",\n  \"app_version\": \"1.0.0\",\n  \"device_model\": \"iPhone 13\"\n}"
        }
      }
    },
    {
      "name": "Get All Support Requests",
      "request": {
        "method": "GET",
        "url": "{{baseUrl}}/api/v1/support-requests",
        "header": [
          {
            "key": "Authorization",
            "value": "Bearer {{jwtToken}}"
          }
        ]
      }
    }
  ]
}
```

---

## Error Handling Examples

### Validation Error (400)

```bash
curl -X POST http://localhost:8080/api/v1/support-request \
  -H "Content-Type: application/json" \
  -d '{"type": "invalid"}'
```

**Response:**

```json
{
  "error": "Key: 'CreateSupportRequestRequest.Type' Error:Tag: 'oneof' ..."
}
```

### Unauthorized (401)

```bash
curl -X GET http://localhost:8080/api/v1/support-requests
```

**Response:**

```json
{
  "error": "Authorization header required"
}
```

### Rate Limited (429)

```bash
# After exceeding rate limit
curl -X POST http://localhost:8080/api/v1/support-request \
  -H "Content-Type: application/json" \
  -d '{"type": "support", "message": "test", "platform": "iOS", "app_version": "1.0", "device_model": "iPhone"}'
```

**Response:**

```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```
