# Release v1.1.0 - Multipart Upload Support

## Summary
This release adds full S3 multipart upload support to the s3-proxy, fixing the "405 Method Not Allowed" error on POST requests.

## Changes

### Added
- **Multipart Upload Operations:**
  - `CreateMultipartUpload` - Initiate multipart uploads (POST with `?uploads`)
  - `UploadPart` - Upload individual parts (PUT with `?uploadId` and `?partNumber`)
  - `CompleteMultipartUpload` - Complete multipart uploads (POST with `?uploadId`)
  - `AbortMultipartUpload` - Abort in-progress uploads (DELETE with `?uploadId`)
  - `ListMultipartUploads` - List all in-progress uploads (GET with `?uploads`)

- **S3-Compatible XML Responses:** All multipart operations return properly formatted S3 XML responses
- **Query Parameter Detection:** Fixed detection of `?uploads` parameter (even with empty value)

### Fixed
- **POST Request Handling:** POST requests are now properly routed to multipart handlers
- **405 Method Not Allowed:** Resolved error when clients attempt multipart uploads

## Technical Details

### Implementation
- Added multipart upload methods to `S3Proxy` interface
- Implemented handlers for all multipart operations in `handlers.go`
- Updated `proxy.go` with AWS SDK multipart upload calls
- Fixed query parameter detection to check for parameter presence, not just value

### Testing
All multipart operations have been tested and verified:
- ✅ InitiateMultipartUpload returns XML with UploadId
- ✅ ListMultipartUploads returns list of in-progress uploads
- ✅ Compatible with S3-compatible storage backends (Wasabi, etc.)

## Usage

### Initiate Multipart Upload
```bash
curl -X POST "http://localhost:8080/file.txt?uploads" \
  -H "Content-Type: text/plain"
```

### Upload Part
```bash
curl -X PUT "http://localhost:8080/file.txt?uploadId=...&partNumber=1" \
  --data-binary @part1.bin
```

### Complete Multipart Upload
```bash
curl -X POST "http://localhost:8080/file.txt?uploadId=..." \
  -H "Content-Type: application/xml" \
  -d '<CompleteMultipartUpload>
    <Part>
      <PartNumber>1</PartNumber>
      <ETag>"etag1"</ETag>
    </Part>
  </CompleteMultipartUpload>'
```

## Breaking Changes
None - this is a feature addition that maintains backward compatibility.

## Migration
No migration required. Existing functionality remains unchanged.

## GitHub Release

To create a GitHub release:

1. Go to https://github.com/jcomo/s3-proxy/releases/new
2. Select tag: `v1.1.0`
3. Title: `v1.1.0 - Multipart Upload Support`
4. Description: Copy from this file
5. Attach binary: `s3-proxy` (from build)
6. Publish release

## Build

```bash
go build -o s3-proxy .
```

The binary is ready for distribution.

