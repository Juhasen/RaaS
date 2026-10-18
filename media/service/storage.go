package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// newS3Client returns a path-style client for the configured S3-compatible
// endpoint (MinIO), or nil when endpoint or credentials are missing.
func newS3Client(cfg *Config) *s3.Client {
	if cfg.S3Endpoint == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" {
		return nil
	}
	region := cfg.S3Region
	if region == "" {
		region = "us-east-1"
	}
	return s3.New(s3.Options{
		Region:       region,
		BaseEndpoint: aws.String(cfg.S3Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		UsePathStyle: true,
	})
}

// publicURL returns the base URL stored in media records.
func publicURL(cfg *Config) string {
	if cfg.S3PublicURL != "" {
		return strings.TrimSuffix(cfg.S3PublicURL, "/")
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(cfg.S3Endpoint, "/"), cfg.S3Bucket)
}

// ensureBucket creates the bucket if missing and makes objects publicly readable
// so browsers can load the stored URLs.
func ensureBucket(ctx context.Context, c *s3.Client, bucket string) error {
	if _, err := c.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
			return fmt.Errorf("create bucket %s: %w", bucket, err)
		}
	}
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucket)
	if _, err := c.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{Bucket: aws.String(bucket), Policy: aws.String(policy)}); err != nil {
		return fmt.Errorf("set policy on bucket %s: %w", bucket, err)
	}
	return nil
}

// bucketReady retries ensureBucket so media can start before MinIO is up.
func bucketReady(c *s3.Client, bucket string) {
	var err error
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = ensureBucket(ctx, c, bucket)
		cancel()
		if err == nil {
			log.Printf("S3 bucket %s ready", bucket)
			return
		}
		time.Sleep(3 * time.Second)
	}
	log.Printf("S3 bucket %s not ready, uploads will fail: %v", bucket, err)
}
