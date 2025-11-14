#!/bin/bash
# Test direct Wasabi S3 access with the credentials

export AWS_ACCESS_KEY_ID="X7SMWFBIMHK761MZDCM4"
export AWS_SECRET_ACCESS_KEY="HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"

echo "=== Testing direct Wasabi S3 access ==="
echo "Endpoint: https://s3.us-east-1.wasabisys.com"
echo "Bucket: sb-zerofs-os"
echo "Region: us-east-1"
echo ""

# Test with curl using AWS signature v4
TEST_FILE="test-direct-$(date +%s).txt"
echo "test data" > /tmp/$TEST_FILE

echo "Attempting to PUT a test file directly to Wasabi..."
curl -v -X PUT \
  "https://s3.us-east-1.wasabisys.com/sb-zerofs-os/$TEST_FILE" \
  -H "Content-Type: text/plain" \
  -H "x-amz-date: $(date -u +%Y%m%dT%H%M%SZ)" \
  --data-binary "@/tmp/$TEST_FILE" \
  2>&1 | grep -E "(HTTP|403|401|200|Signature|Access)"

rm -f /tmp/$TEST_FILE
