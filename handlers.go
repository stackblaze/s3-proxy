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
		// Check if this is a LIST operation (S3 ListObjectsV2)
		if r.URL.Query().Get("list-type") == "2" {
			handleList(proxy, r, w, bucketName)
			return
		}

		path := r.URL.Path
		
		// Handle bucket name in path (e.g., /sd-fs-proxy/...)
		// Since we're a single-bucket proxy, strip the bucket name if present
		path = strings.Trim(path, "/")
		parts := strings.Split(path, "/")
		
		// If path starts with the bucket name, remove it
		// This handles cases like /sd-fs-proxy/path/to/file
		if len(parts) > 0 && parts[0] == bucketName {
			// Remove the bucket name from the path
			parts = parts[1:]
			if len(parts) == 0 {
				path = "/"
			} else {
				path = "/" + strings.Join(parts, "/")
			}
		} else if len(parts) > 0 {
			// Path doesn't start with bucket name, use as-is
			path = "/" + strings.Join(parts, "/")
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

		switch r.Method {
		case http.MethodGet, http.MethodHead:
			handleGet(proxy, key, w, r)
		case http.MethodPut:
			handlePut(proxy, key, w, r)
		case http.MethodDelete:
			handleDelete(proxy, key, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleGet(proxy S3Proxy, key string, w http.ResponseWriter, r *http.Request) {
	// Check for Range header
	rangeHeader := r.Header.Get("Range")
	
	var obj *s3.GetObjectOutput
	var err error
	
	if rangeHeader != "" {
		obj, err = proxy.GetWithRange(key, rangeHeader)
	} else {
		obj, err = proxy.Get(key)
	}
	
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

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

	// Set status code for partial content (Range requests)
	// Headers must be set before WriteHeader
	if rangeHeader != "" && obj.ContentRange != nil {
		w.WriteHeader(http.StatusPartialContent)
	}
	// For regular requests, status 200 is set automatically by http.ResponseWriter

	if r.Method == http.MethodHead {
		return
	}

	io.Copy(w, obj.Body)
}

func handleDelete(proxy S3Proxy, key string, w http.ResponseWriter, r *http.Request) {
	result, err := proxy.Delete(key)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchBucket:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// S3 DeleteObject returns 204 No Content on success
	if result.DeleteMarker != nil && *result.DeleteMarker {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
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
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchBucket:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return ETag in response
	if result.ETag != nil {
		w.Header().Set("ETag", *result.ETag)
	}
	w.WriteHeader(http.StatusOK)
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
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchBucket:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
