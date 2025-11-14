# Wasabi S3 Signature Issue - Root Cause Analysis

## Problem Summary
ZeroFS fails to start with `SignatureDoesNotMatch` error when using s3-proxy to connect to Wasabi S3.

## Root Cause
The issue is a **signature calculation mismatch** caused by the proxy architecture:

### How it Currently Works:
1. **ZeroFS** signs requests with AWS SigV4 for `http://localhost:8080`
2. **s3-proxy** receives the signed request
3. **s3-proxy** extracts the object key and creates a **NEW** S3 request using AWS SDK
4. **s3-proxy** signs the new request with credentials for `https://s3.us-east-1.wasabisys.com`
5. **Wasabi** receives a request with signature calculated for a different endpoint

### Why This Fails:
- AWS SigV4 signature includes the **endpoint hostname** in the signature calculation
- ZeroFS calculates: `Signature(localhost:8080, key, method, headers, body)`
- Proxy needs: `Signature(s3.us-east-1.wasabisys.com, key, method, headers, body)`
- These signatures are **fundamentally different** even with the same credentials

## Testing Results

### ✅ Direct Wasabi Access Works
```bash
mc alias set wasabi-test https://s3.us-east-1.wasabisys.com X7SMWFBIMHK761MZDCM4 HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G
mc ls wasabi-test/sb-zerofs-os/  # SUCCESS
```

### ✅ Proxy Works for Simple Requests
```bash
curl -X PUT http://localhost:8080/test-file -d "test"  # HTTP 200
```

### ❌ Proxy Fails for Pre-Signed Requests
When ZeroFS sends AWS-signed requests to the proxy, the proxy re-signs them, causing signature mismatch.

## Attempted Fixes

### v1.1.4 - Endpoint Normalization
- Added `normalizeWasabiEndpoint()` to convert `s3.wasabisys.com` → `s3.us-east-1.wasabisys.com`
- **Result**: Fixed endpoint format, but signature issue persists

### v1.1.5 - Error Handling
- Added `handleS3Error()` to correctly pass through 403 errors
- **Result**: Improved error transparency, but root cause remains

## Possible Solutions

### Option 1: True Transparent Proxy (Recommended)
Modify s3-proxy to:
- Accept pre-signed requests from clients
- Forward them directly to S3 without re-signing
- Only modify the `Host` header to point to the actual S3 endpoint

**Pros**: Works with any S3 client, maintains original signatures
**Cons**: Requires significant proxy refactoring

### Option 2: Unsigned Client Requests
Configure ZeroFS to send unsigned requests to the proxy:
- Remove credentials from ZeroFS config
- Proxy signs all requests

**Pros**: Simple, works with current proxy implementation
**Cons**: ZeroFS tries to fetch credentials from EC2 metadata service (169.254.169.254)

### Option 3: Direct Wasabi Connection
Skip the proxy entirely:
- Configure ZeroFS to connect directly to Wasabi
- Use `endpoint = "https://s3.us-east-1.wasabisys.com"`

**Pros**: Eliminates signature mismatch
**Cons**: Defeats the purpose of having a proxy

## Current Status
- s3-proxy v1.1.5 is functional for simple requests
- Signature mismatch prevents use with S3 clients that pre-sign requests (like ZeroFS)
- Credentials are valid and work with direct Wasabi access

## Recommendation
Implement **Option 1** (True Transparent Proxy) to make s3-proxy compatible with all S3 clients.
