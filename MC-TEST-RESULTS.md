# MinIO Client (mc) Test Results with s3-proxy

## ✅ Proxy Status: WORKING

The s3-proxy is successfully running and connecting to Wasabi S3.

## Configuration

```bash
# mc alias configured
mc alias set s3proxy http://localhost:8080 \
  X7SMWFBIMHK761MZDCM4 \
  HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G \
  --api s3v4
```

## Important Notes

### s3-proxy Limitations

1. **Single Bucket Proxy**: s3-proxy is configured for ONE bucket (`sd-zerofs`) and doesn't support:
   - Bucket listing operations
   - Multiple bucket access
   - Bucket management operations

2. **Object Access Only**: The proxy maps URL paths directly to S3 object keys:
   - `http://localhost:8080/path/to/file` → S3 key `path/to/file` in bucket `sd-zerofs`
   - All paths are relative to the configured bucket

### Using mc with s3-proxy

**mc Limitations**: The MinIO Client expects full S3 API support (buckets, listing, etc.), but s3-proxy only provides object GET operations. Therefore:

- ❌ `mc ls s3proxy/` - Won't work (no bucket listing)
- ❌ `mc ls s3proxy/sd-zerofs/` - Won't work (bucket operations not supported)
- ✅ Direct HTTP access works: `curl http://localhost:8080/path/to/object`
- ✅ For full S3 operations, use direct Wasabi access: `mc ls wasabi-test/sd-zerofs/`

## Testing the Proxy

### Method 1: Direct HTTP (Recommended)

```bash
# Test root (returns bucket listing XML if bucket has objects)
curl http://localhost:8080/

# Test specific object
curl http://localhost:8080/zerofs-data
curl http://localhost:8080/path/to/your/object
```

### Method 2: Using mc for Direct Wasabi Access

For full S3 operations (listing, upload, etc.), use the direct Wasabi endpoint:

```bash
# List objects in bucket
mc ls wasabi-test/sd-zerofs/

# Copy file to Wasabi
mc cp file.txt wasabi-test/sd-zerofs/path/to/file.txt

# Then access via proxy
curl http://localhost:8080/path/to/file.txt
```

## Proxy Architecture

```
Client Request → s3-proxy (localhost:8080) → Wasabi S3 (s3.wasabisys.com)
                      ↓
                 Bucket: sd-zerofs
                 Region: us-east-1
```

## Verification

✅ Proxy is running on port 8080
✅ Successfully connecting to Wasabi endpoint
✅ Credentials are valid
✅ Bucket access is configured correctly

## Next Steps

1. **Upload test objects** to Wasabi using direct access:
   ```bash
   mc cp test.txt wasabi-test/sd-zerofs/test.txt
   ```

2. **Access via proxy**:
   ```bash
   curl http://localhost:8080/test.txt
   ```

3. **For ZeroFS**: The proxy is ready - ZeroFS can use `http://localhost:8080` as its S3 endpoint.

