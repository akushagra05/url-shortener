
#!/bin/bash

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

echo "==================================="
echo "URL Shortener API Test Suite"
echo "==================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

test_count=0
pass_count=0

# Helper function
test_endpoint() {
  test_count=$((test_count + 1))
  echo ""
  echo -e "${YELLOW}Test $test_count: $1${NC}"
  echo "-----------------------------------"
}

echo ""
echo "Starting tests against $BASE_URL"
echo ""

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

# Test 2: Create Short URL (Auto-generated)
test_endpoint "Create Short URL (Auto-generated)"
response=$(curl -s -X POST "$API_BASE/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com/search?q=golang"}')
if echo "$response" | grep -q '"success":true'; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
  # Extract short code for later tests
  SHORT_CODE=$(echo "$response" | grep -o '"short_code":"[^"]*"' | cut -d'"' -f4)
  echo "Generated short code: $SHORT_CODE"
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 3: Create Custom Alias
test_endpoint "Create Short URL (Custom Alias)"
response=$(curl -s -X POST "$API_BASE/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com", "custom_alias": "github2024"}')
if echo "$response" | grep -q '"success":true'; then
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
  -d '{"url": "not-a-valid-url"}')
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
if echo "$response" | grep -q "RESERVED"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 6: Duplicate Alias
test_endpoint "Duplicate Alias Prevention"
response=$(curl -s -X POST "$API_BASE/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://different-url.com", "custom_alias": "github2024"}')
if echo "$response" | grep -q "ALREADY_EXISTS"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 7: Get URL Details
if [ ! -z "$SHORT_CODE" ]; then
  test_endpoint "Get URL Details"
  response=$(curl -s -X GET "$API_BASE/url/$SHORT_CODE")
  if echo "$response" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ PASS${NC}"
    pass_count=$((pass_count + 1))
  else
    echo -e "${RED}✗ FAIL${NC}"
  fi
  echo "Response: $response"
fi

# Test 8: Get URL Analytics
test_endpoint "Get URL Analytics"
response=$(curl -s -X GET "$API_BASE/url/github2024/analytics")
if echo "$response" | grep -q '"success":true'; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 9: Test Redirect
test_endpoint "URL Redirection"
http_code=$(curl -s -o /dev/null -w "%{http_code}" -L "$BASE_URL/github2024")
if [ "$http_code" = "200" ]; then
  echo -e "${GREEN}✓ PASS${NC}"
  echo "Redirect successful (HTTP $http_code)"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
  echo "Got HTTP $http_code"
fi

# Test 10: Non-existent URL
test_endpoint "Non-existent URL (404)"
response=$(curl -s -X GET "$API_BASE/url/nonexistent123")
if echo "$response" | grep -q "NOT_FOUND"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 11: Delete URL
test_endpoint "Delete URL"
response=$(curl -s -X DELETE "$API_BASE/url/github2024")
if echo "$response" | grep -q '"success":true'; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Test 12: Access Deleted URL
test_endpoint "Access Deleted URL (should fail)"
response=$(curl -s -X GET "$API_BASE/url/github2024")
if echo "$response" | grep -q "NOT_FOUND"; then
  echo -e "${GREEN}✓ PASS${NC}"
  pass_count=$((pass_count + 1))
else
  echo -e "${RED}✗ FAIL${NC}"
fi
echo "Response: $response"

# Summary
echo ""
echo "==================================="
if [ $pass_count -eq $test_count ]; then
  echo -e "${GREEN}✓ All Tests Passed!${NC}"
else
  echo -e "${YELLOW}⚠ Some Tests Failed${NC}"
fi
echo "Results: $pass_count/$test_count passed"
echo "==================================="
echo ""
