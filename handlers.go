package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func NewSSLRedirectHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme != "https" {
			dest := "https://" + r.Host + r.URL.Path
			if r.URL.RawQuery != "" {
				dest += "?" + r.URL.RawQuery
			}

			http.Redirect(w, r, dest, http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	}
}

type HostDispatchingHandler struct {
	hosts map[string]http.Handler
}

func NewHostDispatchingHandler() *HostDispatchingHandler {
	return &HostDispatchingHandler{
		hosts: make(map[string]http.Handler),
	}
}

func (h *HostDispatchingHandler) HandleHost(host string, handler http.Handler) {
	h.hosts[host] = handler
}

func (h *HostDispatchingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := h.hosts[getHost(r)]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	handler.ServeHTTP(w, r)
}

func NewBasicAuthHandler(users []User, next http.Handler) http.HandlerFunc {
	m := make(map[string]string)
	for _, u := range users {
		m[u.Name] = u.Password
	}

	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			challenge(w, r)
			return
		}

		p, ok := m[username]
		if !ok {
			challenge(w, r)
			return
		}

		if password != p {
			challenge(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func NewWebsiteHandler(next http.Handler, cfg *s3.GetBucketWebsiteOutput) http.HandlerFunc {
	suffix := cfg.IndexDocument.Suffix

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "" || path[len(path)-1] == '/' {
			r.URL.Path += *suffix
		}

		next.ServeHTTP(w, r)
	}
}

func NewProxyHandler(proxy S3Proxy, prefix string, bucketName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for multipart upload operations FIRST (before path normalization)
		query := r.URL.Query()
		_, hasUploads := query["uploads"] // Check if parameter exists (even if empty)
		uploadId := query.Get("uploadId")
		partNumber := query.Get("partNumber")

		// InitiateMultipartUpload: POST with ?uploads (parameter may be empty, just needs to be present)
		if hasUploads && r.Method == http.MethodPost {
			handleCreateMultipartUpload(proxy, r, w)
			return
		}

		// ListMultipartUploads: GET with ?uploads
		if hasUploads && r.Method == http.MethodGet {
			handleListMultipartUploads(proxy, r, w, bucketName)
			return
		}

		// CompleteMultipartUpload: POST with ?uploadId
		if uploadId != "" && r.Method == http.MethodPost {
			handleCompleteMultipartUpload(proxy, r, w)
			return
		}

		// AbortMultipartUpload: DELETE with ?uploadId
		if uploadId != "" && r.Method == http.MethodDelete {
			handleAbortMultipartUpload(proxy, r, w)
			return
		}

		// UploadPart: PUT with ?uploadId and ?partNumber
		if uploadId != "" && partNumber != "" && r.Method == http.MethodPut {
			handleUploadPart(proxy, r, w)
			return
		}

		// Check if this is a LIST operation (S3 ListObjectsV2)
		if r.URL.Query().Get("list-type") == "2" {
			handleList(proxy, r, w, bucketName)
			return
		}

		// Extract key from path for regular operations
		path := r.URL.Path
		
		// Handle bucket name in path (e.g., /sd-zerofs/...)
		// Since we're a single-bucket proxy, strip the bucket name if present
		path = strings.Trim(path, "/")
		parts := strings.Split(path, "/")
		
		// If path starts with what looks like a bucket name, remove it
		// This handles cases like /sd-zerofs/path/to/file
		if len(parts) > 1 {
			// Assume first part might be bucket name, but we can't be sure
			// For now, use the full path and let S3 handle it
			path = "/" + strings.Join(parts, "/")
		} else if len(parts) == 1 && parts[0] != "" {
			// Single part - could be bucket name or object key
			path = "/" + parts[0]
		} else {
			path = "/"
		}
		
		if prefix != "" {
			path = "/" + prefix + path
		}

		// Normalize path: remove leading slash for S3 key (S3 keys don't start with /)
		key := path
		if len(key) > 0 && key[0] == '/' {
			key = key[1:]
		}

		// Regular operations
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			handleGet(proxy, key, w, r)
		case http.MethodPut:
			handlePut(proxy, key, w, r)
		case http.MethodPost:
			// POST without multipart params - treat as regular PUT
			handlePut(proxy, key, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleGet(proxy S3Proxy, key string, w http.ResponseWriter, r *http.Request) {
	obj, err := proxy.Get(key)
	if err != nil {
		handleS3Error(w, err)
		return
	}

	setHeader(w, "Cache-Control", s2s(obj.CacheControl))
	setHeader(w, "Content-Disposition", s2s(obj.ContentDisposition))
	setHeader(w, "Content-Encoding", s2s(obj.ContentEncoding))
	setHeader(w, "Content-Language", s2s(obj.ContentLanguage))
	setHeader(w, "Content-Length", i2s(obj.ContentLength))
	setHeader(w, "Content-Range", s2s(obj.ContentRange))
	setHeader(w, "Content-Type", s2s(obj.ContentType))
	setHeader(w, "ETag", s2s(obj.ETag))
	setHeader(w, "Expires", s2s(obj.Expires))
	setHeader(w, "Last-Modified", t2s(obj.LastModified))

	if r.Method == http.MethodHead {
		return
	}

	io.Copy(w, obj.Body)
}

func handlePut(proxy S3Proxy, key string, w http.ResponseWriter, r *http.Request) {
	// Check conditional write headers
	if !checkConditionalWrite(proxy, key, r, w) {
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Put object to S3
	result, err := proxy.Put(key, bytes.NewReader(body), contentType)
	if err != nil {
		handleS3Error(w, err)
		return
	}

	// Return ETag in response
	if result.ETag != nil {
		w.Header().Set("ETag", *result.ETag)
	}
	w.WriteHeader(http.StatusOK)
}

func handleCreateMultipartUpload(proxy S3Proxy, r *http.Request, w http.ResponseWriter) {
	key := extractKeyFromPath(r.URL.Path)
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := proxy.CreateMultipartUpload(key, contentType)
	if err != nil {
		handleS3Error(w, err)
		return
	}

	// Return XML response
	type InitiateMultipartUploadResult struct {
		XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
		Xmlns    string   `xml:"xmlns,attr"`
		Bucket   string   `xml:"Bucket"`
		Key      string   `xml:"Key"`
		UploadId string   `xml:"UploadId"`
	}

	response := InitiateMultipartUploadResult{
		Xmlns:    "http://s3.amazonaws.com/doc/2006-03-01/",
		Bucket:   "",
		Key:      aws.StringValue(result.Key),
		UploadId: aws.StringValue(result.UploadId),
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode XML response", http.StatusInternalServerError)
		return
	}
}

func handleUploadPart(proxy S3Proxy, r *http.Request, w http.ResponseWriter) {
	key := extractKeyFromPath(r.URL.Path)
	uploadId := r.URL.Query().Get("uploadId")
	partNumberStr := r.URL.Query().Get("partNumber")

	partNumber, err := strconv.ParseInt(partNumberStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid partNumber", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	result, err := proxy.UploadPart(key, uploadId, partNumber, bytes.NewReader(body))
	if err != nil {
		handleS3Error(w, err)
		return
	}

	// Return ETag in response
	if result.ETag != nil {
		w.Header().Set("ETag", *result.ETag)
	}
	w.WriteHeader(http.StatusOK)
}

func handleCompleteMultipartUpload(proxy S3Proxy, r *http.Request, w http.ResponseWriter) {
	key := extractKeyFromPath(r.URL.Path)
	uploadId := r.URL.Query().Get("uploadId")

	// Parse the CompleteMultipartUpload XML from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	type CompleteMultipartUpload struct {
		XMLName xml.Name `xml:"CompleteMultipartUpload"`
		Parts   []struct {
			PartNumber int64  `xml:"PartNumber"`
			ETag       string `xml:"ETag"`
		} `xml:"Part"`
	}

	var upload CompleteMultipartUpload
	if err := xml.Unmarshal(body, &upload); err != nil {
		http.Error(w, "Invalid XML in request body", http.StatusBadRequest)
		return
	}

	// Convert to S3 CompletedPart format
	parts := make([]*s3.CompletedPart, len(upload.Parts))
	for i, p := range upload.Parts {
		parts[i] = &s3.CompletedPart{
			PartNumber: aws.Int64(p.PartNumber),
			ETag:       aws.String(p.ETag),
		}
	}

	result, err := proxy.CompleteMultipartUpload(key, uploadId, parts)
	if err != nil {
		handleS3Error(w, err)
		return
	}

	// Return XML response
	type CompleteMultipartUploadResult struct {
		XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
		Xmlns    string   `xml:"xmlns,attr"`
		Location string   `xml:"Location"`
		Bucket   string   `xml:"Bucket"`
		Key      string   `xml:"Key"`
		ETag     string   `xml:"ETag"`
	}

	response := CompleteMultipartUploadResult{
		Xmlns:    "http://s3.amazonaws.com/doc/2006-03-01/",
		Location: "",
		Bucket:   "",
		Key:      aws.StringValue(result.Key),
		ETag:     strings.Trim(aws.StringValue(result.ETag), "\""),
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode XML response", http.StatusInternalServerError)
		return
	}
}

func handleAbortMultipartUpload(proxy S3Proxy, r *http.Request, w http.ResponseWriter) {
	key := extractKeyFromPath(r.URL.Path)
	uploadId := r.URL.Query().Get("uploadId")

	_, err := proxy.AbortMultipartUpload(key, uploadId)
	if err != nil {
		handleS3Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleListMultipartUploads(proxy S3Proxy, r *http.Request, w http.ResponseWriter, bucketName string) {
	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")
	maxUploadsStr := r.URL.Query().Get("max-uploads")

	var maxUploads int64 = 1000
	if maxUploadsStr != "" {
		if parsed, err := strconv.ParseInt(maxUploadsStr, 10, 64); err == nil {
			maxUploads = parsed
		}
	}

	result, err := proxy.ListMultipartUploads(prefix, delimiter, maxUploads)
	if err != nil {
		handleS3Error(w, err)
		return
	}

	// Build XML response
	type Upload struct {
		Key       string `xml:"Key"`
		UploadId  string `xml:"UploadId"`
		Initiated string `xml:"Initiated"`
	}

	type ListMultipartUploadsResult struct {
		XMLName            xml.Name `xml:"ListMultipartUploadsResult"`
		Xmlns              string   `xml:"xmlns,attr"`
		Bucket             string   `xml:"Bucket"`
		KeyMarker          string   `xml:"KeyMarker,omitempty"`
		UploadIdMarker     string   `xml:"UploadIdMarker,omitempty"`
		NextKeyMarker      string   `xml:"NextKeyMarker,omitempty"`
		NextUploadIdMarker string   `xml:"NextUploadIdMarker,omitempty"`
		MaxUploads         int64    `xml:"MaxUploads"`
		IsTruncated        bool     `xml:"IsTruncated"`
		Prefix             string   `xml:"Prefix,omitempty"`
		Delimiter          string   `xml:"Delimiter,omitempty"`
		Uploads            []Upload `xml:"Upload,omitempty"`
	}

	response := ListMultipartUploadsResult{
		Xmlns:       "http://s3.amazonaws.com/doc/2006-03-01/",
		Bucket:      bucketName,
		MaxUploads:  maxUploads,
		IsTruncated: aws.BoolValue(result.IsTruncated),
	}

	if prefix != "" {
		response.Prefix = prefix
	}
	if delimiter != "" {
		response.Delimiter = delimiter
	}

	for _, upload := range result.Uploads {
		response.Uploads = append(response.Uploads, Upload{
			Key:       aws.StringValue(upload.Key),
			UploadId:  aws.StringValue(upload.UploadId),
			Initiated: upload.Initiated.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode XML response", http.StatusInternalServerError)
		return
	}
}

func extractKeyFromPath(path string) string {
	// Remove leading slash
	key := strings.Trim(path, "/")
	
	// Handle bucket name in path
	parts := strings.Split(key, "/")
	if len(parts) > 1 {
		// Assume first part might be bucket name
		key = strings.Join(parts[1:], "/")
	}
	
	return key
}

func checkConditionalWrite(proxy S3Proxy, key string, r *http.Request, w http.ResponseWriter) bool {
	// Get current object metadata (if exists)
	headObj, err := proxy.Head(key)
	objectExists := err == nil

	// Check If-Match: ETag must match
	if ifMatch := r.Header.Get("If-Match"); ifMatch != "" {
		if !objectExists {
			w.WriteHeader(http.StatusPreconditionFailed)
			return false
		}
		currentETag := strings.Trim(*headObj.ETag, "\"")
		expectedETag := strings.Trim(ifMatch, "\"")
		if currentETag != expectedETag {
			w.WriteHeader(http.StatusPreconditionFailed)
			return false
		}
	}

	// Check If-None-Match: ETag must NOT match (for create-only)
	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		if objectExists {
			currentETag := strings.Trim(*headObj.ETag, "\"")
			expectedETag := strings.Trim(ifNoneMatch, "\"")
			// If-None-Match: * means object must not exist
			if ifNoneMatch == "*" || currentETag == expectedETag {
				w.WriteHeader(http.StatusPreconditionFailed)
				return false
			}
		}
	}

	// Check If-Modified-Since: Object must be modified after this date
	if ifModifiedSince := r.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
		if !objectExists {
			w.WriteHeader(http.StatusPreconditionFailed)
			return false
		}
		modifiedSince, err := http.ParseTime(ifModifiedSince)
		if err == nil && headObj.LastModified != nil {
			if !headObj.LastModified.After(modifiedSince) {
				w.WriteHeader(http.StatusPreconditionFailed)
				return false
			}
		}
	}

	// Check If-Unmodified-Since: Object must NOT be modified after this date
	if ifUnmodifiedSince := r.Header.Get("If-Unmodified-Since"); ifUnmodifiedSince != "" {
		if !objectExists {
			// Object doesn't exist, so condition passes
			return true
		}
		unmodifiedSince, err := http.ParseTime(ifUnmodifiedSince)
		if err == nil && headObj.LastModified != nil {
			if headObj.LastModified.After(unmodifiedSince) {
				w.WriteHeader(http.StatusPreconditionFailed)
				return false
			}
		}
	}

	return true
}

func handleList(proxy S3Proxy, r *http.Request, w http.ResponseWriter, bucketName string) {
	// Parse query parameters
	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")
	maxKeysStr := r.URL.Query().Get("max-keys")
	continuationToken := r.URL.Query().Get("continuation-token")
	startAfter := r.URL.Query().Get("start-after")

	var maxKeys int64 = 1000 // Default
	if maxKeysStr != "" {
		if parsed, err := strconv.ParseInt(maxKeysStr, 10, 64); err == nil {
			maxKeys = parsed
		}
	}

	// List objects
	result, err := proxy.ListObjects(prefix, delimiter, maxKeys, continuationToken)
	if err != nil {
		handleS3Error(w, err)
		return
	}

	// Build XML response (S3 ListObjectsV2 format)
	type ListEntry struct {
		Key          string `xml:"Key"`
		LastModified string `xml:"LastModified"`
		ETag         string `xml:"ETag"`
		Size         int64  `xml:"Size"`
		StorageClass string `xml:"StorageClass"`
	}

	type CommonPrefix struct {
		Prefix string `xml:"Prefix"`
	}

	type ListBucketResult struct {
		XMLName            xml.Name       `xml:"ListBucketResult"`
		Xmlns              string         `xml:"xmlns,attr"`
		Name               string         `xml:"Name"`
		Prefix             string         `xml:"Prefix,omitempty"`
		StartAfter         string         `xml:"StartAfter,omitempty"`
		MaxKeys            int64          `xml:"MaxKeys"`
		Delimiter          string         `xml:"Delimiter,omitempty"`
		IsTruncated        bool           `xml:"IsTruncated"`
		NextContinuationToken string      `xml:"NextContinuationToken,omitempty"`
		ContinuationToken  string         `xml:"ContinuationToken,omitempty"`
		Contents           []ListEntry    `xml:"Contents,omitempty"`
		CommonPrefixes     []CommonPrefix `xml:"CommonPrefixes,omitempty"`
	}

	response := ListBucketResult{
		Xmlns:       "http://s3.amazonaws.com/doc/2006-03-01/",
		Name:        bucketName, // Bucket name
		MaxKeys:     maxKeys,
		IsTruncated: aws.BoolValue(result.IsTruncated),
	}

	if prefix != "" {
		response.Prefix = prefix
	}
	if startAfter != "" {
		response.StartAfter = startAfter
	}
	if delimiter != "" {
		response.Delimiter = delimiter
	}
	if continuationToken != "" {
		response.ContinuationToken = continuationToken
	}
	if result.NextContinuationToken != nil {
		response.NextContinuationToken = *result.NextContinuationToken
	}

	// Add objects
	for _, obj := range result.Contents {
		entry := ListEntry{
			Key:          aws.StringValue(obj.Key),
			LastModified: obj.LastModified.Format(time.RFC3339),
			ETag:         strings.Trim(aws.StringValue(obj.ETag), "\""),
			Size:         aws.Int64Value(obj.Size),
		}
		if obj.StorageClass != nil {
			entry.StorageClass = *obj.StorageClass
		}
		response.Contents = append(response.Contents, entry)
	}

	// Add common prefixes (for delimiter-based listing)
	for _, prefix := range result.CommonPrefixes {
		response.CommonPrefixes = append(response.CommonPrefixes, CommonPrefix{
			Prefix: aws.StringValue(prefix.Prefix),
		})
	}

	// Set content type and write XML
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)

	// Write XML declaration
	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode XML response", http.StatusInternalServerError)
		return
	}
}

func challenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+getHost(r)+`"`)
	http.Error(w, "", http.StatusUnauthorized)
}

func getHost(r *http.Request) string {
	host := r.Header.Get("Host")
	if host == "" {
		host = r.Host
	}

	return host
}

func s2s(s *string) string {
	if s != nil {
		return *s
	} else {
		return ""
	}
}

func i2s(i *int64) string {
	if i != nil {
		return strconv.FormatInt(*i, 10)
	} else {
		return ""
	}
}

func t2s(t *time.Time) string {
	if t != nil {
		return t.UTC().Format(http.TimeFormat)
	} else {
		return ""
	}
}

func setHeader(w http.ResponseWriter, key, value string) {
	if value != "" {
		w.Header().Add(key, value)
	}
}

// handleS3Error properly maps AWS S3 errors to HTTP status codes
// and passes through the original S3 error status codes (like 403 for SignatureDoesNotMatch)
func handleS3Error(w http.ResponseWriter, err error) {
	// Check if it's an AWS RequestFailure (has HTTP status code)
	if reqErr, ok := err.(awserr.RequestFailure); ok {
		statusCode := reqErr.StatusCode()
		// Pass through the original HTTP status code from S3
		// This ensures 403 SignatureDoesNotMatch is returned as 403, not 500
		http.Error(w, err.Error(), statusCode)
		return
	}

	// Check if it's a generic AWS error
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey:
			http.Error(w, err.Error(), http.StatusNotFound)
		case "SignatureDoesNotMatch", "InvalidAccessKeyId", "AccessDenied":
			// These should be 403, but if RequestFailure wasn't available, use 403
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			// For other AWS errors, try to infer status or use 500
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Non-AWS error
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
