package controllers

import (
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

	if err := globalStore.CreateBlog(&blog); err != nil {
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

	blogs, total, err := globalStore.GetBlogs(page, limit, publishedOnly)
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

	blog, err := globalStore.GetBlog(uint(id))
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
	blog, err := globalStore.GetBlog(uint(id))
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

	if err := globalStore.UpdateBlog(blog); err != nil {
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
	blog, err := globalStore.GetBlog(uint(id))
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

	if err := globalStore.DeleteBlog(uint(id)); err != nil {
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
	blog, err := globalStore.GetBlog(uint(id))
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
	if err := globalStore.UpdateBlog(blog); err != nil {
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
