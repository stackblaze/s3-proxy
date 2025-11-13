#!/bin/bash

# Test script for s3-proxy using ZeroFS configuration
# Based on your ZeroFS config file

echo "Setting up s3-proxy with ZeroFS credentials..."

# Extract credentials from ZeroFS config
export S3PROXY_AWS_KEY="X7SMWFBIMHK761MZDCM4"
export S3PROXY_AWS_SECRET="HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
export S3PROXY_AWS_REGION="us-east-1"
export S3PROXY_AWS_BUCKET="sd-zerofs"

# Wasabi endpoint (adjust if using different region)
# Common Wasabi endpoints:
#   us-east-1: s3.wasabisys.com
#   us-east-2: s3.us-east-2.wasabisys.com
#   us-west-1: s3.us-west-1.wasabisys.com
#   eu-central-1: s3.eu-central-1.wasabisys.com
export S3PROXY_AWS_ENDPOINT="https://s3.wasabisys.com"

# Optional: Enable CORS if needed
# export S3PROXY_OPTION_CORS="true"

# Optional: Enable gzip compression
# export S3PROXY_OPTION_GZIP="true"

echo "Configuration:"
echo "  Bucket: $S3PROXY_AWS_BUCKET"
echo "  Region: $S3PROXY_AWS_REGION"
echo "  Endpoint: $S3PROXY_AWS_ENDPOINT"
echo "  Access Key: ${S3PROXY_AWS_KEY:0:10}..."
echo ""
echo "Starting s3-proxy on port 8080..."
echo "Press Ctrl+C to stop"
echo ""

# Start the proxy in background
./s3-proxy -port 8080 &
PROXY_PID=$!

# Wait a moment for server to start
sleep 2

# Test the proxy
echo "Testing proxy..."
echo ""

# Test 1: List root (might fail if bucket is empty or requires specific path)
echo "Test 1: Testing root path /"
curl -v http://localhost:8080/ 2>&1 | head -20
echo ""
echo ""

# Test 2: Test with a specific path (adjust based on your bucket contents)
echo "Test 2: Testing path /zerofs-data (from your ZeroFS config)"
curl -v http://localhost:8080/zerofs-data 2>&1 | head -20
echo ""
echo ""

# Test 3: Simple HEAD request
echo "Test 3: HEAD request to root"
curl -I http://localhost:8080/ 2>&1
echo ""

echo ""
echo "Proxy is running. Test it with:"
echo "  curl http://localhost:8080/<path-to-object>"
echo ""
echo "To stop the proxy, run: kill $PROXY_PID"
echo "Or press Ctrl+C and run: pkill s3-proxy"

# Wait for user interrupt
wait $PROXY_PID

