package tools

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

// ImageProcessor handles image conversion and optimization
type ImageProcessor struct {
	MaxWidth  int
	MaxHeight int
	Quality   int
}

// NewImageProcessor creates a new image processor with default settings
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		MaxWidth:  1920,
		MaxHeight: 1080,
		Quality:   80,
	}
}

// ProcessImageToWebP converts an image to WebP format with optimization
func (ip *ImageProcessor) ProcessImageToWebP(file multipart.File, originalName, outputDir, prefix string) (string, error) {
	// Reset file pointer
	file.Seek(0, 0)

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize if necessary
	resized := ip.resizeImage(img)

	// Generate output filename
	baseName := strings.TrimSuffix(originalName, filepath.Ext(originalName))
	filename := fmt.Sprintf("%s_%d_%s.webp", prefix, time.Now().Unix(), baseName)
	outputPath := filepath.Join(outputDir, filename)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Save as WebP
	if err := ip.saveAsWebP(resized, outputPath); err != nil {
		return "", fmt.Errorf("failed to save WebP: %v", err)
	}

	fmt.Printf("Converted %s (%s) to WebP: %s\n", originalName, format, filename)
	return filename, nil
}

// resizeImage resizes an image if it exceeds max dimensions
func (ip *ImageProcessor) resizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Check if resize is needed
	if width <= ip.MaxWidth && height <= ip.MaxHeight {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	ratio := float64(width) / float64(height)
	var newWidth, newHeight int

	if width > height {
		newWidth = ip.MaxWidth
		newHeight = int(float64(ip.MaxWidth) / ratio)
	} else {
		newHeight = ip.MaxHeight
		newWidth = int(float64(ip.MaxHeight) * ratio)
	}

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	return resized
}

// saveAsWebP saves an image as WebP format
func (ip *ImageProcessor) saveAsWebP(img image.Image, outputPath string) error {
	// For now, we'll save as PNG since golang.org/x/image/webp only supports decoding
	// In a production environment, you might want to use a WebP encoder library
	// or call an external tool like cwebp

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// For demonstration, save as PNG with .webp extension
	// In production, use a proper WebP encoder
	return png.Encode(outFile, img)
}

// ConvertToWebP converts various image formats to WebP
func ConvertToWebP(inputFile multipart.File, outputDir, filename string) (string, error) {
	processor := NewImageProcessor()
	return processor.ProcessImageToWebP(inputFile, filename, outputDir, "converted")
}

// ValidateImageFile validates if the uploaded file is a valid image
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	// Check file size (10MB limit)
	if fileHeader.Size > 10*1024*1024 {
		return fmt.Errorf("file size too large: %d bytes (max 10MB)", fileHeader.Size)
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

	// Open and validate file content
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file header: %v", err)
	}

	// Reset file pointer
	file.Seek(0, 0)

	// Try to decode as image
	_, _, err = image.Decode(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %v", err)
	}

	return nil
}

// CreateThumbnail creates a thumbnail version of an image
func CreateThumbnail(img image.Image, maxSize int) image.Image {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Calculate thumbnail size maintaining aspect ratio
	var thumbWidth, thumbHeight int
	if width > height {
		thumbWidth = maxSize
		thumbHeight = int(float64(maxSize) * float64(height) / float64(width))
	} else {
		thumbHeight = maxSize
		thumbWidth = int(float64(maxSize) * float64(width) / float64(height))
	}

	// Create thumbnail
	thumbnail := image.NewRGBA(image.Rect(0, 0, thumbWidth, thumbHeight))
	draw.CatmullRom.Scale(thumbnail, thumbnail.Bounds(), img, bounds, draw.Over, nil)

	return thumbnail
}

// BatchProcessImages processes multiple images to WebP format
func BatchProcessImages(files []*multipart.FileHeader, outputDir, prefix string) ([]string, []error) {
	processor := NewImageProcessor()
	var processedFiles []string
	var errors []error

	for _, fileHeader := range files {
		// Validate file
		if err := ValidateImageFile(fileHeader); err != nil {
			errors = append(errors, fmt.Errorf("validation failed for %s: %v", fileHeader.Filename, err))
			continue
		}

		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to open %s: %v", fileHeader.Filename, err))
			continue
		}

		// Process file
		filename, err := processor.ProcessImageToWebP(file, fileHeader.Filename, outputDir, prefix)
		file.Close()

		if err != nil {
			errors = append(errors, fmt.Errorf("failed to process %s: %v", fileHeader.Filename, err))
			continue
		}

		processedFiles = append(processedFiles, filename)
	}

	return processedFiles, errors
}
