
# URL Shortener

A scalable URL shortening service with Snowflake ID generation, Redis caching, and analytics.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)](https://postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D?logo=redis)](https://redis.io/)

**GitHub:** [akushagra05/url-shortener](https://github.com/akushagra05/url-shortener)

---

## Quick Start

```bash
docker-compose up -d
```

Backend: http://localhost:8080 | Frontend: http://localhost:5173

---

## Features

- Short URL generation (7-8 characters)
- Custom aliases and URL expiration
- Click tracking and analytics
- Rate limiting (10 creates/min, 100 redirects/min)
- Redis caching for sub-50ms response time

---

## Architecture

```
Clients → Gin Router → Handlers → Services → Repository
                                              ↓
                                    Redis + PostgreSQL
```

**Core components:**
- **Snowflake IDs:** Distributed ID generation, 4.2M IDs/sec capacity
- **Two-tier cache:** Redis (hot data) + PostgreSQL (persistent storage)
- **Async processing:** Background worker syncs click counts every 5 minutes
- **Rate limiting:** Redis-backed token bucket algorithm

📖 **[Read detailed architecture docs →](docs/ARCHITECTURE.md)**

---

## API

```bash
# Create short URL
POST /api/v1/shorten
{
  "url": "https://example.com",
  "custom_alias": "sale",      # Optional
  "expires_in": "24h"           # Optional: 5m, 1h, 7d, 30d
}

# Redirect
GET /{short_code}              # Returns 302 Found

# Get analytics
GET /api/v1/url/{short_code}/analytics

# Health check
GET /health
```

📖 **[Complete API documentation →](docs/API.md)**

---

## Performance

| Metric | Result |
|--------|--------|
| Redirect latency (cached) | 5-10ms |
| Redirect latency (DB) | 30-50ms |
| Cache hit ratio | 95%+ |
| Throughput | 10K req/s |

---

## Tech Stack

- **Backend:** Go 1.21, Gin, GORM
- **Database:** PostgreSQL 15
- **Cache:** Redis 7
- **Frontend:** React 18, Vite, Tailwind CSS
- **Deployment:** Docker Compose

---

## Setup

### Docker (30 seconds)
```bash
docker-compose up -d
```

### Manual Setup
```bash
# Backend
cd backend && go run main.go

# Frontend  
cd frontend && npm run dev
```

📖 **[Detailed setup guide →](docs/SETUP.md)**

---

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - System design, scalability, trade-offs
- **[API Reference](docs/API.md)** - Endpoints, errors, rate limits
- **[Setup Guide](docs/SETUP.md)** - Docker, manual, troubleshooting
- **[Screenshots](docs/SCREENSHOTS.md)** - Visual documentation guide

---

## Project Structure

```
url-shortener/
├── backend/
│   ├── internal/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── service/        # Business logic
│   │   ├── repository/     # Database operations
│   │   └── models/         # Data models
│   ├── cache/              # Redis layer
│   ├── middleware/         # CORS, rate limiting
│   ├── pkg/snowflake/      # ID generation
│   └── workers/            # Background jobs
├── frontend/               # React application
└── docs/                   # Documentation
```

---

## Testing

```bash
# Backend tests
cd backend && go test ./...

# With coverage
go test -cover ./...
```

---

## License

MIT License

---

## Author

Built by [@akushagra05](https://github.com/akushagra05)
