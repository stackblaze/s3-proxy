# Wasabi S3 Signature Fix - Summary

## Issue
ZeroFS was failing to start with `PermissionDenied` errors when trying to write the bucket ID marker to S3 via s3-proxy. The root cause was `SignatureDoesNotMatch` (403) errors from Wasabi S3.

## Root Cause
Wasabi S3 requires **region-specific endpoints** for proper signature calculation:
- ❌ Generic: `https://s3.wasabisys.com`
- ✅ Region-specific: `https://s3.us-east-1.wasabisys.com`

Using the generic endpoint causes AWS SDK signature calculations to fail, resulting in 403 errors.

## Solution
Added automatic endpoint normalization in `proxy.go`:

```go
func normalizeWasabiEndpoint(endpoint, region string) string {
    // Converts s3.wasabisys.com -> s3.{region}.wasabisys.com
    if strings.Contains(endpoint, "wasabisys.com") {
        if strings.Contains(endpoint, "s3.wasabisys.com") {
            return strings.Replace(endpoint, "s3.wasabisys.com", "s3."+region+".wasabisys.com", 1)
        }
        if strings.HasPrefix(endpoint, "https://s3.wasabisys.com") {
            return "https://s3." + region + ".wasabisys.com"
        }
    }
    return endpoint
}
```

## Testing Results
✅ Endpoint normalization works correctly:
- Input: `https://s3.wasabisys.com` + Region: `us-east-1`
- Output: `https://s3.us-east-1.wasabisys.com`

✅ PUT requests now return HTTP 200 (previously 403)

## Configuration
Users can continue using the generic endpoint format:
```bash
export S3PROXY_AWS_ENDPOINT="https://s3.wasabisys.com"
```

The proxy automatically converts it to the region-specific format based on `S3PROXY_AWS_REGION`.

## Releases
- **v1.1.4**: Initial Wasabi signature fix
- **v1.1.4+debug**: Added debug logging to verify endpoint normalization

## Next Steps for Users
1. Update to v1.1.4 or later
2. Restart s3-proxy
3. ZeroFS should now start successfully without `PermissionDenied` errors

## Backward Compatibility
✅ Other S3-compatible services (MinIO, AWS S3, etc.) are unaffected
✅ Region-specific Wasabi endpoints are passed through unchanged
✅ No configuration changes required

