package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	db "github.com/mohammedrefaat/hamber/Db"
	"github.com/mohammedrefaat/hamber/utils"
)

type CreateProductRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	Price         float64  `json:"price" binding:"required"`
	DiscountPrice float64  `json:"discount_price"`
	Quantity      int      `json:"quantity" binding:"required"`
	SKU           string   `json:"sku" binding:"required"`
	Category      string   `json:"category"`
	Brand         string   `json:"brand"`
	Photos        []string `json:"photos"` // Array of base64 encoded images
	Weight        float64  `json:"weight"`
	Tags          string   `json:"tags"`
	Fevorite      bool     `json:"fevorite"`
}

type UpdateProductRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	DiscountPrice float64  `json:"discount_price"`
	Quantity      int      `json:"quantity"`
	SKU           string   `json:"sku"`
	Category      string   `json:"category"`
	Brand         string   `json:"brand"`
	Photos        []string `json:"photos"` // Array of base64 encoded images
	Weight        float64  `json:"weight"`
	Tags          string   `json:"tags"`
	Favorite      bool     `json:"favorite"`
}

type ProductResponse struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	DiscountPrice float64  `json:"discount_price"`
	Quantity      int      `json:"quantity"`
	SKU           string   `json:"sku"`
	Category      string   `json:"category"`
	Brand         string   `json:"brand"`
	Photos        []string `json:"photos"` // Array of base64 encoded images
	Weight        float64  `json:"weight"`
	Tags          string   `json:"tags"`
	UserID        uint     `json:"user_id"`
	IsActive      bool     `json:"is_active"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	Favorite      bool     `json:"favorite"`
}

// CreateProduct creates a new product with base64 photos
func CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	photoService := globalStore.PhotoSrv

	// Upload photos to MinIO and get URLs
	var photoURLs []string
	if len(req.Photos) > 0 {
		uploadedURLs, err := uploadBase64Photos(ctx, photoService, req.Photos, db.CategoryPackage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upload photos: " + err.Error(),
			})
			return
		}
		photoURLs = uploadedURLs
	}

	// Convert photo URLs to JSON string for storage
	imagesJSON := "[]"
	if len(photoURLs) > 0 {
		imagesBytes, _ := json.Marshal(photoURLs)
		imagesJSON = string(imagesBytes)
	}

	product := dbmodels.Product{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		DiscountPrice: req.DiscountPrice,
		Quantity:      req.Quantity,
		SKU:           req.SKU,
		Category:      req.Category,
		Brand:         req.Brand,
		Images:        imagesJSON,
		Weight:        req.Weight,
		Tags:          req.Tags,
		UserID:        userID,
		IsActive:      true,
		Favorite:      req.Fevorite,
	}

	if err := globalStore.StStore.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Convert response with base64 photos
	response, err := convertProductToResponse(ctx, photoService, &product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product photos"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"product": response,
		"message": "Product created successfully",
	})
}

// GetProducts retrieves products with pagination and returns base64 photos
func GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")

	userID, _ := utils.GetUserIDFromContext(c)

	products, total, err := globalStore.StStore.GetProducts(page, limit, userID, category, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	ctx := context.Background()
	photoService := globalStore.PhotoSrv

	// Convert products to response format with base64 photos
	var productResponses []ProductResponse
	for _, product := range products {
		response, err := convertProductToResponse(ctx, photoService, &product)
		if err != nil {
			// Log error but continue with other products
			fmt.Printf("Warning: Failed to fetch photos for product %d: %v\n", product.ID, err)
			continue
		}
		productResponses = append(productResponses, *response)
	}

	c.JSON(http.StatusOK, gin.H{
		"products":    productResponses,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// GetProduct retrieves a single product with base64 photos
func GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := globalStore.StStore.GetProduct(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	ctx := context.Background()
	photoService := globalStore.PhotoSrv

	response, err := convertProductToResponse(ctx, photoService, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product photos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": response})
}

// UpdateProduct updates an existing product with base64 photos
func UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	product, err := globalStore.StStore.GetProduct(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if product.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this product"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	photoService := globalStore.PhotoSrv

	// Update basic fields
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Quantity >= 0 {
		product.Quantity = req.Quantity
	}
	if req.SKU != "" {
		product.SKU = req.SKU
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Brand != "" {
		product.Brand = req.Brand
	}
	if req.Weight >= 0 {
		product.Weight = req.Weight
	}
	if req.Tags != "" {
		product.Tags = req.Tags
	}
	if req.DiscountPrice >= 0 {
		product.DiscountPrice = req.DiscountPrice
	}
	if req.Favorite {
		product.Favorite = req.Favorite
	}
	// Handle photo updates
	if req.Photos != nil {
		// Delete old photos from MinIO
		var oldPhotoURLs []string
		if product.Images != "" && product.Images != "[]" {
			json.Unmarshal([]byte(product.Images), &oldPhotoURLs)
			for _, url := range oldPhotoURLs {
				fileName := extractFileNameFromURL(url)
				if fileName != "" {
					photoService.DeletePhoto(ctx, fileName)
				}
			}
		}

		// Upload new photos
		var photoURLs []string
		if len(req.Photos) > 0 {
			uploadedURLs, err := uploadBase64Photos(ctx, photoService, req.Photos, db.CategoryPackage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to upload photos: " + err.Error(),
				})
				return
			}
			photoURLs = uploadedURLs
		}

		// Update images JSON
		imagesJSON := "[]"
		if len(photoURLs) > 0 {
			imagesBytes, _ := json.Marshal(photoURLs)
			imagesJSON = string(imagesBytes)
		}
		product.Images = imagesJSON
	}

	if err := globalStore.StStore.UpdateProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	response, err := convertProductToResponse(ctx, photoService, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product photos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product": response,
		"message": "Product updated successfully",
	})
}

// DeleteProduct soft deletes a product
func DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	product, err := globalStore.StStore.GetProduct(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if product.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this product"})
		return
	}

	// Optional: Delete photos from MinIO when deleting product
	ctx := context.Background()
	photoService := globalStore.PhotoSrv
	var photoURLs []string
	if product.Images != "" && product.Images != "[]" {
		json.Unmarshal([]byte(product.Images), &photoURLs)
		for _, url := range photoURLs {
			fileName := extractFileNameFromURL(url)
			if fileName != "" {
				photoService.DeletePhoto(ctx, fileName)
			}
		}
	}

	if err := globalStore.StStore.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// UpdateProductQuantity updates product quantity
func UpdateProductQuantity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	product, err := globalStore.StStore.GetProduct(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if product.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this product"})
		return
	}

	if err := globalStore.StStore.UpdateProductQuantity(uint(id), req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quantity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product quantity updated successfully"})
}

// Helper Functions

// uploadBase64Photos uploads an array of base64 encoded photos to MinIO
func uploadBase64Photos(ctx context.Context, photoService *db.PhotoSrv, base64Photos []string, category db.PhotoCategory) ([]string, error) {
	var photoURLs []string

	for i, base64Photo := range base64Photos {
		// Remove data URL prefix if present (e.g., "data:image/jpeg;base64,")
		base64Data := base64Photo
		if idx := strings.Index(base64Photo, ","); idx != -1 {
			base64Data = base64Photo[idx+1:]
		}

		// Decode base64
		photoData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 photo %d: %v", i, err)
		}

		// Determine content type from base64 prefix
		contentType := "image/jpeg" // default
		if strings.Contains(base64Photo, "data:image/png") {
			contentType = "image/png"
		} else if strings.Contains(base64Photo, "data:image/webp") {
			contentType = "image/webp"
		}

		// Determine file extension
		ext := ".jpg"
		if contentType == "image/png" {
			ext = ".png"
		} else if contentType == "image/webp" {
			ext = ".webp"
		}

		// Generate filename
		fileName := fmt.Sprintf("product_%d%s", i, ext)

		// Upload to MinIO
		reader := bytes.NewReader(photoData)
		result, err := photoService.UploadFromReader(ctx, reader, fileName, int64(len(photoData)), contentType, category)
		if err != nil {
			return nil, fmt.Errorf("failed to upload photo %d: %v", i, err)
		}

		photoURLs = append(photoURLs, result.URL)
	}

	return photoURLs, nil
}

// downloadPhotoAsBase64 downloads a photo from MinIO and converts to base64
func downloadPhotoAsBase64(ctx context.Context, photoService *db.PhotoSrv, photoURL string) (string, error) {
	fileName := extractFileNameFromURL(photoURL)
	if fileName == "" {
		return "", fmt.Errorf("invalid photo URL")
	}

	// Get photo from MinIO
	reader, err := photoService.GetPhoto(ctx, fileName)
	if err != nil {
		return "", fmt.Errorf("failed to get photo: %v", err)
	}
	defer reader.Close()

	// Read photo data
	photoData, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read photo data: %v", err)
	}

	// Determine content type from file extension
	contentType := "image/jpeg"
	if strings.HasSuffix(fileName, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(fileName, ".webp") {
		contentType = "image/webp"
	}

	// Encode to base64 with data URL prefix
	base64Data := base64.StdEncoding.EncodeToString(photoData)
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Data), nil
}

// convertProductToResponse converts a product with photo URLs to base64 photos
func convertProductToResponse(ctx context.Context, photoService *db.PhotoSrv, product *dbmodels.Product) (*ProductResponse, error) {
	var photoURLs []string
	if product.Images != "" && product.Images != "[]" {
		json.Unmarshal([]byte(product.Images), &photoURLs)
	}

	// Convert photo URLs to base64
	var base64Photos []string
	for _, url := range photoURLs {
		base64Photo, err := downloadPhotoAsBase64(ctx, photoService, url)
		if err != nil {
			fmt.Printf("Warning: Failed to convert photo to base64: %v\n", err)
			continue
		}
		base64Photos = append(base64Photos, base64Photo)
	}

	return &ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		Quantity:      product.Quantity,
		SKU:           product.SKU,
		Category:      product.Category,
		Brand:         product.Brand,
		Photos:        base64Photos,
		Weight:        product.Weight,
		Tags:          product.Tags,
		UserID:        product.UserID,
		IsActive:      product.IsActive,
		CreatedAt:     product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Favorite:      product.Favorite,
		DiscountPrice: product.DiscountPrice,
	}, nil
}
