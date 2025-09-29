package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	db "github.com/mohammedrefaat/hamber/Db"
)

type CreateBlogRequest struct {
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Summary     string `json:"summary"`
	Slug        string `json:"slug" binding:"required"`
	IsPublished bool   `json:"is_published"`
}

type UpdateBlogRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Summary     string `json:"summary"`
	Slug        string `json:"slug"`
	IsPublished bool   `json:"is_published"`
}

// CreateBlog creates a new blog post
func CreateBlog(c *gin.Context) {
	var req CreateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	blog := dbmodels.Blog{
		Title:       req.Title,
		Content:     req.Content,
		Summary:     req.Summary,
		Slug:        req.Slug,
		AuthorID:    userID.(uint),
		IsPublished: req.IsPublished,
		Photos:      "[]", // Empty JSON array initially
	}

	if req.IsPublished {
		now := time.Now()
		blog.PublishedAt = &now
	}

	if err := globalStore.StStore.CreateBlog(&blog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create blog post",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"blog": blog,
	})
}

// GetBlogs returns all blog posts with pagination
func GetBlogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	publishedOnly := c.DefaultQuery("published", "false") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	blogs, total, err := globalStore.StStore.GetBlogs(page, limit, publishedOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch blogs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blogs": blogs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetBlog returns a specific blog post
func GetBlog(c *gin.Context) {
	blogID := c.Param("id")
	id, err := strconv.ParseUint(blogID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid blog ID",
		})
		return
	}

	blog, err := globalStore.StStore.GetBlog(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Blog not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blog": blog,
	})
}

// UpdateBlog updates an existing blog post
func UpdateBlog(c *gin.Context) {
	blogID := c.Param("id")
	id, err := strconv.ParseUint(blogID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid blog ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing blog
	blog, err := globalStore.StStore.GetBlog(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Blog not found",
		})
		return
	}

	// Check if user is the author or admin
	if blog.AuthorID != userID.(uint) {
		userRole, _ := c.Get("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You can only edit your own blog posts",
			})
			return
		}
	}

	var req UpdateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update fields
	if req.Title != "" {
		blog.Title = req.Title
	}
	if req.Content != "" {
		blog.Content = req.Content
	}
	if req.Summary != "" {
		blog.Summary = req.Summary
	}
	if req.Slug != "" {
		blog.Slug = req.Slug
	}

	// Handle publishing
	if req.IsPublished && !blog.IsPublished {
		blog.IsPublished = true
		now := time.Now()
		blog.PublishedAt = &now
	} else if !req.IsPublished {
		blog.IsPublished = false
		blog.PublishedAt = nil
	}

	if err := globalStore.StStore.UpdateBlog(blog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update blog post",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blog": blog,
	})
}

// DeleteBlog deletes a blog post
func DeleteBlog(c *gin.Context) {
	blogID := c.Param("id")
	id, err := strconv.ParseUint(blogID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid blog ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing blog
	blog, err := globalStore.StStore.GetBlog(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Blog not found",
		})
		return
	}

	// Check if user is the author or admin
	if blog.AuthorID != userID.(uint) {
		userRole, _ := c.Get("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You can only delete your own blog posts",
			})
			return
		}
	}

	if err := globalStore.StStore.DeleteBlog(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete blog post",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog post deleted successfully",
	})
}

// UploadBlogPhoto uploads and converts photos to WebP format
func UploadBlogPhoto(c *gin.Context) {
	blogID := c.Param("id")
	id, err := strconv.ParseUint(blogID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid blog ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing blog
	blog, err := globalStore.StStore.GetBlog(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Blog not found",
		})
		return
	}

	// Check if user is the author or admin
	if blog.AuthorID != userID.(uint) {
		userRole, _ := c.Get("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You can only upload photos to your own blog posts",
			})
			return
		}
	}

	// Parse multipart form
	err = c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	files := c.Request.MultipartForm.File["photos"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No photos uploaded",
		})
		return
	}

	var uploadedPhotos []string

	// Get existing photos
	var existingPhotos []string
	if blog.Photos != "" && blog.Photos != "[]" {
		json.Unmarshal([]byte(blog.Photos), &existingPhotos)
	}

	for _, fileHeader := range files {
		// Validate file type
		if !isValidImageType(fileHeader) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid file type for %s. Only JPG, PNG, and WebP are allowed", fileHeader.Filename),
			})
			return
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to open uploaded file",
			})
			return
		}
		defer file.Close()

		// Convert and save as WebP
		webpURL, err := convertAndSaveAsWebP(file, fileHeader, fmt.Sprintf("blog_%d", id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process image: " + err.Error(),
			})
			return
		}

		uploadedPhotos = append(uploadedPhotos, webpURL)
	}

	// Combine with existing photos
	allPhotos := append(existingPhotos, uploadedPhotos...)
	photosJSON, _ := json.Marshal(allPhotos)
	blog.Photos = string(photosJSON)

	// Update blog with new photos
	if err := globalStore.StStore.UpdateBlog(blog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update blog with new photos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Photos uploaded successfully",
		"uploaded_photos": uploadedPhotos,
		"total_photos":    len(allPhotos),
	})
}

// Helper functions
func isValidImageType(fileHeader *multipart.FileHeader) bool {
	validTypes := []string{".jpg", ".jpeg", ".png", ".webp"}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	for _, validType := range validTypes {
		if ext == validType {
			return true
		}
	}
	return false
}

func convertAndSaveAsWebP(file multipart.File, fileHeader *multipart.FileHeader, prefix string) (string, error) {
	// This is a placeholder - you'll need to implement the actual conversion
	// using your existing tools.DecodeWebP and image processing capabilities

	// For now, return a mock URL
	filename := fmt.Sprintf("%s_%d_%s.webp", prefix, time.Now().Unix(),
		strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename)))

	// In a real implementation, you would:
	// 1. Read the image using image.Decode()
	// 2. Convert it to WebP format
	// 3. Save it to your storage (MinIO, local filesystem, etc.)
	// 4. Return the URL/path where it's stored

	return fmt.Sprintf("/uploads/blogs/%s", filename), nil
}

// UploadBlogPhotoV2 - Updated version using MinIO service
func UploadBlogPhotoV2(c *gin.Context) {
	blogID := c.Param("id")
	id, err := strconv.ParseUint(blogID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid blog ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing blog
	blog, err := globalStore.StStore.GetBlog(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Blog not found",
		})
		return
	}

	// Check if user is the author or admin
	if blog.AuthorID != userID.(uint) {
		userRole, _ := c.Get("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You can only upload photos to your own blog posts",
			})
			return
		}
	}

	// Parse multipart form
	err = c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	files := c.Request.MultipartForm.File["photos"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No photos uploaded",
		})
		return
	}

	// Get photo service
	photoService := globalStore.PhotoSrv
	ctx := context.Background()

	// Upload photos to MinIO
	uploadResults, err := photoService.UploadMultiplePhotos(ctx, files, db.CategoryBlog)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload photos: " + err.Error(),
		})
		return
	}

	// Get existing photos
	var existingPhotos []string
	if blog.Photos != "" && blog.Photos != "[]" {
		json.Unmarshal([]byte(blog.Photos), &existingPhotos)
	}

	// Extract URLs from upload results
	var newPhotoURLs []string
	for _, result := range uploadResults {
		newPhotoURLs = append(newPhotoURLs, result.URL)
	}

	// Combine with existing photos
	allPhotos := append(existingPhotos, newPhotoURLs...)
	photosJSON, _ := json.Marshal(allPhotos)
	blog.Photos = string(photosJSON)

	// Update blog with new photos
	if err := globalStore.StStore.UpdateBlog(blog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update blog with new photos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Photos uploaded successfully",
		"uploaded_photos": uploadResults,
		"total_photos":    len(allPhotos),
	})
}

// UploadAvatarPhoto uploads user avatar
func UploadAvatarPhoto(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get user
	user, err := globalStore.StStore.GetUser(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get avatar file",
		})
		return
	}
	defer file.Close()

	// Get photo service

	ctx := context.Background()
	photoService := globalStore.PhotoSrv
	// Delete old avatar if exists
	if user.Avatar != "" {
		// Extract filename from URL
		oldFileName := extractFileNameFromURL(user.Avatar)
		if oldFileName != "" {
			photoService.DeletePhoto(ctx, oldFileName)
		}
	}

	// Upload new avatar
	result, err := photoService.UploadPhoto(ctx, file, header, db.CategoryAvatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload avatar: " + err.Error(),
		})
		return
	}

	// Update user avatar
	user.Avatar = result.URL
	if err := globalStore.StStore.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user avatar",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar uploaded successfully",
		"avatar":  result,
	})
}

// DeleteBlogPhoto deletes a specific photo from a blog
func DeleteBlogPhoto(c *gin.Context) {
	blogID := c.Param("id")
	id, err := strconv.ParseUint(blogID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid blog ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing blog
	blog, err := globalStore.StStore.GetBlog(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Blog not found",
		})
		return
	}

	// Check if user is the author or admin
	if blog.AuthorID != userID.(uint) {
		userRole, _ := c.Get("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You can only delete photos from your own blog posts",
			})
			return
		}
	}

	var req struct {
		PhotoURL string `json:"photo_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get existing photos
	var existingPhotos []string
	if blog.Photos != "" && blog.Photos != "[]" {
		json.Unmarshal([]byte(blog.Photos), &existingPhotos)
	}

	// Find and remove the photo URL
	newPhotos := []string{}
	photoToDelete := ""
	for _, url := range existingPhotos {
		if url == req.PhotoURL {
			photoToDelete = url
		} else {
			newPhotos = append(newPhotos, url)
		}
	}

	if photoToDelete == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Photo not found in blog",
		})
		return
	}

	// Delete from MinIO
	photoService := globalStore.PhotoSrv
	ctx := context.Background()
	fileName := extractFileNameFromURL(photoToDelete)
	if fileName != "" {
		if err := photoService.DeletePhoto(ctx, fileName); err != nil {
			// Log error but continue to update database
			fmt.Printf("Warning: Failed to delete photo from MinIO: %v\n", err)
		}
	}

	// Update blog
	photosJSON, _ := json.Marshal(newPhotos)
	blog.Photos = string(photosJSON)
	if err := globalStore.StStore.UpdateBlog(blog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update blog",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Photo deleted successfully",
		"total_photos": len(newPhotos),
	})
}

// GetPhotoStats returns statistics about photo storage
func GetPhotoStats(c *gin.Context) {
	photoService := globalStore.PhotoSrv
	ctx := context.Background()

	stats, err := photoService.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get photo stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetPhotoPresignedURL generates a presigned URL for private photo access
func GetPhotoPresignedURL(c *gin.Context) {
	fileName := c.Query("file_name")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_name query parameter is required",
		})
		return
	}

	expiryStr := c.DefaultQuery("expiry", "3600") // Default 1 hour
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid expiry value",
		})
		return
	}

	photoService := globalStore.PhotoSrv
	ctx := context.Background()

	url, err := photoService.GetPhotoURL(ctx, fileName, time.Duration(expiry)*time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate presigned URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":        url,
		"expires_in": expiry,
	})
}

// Helper function to extract filename from URL
func extractFileNameFromURL(url string) string {
	// Expected format: http://endpoint/bucket/category/filename
	parts := splitURL(url)
	if len(parts) >= 2 {
		// Return category/filename
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return ""
}

func splitURL(url string) []string {
	// Remove protocol
	url = removeProtocol(url)

	// Split by /
	parts := []string{}
	for _, part := range splitString(url, '/') {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func removeProtocol(url string) string {
	if len(url) > 7 && url[:7] == "http://" {
		return url[7:]
	}
	if len(url) > 8 && url[:8] == "https://" {
		return url[8:]
	}
	return url
}

func splitString(s string, sep rune) []string {
	var parts []string
	var current string
	for _, char := range s {
		if char == sep {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	parts = append(parts, current)
	return parts
}

// CreateBlogWithPhotos creates a blog post with photos in one request
func CreateBlogWithPhotos(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Extract blog data from form
	title := c.PostForm("title")
	content := c.PostForm("content")
	summary := c.PostForm("summary")
	slug := c.PostForm("slug")
	isPublishedStr := c.PostForm("is_published")

	// Validate required fields
	if title == "" || content == "" || slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Title, content, and slug are required",
		})
		return
	}

	// Parse is_published
	isPublished := false
	if isPublishedStr == "true" {
		isPublished = true
	}

	// Create blog
	blog := dbmodels.Blog{
		Title:       title,
		Content:     content,
		Summary:     summary,
		Slug:        slug,
		AuthorID:    userID.(uint),
		IsPublished: isPublished,
		Photos:      "[]", // Initialize as empty array
	}

	if isPublished {
		now := time.Now()
		blog.PublishedAt = &now
	}

	// Save blog to get ID
	if err := globalStore.StStore.CreateBlog(&blog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create blog post",
		})
		return
	}

	// Get uploaded photos
	files := c.Request.MultipartForm.File["photos"]

	// If photos were uploaded, process them
	if len(files) > 0 {
		photoService := globalStore.PhotoSrv
		ctx := context.Background()

		// Upload photos to MinIO
		uploadResults, err := photoService.UploadMultiplePhotos(ctx, files, db.CategoryBlog)
		if err != nil {
			// Log error but don't fail the blog creation
			fmt.Printf("Warning: Failed to upload photos: %v\n", err)
		} else {
			// Extract URLs from upload results
			var photoURLs []string
			for _, result := range uploadResults {
				photoURLs = append(photoURLs, result.URL)
			}

			// Update blog with photo URLs
			photosJSON, _ := json.Marshal(photoURLs)
			blog.Photos = string(photosJSON)

			// Update blog with photos
			if err := globalStore.StStore.UpdateBlog(&blog); err != nil {
				fmt.Printf("Warning: Failed to update blog with photos: %v\n", err)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"blog":    blog,
		"message": "Blog created successfully",
	})
}
