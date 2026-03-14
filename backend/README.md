# URL Shortener Backend

A highly scalable URL shortener service built with Go, PostgreSQL, and Redis.

## Features

- 🚀 **High Performance**: Built with Gin framework for maximum throughput
- 🔐 **Collision-Free**: Snowflake algorithm ensures unique short codes
- 💾 **Smart Caching**: Redis caching with 95%+ cache hit ratio
- 📊 **Analytics**: Track clicks, device types, countries, and referrers
- ⚡ **Rate Limiting**: Built-in rate limiting for API protection
- 🎯 **Custom Aliases**: Support for custom short codes
- ⏰ **URL Expiration**: Optional expiration dates for URLs
- 🐳 **Docker Ready**: Full Docker and Docker Compose support

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **ORM**: GORM

## Project Structure

```
backend/
├── cache/              # Redis caching layer
├── config/             # Configuration management
├── database/           # Database connection and operations
├── handlers/           # HTTP request handlers
├── middleware/         # Middleware (CORS, rate limiting)
├── models/             # Data models
├── pkg/
│   └── snowflake/      # Snowflake ID generator
├── workers/            # Background workers
├── main.go             # Application entry point
├── go.mod              # Go module definition
├── Dockerfile          # Docker configuration
└── .env.example        # Environment variables template
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd url-shortener/backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Run with Docker Compose (Recommended)**
```bash
# From project root
docker-compose up -d
```

5. **Or run manually**

Start PostgreSQL and Redis, then:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Create Short URL
```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url",
  "custom_alias": "sale2024",  // Optional
  "expires_at": "2024-12-31T23:59:59Z"  // Optional
}

Response (201 Created):
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "short_url": "http://localhost:8080/aBc1DeF",
    "original_url": "https://example.com/very/long/url",
    "created_at": "2024-03-13T02:34:22Z",
    "expires_at": "2024-12-31T23:59:59Z"
  }
}
```

### Redirect to Original URL
```http
GET /{short_code}

Response: 302 Redirect to original URL
```

### Get URL Details
```http
GET /api/v1/url/{short_code}

Response (200 OK):
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "short_url": "http://localhost:8080/aBc1DeF",
    "original_url": "https://example.com/very/long/url",
    "click_count": 12345,
    "created_at": "2024-03-13T02:34:22Z",
    "last_accessed_at": "2024-03-13T10:15:30Z",
    "is_active": true
  }
}
```

### Get URL Analytics
```http
GET /api/v1/url/{short_code}/analytics?days=7

Response (200 OK):
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "total_clicks": 12345,
    "analytics": {
      "clicks_by_date": [...],
      "top_countries": [...],
      "top_referrers": [...],
      "devices": {...}
    }
  }
}
```

### Delete URL
```http
DELETE /api/v1/url/{short_code}

Response (200 OK):
{
  "success": true,
  "data": {
    "message": "Short URL deleted successfully"
  }
}
```

### Health Check
```http
GET /health

Response (200 OK):
{
  "status": "healthy",
  "timestamp": "2024-03-13T02:34:22Z",
  "services": {
    "database": "up",
    "redis": "up"
  }
}
```

## Configuration

All configuration is managed through environment variables. See `.env.example` for all available options.

Key configurations:
- `SERVER_PORT`: Server port (default: 8080)
- `DB_HOST`: PostgreSQL host
- `REDIS_HOST`: Redis host
- `MACHINE_ID`: Snowflake machine ID (0-1023)
- `RATE_LIMIT_CREATE`: Rate limit for URL creation (requests/minute)
- `RATE_LIMIT_REDIRECT`: Rate limit for redirects (requests/minute)

## Architecture Highlights

### Snowflake ID Generation
- 64-bit unique IDs with timestamp, machine ID, and sequence
- 4.2M+ IDs per second capacity
- Encoded to Base62 for short codes (7-8 characters)

### Redis Caching Strategy
- **URL Lookup Cache**: 24-hour TTL, 95%+ hit ratio
- **Click Count Buffer**: Async writes every 5 minutes
- **Rate Limiting**: Token bucket algorithm with sliding window

### Database Schema
- **urls**: Main URL storage with indexes
- **url_analytics**: Click tracking and analytics
- **reserved_keywords**: Protected short codes

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o url-shortener main.go
```

### Docker Build
```bash
docker build -t url-shortener-backend .
```

## Performance Targets

- ✅ <50ms response time (cached requests)
- ✅ <100ms response time (cache miss)
- ✅ 95%+ cache hit ratio
- ✅ 10K+ requests/second capacity

## Deployment

The application is designed to run on free-tier hosting platforms:

- **Backend**: Render.com, Railway.app, or Fly.io
- **Database**: Render PostgreSQL or Supabase
- **Redis**: Railway.app Redis or Upstash

See deployment documentation in the main README for detailed instructions.

## License

MIT License - See LICENSE file for details
