
# URL Shortener - Backend Architecture

## Overview
A highly scalable URL shortener service built with clean architecture principles, featuring a layered design pattern for maintainability and testability.

## Tech Stack

### Core Technologies
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **ORM**: GORM
- **ID Generation**: Snowflake Algorithm

### Key Libraries
- `gin-gonic/gin` - HTTP framework
- `gorm.io/gorm` - ORM
- `go-redis/redis/v8` - Redis client
- `joho/godotenv` - Environment variables

## Project Structure

```
backend/
├── internal/              # Core application code (layered architecture)
│   ├── models/           # Data models and DTOs
│   │   └── url.go       # URL entities, request/response types
│   ├── repository/       # Data access layer
│   │   └── url_repository.go  # Database operations with interface
│   ├── service/          # Business logic layer
│   │   └── url_service.go     # Core business rules
│   └── handler/          # HTTP handlers layer
│       └── url_handler.go     # Request/response handling
│
├── cache/                # Redis utilities
│   └── redis.go         # Cache operations
│
├── database/             # Database utilities
│   └── database.go      # DB initialization and migrations
│
├── config/               # Configuration management
│   └── config.go        # App configuration
│
├── middleware/           # HTTP middleware
│   ├── cors.go          # CORS handling
│   ├── ratelimit.go     # Rate limiting
│   └── logger.go        # Request logging
│
├── workers/              # Background jobs
│   └── click_sync.go    # Sync Redis clicks to PostgreSQL
│
├── pkg/                  # Shared packages
│   └── snowflake/       # Snowflake ID generator
│       └── snowflake.go
│
├── main.go              # Application entry point
├── .env                 # Environment variables
├── Dockerfile           # Docker configuration
└── go.mod              # Go dependencies
```

## Architecture Layers

### 1. Handler Layer (`internal/handler/`)
**Responsibility**: HTTP request/response handling

```go
type URLHandler struct {
    service service.URLService
}
```

**Functions**:
- Parse HTTP requests
- Validate input format
- Call service layer
- Format HTTP responses
- Handle errors

**Example**:
```go
func (h *URLHandler) CreateShortURL(c *gin.Context) {
    var req models.CreateURLRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, models.APIResponse{...})
        return
    }
    
    result, err := h.service.CreateShortURL(c.Request.Context(), &req)
    c.JSON(201, models.APIResponse{Success: true, Data: result})
}
```

### 2. Service Layer (`internal/service/`)
**Responsibility**: Business logic and orchestration

```go
type URLService interface {
    CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.CreateURLResponse, error)
    GetOriginalURL(ctx context.Context, shortCode string) (string, error)
    GetURLDetails(ctx context.Context, shortCode string) (*models.URLDetailsResponse, error)
    GetAnalytics(ctx context.Context, shortCode string) (*models.URLAnalyticsResponse, error)
    DeleteURL(ctx context.Context, shortCode string) error
}
```

**Responsibilities**:
- URL validation (format, length, domain checks)
- Reserved keyword checking
- Snowflake ID generation for short codes
- Custom alias validation
- Business rules enforcement
- Cache coordination
- Error handling with business context

**Example**:
```go
func (s *urlService) CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.CreateURLResponse, error) {
    // 1. Validate URL
    if err := s.validateURL(req.URL); err != nil {
        return nil, err
    }
    
    // 2. Check reserved keywords
    if req.CustomAlias != "" && s.isReservedKeyword(req.CustomAlias) {
        return nil, errors.New("RESERVED_KEYWORD")
    }
    
    // 3. Generate or use custom short code
    shortCode := req.CustomAlias
    if shortCode == "" {
        shortCode, _ = s.snowflakeGen.GenerateString()
    }
    
    // 4. Create URL via repository
    urlRecord := &models.URL{...}
    if err := s.repo.Create(ctx, urlRecord); err != nil {
        return nil, err
    }
    
    return response, nil
}
```

### 3. Repository Layer (`internal/repository/`)
**Responsibility**: Data access and persistence

```go
type URLRepository interface {
    Create(ctx context.Context, url *models.URL) error
    FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
    Update(ctx context.Context, url *models.URL) error
    Delete(ctx context.Context, shortCode string) error
    IncrementClickCount(ctx context.Context, id uint) error
    IsReservedKeyword(ctx context.Context, keyword string) (bool, error)
    GetAnalytics(ctx context.Context, shortCode string) (*models.URLAnalyticsResponse, error)
}
```

**Responsibilities**:
- CRUD operations
- Database queries
- Transaction management
- Data mapping
- No business logic

**Example**:
```go
func (r *urlRepository) Create(ctx context.Context, url *models.URL) error {
    return r.db.WithContext(ctx).Create(url).Error
}

func (r *urlRepository) FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
    var url models.URL
    err := r.db.WithContext(ctx).
        Where("short_code = ? AND is_active = ?", shortCode, true).
        First(&url).Error
    return &url, err
}
```

### 4. Models Layer (`internal/models/`)
**Responsibility**: Data structures and contracts

**Types**:
- Database entities (`URL`, `URLAnalytics`, `ReservedKeyword`)
- Request DTOs (`CreateURLRequest`)
- Response DTOs (`CreateURLResponse`, `URLDetailsResponse`)
- API responses (`APIResponse`, `APIError`)

```go
type URL struct {
    ID             uint           `gorm:"primaryKey"`
    ShortCode      string         `gorm:"uniqueIndex;size:20;not null"`
    OriginalURL    string         `gorm:"type:text;not null"`
    CustomAlias    bool           `gorm:"default:false"`
    CreatedAt      time.Time
    ExpiresAt      *time.Time
    ClickCount     int64          `gorm:"default:0"`
    LastAccessedAt *time.Time
    IsActive       bool           `gorm:"default:true"`
}
```

## Database Schema

### URLs Table
```sql
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    custom_alias BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    click_count BIGINT DEFAULT 0,
    last_accessed_at TIMESTAMP,
    user_id VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    
    CONSTRAINT short_code_length CHECK (LENGTH(short_code) >= 4 AND LENGTH(short_code) <= 20)
);

CREATE INDEX idx_short_code ON urls(short_code);
CREATE INDEX idx_active_urls ON urls(is_active) WHERE is_active = TRUE;
```

### URL Analytics Table
```sql
CREATE TABLE url_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) NOT NULL,
    accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    country_code CHAR(2),
    city VARCHAR(100),
    device_type VARCHAR(20),
    
    FOREIGN KEY (short_code) REFERENCES urls(short_code) ON DELETE CASCADE
);

CREATE INDEX idx_analytics_short_code_time ON url_analytics(short_code, accessed_at DESC);
```

### Reserved Keywords Table
```sql
CREATE TABLE reserved_keywords (
    keyword VARCHAR(20) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Snowflake ID Generation

### Algorithm Structure (64 bits)
```
┌─────────────┬──────────────┬─────────────┐
│  Timestamp  │  Machine ID  │  Sequence   │
│  (41 bits)  │  (10 bits)   │  (12 bits)  │
└─────────────┴──────────────┴─────────────┘
```

**Capacity**:
- 1024 concurrent servers
- 4096 IDs per millisecond per server
- ~69 years of unique IDs
- ~4.2M IDs/second globally

**Implementation**:
```go
type Generator struct {
    mu            sync.Mutex
    epoch         int64  // Custom epoch (milliseconds)
    machineID     int64  // 0-1023
    sequence      int64  // 0-4095
    lastTimestamp int64
}
```

**Benefits**:
- Guaranteed uniqueness across distributed systems
- Time-sortable IDs
- No database coordination needed
- High performance (in-memory generation)

## Redis Caching Strategy

### Click Counter Cache
```
Key: "clicks:{short_code}"
Value: Integer counter
TTL: None (persistent until synced)
```

**Purpose**: Buffer click counts to reduce database writes

**Sync Pattern**:
1. Increment counter in Redis on each access
2. Background worker syncs to PostgreSQL every 5 minutes
3. Batch update database with accumulated counts
4. Clear Redis counters after successful sync

### Rate Limiting Cache
```
Key: "ratelimit:{ip}:{endpoint}"
Value: Request count
TTL: 60 seconds
```

**Limits**:
- URL creation: 10 requests/minute per IP
- URL redirect: 100 requests/minute per IP

## API Endpoints

### 1. Create Short URL
```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url",
  "custom_alias": "sale2024",      // optional
  "expires_at": "2024-12-31T23:59:59Z"  // optional
}

Response 201:
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "short_url": "http://localhost:8080/aBc1DeF",
    "original_url": "https://example.com/very/long/url",
    "created_at": "2024-03-13T10:30:00Z"
  }
}
```

### 2. Redirect to Original URL
```http
GET /{short_code}

Response 302:
Location: https://example.com/very/long/url
```

### 3. Get URL Details
```http
GET /api/v1/url/{short_code}

Response 200:
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "original_url": "https://example.com/very/long/url",
    "click_count": 142,
    "created_at": "2024-03-13T10:30:00Z",
    "last_accessed_at": "2024-03-14T15:20:00Z"
  }
}
```

### 4. Get Analytics
```http
GET /api/v1/url/{short_code}/analytics

Response 200:
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "original_url": "https://example.com/very/long/url",
    "click_count": 142,
    "created_at": "2024-03-13T10:30:00Z"
  }
}
```

### 5. Delete URL
```http
DELETE /api/v1/url/{short_code}

Response 200:
{
  "success": true,
  "message": "URL deleted successfully"
}
```

### 6. Health Check
```http
GET /health

Response 200:
{
  "status": "healthy",
  "timestamp": "2024-03-14T20:00:00Z"
}
```

## Validation Rules

### URL Validation
- Must start with `http://` or `https://`
- Maximum length: 2048 characters
- Must be valid URL format
- Domain must not be in blocked list

### Custom Alias Validation
- Length: 4-20 characters
- Allowed: a-z, A-Z, 0-9, hyphen, underscore
- Must not be reserved keyword
- Must be unique in database

### Reserved Keywords
```
api, admin, dashboard, health, metrics,
login, logout, register, docs, swagger,
static, assets, public, www, app,
v1, v2, version, status, ping
```

## Error Handling

### Error Codes
- `INVALID_URL` - URL format is invalid
- `ALIAS_ALREADY_EXISTS` - Custom alias is taken
- `RESERVED_KEYWORD` - Alias is a reserved word
- `URL_NOT_FOUND` - Short code doesn't exist
- `URL_EXPIRED` - Short URL has expired
- `RATE_LIMIT_EXCEEDED` - Too many requests

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "INVALID_URL",
    "message": "URL must start with http:// or https://"
  }
}
```

## Middleware Stack

1. **Logger** - Request/response logging
2. **CORS** - Cross-origin resource sharing
3. **Rate Limiter** - Request throttling per IP
4. **Recovery** - Panic recovery with logging

## Background Workers

### Click Counter Sync Worker
**Schedule**: Every 5 minutes

**Process**:
1. Fetch all Redis click counters
2. Batch update PostgreSQL with accumulated counts
3. Clear Redis counters on success
4. Log sync statistics

```go
func SyncClickCounters() error {
    counters, err := cache.GetAllClickCounters()
    
    for shortCode, count := range counters {
        database.UpdateClickCount(shortCode, count)
    }
    
    cache.FlushClickCounters(shortCodes)
}
```

## Configuration

### Environment Variables
```bash
# Server
PORT=8080
BASE_URL=http://localhost:8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=urlshortener

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Snowflake
MACHINE_ID=1
```

## Deployment

### Docker Compose
```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
  
  postgres:
    image: postgres:15
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  redis:
    image: redis:7
```

### Health Checks
- Database connectivity check on startup
- Redis connectivity check on startup
- `/health` endpoint for load balancer

## Performance Characteristics

### Latency Targets
- URL Creation: <100ms (p95)
- URL Redirect: <50ms (p95, cached)
- URL Redirect: <100ms (p95, cache miss)

### Scalability
- Horizontal scaling via multiple app instances
- Snowflake ensures unique IDs across instances
- Redis for distributed caching
- PostgreSQL read replicas for high read load

### Capacity
- 4.2M short URLs per second (Snowflake limit)
- Limited by database write throughput for analytics
- Redis click buffering reduces database load by 99%

## Testing Strategy

### Unit Tests
- Service layer business logic
- Repository layer database operations
- Snowflake ID generation

### Integration Tests
- API endpoint testing
- Database transaction testing
- Redis cache behavior

### Load Tests
- Concurrent URL creation
- High-frequency redirects
- Rate limiter behavior

## Future Enhancements

1. **Advanced Analytics**
   - Geographic visualization
   - Time-series click patterns
   - Device/browser statistics
   - Referrer tracking

2. **User Management**
   - User authentication
   - Personal dashboards
   - URL organization (folders/tags)
   - Usage quotas

3. **Premium Features**
   - Custom domains
   - QR code generation
   - Link expiration
   - Password protection

4. **Monitoring**
   - Prometheus metrics
   - Grafana dashboards
   - Alert rules
   - Performance tracking

## Security Considerations

1. **Rate Limiting** - Prevents abuse
2. **URL Validation** - Blocks malicious URLs
3. **Reserved Keywords** - Protects system routes
4. **Input Sanitization** - Prevents injection attacks
5. **HTTPS Only** - In production deployment

## Conclusion

This architecture provides a solid foundation for a production-ready URL shortener service with clear separation of concerns, testability, and scalability. The layered design makes it easy to maintain, extend, and test each component independently.
