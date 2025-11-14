#!/bin/bash
# Test multipart upload signature handling

echo "=== Testing Multipart Upload with Signed Requests ==="
echo ""

KEY="test-multipart-$(date +%s).txt"
echo "Using key: $KEY"
echo ""

# Test 1: InitiateMultipartUpload with fake signature
echo "Test 1: InitiateMultipartUpload with FAKE signature"
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "http://localhost:8080/${KEY}?uploads" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=FAKE/20231114/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-date, Signature=fakesignature" \
  -H "Content-Type: application/octet-stream")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | grep -v "HTTP_CODE:")

echo "HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
  echo "✅ InitiateMultipartUpload succeeded (proxy strips fake auth)"
  UPLOAD_ID=$(echo "$BODY" | grep -oP '<UploadId>\K[^<]+')
  echo "Upload ID: $UPLOAD_ID"
else
  echo "❌ InitiateMultipartUpload failed"
  echo "Response: $BODY"
fi

echo ""
echo "=== Testing with real ZeroFS multipart upload ==="
echo "Check ZeroFS logs for SignatureDoesNotMatch errors during large file uploads"
