#!/bin/bash

# URL Shortener API Test Script
# This script tests all the main endpoints of the URL shortener

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

echo "🔗 URL Shortener API Test Suite"
echo "================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Test 1: Health Check
echo "Test 1: Health Check"
echo "---------------------"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" = "200" ]; then
    print_result 0 "Health check passed"
    echo "$body" | jq '.'
else
    print_result 1 "Health check failed (Status: $status_code)"
fi
echo ""

# Test 2: Create Short URL
echo "Test 2: Create Short URL"
echo "------------------------"
response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/shorten" \
    -H "Content-Type: application/json" \
    -d '{
        "url": "https://github.com/example/url-shortener"
    }')
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" = "201" ]; then
    print_result 0 "URL creation successful"
    echo "$body" | jq '.'
    # Extract short_code for later tests
    SHORT_CODE=$(echo "$body" | jq -r '.data.short_code')
    echo -e "${YELLOW}Short code: $SHORT_CODE${NC}"
else
    print_result 1 "URL creation failed (Status: $status_code)"
    SHORT_CODE="test123"
fi
echo ""

# Test 3: Create Custom Alias
echo "Test 3: Create Custom Alias"
echo "----------------------------"
CUSTOM_ALIAS="interview$(date +%s)"
response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/shorten" \
    -H "Content-Type: application/json" \
    -d "{
        \"url\": \"https://example.com/custom\",
        \"custom_alias\": \"$CUSTOM_ALIAS\"
    }")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" = "201" ]; then
    print_result 0 "Custom alias creation successful"
    echo "$body" | jq '.'
else
    print_result 1 "Custom alias creation failed (Status: $status_code)"
fi
echo ""

# Test 4: Get URL Details
echo "Test 4: Get URL Details"
echo "-----------------------"
response=$(curl -s -w "\n%{http_code}" "$API_URL/url/$SHORT_CODE")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" = "200" ]; then
    print_result 0 "URL details retrieved successfully"
    echo "$body" | jq '.'
else
    print_result 1 "URL details retrieval failed (Status: $status_code)"
fi
echo ""

# Test 5: Test Redirect
echo "Test 5: Test Redirect"
echo "---------------------"
response=$(curl -s -w "\n%{http_code}" -I "$BASE_URL/$SHORT_CODE")
status_code=$(echo "$response" | grep -oP 'HTTP/\d+\.\d+ \K\d+' | head -n 1)

if [ "$status_code" = "302" ] || [ "$status_code" = "301" ]; then
    print_result 0 "Redirect working (Status: $status_code)"
    echo "$response" | grep -i "location:"
else
    print_result 1 "Redirect failed (Status: $status_code)"
fi
echo ""

# Test 6: Invalid URL Test
echo "Test 6: Invalid URL Test"
echo "------------------------"
response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/shorten" \
    -H "Content-Type: application/json" \
    -d '{
        "url": "not-a-valid-url"
    }')
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" = "400" ]; then
    print_result 0 "Invalid URL properly rejected"
    echo "$body" | jq '.'
else
    print_result 1 "Invalid URL test failed (Status: $status_code)"
fi
echo ""

# Test 7: Get Analytics
echo "Test 7: Get Analytics"
echo "---------------------"
response=$(curl -s -w "\n%{http_code}" "$API_URL/url/$SHORT_CODE/analytics?days=7")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" = "200" ]; then
    print_result 0 "Analytics retrieved successfully"
    echo "$body" | jq '.'
else
    print_result 1 "Analytics retrieval failed (Status: $status_code)"
fi
echo ""

# Test 8: Rate Limiting
echo "Test 8: Rate Limiting Test"
echo "---------------------------"
echo "Making 15 rapid requests to test rate limiting..."
success_count=0
rate_limited_count=0

for i in {1..15}; do
    response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/shorten" \
        -H "Content-Type: application/json" \
        -d "{
            \"url\": \"https://example.com/test$i\"
        }")
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "201" ]; then
        ((success_count++))
    elif [ "$status_code" = "429" ]; then
        ((rate_limited_count++))
    fi
done

echo "Successful requests: $success_count"
echo "Rate limited requests: $rate_limited_count"

if [ $rate_limited_count -gt 0 ]; then
    print_result 0 "Rate limiting is working"
else
    echo -e "${YELLOW}⚠ Rate limiting not triggered (limit may be high)${NC}"
fi
echo ""

# Test 9: Delete URL
echo "Test 9: Delete URL"
echo "------------------"
response=$(curl -s -w "\n%{http_code}" -X DELETE "$API_URL/url/$CUSTOM_ALIAS")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" = "200" ]; then
    print_result 0 "URL deleted successfully"
    echo "$body" | jq '.'
else
    print_result 1 "URL deletion failed (Status: $status_code)"
fi
echo ""

# Test 10: Access Deleted URL
echo "Test 10: Access Deleted URL"
echo "----------------------------"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/$CUSTOM_ALIAS")
status_code=$(echo "$response" | tail -n 1)

if [ "$status_code" = "404" ]; then
    print_result 0 "Deleted URL properly returns 404"
else
    print_result 1 "Deleted URL test failed (Status: $status_code)"
fi
echo ""

echo "================================"
echo "🏁 Test Suite Complete!"
echo "================================"
