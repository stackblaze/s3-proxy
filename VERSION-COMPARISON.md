# Version Comparison: v1.0.0 vs v1.1.5

## User Report
"current version broke how zerofs initialize the database - version 1 used to work"

## Key Changes Between v1.0.0 and v1.1.5

### v1.1.0 - Multipart Upload Support
- Added routing for multipart upload operations (POST with `?uploads`, `?uploadId`, `?partNumber`)
- Multipart checks happen **BEFORE** path normalization
- This could potentially interfere with ZeroFS requests

### v1.1.4 - Wasabi Endpoint Normalization
- Added `normalizeWasabiEndpoint()` to convert generic endpoints to region-specific ones
- `https://s3.wasabisys.com` → `https://s3.us-east-1.wasabisys.com`

### v1.1.5 - Error Handling Changes
- Replaced individual error handling with `handleS3Error()` function
- Now passes through actual S3 HTTP status codes (403, 404, etc.)

## Current ZeroFS Behavior

### ✅ What Works
- Bucket ID marker creation
- Storage compatibility check (conditional writes)
- Encryption key initialization
- Basic PUT/GET operations through proxy

### ❌ What Fails
- SlateDB database initialization with error: `LatestManifestMissing`
- This is a **ZeroFS/SlateDB internal error**, not a proxy error

## Question for User
**Did ZeroFS complete full initialization with v1.0.0?**
- Did it get past the "opening SlateDB database" step?
- Did the NFS/9P/NBD servers start successfully?
- Or did it also fail with database errors?

## Hypothesis
The "LatestManifestMissing" error suggests ZeroFS is trying to open an existing database that doesn't have a manifest file. This could be:
1. A ZeroFS version incompatibility (new ZeroFS with old database format)
2. A proxy behavior change that affects how ZeroFS initializes its database
3. A missing S3 operation that ZeroFS needs during initialization

## Next Steps
1. Confirm if v1.0.0 actually worked end-to-end with ZeroFS
2. If yes, identify which specific change broke it
3. Test each version (v1.1.0, v1.1.1, etc.) to find the breaking change
