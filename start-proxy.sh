#!/bin/bash

# Start s3-proxy with sd-fs-proxy bucket configuration

export S3PROXY_AWS_KEY="X7SMWFBIMHK761MZDCM4"
export S3PROXY_AWS_SECRET="HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
export S3PROXY_AWS_REGION="us-east-1"
export S3PROXY_AWS_BUCKET="sd-fs-proxy"
export S3PROXY_AWS_ENDPOINT="https://s3.wasabisys.com"

echo "Starting s3-proxy with configuration:"
echo "  Bucket: $S3PROXY_AWS_BUCKET"
echo "  Region: $S3PROXY_AWS_REGION"
echo "  Endpoint: $S3PROXY_AWS_ENDPOINT"
echo "  Port: 8080"
echo ""

cd "$(dirname "$0")"
./s3-proxy -port 8080

