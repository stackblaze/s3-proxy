#!/bin/bash
# Test Wasabi access with AWS CLI

export AWS_ACCESS_KEY_ID="X7SMWFBIMHK761MZDCM4"
export AWS_SECRET_ACCESS_KEY="HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
export AWS_DEFAULT_REGION="us-east-1"

echo "=== Testing Wasabi S3 with AWS CLI ==="
echo ""
echo "Test 1: List buckets"
aws s3 ls --endpoint-url https://s3.us-east-1.wasabisys.com 2>&1 | head -10

echo ""
echo "Test 2: List objects in sb-zerofs-os bucket"
aws s3 ls s3://sb-zerofs-os/ --endpoint-url https://s3.us-east-1.wasabisys.com 2>&1 | head -10

echo ""
echo "Test 3: Try to PUT a test file"
echo "test data" > /tmp/test-wasabi.txt
aws s3 cp /tmp/test-wasabi.txt s3://sb-zerofs-os/test-wasabi.txt --endpoint-url https://s3.us-east-1.wasabisys.com 2>&1
rm -f /tmp/test-wasabi.txt
