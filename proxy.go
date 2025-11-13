package main

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Proxy interface {
	Get(key string) (*s3.GetObjectOutput, error)
	GetWithRange(key string, rangeHeader string) (*s3.GetObjectOutput, error)
	Put(key string, body io.ReadSeeker, contentType string) (*s3.PutObjectOutput, error)
	Delete(key string) (*s3.DeleteObjectOutput, error)
	Head(key string) (*s3.HeadObjectOutput, error)
	ListObjects(prefix string, delimiter string, maxKeys int64, continuationToken string) (*s3.ListObjectsV2Output, error)
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
		cfg.Endpoint = aws.String(endpoint)
		cfg.S3ForcePathStyle = aws.Bool(true) // Required for custom endpoints
	}
	
	sess := session.Must(session.NewSession(cfg))

	return &RealS3Proxy{
		bucket: bucket,
		s3:     s3.New(sess),
	}
}

func (p *RealS3Proxy) Get(key string) (*s3.GetObjectOutput, error) {
	req := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}

	return p.s3.GetObject(req)
}

func (p *RealS3Proxy) GetWithRange(key string, rangeHeader string) (*s3.GetObjectOutput, error) {
	req := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}
	
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

func (p *RealS3Proxy) Delete(key string) (*s3.DeleteObjectOutput, error) {
	req := &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}

	return p.s3.DeleteObject(req)
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

func (p *RealS3Proxy) GetWebsiteConfig() (*s3.GetBucketWebsiteOutput, error) {
	req := &s3.GetBucketWebsiteInput{
		Bucket: aws.String(p.bucket),
	}

	return p.s3.GetBucketWebsite(req)
}
