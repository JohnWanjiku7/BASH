package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

type S3Uploader struct {
	Client     *s3.Client
	Uploader   *manager.Uploader
	BucketName string
}

func NewS3Uploader(accessKey string, secretKey string, region string) (*S3Uploader, error) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Create static credentials using aws.Credentials
	creds := aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
			Source:          "environment",
		}, nil
	}))

	// Load AWS configuration with the static credentials and region
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create an S3 client and uploader
	client := s3.NewFromConfig(awsConfig)
	uploader := manager.NewUploader(client)

	bucketName := "bash-bucket-test-ct" // Replace with your bucket name

	return &S3Uploader{
		Client:     client,
		Uploader:   uploader,
		BucketName: bucketName,
	}, nil
}

func (uploader *S3Uploader) UploadImage(fileHeader multipart.FileHeader, ctx context.Context) (string, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Generate a unique file name
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))

	// Upload to S3
	uploadResult, err := uploader.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(uploader.BucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return the S3 file location
	return uploadResult.Location, nil
}
