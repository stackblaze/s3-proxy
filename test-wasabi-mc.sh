#!/bin/bash
# Test Wasabi S3 with MinIO Client (mc)

echo "=== Testing Wasabi S3 with mc client ==="
echo ""

# Configure mc alias for Wasabi
echo "Configuring mc alias for Wasabi..."
mc alias set wasabi-test https://s3.us-east-1.wasabisys.com X7SMWFBIMHK761MZDCM4 HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G

echo ""
echo "Test 1: List buckets"
mc ls wasabi-test/ 2>&1 | head -10

echo ""
echo "Test 2: List objects in sb-zerofs-os bucket"
mc ls wasabi-test/sb-zerofs-os/ 2>&1 | head -10

echo ""
echo "Test 3: PUT a test file"
echo "test data from mc" > /tmp/test-mc.txt
mc cp /tmp/test-mc.txt wasabi-test/sb-zerofs-os/test-mc-$(date +%s).txt 2>&1
rm -f /tmp/test-mc.txt

echo ""
echo "Test 4: Test via s3-proxy (localhost:8080)"
mc alias set s3proxy http://localhost:8080 "" "" --api S3v4
mc ls s3proxy/sb-zerofs-os/ 2>&1 | head -10
