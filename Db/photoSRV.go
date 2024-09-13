package db

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type PhotoSrv struct {
	PhotoClientt *minio.Client
}

// Initialize MinIO client
func initMinIO(endpoint, accessKey, secretKey string, useSSL bool) *minio.Client {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}
	return minioClient
}

// Create a bucket in MinIO if it doesn't exist
func (c *PhotoSrv) createBucketIfNotExists(ctx context.Context, bucketName string, opts *minio.MakeBucketOptions) {
	err := c.PhotoClientt.MakeBucket(ctx, bucketName, *opts)
	if err != nil {
		exists, errBucketExists := c.PhotoClientt.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			fmt.Printf("Bucket %s already exists\n", bucketName)
		} else {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}
}

// Upload a file to MinIO
func (c *PhotoSrv) uploadToMinIO(bucketName, objectName string, file multipart.File, fileSize int64, contentType string) error {
	_, err := c.PhotoClientt.PutObject(nil, bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to MinIO: %v", err)
	}
	return nil
}

// Download a file from MinIO
func (c *PhotoSrv) downloadFromMinIO(bucketName, objectName string) (*minio.Object, error) {
	object, err := c.PhotoClientt.GetObject(nil, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from MinIO: %v", err)
	}
	return object, nil
}
