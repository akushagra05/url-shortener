
# API Testing Guide

Complete guide to test all endpoints using cURL or Postman.

## Prerequisites

1. **Start the backend server:**
```bash
cd backend
go run main.go
```

2. **Ensure services are running:**
- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`
- Backend on `http://localhost:8080`

## Test Endpoints

### 1. Health Check
**Purpose**: Verify server is running

```bash
curl -X GET http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-03-14T20:15:30Z"
}
```

---

### 2. Create Short URL (Auto-generated)
**Purpose**: Create a URL with auto-generated short code

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.google.com/search?q=golang+backend+interview"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "short_code": "aBc1DeF",
    "short_url": "http://localhost:8080/aBc1DeF",
    "original_url": "https://www.google.com/search?q=golang+backend+interview",
    "custom_alias": false,
    "created_at": "2024-03-14T20:15:30Z"
  }
}
```

---

### 3. Create Short URL (Custom Alias)
**Purpose**: Create URL with custom short code

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://github.com/yourusername/url-shortener",
    "custom_alias": "github2024"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "short_code": "github2024",
    "short_url": "http://localhost:8080/github2024",
    "original_url": "https://github.com/yourusername/url-shortener",
    "custom_alias": true,
    "created_at": "2024-03-14T20:16:00Z"
  }
}
```

---

### 4. Create Short URL with Expiration
**Purpose**: Create URL that expires after a date

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/limited-time-offer",
    "custom_alias": "sale2024",
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "short_code": "sale2024",
    "short_url": "http://localhost:8080/sale2024",
    "original_url": "https://example.com/limited-time-offer",
    "custom_alias": true,
    "created_at": "2024-03-14T20:17:00Z",
    "expires_at": "2024-12-31T23:59:59Z"
  }
}
```

---

### 5. Test Invalid URL
**Purpose**: Verify URL validation

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "not-a-valid-url"
  }'
```

**Expected Response (400):**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_URL",
    "message": "URL must start with http:// or https://"
  }
}
```

---

### 6. Test Reserved Keyword
**Purpose**: Verify reserved keywords are blocked

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "custom_alias": "api"
  }'
```

**Expected Response (400):**
```json
{
  "success": false,
  "error": {
    "code": "RESERVED_KEYWORD",
    "message": "This alias is reserved and cannot be used"
  }
}
```

---

### 7. Test Duplicate Custom Alias
**Purpose**: Verify uniqueness constraint

First, create a URL:
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "custom_alias": "test123"
  }'
```

Then try to create another with same alias:
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://different-url.com",
    "custom_alias": "test123"
  }'
```

**Expected Response (409):**
```json
{
  "success": false,
  "error": {
    "code": "ALIAS_ALREADY_EXISTS",
    "message": "Custom alias 'test123' is already taken"
  }
}
```

---

### 8. Redirect to Original URL
**Purpose**: Test URL redirection (replace {short_code} with actual code)

```bash
curl -L -X GET http://localhost:8080/{short_code}
```

**Example:**
```bash
curl -L -X GET http://localhost:8080/github2024
```

**Expected**: Should redirect (302) to the original URL

---

### 9. Get URL Details
**Purpose**: Retrieve URL information

```bash
curl -X GET http://localhost:8080/api/v1/url/{short_code}
```

**Example:**
```bash
curl -X GET http://localhost:8080/api/v1/url/github2024
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "short_code": "github2024",
    "short_url": "http://localhost:8080/github2024",
    "original_url": "https://github.com/yourusername/url-shortener",
    "click_count": 5,
    "created_at": "2024-03-14T20:16:00Z",
    "last_accessed_at": "2024-03-14T20:25:00Z",
    "is_active": true
  }
}
```

---

### 10. Get URL Analytics
**Purpose**: Retrieve URL analytics

```bash
curl -X GET http://localhost:8080/api/v1/url/{short_code}/analytics
```

**Example:**
```bash
curl -X GET http://localhost:8080/api/v1/url/github2024/analytics
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "short_code": "github2024",
    "original_url": "https://github.com/yourusername/url-shortener",
    "click_count": 5,
    "created_at": "2024-03-14T20:16:00Z",
    "last_accessed": "2024-03-14T20:25:00Z"
  }
}
```

---

### 11. Test Non-existent URL
**Purpose**: Test 404 handling

```bash
curl -X GET http://localhost:8080/api/v1/url/nonexistent
```

**Expected Response (404):**
```json
{
  "success": false,
  "error": {
    "code": "URL_NOT_FOUND",
    "message": "Short URL not found"
  }
}
```

---

### 12. Delete URL
**Purpose**: Soft delete a URL

```bash
curl -X DELETE http://localhost:8080/api/v1/url/{short_code}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/url/test123
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Short URL deleted successfully"
}
```

---

### 13. Test Rate Limiting
**Purpose**: Verify rate limiting works

Run this script to make multiple requests quickly:

```bash
#!/bin/bash
for i in {1..15}; do
  echo "Request $i:"
  curl -X POST http://localhost:8080/api/v1/shorten \
    -H "Content-Type: application/json" \
    -d "{\"url\": \"https://example.com/test$i\"}"
  echo ""
done
```

**Expected**: After 10 requests, you should see rate limit error:
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later."
  }
}
```

---

## Complete Test Scenario

### Scenario: Create, Access, Analyze, Delete

```bash
# 1. Create a short URL
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.example.com/very/long/url/path",
    "custom_alias": "demo2024"
  }')
echo "Created: $RESPONSE"

# 2. Access it multiple times (simulate clicks)
for i in {1..5}; do
  curl -L -X GET http://localhost:8080/demo2024 > /dev/null 2>&1
  echo "Click $i"
done

# 3. Get URL details
curl -X GET http://localhost:8080/api/v1/url/demo2024
echo ""

# 4. Get analytics
curl -X GET http://localhost:8080/api/v1/url/demo2024/analytics
echo ""

# 5. Delete the URL
curl -X DELETE http://localhost:8080/api/v1/url/demo2024
echo ""

# 6. Try to access deleted URL
curl -X GET http://localhost:8080/api/v1/url/demo2024
```

---

## Postman Collection

### Import These Requests:

1. **Create Collection**: "URL Shortener API Tests"

2. **Set Environment Variables**:
   - `base_url`: `http://localhost:8080`
   - `api_base`: `{{base_url}}/api/v1`

3. **Add Requests**:

#### Health Check
```
GET {{base_url}}/health
```

#### Create Short URL
```
POST {{api_base}}/shorten
Headers: Content-Type: application/json
Body (JSON):
{
  "url": "https://www.google.com",
  "custom_alias": "google"
}
```

#### Get URL Details
```
GET {{api_base}}/url/google
```

#### Redirect
```
GET {{base_url}}/google
```

#### Delete URL
```
DELETE {{api_base}}/url/google
```

---

## Automated Test Script

Save this as `test-api.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

echo "==================================="
echo "URL Shortener API Test Suite"
echo "==================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

test_count=0
pass_count=0

# Helper function
test_endpoint() {
  test_count=$((test_count + 1))
  echo ""
  echo "Test $test_count: $1"
  echo "-----------------------------------"
}

# Test 1: Health Check
test_endpoint "Health Check"
response=$(curl -s -X GET "$BASE_URL/health")
if echo "$response" | grep -q "healthy"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 2: Create Short URL
test_endpoint "Create Short URL (Auto)"
response=$(curl -s -X POST "$API_BASE/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com"}')
if echo "$response" | grep -q "success.*true"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
  # Extract short code for later tests
  SHORT_CODE=$(echo "$response" | grep -o '"short_code":"[^"]*"' | cut -d'"' -f4)
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 3: Create Custom Alias
test_endpoint "Create Short URL (Custom Alias)"
response=$(curl -s -X POST "$API_BASE/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com", "custom_alias": "gh2024"}')
if echo "$response" | grep -q "success.*true"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 4: Invalid URL
test_endpoint "Invalid URL Validation"
response=$(curl -s -X POST "$API_BASE/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-url"}')
if echo "$response" | grep -q "INVALID_URL"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 5: Reserved Keyword
test_endpoint "Reserved Keyword Rejection"
response=$(curl -s -X POST "$API_BASE/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "custom_alias": "api"}')
if echo "$response" | grep -q "RESERVED_KEYWORD"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 6: Get URL Details
if [ ! -z "$SHORT_CODE" ]; then
  test_endpoint "Get URL Details"
  response=$(curl -s -X GET "$API_BASE/url/$SHORT_CODE")
  if echo "$response" | grep -q "success.*true"; then
    echo -e "${GREEN}✓ PASS${NC}"
    pass_count=$((pass_count + 1))
  else
    echo -e "${RED}✗ FAIL${NC}"
  fi
  echo "Response: $response"
fi

# Summary
echo ""
echo "==================================="
echo "Test Results: $pass_count/$test_count passed"
echo "==================================="
```

Make it executable and run:
```bash
chmod +x test-api.sh
./test-api.sh
```

---

## Testing from Frontend

Once backend is verified, test from the React frontend:

1. **Start Frontend**:
```bash
cd frontend
npm run dev
```

2. **Open Browser**: `http://localhost:5173`

3. **Test Flow**:
   - Enter a long URL
   - Optionally add custom alias
   - Click "Shorten URL"
   - Copy short URL
   - Open short URL in new tab (should redirect)
   - Check analytics/click count

---

## Common Issues & Solutions

### Issue: Connection Refused
**Solution**: Ensure backend is running on port 8080

### Issue: Database Error
**Solution**: Check PostgreSQL is running and database exists

### Issue: Redis Error
**Solution**: Ensure Redis is running on port 6379

### Issue: CORS Error (from frontend)
**Solution**: CORS middleware should allow `http://localhost:5173`

### Issue: Rate Limit Hit
**Solution**: Wait 60 seconds or restart Redis to clear counters

---

## Success Criteria

✅ All endpoints return correct status codes  
✅ URL validation works properly  
✅ Custom aliases are enforced  
✅ Reserved keywords are blocked  
✅ Redirects work correctly  
✅ Click counting increments  
✅ Rate limiting prevents abuse  
✅ Error messages are clear and helpful

---

## Next Steps

1. ✅ Test all endpoints manually
2. ✅ Run automated test script
3. ✅ Test from frontend
4. ✅ Take screenshots for documentation
5. ✅ Document any issues found
6. ✅ Ready for submission!
