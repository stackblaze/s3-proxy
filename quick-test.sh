#!/bin/bash
# Quick test - just start the proxy with ZeroFS config

export S3PROXY_AWS_KEY="X7SMWFBIMHK761MZDCM4"
export S3PROXY_AWS_SECRET="HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
export S3PROXY_AWS_REGION="us-east-1"
export S3PROXY_AWS_BUCKET="sd-zerofs"
export S3PROXY_AWS_ENDPOINT="https://s3.wasabisys.com"

echo "Starting s3-proxy with ZeroFS configuration..."
echo "Access at: http://localhost:8080"
echo "Press Ctrl+C to stop"
./s3-proxy -port 8080
