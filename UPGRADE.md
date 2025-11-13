# Upgrading S3-Proxy

## Quick Upgrade

If you installed using the install script, simply run it again to get the latest version:

```bash
curl -sSfL https://raw.githubusercontent.com/stackblaze/s3-proxy/main/scripts/install.sh | sh
```

The script will automatically:
- Download the latest release
- Replace your existing binary
- Preserve your configuration

## Manual Upgrade

### Option 1: Download Latest Binary

```bash
# Stop the running proxy (if using systemd)
sudo systemctl stop s3-proxy

# Download latest release
VERSION="v1.1.0"  # or "latest" for newest
curl -L -o s3-proxy https://github.com/stackblaze/s3-proxy/releases/download/${VERSION}/s3-proxy

# Make executable
chmod +x s3-proxy

# Replace existing binary
sudo mv s3-proxy /usr/local/bin/s3-proxy  # or wherever you installed it

# Restart the proxy
sudo systemctl start s3-proxy
```

### Option 2: Using GitHub Releases API

```bash
# Get latest release version
LATEST=$(curl -s https://api.github.com/repos/stackblaze/s3-proxy/releases/latest | grep tag_name | cut -d '"' -f 4)

# Download binary
curl -L -o s3-proxy https://github.com/stackblaze/s3-proxy/releases/download/${LATEST}/s3-proxy

# Install
chmod +x s3-proxy
sudo mv s3-proxy /usr/local/bin/s3-proxy
```

### Option 3: Docker Users

```bash
# Pull latest image
docker pull ghcr.io/stackblaze/s3-proxy:latest

# Restart container
docker restart s3-proxy
```

## What's New in v1.1.0

- ✅ **Multipart Upload Support** - Full S3 multipart upload API
- ✅ **POST Request Handling** - Fixed 405 errors on POST requests
- ✅ **Backward Compatible** - No breaking changes, drop-in replacement

## Verification

After upgrading, verify the new version:

```bash
# Check version (if available)
./s3-proxy --version

# Or test multipart upload support
curl -X POST "http://localhost:8080/test.txt?uploads" \
  -H "Content-Type: text/plain"
```

You should get an XML response with an `UploadId` if multipart uploads are working.

## Rollback

If you need to rollback to a previous version:

```bash
# Download specific version
VERSION="v1.0.0"  # or your previous version
curl -L -o s3-proxy https://github.com/stackblaze/s3-proxy/releases/download/${VERSION}/s3-proxy
chmod +x s3-proxy
sudo mv s3-proxy /usr/local/bin/s3-proxy
sudo systemctl restart s3-proxy
```

## Notes

- **No Configuration Changes Required** - v1.1.0 is fully backward compatible
- **Zero Downtime** - You can upgrade without changing your configuration
- **Same Environment Variables** - All existing env vars work the same way

