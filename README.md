# S3 Proxy

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight S3-compatible proxy optimized for ZeroFS. Supports Range requests, DELETE operations, and conditional writes.

## Quick Start

### Docker

```bash
docker run -d -p 8080:8080 \
  -e S3PROXY_AWS_KEY=your-key \
  -e S3PROXY_AWS_SECRET=your-secret \
  -e S3PROXY_AWS_REGION=us-east-1 \
  -e S3PROXY_AWS_BUCKET=your-bucket \
  -e S3PROXY_AWS_ENDPOINT=https://s3.wasabisys.com \
  ghcr.io/stackblaze/s3-proxy:latest
```

### Install Binary

```bash
curl -sSfL https://raw.githubusercontent.com/stackblaze/s3-proxy/main/scripts/install.sh | sh
```

Or download from [releases](https://github.com/stackblaze/s3-proxy/releases).

### Local Build

```bash
git clone https://github.com/stackblaze/s3-proxy.git && cd s3-proxy
go build -o s3-proxy && ./s3-proxy -port 8080
```

## Configuration

### Single Mode

Environment variables:
- `S3PROXY_AWS_KEY` (required)
- `S3PROXY_AWS_SECRET` (required)
- `S3PROXY_AWS_REGION` (default: us-east-1)
- `S3PROXY_AWS_BUCKET` (required)
- `S3PROXY_AWS_ENDPOINT` (optional)

### Multi Mode

Set `S3PROXY_CONFIG` as YAML or JSON array:

```yaml
- host: wasabi.localhost
  awsKey: key
  awsSecret: secret
  awsRegion: us-east-1
  awsBucket: bucket
  awsEndpoint: https://s3.wasabisys.com
```

### Hot-Reload

Use `-config-file` for real-time updates without restart:

```bash
./s3-proxy -config-file config.yaml -port 8080
```

Configuration reloads automatically on file changes (~100ms).

## ZeroFS

```toml
[aws]
access_key_id = "your-key"
secret_access_key = "your-secret"
endpoint = "http://localhost:8080"
allow_http = "true"
```

## Features

- Range requests, DELETE, conditional writes
- Multi-bucket/backend support
- YAML config with hot-reload
- Optimized for ZeroFS

## Testing

```bash
go test ./...        # Run all tests
make test           # Using Makefile
```

## License

MIT
