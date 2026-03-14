
# 🏗️ Architecture Deep Dive

## System Design

```
┌─────────┐
│ Clients │
└────┬────┘
     │
┌────▼────────┐
│ Gin Router  │  ← Rate Limiting, CORS
└────┬────────┘
     │
┌────▼────────┐
│  Handlers   │  ← Request validation
└────┬────────┘
     │
┌────▼────────┐
│  Services   │  ← Business logic
└────┬────────┘
     │
┌────▼────────────────┐
│ Repository + Cache  │
└──┬─────────────┬────┘
   │             │
┌──▼───┐    ┌───▼────┐
│Redis │    │Postgres│
└──────┘    └────────┘
```

## Key Technical Decisions

### 1. Snowflake ID Generation
**Why not UUID?**
- Sequential IDs for better DB performance (B-tree index friendly)
- Sortable by time (helps with analytics)
- Shorter encoded form (7-8 chars base62)

**Structure (64 bits):**
```
[41 bits timestamp][10 bits machine][12 bits sequence]
= 4096 IDs/ms/machine × 1024 machines = 4.2M IDs/sec
```

**Collision Prevention:** Distributed unique IDs, no DB roundtrip needed.

### 2. Two-Tier Caching Strategy

**Redis (Hot Cache):**
```go
// Cache URL for 24h
SET url:{code} {json} EX 86400

// Buffer clicks for batch write
INCR clicks:{code}
```

**PostgreSQL (Cold Storage):**
- Persistent storage
- Analytics queries
- Background sync from Redis

**Trade-off:** Eventual consistency (acceptable for click counts). Max 5min delay.

### 3. Click Count Optimization

**Problem:** DB write on every redirect kills performance.

**Solution:**
```
Click → INCR clicks:{code} in Redis (< 1ms)
       ↓
Background worker syncs every 5min
       ↓
Batch UPDATE urls SET click_count = click_count + {redis_count}
```

**Result:** 100x throughput improvement.

### 4. Rate Limiting (Token Bucket)

```redis
SET ratelimit:{ip}:{endpoint} 1 EX 60
INCR ratelimit:{ip}:{endpoint}
GET → if count > limit → 429 Too Many Requests
```

**Limits:**
- Create URL: 10/min per IP
- Redirect: 100/min per IP

**Why not DB-based?** Redis atomic operations are faster than DB transactions.

### 5. Redirect Strategy (302 vs 301)

**301 Permanent:** Browser caches forever → Can't track clicks, can't expire URLs

**302 Temporary:** Browser always checks server → Trackable, expirable

**Decision:** Use 302 Found for flexibility.

## Scalability Plan

### Current (Single Instance)
- Handles: 10K req/s
- Bottleneck: DB writes

### Scale to 100K req/s
1. Horizontal scaling (Kubernetes)
2. Read replicas for PostgreSQL
3. Redis Cluster (sharding)

### Scale to 1M req/s
1. CDN for static redirects
2. Database sharding by short_code hash
3. Separate analytics service
4. Message queue (Kafka) for click events

## Database Design

### Indexes Strategy
```sql
-- B-tree for lookups
CREATE INDEX idx_short_code ON urls(short_code);

-- Filtered for active URLs
CREATE INDEX idx_active_urls ON urls(is_active) WHERE is_active = TRUE;

-- Composite for analytics queries
CREATE INDEX idx_analytics_time ON url_analytics(short_code, accessed_at DESC);
```

### Partitioning (Future)
```sql
-- Partition analytics by month
CREATE TABLE url_analytics_2024_03 PARTITION OF url_analytics
FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');
```

**Why?** Analytics table grows unbounded. Partitioning enables:
- Fast queries (scan only relevant partition)
- Easy archival (DROP old partitions)

## Performance Metrics

| Operation | Latency | Notes |
|-----------|---------|-------|
| Create URL | 15-25ms | DB write + Redis set |
| Redirect (cached) | 5-10ms | Redis GET + 302 |
| Redirect (miss) | 30-50ms | DB query + cache + 302 |
| Analytics query | 50-100ms | DB aggregation |

**Cache hit ratio:** 95%+ (Redis handles 95% of redirects, DB only 5%)

## Security Considerations

1. **SQL Injection:** Prevented via GORM ORM
2. **Reserved Keywords:** Blocked common paths (api, admin, etc.)
3. **Input Validation:** Regex for custom aliases, URL parsing
4. **Rate Limiting:** Prevents abuse
5. **Expiry Enforcement:** UTC-based time comparison (timezone-safe)

## Trade-offs & Future Improvements

### Current Limitations
- No analytics retention policy (table grows unbounded)
- No URL ownership (anyone can delete any URL)
- Single Redis instance (SPOF)

### If Building at Scale
1. **Add authentication** - JWT tokens, API keys
2. **URL ownership** - User accounts, team workspaces
3. **Analytics pipeline** - Stream to data warehouse (BigQuery/Redshift)
4. **Redis Sentinel** - Automatic failover
5. **Circuit breakers** - Graceful degradation if Redis dies
6. **Prometheus metrics** - Monitoring & alerting

---

