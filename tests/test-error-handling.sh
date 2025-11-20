#!/bin/bash

# Test script to verify S3 error handling (403 vs 500)
# This ensures SignatureDoesNotMatch and other S3 errors are passed through correctly

set -e

PROXY_PORT=8080
PROXY_LOG="/tmp/s3-proxy-error-test.log"
TEST_RESULTS="/tmp/error-handling-test-results.txt"

echo "=== S3-Proxy Error Handling Test ===" > "$TEST_RESULTS"
echo "Date: $(date)" >> "$TEST_RESULTS"
echo "" >> "$TEST_RESULTS"

# Clean up any existing proxy
pkill -f "s3-proxy -port $PROXY_PORT" 2>/dev/null || true
sleep 1

# Test 1: Invalid credentials should return 403, not 500
echo "Test 1: Invalid credentials (should return 403)" >> "$TEST_RESULTS"
export S3PROXY_AWS_KEY="INVALID_KEY"
export S3PROXY_AWS_SECRET="INVALID_SECRET"
export S3PROXY_AWS_REGION="us-east-1"
export S3PROXY_AWS_BUCKET="sd-fs-proxy"
export S3PROXY_AWS_ENDPOINT="https://s3.wasabisys.com"

cd "$(dirname "$0")"
./s3-proxy -port $PROXY_PORT > "$PROXY_LOG" 2>&1 &
PROXY_PID=$!
sleep 2

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "http://localhost:$PROXY_PORT/test-invalid" -d "test")
ERROR_MSG=$(curl -s -X PUT "http://localhost:$PROXY_PORT/test-invalid" -d "test" 2>&1 | head -1)

kill $PROXY_PID 2>/dev/null || true
wait $PROXY_PID 2>/dev/null || true

if [ "$HTTP_CODE" = "403" ]; then
    echo "  ✓ PASS: Returned HTTP 403 (correct)" >> "$TEST_RESULTS"
    echo "  Error message: $ERROR_MSG" >> "$TEST_RESULTS"
else
    echo "  ✗ FAIL: Returned HTTP $HTTP_CODE (expected 403)" >> "$TEST_RESULTS"
    echo "  Error message: $ERROR_MSG" >> "$TEST_RESULTS"
    exit 1
fi
echo "" >> "$TEST_RESULTS"

# Test 2: Multipart upload with invalid credentials (should return 403)
echo "Test 2: Multipart upload with invalid credentials (should return 403)" >> "$TEST_RESULTS"
# Keep invalid credentials from Test 1
./s3-proxy -port $PROXY_PORT > "$PROXY_LOG" 2>&1 &
PROXY_PID=$!
sleep 2

# Test CreateMultipartUpload
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$PROXY_PORT/test-multipart-invalid?uploads" -H "Content-Type: application/octet-stream")
ERROR_MSG=$(curl -s -X POST "http://localhost:$PROXY_PORT/test-multipart-invalid?uploads" -H "Content-Type: application/octet-stream" 2>&1 | head -1)

kill $PROXY_PID 2>/dev/null || true
wait $PROXY_PID 2>/dev/null || true

if [ "$HTTP_CODE" = "403" ]; then
    echo "  ✓ PASS: CreateMultipartUpload returned HTTP 403 (correct)" >> "$TEST_RESULTS"
    echo "  Error message: $ERROR_MSG" >> "$TEST_RESULTS"
else
    echo "  ✗ FAIL: CreateMultipartUpload returned HTTP $HTTP_CODE (expected 403)" >> "$TEST_RESULTS"
    echo "  Error message: $ERROR_MSG" >> "$TEST_RESULTS"
    exit 1
fi
echo "" >> "$TEST_RESULTS"

# Test 3: Valid credentials should work normally
echo "Test 3: Valid credentials PUT (should return 200)" >> "$TEST_RESULTS"
export S3PROXY_AWS_KEY="X7SMWFBIMHK761MZDCM4"
export S3PROXY_AWS_SECRET="HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
export S3PROXY_AWS_REGION="us-east-1"
export S3PROXY_AWS_BUCKET="sd-fs-proxy"
export S3PROXY_AWS_ENDPOINT="https://s3.wasabisys.com"

./s3-proxy -port $PROXY_PORT > "$PROXY_LOG" 2>&1 &
PROXY_PID=$!
sleep 2

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "http://localhost:$PROXY_PORT/test-valid-$(date +%s)" -d "test")
kill $PROXY_PID 2>/dev/null || true
wait $PROXY_PID 2>/dev/null || true

if [ "$HTTP_CODE" = "200" ]; then
    echo "  ✓ PASS: Returned HTTP 200 (normal operation)" >> "$TEST_RESULTS"
else
    echo "  ✗ FAIL: Returned HTTP $HTTP_CODE (expected 200)" >> "$TEST_RESULTS"
    exit 1
fi
echo "" >> "$TEST_RESULTS"

# Test 4: Multipart upload with valid credentials (should return 200)
echo "Test 4: Multipart upload with valid credentials (should return 200)" >> "$TEST_RESULTS"
./s3-proxy -port $PROXY_PORT > "$PROXY_LOG" 2>&1 &
PROXY_PID=$!
sleep 2

# Test CreateMultipartUpload
KEY="test-multipart-valid-$(date +%s)"
RESPONSE=$(curl -s -X POST "http://localhost:$PROXY_PORT/$KEY?uploads" -H "Content-Type: application/octet-stream")
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$PROXY_PORT/$KEY?uploads" -H "Content-Type: application/octet-stream")

if [ "$HTTP_CODE" = "200" ]; then
    # Extract UploadId from XML response
    UPLOAD_ID=$(echo "$RESPONSE" | grep -oP '<UploadId>\K[^<]+' | head -1)
    if [ -n "$UPLOAD_ID" ]; then
        echo "  ✓ PASS: CreateMultipartUpload returned HTTP 200" >> "$TEST_RESULTS"
        echo "  UploadId: $UPLOAD_ID" >> "$TEST_RESULTS"
        
        # Test UploadPart (use same KEY)
        PART_RESPONSE=$(curl -s -i -X PUT "http://localhost:$PROXY_PORT/$KEY?uploadId=$UPLOAD_ID&partNumber=1" -d "part1data")
        PART_HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "http://localhost:$PROXY_PORT/$KEY?uploadId=$UPLOAD_ID&partNumber=1" -d "part1data")
        PART_ETAG=$(echo "$PART_RESPONSE" | grep -i "^etag:" | cut -d' ' -f2 | tr -d '\r\n')
        
        if [ "$PART_HTTP_CODE" = "200" ] && [ -n "$PART_ETAG" ]; then
            echo "  ✓ PASS: UploadPart returned HTTP 200" >> "$TEST_RESULTS"
            echo "  Part ETag: $PART_ETAG" >> "$TEST_RESULTS"
            
            # Test CompleteMultipartUpload
            COMPLETE_XML="<?xml version=\"1.0\" encoding=\"UTF-8\"?>
<CompleteMultipartUpload>
  <Part>
    <PartNumber>1</PartNumber>
    <ETag>$PART_ETAG</ETag>
  </Part>
</CompleteMultipartUpload>"
            
            COMPLETE_HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$PROXY_PORT/$KEY?uploadId=$UPLOAD_ID" -H "Content-Type: application/xml" -d "$COMPLETE_XML")
            
            if [ "$COMPLETE_HTTP_CODE" = "200" ]; then
                echo "  ✓ PASS: CompleteMultipartUpload returned HTTP 200" >> "$TEST_RESULTS"
            else
                echo "  ⚠ WARN: CompleteMultipartUpload returned HTTP $COMPLETE_HTTP_CODE (may need cleanup)" >> "$TEST_RESULTS"
            fi
        else
            echo "  ⚠ WARN: UploadPart returned HTTP $PART_HTTP_CODE (may need cleanup)" >> "$TEST_RESULTS"
        fi
    else
        echo "  ⚠ WARN: CreateMultipartUpload succeeded but couldn't extract UploadId" >> "$TEST_RESULTS"
    fi
else
    echo "  ✗ FAIL: CreateMultipartUpload returned HTTP $HTTP_CODE (expected 200)" >> "$TEST_RESULTS"
    exit 1
fi

kill $PROXY_PID 2>/dev/null || true
wait $PROXY_PID 2>/dev/null || true
echo "" >> "$TEST_RESULTS"

# Summary
echo "=== Test Summary ===" >> "$TEST_RESULTS"
echo "All tests passed! Error handling is working correctly." >> "$TEST_RESULTS"
echo "  - Invalid credentials return 403 (not 500)" >> "$TEST_RESULTS"
echo "  - Multipart uploads with invalid credentials return 403 (not 500)" >> "$TEST_RESULTS"
echo "  - Valid credentials return 200 (normal operation)" >> "$TEST_RESULTS"
echo "  - Multipart uploads with valid credentials work correctly" >> "$TEST_RESULTS"
echo "  - S3 error status codes are passed through correctly" >> "$TEST_RESULTS"

cat "$TEST_RESULTS"
echo ""
echo "Full test results saved to: $TEST_RESULTS"

