# Testing s3-proxy with ZeroFS Configuration

## Quick Start

### 1. Set Environment Variables

Based on your ZeroFS config:

```bash
export S3PROXY_AWS_KEY="X7SMWFBIMHK761MZDCM4"
export S3PROXY_AWS_SECRET="HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
export S3PROXY_AWS_REGION="us-east-1"
export S3PROXY_AWS_BUCKET="sd-zerofs"
export S3PROXY_AWS_ENDPOINT="https://s3.wasabisys.com"
```

### 2. Start the Proxy

```bash
./s3-proxy -port 8080
```

### 3. Test the Proxy

```bash
# Test root path
curl http://localhost:8080/

# Test specific object (adjust path based on your bucket)
curl http://localhost:8080/zerofs-data

# Test with verbose output
curl -v http://localhost:8080/zerofs-data
```

## Wasabi Endpoints

If you're using a different Wasabi region, update `S3PROXY_AWS_ENDPOINT`:

- **us-east-1**: `https://s3.wasabisys.com`
- **us-east-2**: `https://s3.us-east-2.wasabisys.com`
- **us-west-1**: `https://s3.us-west-1.wasabisys.com`
- **eu-central-1**: `https://s3.eu-central-1.wasabisys.com`

## Using the Test Script

Run the automated test script:

```bash
./test-zerofs-config.sh
```

This will:
1. Set all environment variables
2. Start the proxy on port 8080
3. Run some basic tests
4. Keep the proxy running for manual testing

## Troubleshooting

### Error: "AWS Key not specified"
- Make sure all environment variables are set
- Check with: `env | grep S3PROXY`

### Error: "NoSuchBucket" or "AccessDenied"
- Verify your bucket name is correct
- Check that your AWS credentials have read access to the bucket
- Verify the Wasabi endpoint matches your region

### Error: Connection refused
- Make sure the proxy is running: `ps aux | grep s3-proxy`
- Check if port 8080 is already in use: `lsof -i :8080`

### Testing with ZeroFS

Once s3-proxy is running on `localhost:8080`, your ZeroFS config should work:

```toml
[aws]
access_key_id = "X7SMWFBIMHK761MZDCM4"
secret_access_key = "HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
endpoint = "http://localhost:8080"  # Points to s3-proxy
allow_http = "true"
```

## Notes

- The proxy maps URL paths directly to S3 object keys
- Example: `http://localhost:8080/path/to/file.txt` → S3 key `path/to/file.txt`
- If your ZeroFS storage URL is `s3://sd-zerofs/zerofs-data`, the object key is `zerofs-data`

