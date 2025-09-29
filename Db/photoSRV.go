package db

import (
	"context"
	"fmt"
	"io"
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
	client     *minio.Client
	bucketName string
	baseURL    string
	useSSL     bool
}

type PhotoUploadResult struct {
	FileName    string    `json:"file_name"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type PhotoCategory string

const (
	CategoryBlog    PhotoCategory = "blogs"
	CategoryAvatar  PhotoCategory = "avatars"
	CategoryPackage PhotoCategory = "packages"
	CategoryGeneral PhotoCategory = "general"
)

// NewPhotoService creates and initializes a new MinIO photo service
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
		client:     minioClient,
		bucketName: bucketName,
		baseURL:    baseURL,
		useSSL:     useSSL,
	}

	// Ensure bucket exists with proper policy
	if err := service.ensureBucketExists(context.Background()); err != nil {
		return nil, err
	}

	return service, nil
}

// ensureBucketExists creates bucket if it doesn't exist and sets public read policy
func (p *PhotoSrv) ensureBucketExists(ctx context.Context) error {
	exists, err := p.client.BucketExists(ctx, p.bucketName)
	if err != nil {
		return fmt.Errorf("error checking bucket existence: %v", err)
	}

	if !exists {
		err = p.client.MakeBucket(ctx, p.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		log.Printf("✓ Bucket '%s' created successfully", p.bucketName)

		if err := p.setBucketPublicReadPolicy(ctx); err != nil {
			log.Printf("⚠ Warning: Failed to set bucket policy: %v", err)
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
	}`, p.bucketName)

	return p.client.SetBucketPolicy(ctx, p.bucketName, policy)
}

// UploadPhoto uploads a single photo to MinIO
func (p *PhotoSrv) UploadPhoto(ctx context.Context, file multipart.File, header *multipart.FileHeader, category PhotoCategory) (*PhotoUploadResult, error) {
	// Validate file
	if err := p.validateImage(header); err != nil {
		return nil, err
	}

	// Generate unique filename with category prefix
	fileName := p.generateFileName(header.Filename, category)

	// Upload to MinIO
	_, err := p.client.PutObject(ctx, p.bucketName, fileName, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload photo: %v", err)
	}

	// Generate public URL
	photoURL := fmt.Sprintf("%s/%s/%s", p.baseURL, p.bucketName, fileName)

	return &PhotoUploadResult{
		FileName:    fileName,
		URL:         photoURL,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
		UploadedAt:  time.Now(),
	}, nil
}

// UploadMultiplePhotos uploads multiple photos concurrently
func (p *PhotoSrv) UploadMultiplePhotos(ctx context.Context, files []*multipart.FileHeader, category PhotoCategory) ([]*PhotoUploadResult, error) {
	type uploadResult struct {
		index  int
		result *PhotoUploadResult
		err    error
	}

	resultChan := make(chan uploadResult, len(files))

	for i, header := range files {
		go func(index int, h *multipart.FileHeader) {
			file, err := h.Open()
			if err != nil {
				resultChan <- uploadResult{index: index, err: fmt.Errorf("failed to open file %s: %v", h.Filename, err)}
				return
			}
			defer file.Close()

			result, err := p.UploadPhoto(ctx, file, h, category)
			resultChan <- uploadResult{index: index, result: result, err: err}
		}(i, header)
	}

	results := make([]*PhotoUploadResult, len(files))
	for i := 0; i < len(files); i++ {
		res := <-resultChan
		if res.err != nil {
			return nil, res.err
		}
		results[res.index] = res.result
	}

	return results, nil
}

// UploadFromReader uploads photo from an io.Reader (useful for processed images)
func (p *PhotoSrv) UploadFromReader(ctx context.Context, reader io.Reader, fileName string, size int64, contentType string, category PhotoCategory) (*PhotoUploadResult, error) {
	fullFileName := p.generateFileName(fileName, category)

	_, err := p.client.PutObject(ctx, p.bucketName, fullFileName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload photo: %v", err)
	}

	photoURL := fmt.Sprintf("%s/%s/%s", p.baseURL, p.bucketName, fullFileName)

	return &PhotoUploadResult{
		FileName:    fullFileName,
		URL:         photoURL,
		Size:        size,
		ContentType: contentType,
		UploadedAt:  time.Now(),
	}, nil
}

// GetPhoto retrieves a photo as a byte stream
func (p *PhotoSrv) GetPhoto(ctx context.Context, fileName string) (io.ReadCloser, error) {
	object, err := p.client.GetObject(ctx, p.bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get photo: %v", err)
	}
	return object, nil
}

// GetPhotoURL returns a presigned URL for private photos
func (p *PhotoSrv) GetPhotoURL(ctx context.Context, fileName string, expiry time.Duration) (string, error) {
	presignedURL, err := p.client.PresignedGetObject(ctx, p.bucketName, fileName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	return presignedURL.String(), nil
}

// GetPublicURL returns the direct public URL (for public buckets)
func (p *PhotoSrv) GetPublicURL(fileName string) string {
	return fmt.Sprintf("%s/%s/%s", p.baseURL, p.bucketName, fileName)
}

// DeletePhoto deletes a photo from MinIO
func (p *PhotoSrv) DeletePhoto(ctx context.Context, fileName string) error {
	err := p.client.RemoveObject(ctx, p.bucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete photo: %v", err)
	}
	return nil
}

// DeleteMultiplePhotos deletes multiple photos
func (p *PhotoSrv) DeleteMultiplePhotos(ctx context.Context, fileNames []string) error {
	objectsCh := make(chan minio.ObjectInfo, len(fileNames))

	go func() {
		defer close(objectsCh)
		for _, fileName := range fileNames {
			objectsCh <- minio.ObjectInfo{Key: fileName}
		}
	}()

	errorCh := p.client.RemoveObjects(ctx, p.bucketName, objectsCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("failed to delete photo %s: %v", err.ObjectName, err.Err)
		}
	}

	return nil
}

// ListPhotos lists all photos in a category
func (p *PhotoSrv) ListPhotos(ctx context.Context, category PhotoCategory) ([]string, error) {
	var photos []string
	prefix := string(category) + "/"

	for object := range p.client.ListObjects(ctx, p.bucketName, minio.ListObjectsOptions{Prefix: prefix}) {
		if object.Err != nil {
			return nil, object.Err
		}
		photos = append(photos, object.Key)
	}

	return photos, nil
}

// PhotoExists checks if a photo exists in MinIO
func (p *PhotoSrv) PhotoExists(ctx context.Context, fileName string) (bool, error) {
	_, err := p.client.StatObject(ctx, p.bucketName, fileName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CopyPhoto copies a photo to a new location
func (p *PhotoSrv) CopyPhoto(ctx context.Context, srcFileName, destFileName string) error {
	src := minio.CopySrcOptions{
		Bucket: p.bucketName,
		Object: srcFileName,
	}

	dest := minio.CopyDestOptions{
		Bucket: p.bucketName,
		Object: destFileName,
	}

	_, err := p.client.CopyObject(ctx, dest, src)
	return err
}

// Internal helper functions

// validateImage validates if the uploaded file is a valid image
func (p *PhotoSrv) validateImage(fileHeader *multipart.FileHeader) error {
	// Check file size (10MB limit)
	const maxSize = 10 * 1024 * 1024
	if fileHeader.Size > maxSize {
		return fmt.Errorf("file size too large: %d bytes (max 10MB)", fileHeader.Size)
	}

	// Check content type
	contentType := fileHeader.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return fmt.Errorf("invalid file type: %s (only images are allowed)", contentType)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
	}

	if !validExts[ext] {
		return fmt.Errorf("invalid file extension: %s", ext)
	}

	return nil
}

// generateFileName generates a unique filename with category prefix
func (p *PhotoSrv) generateFileName(originalName string, category PhotoCategory) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(filepath.Base(originalName), ext)

	// Sanitize filename
	name = sanitizeFileName(name)

	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]

	return fmt.Sprintf("%s/%s_%s_%s%s", category, name, timestamp, uniqueID, ext)
}

// sanitizeFileName removes invalid characters from filename
func sanitizeFileName(name string) string {
	// Replace spaces and special characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, name)

	// Limit length
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

// isValidImageType checks if content type is a valid image type
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}

// GetStats returns statistics about photos in MinIO
func (p *PhotoSrv) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	categories := []PhotoCategory{CategoryBlog, CategoryAvatar, CategoryPackage, CategoryGeneral}

	for _, category := range categories {
		photos, err := p.ListPhotos(ctx, category)
		if err != nil {
			return nil, err
		}
		stats[string(category)] = len(photos)
	}

	return stats, nil
}
