
# 📡 API Reference

Base URL: `http://localhost:8080/api/v1`

## Endpoints

### 1. Create Short URL
```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://example.com/long/url",
  "custom_alias": "summer-sale",        // Optional: 4-20 chars
  "expires_in": "24h",                   // Optional: 5m, 1h, 7d, 30d
  "expires_at": "2024-12-31T23:59:59Z"  // Optional: Specific datetime
}
```

**Response 201:**
```json
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "short_url": "http://localhost:8080/aBc1DeF",
    "original_url": "https://example.com/long/url",
    "created_at": "2024-03-13T10:30:00Z",
    "expires_at": "2024-03-14T10:30:00Z"
  }
}
```

**Errors:**
- `400` - Invalid URL format
- `409` - Custom alias already taken
- `422` - Reserved keyword (api, admin, etc.)
- `429` - Rate limit exceeded (10 req/min)

---

### 2. Redirect
```http
GET /{short_code}
```

**Response:**
- `302 Found` → Redirects to original URL
- `404 Not Found` → URL doesn't exist
- `410 Gone` → URL expired
- `429` - Rate limit (100 req/min)

**Headers set:**
```
Location: https://original-url.com
Cache-Control: no-cache
```

---

### 3. Get URL Details
```http
GET /api/v1/url/{short_code}
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "original_url": "https://example.com",
    "click_count": 142,
    "created_at": "2024-03-13T10:30:00Z",
    "last_accessed_at": "2024-03-15T08:15:00Z",
    "is_active": true,
    "expires_at": null
  }
}
```

---

### 4. Get Analytics
```http
GET /api/v1/url/{short_code}/analytics?days=30
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "total_clicks": 142,
    "analytics": {
      "clicks_by_date": [
        {"date": "2024-03-14", "count": 50},
        {"date": "2024-03-15", "count": 92}
      ],
      "top_countries": [
        {"country": "US", "count": 80},
        {"country": "IN", "count": 62}
      ],
      "devices": {
        "mobile": 85,
        "desktop": 50,
        "tablet": 7
      }
    }
  }
}
```

**Note:** Detailed analytics (countries, devices) are placeholders in current MVP.

---

### 5. Delete URL
```http
DELETE /api/v1/url/{short_code}
```

**Response 200:**
```json
{
  "success": true,
  "message": "URL deleted successfully"
}
```

**Note:** Soft delete - sets `is_active = false`, doesn't remove from DB.

---

### 6. Health Check
```http
GET /health
```

**Response 200:**
```json
{
  "status": "healthy",
  "timestamp": "2024-03-15T10:30:00Z",
  "services": {
    "database": "up",
    "redis": "up"
  }
}
```

---

## Error Response Format

All errors follow this structure:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

**Error Codes:**
- `INVALID_URL` - Malformed URL
- `INVALID_REQUEST` - Missing required fields
- `ALIAS_ALREADY_EXISTS` - Custom alias taken
- `ALIAS_RESERVED` - Reserved keyword used
- `URL_NOT_FOUND` - Short code doesn't exist
- `URL_EXPIRED` - URL past expiration
- `RATE_LIMIT_EXCEEDED` - Too many requests

---

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| POST /shorten | 10 requests | 1 minute |
| GET /{code} | 100 requests | 1 minute |
| Others | No limit | - |

**Rate limit headers:**
```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1710504000
```

---

## Testing with cURL

```bash
# Create short URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://google.com"}'

# Get URL details
curl http://localhost:8080/api/v1/url/abc123

# Test redirect
curl -I http://localhost:8080/abc123

# Get analytics
curl http://localhost:8080/api/v1/url/abc123/analytics

# Delete URL
curl -X DELETE http://localhost:8080/api/v1/url/abc123
```

---

**Complete Postman collection:** [Coming soon]
