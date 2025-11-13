# S3 Proxy

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight S3-compatible proxy optimized for ZeroFS. Proxies S3 requests to backend storage (Wasabi, AWS S3, Backblaze, etc.) with support for Range requests, DELETE operations, and conditional writes.

**Open Source** - Licensed under MIT License

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

**Single Mode** (one bucket):
- `S3PROXY_AWS_KEY` - S3 access key (required)
- `S3PROXY_AWS_SECRET` - S3 secret key (required)
- `S3PROXY_AWS_REGION` - S3 region (default: us-east-1)
- `S3PROXY_AWS_BUCKET` - S3 bucket name (required)
- `S3PROXY_AWS_ENDPOINT` - S3 endpoint URL (optional, for Wasabi, Backblaze, MinIO, etc.)

**Multi Mode** (multiple buckets/backends):
- `S3PROXY_CONFIG` - YAML or JSON array of configurations (see example below)

### Multiple Buckets / Backends

Configure multiple buckets with different backends (Backblaze, Wasabi, AWS S3, etc.):

**YAML Format (Recommended):**
```yaml
- host: wasabi.localhost
  awsKey: wasabi-access-key
  awsSecret: wasabi-secret-key
  awsRegion: us-east-1
  awsBucket: my-wasabi-bucket
  awsEndpoint: https://s3.wasabisys.com

- host: backblaze.localhost
  awsKey: backblaze-key-id
  awsSecret: backblaze-application-key
  awsRegion: us-west-004
  awsBucket: my-backblaze-bucket
  awsEndpoint: https://s3.us-west-004.backblazeb2.com

- host: aws.localhost
  awsKey: aws-access-key
  awsSecret: aws-secret-key
  awsRegion: us-east-1
  awsBucket: my-aws-bucket
```

**JSON Format (also supported):**
```bash
export S3PROXY_CONFIG='[{"host":"wasabi.localhost","awsKey":"key","awsSecret":"secret","awsRegion":"us-east-1","awsBucket":"bucket","awsEndpoint":"https://s3.wasabisys.com"}]'
```

**Docker Example (Environment Variable):**
```bash
docker run -d \
  -p 8080:8080 \
  -e S3PROXY_CONFIG='[{"host":"wasabi.localhost","awsKey":"key","awsSecret":"secret","awsRegion":"us-east-1","awsBucket":"bucket","awsEndpoint":"https://s3.wasabisys.com"}]' \
  stackblaze/s3-proxy:latest
```

### Hot-Reload Configuration (No Restart Required)

Use a config file for real-time configuration updates without restarting:

```bash
# Create YAML config file
cat > config.yaml << 'EOF'
- host: wasabi.localhost
  awsKey: wasabi-key
  awsSecret: wasabi-secret
  awsRegion: us-east-1
  awsBucket: my-bucket
  awsEndpoint: https://s3.wasabisys.com
EOF

# Start with config file (auto-reloads on changes)
./s3-proxy -config-file config.yaml -port 8080
```

**Docker with Config File:**
```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  stackblaze/s3-proxy:latest \
  -config-file /app/config.yaml
```

**How it works:**
- Edit `config.yaml` and save
- Configuration reloads automatically within ~100ms
- No proxy restart needed
- Changes take effect immediately

**Note:** Multi-mode routes requests based on the `Host` header. Configure DNS or use `/etc/hosts` to point different hostnames to the proxy.

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
