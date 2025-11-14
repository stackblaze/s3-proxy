package main

import (
	"fmt"
	"os"
	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	key := "X7SMWFBIMHK761MZDCM4"
	secret := "HeCjI9zsWe6lemh42fmCCugfyF06f7zXlyb9VY0G"
	region := "us-east-1"
	bucket := "sb-zerofs-os"
	endpoint := "https://s3.us-east-1.wasabisys.com"
	
	fmt.Println("=== Testing Wasabi S3 Credentials ===")
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Region: %s\n", region)
	fmt.Printf("Bucket: %s\n", bucket)
	fmt.Printf("Key: %s\n", key)
	fmt.Println()
	
	cfg := &aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}
	
	sess, err := session.NewSession(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to create session: %v\n", err)
		os.Exit(1)
	}
	
	svc := s3.New(sess)
	
	// Test 1: List buckets
	fmt.Println("Test 1: List buckets")
	listResult, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		fmt.Printf("❌ ListBuckets failed: %v\n", err)
	} else {
		fmt.Printf("✅ ListBuckets succeeded, found %d buckets\n", len(listResult.Buckets))
		for _, b := range listResult.Buckets {
			fmt.Printf("   - %s\n", *b.Name)
		}
	}
	fmt.Println()
	
	// Test 2: List objects in bucket
	fmt.Println("Test 2: List objects in bucket")
	listObjResult, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(5),
	})
	if err != nil {
		fmt.Printf("❌ ListObjects failed: %v\n", err)
	} else {
		fmt.Printf("✅ ListObjects succeeded, found %d objects\n", len(listObjResult.Contents))
	}
	fmt.Println()
	
	// Test 3: PUT object
	fmt.Println("Test 3: PUT test object")
	testKey := "test-credentials-check.txt"
	testData := []byte("test data from credential check")
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(testKey),
		Body:   aws.ReadSeekCloser(bytes.NewReader(testData)),
	})
	if err != nil {
		fmt.Printf("❌ PutObject failed: %v\n", err)
	} else {
		fmt.Printf("✅ PutObject succeeded\n")
	}
}
