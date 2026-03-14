
# ✅ Interview Presentation Checklist

## 📸 Screenshots to Capture (8 Required)

Save all in `docs/screenshots/` folder:

### Must-Have Screenshots:
- [ ] `homepage.png` - Main UI with URL creation form
- [ ] `url-created.png` - Successful URL creation with short link
- [ ] `url-list.png` - Dashboard showing multiple URLs with click counts
- [ ] `analytics.png` - Analytics modal open showing stats
- [ ] `redis-cache.png` - Terminal showing Redis cache (KEYS *, GET url:*)
- [ ] `database.png` - PostgreSQL schema (\d urls)
- [ ] `expired-url.png` - 410 Gone response for expired URL
- [ ] `docker.png` - docker-compose ps showing services running

### How to Capture:
```bash
# Redis cache proof
redis-cli
> KEYS *
> GET "url:abc123"
> GET "clicks:abc123"

# Database schema
psql -U postgres -d url_shortener
\d urls
\d url_analytics

# Expired URL test
curl -I http://localhost:8080/abc123
# Should show: HTTP/1.1 410 Gone

# Docker running
docker-compose ps
```

---

## 📝 Documentation Review

- [ ] README.md - Main overview, quick to scan
- [ ] docs/ARCHITECTURE.md - Deep technical details
- [ ] docs/API.md - Complete API reference
- [ ] docs/SETUP.md - Setup instructions
- [ ] docs/SCREENSHOTS.md - Screenshot guide

---

## 🧪 Pre-Interview Testing

Run these to ensure everything works:

```bash
# 1. Health check
curl http://localhost:8080/health
# Should return: {"status":"healthy"}

# 2. Create URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://google.com"}'
# Should return 201 with short_url

# 3. Test redirect
curl -I http://localhost:8080/{short_code}
# Should return 302 Found

# 4. Test expiry
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","expires_in":"1m"}'
# Wait 1 minute, then:
curl -I http://localhost:8080/{short_code}
# Should return 410 Gone

# 5. Test rate limiting
for i in {1..12}; do curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'; done
# After 10 requests: Should return 429 Too Many Requests

# 6. Test custom alias
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","custom_alias":"test123"}'
# Should create http://localhost:8080/test123

# 7. Test reserved keywords
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","custom_alias":"api"}'
# Should return 422 with ALIAS_RESERVED

# 8. Verify Redis caching
redis-cli KEYS "*"
# Should show: url:*, clicks:*, ratelimit:*

# 9. Check click count increment
# Create URL, visit it 3 times, check analytics
curl http://localhost:8080/api/v1/url/{short_code}/analytics
# total_clicks should show 3

# 10. Frontend works
# Open http://localhost:5173
# Try creating URLs through UI
```

---

## 🎯 Key Talking Points for Interview

### Backend Architecture Strengths:
1. **Snowflake IDs** - No central bottleneck, 4.2M IDs/sec capacity
2. **Two-tier caching** - Redis for speed, PostgreSQL for durability
3. **Async click tracking** - 100x throughput vs synchronous writes
4. **Rate limiting** - Redis-backed token bucket, prevents abuse
5. **Soft deletes** - Data safety, audit trails
6. **Horizontal scalability** - Stateless design, K8s ready

### Trade-offs Made:
1. **Eventual consistency** for clicks (5min delay) - Performance > Real-time
2. **302 redirect** over 301 - Flexibility > Browser caching
3. **In-memory buffer** - Risk of data loss, mitigated by frequent syncs

### Scalability Path:
- **10K req/s** - Single instance (current)
- **100K req/s** - Horizontal scaling + Redis Cluster
- **1M req/s** - CDN + DB sharding + analytics pipeline

### What You'd Add at Scale:
1. Authentication (JWT, API keys)
2. Redis Sentinel (auto-failover)
3. Circuit breakers (graceful degradation)
4. Prometheus metrics + Grafana
5. Analytics pipeline (stream to data warehouse)

---

## 💼 Final Touches

- [ ] Update GitHub repo link in all docs
- [ ] Add .gitignore for sensitive files
- [ ] Clean up any TODO comments in code
- [ ] Ensure .env.example is up to date
- [ ] Test Docker setup on clean machine (if possible)
- [ ] Prepare to demo live (have Redis, DB, server running)

---

## 🎤 Demo Flow Suggestion

1. **Start with README** (30 sec)
   - "Production-ready URL shortener with Snowflake IDs and Redis caching"
   
2. **Quick architecture overview** (1 min)
   - Show diagram, explain caching strategy
   
3. **Live demo** (2 min)
   - Create URL via Postman
   - Show Redis cache (redis-cli GET)
   - Test redirect
   - Show analytics
   
4. **Code walkthrough** (2 min)
   - Snowflake generation
   - Layered architecture
   - Async click worker
   
5. **Scalability discussion** (1 min)
   - Current capacity
   - How to scale to 100K req/s
   
6. **Questions** (3 min)

---

## 🔍 Common Interview Questions - Prepare Answers

**Q: Why Snowflake over auto-increment or UUID?**
A: Sequential for DB performance, sortable for analytics, distributed (no bottleneck), shorter encoding.

**Q: What if Redis goes down?**
A: App continues working, queries hit DB directly. Click counts buffer lost (max 5min data). Solution at scale: Redis Sentinel for auto-failover.

**Q: How to prevent duplicate short codes?**
A: Snowflake guarantees uniqueness via timestamp + machine ID + sequence. For custom aliases: DB unique constraint.

**Q: How to scale to 1 million users?**
A: 1) Horizontal scaling (K8s), 2) Redis Cluster, 3) DB read replicas + sharding, 4) CDN for hot URLs, 5) Analytics to data warehouse.

**Q: Why eventual consistency for clicks?**
A: Click stats don't need real-time accuracy. Buffering in Redis gives 100x throughput. Acceptable trade-off: 5min delay vs DB write on every click.

---

**Good luck with your interview! 🚀**

Remember: Focus on backend architecture, scalability, and trade-off analysis. That's what differentiates senior engineers!
