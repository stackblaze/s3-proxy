# S3 Proxy

A lightweight S3-compatible proxy optimized for ZeroFS. Proxies S3 requests to backend storage (Wasabi, AWS S3, etc.) with support for Range requests, DELETE operations, and conditional writes.

## Quick Start

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e AWS_ACCESS_KEY_ID=your-access-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret-key \
  -e AWS_REGION=us-east-1 \
  -e AWS_BUCKET=your-bucket \
  -e AWS_ENDPOINT=https://s3.wasabisys.com \
  stackblaze/s3-proxy:latest
```

### Local Build

```bash
# Clone and build
git clone https://github.com/stackblaze/s3-proxy.git
cd s3-proxy
go build -o s3-proxy

# Run
./s3-proxy -port 8080
```

### Environment Variables

- `AWS_ACCESS_KEY_ID` - S3 access key (required)
- `AWS_SECRET_ACCESS_KEY` - S3 secret key (required)
- `AWS_REGION` - S3 region (default: us-east-1)
- `AWS_BUCKET` - S3 bucket name (required)
- `AWS_ENDPOINT` - S3 endpoint URL (optional, for Wasabi, MinIO, etc.)

### ZeroFS Configuration

```toml
[aws]
access_key_id = "your-access-key"
secret_access_key = "your-secret-key"
endpoint = "http://localhost:8080"  # Proxy endpoint
allow_http = "true"
```

## Features

- ✅ Range request support (partial file reads)
- ✅ DELETE method support
- ✅ Conditional writes (If-None-Match)
- ✅ Path handling for single-bucket proxy
- ✅ Optimized for ZeroFS NBD devices

## License

MIT
