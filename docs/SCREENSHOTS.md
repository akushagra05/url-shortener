
# 📸 Screenshot Guide

Quick guide for what screenshots to capture for documentation.

## Required Screenshots (8 total)

### 1. `homepage.png` - Main Interface
- Full app showing URL creation form
- All input fields visible (URL, custom alias, expiry options)

### 2. `url-created.png` - Success Flow
- Successful URL creation
- Short link displayed with copy button

### 3. `url-list.png` - Dashboard
- Multiple URLs listed
- Click counts visible
- Action buttons (Analytics, Delete)

### 4. `analytics.png` - Analytics Modal
- Open modal showing click statistics
- Total clicks counter

### 5. `redis-cache.png` - Caching Proof
```bash
redis-cli
> KEYS *
> GET "url:abc123"
> GET "clicks:abc123"
```

### 6. `database.png` - Schema
```sql
\d urls
```
Show table structure with indexes

### 7. `expired-url.png` - Expiry Handling
- Terminal/Postman showing 410 Gone response
- Expired URL error message

### 8. `docker.png` - Easy Setup
```bash
docker-compose ps
```
All services running

## Where to Add

Save all screenshots in `docs/screenshots/` folder.

Main README will embed key ones inline.
