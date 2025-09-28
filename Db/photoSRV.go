package db

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type PhotoSrv struct {
	PhotoClient *minio.Client
	BucketName  string
	BaseURL     string // Your MinIO endpoint URL for generating photo links
}

type PhotoUploadResult struct {
	FileName    string    `json:"file_name"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// NewPhotoService creates a new photo service instance
func NewPhotoService(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*PhotoSrv, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	baseURL := "http://"
	if useSSL {
		baseURL = "https://"
	}
	baseURL += endpoint

	service := &PhotoSrv{
		PhotoClient: minioClient,
		BucketName:  bucketName,
		BaseURL:     baseURL,
	}

	// Create bucket if not exists
	if err := service.ensureBucketExists(context.Background()); err != nil {
		return nil, err
	}

	return service, nil
}

// ensureBucketExists creates bucket if it doesn't exist
func (p *PhotoSrv) ensureBucketExists(ctx context.Context) error {
	exists, err := p.PhotoClient.BucketExists(ctx, p.BucketName)
	if err != nil {
		return fmt.Errorf("error checking bucket existence: %v", err)
	}

	if !exists {
		err = p.PhotoClient.MakeBucket(ctx, p.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		log.Printf("Bucket %s created successfully", p.BucketName)

		// Set bucket policy to allow public read access for photos
		if err := p.setBucketPublicReadPolicy(ctx); err != nil {
			log.Printf("Warning: Failed to set bucket policy: %v", err)
		}
	}

	return nil
}

// setBucketPublicReadPolicy sets the bucket policy to allow public read access
func (p *PhotoSrv) setBucketPublicReadPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, p.BucketName)

	return p.PhotoClient.SetBucketPolicy(ctx, p.BucketName, policy)
}

// UploadPhoto uploads a photo and returns the public URL
func (p *PhotoSrv) UploadPhoto(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*PhotoUploadResult, error) {
	// Validate file type
	if !isValidImageType(header.Header.Get("Content-Type")) {
		return nil, fmt.Errorf("invalid file type: only images are allowed")
	}

	// Generate unique filename
	fileName := generateUniqueFileName(header.Filename)

	// Get file size
	fileSize := header.Size

	// Upload to MinIO
	_, err := p.PhotoClient.PutObject(ctx, p.BucketName, fileName, file, fileSize, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload photo: %v", err)
	}

	// Generate public URL
	photoURL := fmt.Sprintf("%s/%s/%s", p.BaseURL, p.BucketName, fileName)

	return &PhotoUploadResult{
		FileName:    fileName,
		URL:         photoURL,
		Size:        fileSize,
		ContentType: header.Header.Get("Content-Type"),
		UploadedAt:  time.Now(),
	}, nil
}

// GetPhotoURL returns a presigned URL for a photo (useful for private buckets)
func (p *PhotoSrv) GetPhotoURL(ctx context.Context, fileName string, expiry time.Duration) (string, error) {
	presignedURL, err := p.PhotoClient.PresignedGetObject(ctx, p.BucketName, fileName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	return presignedURL.String(), nil
}

// DeletePhoto deletes a photo from storage
func (p *PhotoSrv) DeletePhoto(ctx context.Context, fileName string) error {
	err := p.PhotoClient.RemoveObject(ctx, p.BucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete photo: %v", err)
	}
	return nil
}

// UploadMultiplePhotos uploads multiple photos concurrently
func (p *PhotoSrv) UploadMultiplePhotos(ctx context.Context, files []*multipart.FileHeader) ([]*PhotoUploadResult, error) {
	results := make([]*PhotoUploadResult, len(files))
	errChan := make(chan error, len(files))

	for i, header := range files {
		go func(index int, h *multipart.FileHeader) {
			file, err := h.Open()
			if err != nil {
				errChan <- fmt.Errorf("failed to open file %s: %v", h.Filename, err)
				return
			}
			defer file.Close()

			result, err := p.UploadPhoto(ctx, file, h)
			if err != nil {
				errChan <- err
				return
			}

			results[index] = result
			errChan <- nil
		}(i, header)
	}

	// Wait for all uploads to complete
	for i := 0; i < len(files); i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	return results, nil
}

// ListPhotos lists all photos in the bucket
func (p *PhotoSrv) ListPhotos(ctx context.Context) ([]string, error) {
	var photos []string

	for object := range p.PhotoClient.ListObjects(ctx, p.BucketName, minio.ListObjectsOptions{}) {
		if object.Err != nil {
			return nil, object.Err
		}
		photos = append(photos, object.Key)
	}

	return photos, nil
}

// Helper functions

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func generateUniqueFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)

	// Create unique filename with timestamp and UUID
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]

	return fmt.Sprintf("%s_%s_%s%s", name, timestamp, uniqueID, ext)
}
