# S3 Proxy

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

## Configuration

**Environment variables:**
- `S3PROXY_AWS_KEY` (required)
- `S3PROXY_AWS_SECRET` (required)
- `S3PROXY_AWS_REGION` (default: us-east-1)
- `S3PROXY_AWS_BUCKET` (required)
- `S3PROXY_AWS_ENDPOINT` (optional)

**Multi-bucket mode:** Set `S3PROXY_CONFIG` as YAML or JSON array. See `examples/` for configuration templates.

**Hot-reload:** Use `-config-file` flag for real-time configuration updates without restart.

## Features

- Range requests, DELETE, conditional writes
- Multi-bucket/backend support
- YAML config with hot-reload
- Optimized for ZeroFS

## License

MIT
