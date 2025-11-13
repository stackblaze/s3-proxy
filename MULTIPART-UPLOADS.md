# S3 Proxy - Multipart Upload Support

The s3-proxy now fully supports S3 multipart uploads, including all required operations.

## Supported Operations

### 1. InitiateMultipartUpload
**Request:** `POST /object-key?uploads`
**Response:** XML with `UploadId`

```bash
curl -X POST "http://localhost:8080/test-file.txt?uploads" \
  -H "Content-Type: text/plain"
```

### 2. UploadPart
**Request:** `PUT /object-key?uploadId=...&partNumber=1`
**Response:** ETag header

```bash
curl -X PUT "http://localhost:8080/test-file.txt?uploadId=...&partNumber=1" \
  --data-binary @part1.bin
```

### 3. CompleteMultipartUpload
**Request:** `POST /object-key?uploadId=...`
**Body:** XML with completed parts
**Response:** XML with final ETag

```bash
curl -X POST "http://localhost:8080/test-file.txt?uploadId=..." \
  -H "Content-Type: application/xml" \
  -d '<CompleteMultipartUpload>
    <Part>
      <PartNumber>1</PartNumber>
      <ETag>"etag1"</ETag>
    </Part>
  </CompleteMultipartUpload>'
```

### 4. AbortMultipartUpload
**Request:** `DELETE /object-key?uploadId=...`
**Response:** 204 No Content

```bash
curl -X DELETE "http://localhost:8080/test-file.txt?uploadId=..."
```

### 5. ListMultipartUploads
**Request:** `GET /?uploads`
**Response:** XML list of in-progress uploads

```bash
curl -X GET "http://localhost:8080/?uploads"
```

## Implementation Details

- **POST Method Support:** The proxy now accepts POST requests for multipart operations
- **Query Parameter Detection:** Uses `query["uploads"]` to detect presence (even if empty value)
- **S3-Compatible XML:** All responses follow S3 XML format specifications
- **Error Handling:** Proper HTTP status codes for S3 errors (404, 500, etc.)

## Testing

All multipart upload operations have been tested and verified to work correctly with S3-compatible storage backends (including Wasabi).

