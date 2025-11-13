#!/bin/bash

# Test script for conditional writes in s3-proxy

PROXY_URL="http://localhost:8080"
TEST_FILE="conditional-test.txt"

echo "=== Testing Conditional Writes ==="
echo ""

# 1. Create initial file
echo "1. Creating initial file..."
echo "Hello, World!" > /tmp/test-content.txt
curl -X PUT -H "Content-Type: text/plain" \
  --data-binary @/tmp/test-content.txt \
  "$PROXY_URL/$TEST_FILE" -v 2>&1 | grep -E "(HTTP|ETag)" | head -5
echo ""

# Get ETag from response
ETAG=$(curl -s -X PUT -H "Content-Type: text/plain" \
  --data-binary @/tmp/test-content.txt \
  "$PROXY_URL/$TEST_FILE" -I | grep -i etag | cut -d' ' -f2 | tr -d '\r\n')
echo "ETag: $ETAG"
echo ""

# 2. Test If-Match (should succeed with correct ETag)
echo "2. Testing If-Match with correct ETag (should succeed)..."
curl -X PUT -H "Content-Type: text/plain" \
  -H "If-Match: $ETAG" \
  --data-binary "Updated content" \
  "$PROXY_URL/$TEST_FILE" -v 2>&1 | grep -E "(HTTP|ETag)" | head -5
echo ""

# 3. Test If-Match (should fail with wrong ETag)
echo "3. Testing If-Match with wrong ETag (should fail with 412)..."
curl -X PUT -H "Content-Type: text/plain" \
  -H "If-Match: wrong-etag-12345" \
  --data-binary "Should not update" \
  "$PROXY_URL/$TEST_FILE" -v 2>&1 | grep -E "(HTTP|412)" | head -3
echo ""

# 4. Get new ETag
NEW_ETAG=$(curl -s -I "$PROXY_URL/$TEST_FILE" | grep -i etag | cut -d' ' -f2 | tr -d '\r\n')
echo "New ETag after update: $NEW_ETAG"
echo ""

# 5. Test If-None-Match: * (create-only, should fail if exists)
echo "5. Testing If-None-Match: * (should fail with 412 if object exists)..."
curl -X PUT -H "Content-Type: text/plain" \
  -H "If-None-Match: *" \
  --data-binary "Should not create" \
  "$PROXY_URL/$TEST_FILE" -v 2>&1 | grep -E "(HTTP|412)" | head -3
echo ""

# 6. Test If-None-Match with specific ETag (should succeed if different)
echo "6. Testing If-None-Match with different ETag (should succeed)..."
curl -X PUT -H "Content-Type: text/plain" \
  -H "If-None-Match: old-etag-12345" \
  --data-binary "Updated with If-None-Match" \
  "$PROXY_URL/$TEST_FILE" -v 2>&1 | grep -E "(HTTP|ETag)" | head -5
echo ""

# 7. Test If-Unmodified-Since
echo "7. Testing If-Unmodified-Since..."
LAST_MOD=$(curl -s -I "$PROXY_URL/$TEST_FILE" | grep -i "last-modified" | cut -d' ' -f2- | tr -d '\r\n')
echo "Last-Modified: $LAST_MOD"
curl -X PUT -H "Content-Type: text/plain" \
  -H "If-Unmodified-Since: $LAST_MOD" \
  --data-binary "Updated with If-Unmodified-Since" \
  "$PROXY_URL/$TEST_FILE" -v 2>&1 | grep -E "(HTTP|ETag)" | head -5
echo ""

# 8. Cleanup
echo "8. Verifying final content..."
curl -s "$PROXY_URL/$TEST_FILE"
echo ""
echo ""

echo "=== Conditional Write Tests Complete ==="

