package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Proxy interface {
	Get(key string, rangeHeader string) (*s3.GetObjectOutput, error)
	Put(key string, body io.ReadSeeker, contentType string) (*s3.PutObjectOutput, error)
	Head(key string) (*s3.HeadObjectOutput, error)
	ListObjects(prefix string, delimiter string, maxKeys int64, continuationToken string) (*s3.ListObjectsV2Output, error)
	CreateMultipartUpload(key string, contentType string) (*s3.CreateMultipartUploadOutput, error)
	UploadPart(key string, uploadId string, partNumber int64, body io.ReadSeeker) (*s3.UploadPartOutput, error)
	CompleteMultipartUpload(key string, uploadId string, parts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error)
	AbortMultipartUpload(key string, uploadId string) (*s3.AbortMultipartUploadOutput, error)
	ListMultipartUploads(prefix string, delimiter string, maxUploads int64) (*s3.ListMultipartUploadsOutput, error)
	GetWebsiteConfig() (*s3.GetBucketWebsiteOutput, error)
}

type RealS3Proxy struct {
	bucket string
	s3     *s3.S3
}

func NewS3Proxy(key, secret, region, bucket, endpoint string) S3Proxy {
	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	}
	
	// Add custom endpoint if provided (for S3-compatible services like Wasabi)
	if endpoint != "" {
		// Wasabi requires region-specific endpoints (e.g., s3.us-east-1.wasabisys.com)
		// If a generic endpoint is provided, convert it to region-specific format
		normalizedEndpoint := normalizeWasabiEndpoint(endpoint, region)
		
		// Debug logging to verify endpoint normalization
		if endpoint != normalizedEndpoint {
			fmt.Printf("Endpoint normalized: %s -> %s (region: %s)\n", endpoint, normalizedEndpoint, region)
		}
		
		cfg.Endpoint = aws.String(normalizedEndpoint)
		cfg.S3ForcePathStyle = aws.Bool(true) // Required for custom endpoints
	}
	
	sess := session.Must(session.NewSession(cfg))

	return &RealS3Proxy{
		bucket: bucket,
		s3:     s3.New(sess),
	}
}

// normalizeWasabiEndpoint converts generic Wasabi endpoints to region-specific ones
// Wasabi requires region-specific endpoints for proper signature calculation
// Example: s3.wasabisys.com -> s3.us-east-1.wasabisys.com
func normalizeWasabiEndpoint(endpoint, region string) string {
	// If endpoint already contains the region, return as-is
	if strings.Contains(endpoint, region) {
		return endpoint
	}
	
	// Check if it's a Wasabi generic endpoint
	if strings.Contains(endpoint, "wasabisys.com") {
		// Replace s3.wasabisys.com with s3.{region}.wasabisys.com
		if strings.Contains(endpoint, "s3.wasabisys.com") {
			return strings.Replace(endpoint, "s3.wasabisys.com", "s3."+region+".wasabisys.com", 1)
		}
		// If it's already https://s3.wasabisys.com, convert to region-specific
		if strings.HasPrefix(endpoint, "https://s3.wasabisys.com") {
			return "https://s3." + region + ".wasabisys.com"
		}
	}
	
	// For other S3-compatible services, return as-is
	return endpoint
}

func (p *RealS3Proxy) Get(key string, rangeHeader string) (*s3.GetObjectOutput, error) {
	req := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}
	
	// Support HTTP Range requests
	if rangeHeader != "" {
		req.Range = aws.String(rangeHeader)
	}

	return p.s3.GetObject(req)
}

func (p *RealS3Proxy) Head(key string) (*s3.HeadObjectOutput, error) {
	req := &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}

	return p.s3.HeadObject(req)
}

func (p *RealS3Proxy) Put(key string, body io.ReadSeeker, contentType string) (*s3.PutObjectOutput, error) {
	req := &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}

	return p.s3.PutObject(req)
}

func (p *RealS3Proxy) ListObjects(prefix string, delimiter string, maxKeys int64, continuationToken string) (*s3.ListObjectsV2Output, error) {
	req := &s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucket),
	}

	if prefix != "" {
		req.Prefix = aws.String(prefix)
	}

	if delimiter != "" {
		req.Delimiter = aws.String(delimiter)
	}

	if maxKeys > 0 {
		req.MaxKeys = aws.Int64(maxKeys)
	}

	if continuationToken != "" {
		req.ContinuationToken = aws.String(continuationToken)
	}

	return p.s3.ListObjectsV2(req)
}

func (p *RealS3Proxy) CreateMultipartUpload(key string, contentType string) (*s3.CreateMultipartUploadOutput, error) {
	req := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	return p.s3.CreateMultipartUpload(req)
}

func (p *RealS3Proxy) UploadPart(key string, uploadId string, partNumber int64, body io.ReadSeeker) (*s3.UploadPartOutput, error) {
	req := &s3.UploadPartInput{
		Bucket:     aws.String(p.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadId),
		PartNumber: aws.Int64(partNumber),
		Body:       body,
	}

	return p.s3.UploadPart(req)
}

func (p *RealS3Proxy) CompleteMultipartUpload(key string, uploadId string, parts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	req := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(p.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
	}

	return p.s3.CompleteMultipartUpload(req)
}

func (p *RealS3Proxy) AbortMultipartUpload(key string, uploadId string) (*s3.AbortMultipartUploadOutput, error) {
	req := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(p.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadId),
	}

	return p.s3.AbortMultipartUpload(req)
}

func (p *RealS3Proxy) ListMultipartUploads(prefix string, delimiter string, maxUploads int64) (*s3.ListMultipartUploadsOutput, error) {
	req := &s3.ListMultipartUploadsInput{
		Bucket: aws.String(p.bucket),
	}

	if prefix != "" {
		req.Prefix = aws.String(prefix)
	}

	if delimiter != "" {
		req.Delimiter = aws.String(delimiter)
	}

	if maxUploads > 0 {
		req.MaxUploads = aws.Int64(maxUploads)
	}

	return p.s3.ListMultipartUploads(req)
}

func (p *RealS3Proxy) GetWebsiteConfig() (*s3.GetBucketWebsiteOutput, error) {
	req := &s3.GetBucketWebsiteInput{
		Bucket: aws.String(p.bucket),
	}

	return p.s3.GetBucketWebsite(req)
}
