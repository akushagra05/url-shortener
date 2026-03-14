
# 🚀 Setup Guide

## Quick Start (Docker) - 30 seconds

```bash
git clone https://github.com/akushagra05/url-shortener
cd url-shortener
docker-compose up -d
```

**Done!** Access at:
- Backend: http://localhost:8080
- Frontend: http://localhost:5173
- PostgreSQL: localhost:5432
- Redis: localhost:6379

---

## Manual Setup (Local Development)

### Prerequisites
- Go 1.25+
- Node.js 18+
- PostgreSQL 17+
- Redis 8.2+

### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Set up environment
cp .env.example .env
```

**Edit `.env`:**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=url_shortener

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

SERVER_PORT=8080
BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:5173

SNOWFLAKE_MACHINE_ID=1
```

**Start PostgreSQL & Redis:**
```bash
# Option 1: Docker
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:17-alpine
docker run -d -p 6379:6379 redis:8-alpine

# Option 2: Homebrew (Mac)
brew install postgresql redis
brew services start postgresql
brew services start redis

# Option 3: apt (Linux)
sudo apt install postgresql redis
sudo systemctl start postgresql redis
```

**Run backend:**
```bash
go run main.go
# OR
go build && ./url-shortener
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

---

## Verify Installation

### Health Check
```bash
curl http://localhost:8080/health
```

Expected:
```json
{
  "status": "healthy",
  "services": {
    "database": "up",
    "redis": "up"
  }
}
```

### Create Test URL
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://google.com"}'
```

### Test Redirect
```bash
curl -I http://localhost:8080/{short_code}
# Should return 302 Found
```

---

## Database Setup (Manual)

If auto-migration fails:

```sql
-- Connect to PostgreSQL
psql -U postgres

-- Create database
CREATE DATABASE url_shortener;

-- Connect to it
\c url_shortener

-- Tables will auto-create on first run
-- Or run migrations manually:
\i backend/migrations/001_init.sql
```

---

## Troubleshooting

### Port already in use
```bash
# Backend (8080)
lsof -ti:8080 | xargs kill -9

# Frontend (5173)
lsof -ti:5173 | xargs kill -9

# PostgreSQL (5432)
lsof -ti:5432 | xargs kill -9
```

### Database connection failed
```bash
# Check PostgreSQL is running
psql -U postgres -c "SELECT version();"

# Check credentials in .env
cat backend/.env
```

### Redis connection failed
```bash
# Test Redis
redis-cli ping
# Should return: PONG

# Check Redis config
redis-cli CONFIG GET bind
```

### Docker issues
```bash
# Reset everything
docker-compose down -v
docker-compose up -d --build

# View logs
docker-compose logs -f
```

---

## Production Deployment

### Environment Variables
```env
# Security
DB_PASSWORD=<strong-password>
REDIS_PASSWORD=<strong-password>

# URLs
BASE_URL=https://short.yourdomain.com
FRONTEND_URL=https://yourdomain.com

# Scale
SNOWFLAKE_MACHINE_ID=<unique-per-instance>
```

### Recommended Platforms
- **Backend:** Render.com, Railway.app, Fly.io
- **Frontend:** Vercel, Netlify
- **Database:** Render PostgreSQL, Supabase
- **Redis:** Railway, Redis Labs

### Docker Production Build
```bash
docker build -t url-shortener-backend ./backend
docker build -t url-shortener-frontend ./frontend
```

---

## Development Tips

### Hot Reload (Backend)
```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
cd backend && air
```

### View Database
```bash
psql -U postgres -d url_shortener

# List tables
\dt

# View URLs
SELECT * FROM urls ORDER BY created_at DESC LIMIT 10;

# View analytics
SELECT * FROM url_analytics LIMIT 10;
```

### Monitor Redis
```bash
redis-cli MONITOR
# Shows all Redis commands in real-time

# Check cache
redis-cli KEYS "url:*"
redis-cli GET "url:abc123"
```

---

**Need help?** Open an issue on GitHub!
