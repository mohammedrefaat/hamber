package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== BANNER MANAGEMENT ==========

type CreateBannerRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Photo       string     `json:"photo" binding:"required"` // Base64 or URL
	Link        string     `json:"link"`
	LinkText    string     `json:"link_text"`
	Position    string     `json:"position"` // top, middle, bottom, sidebar
	Priority    int        `json:"priority"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	TargetRoles string     `json:"target_roles"` // JSON array
	TargetUsers string     `json:"target_users"` // JSON array
}

// CreateBanner godoc
// @Summary      Create banner (Admin)
// @Description  Create a new promotional banner
// @Tags         Banners
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateBannerRequest true "Banner details"
// @Success      201 {object} map[string]interface{} "Banner created"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /banners [post]
func CreateBanner(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create banners"})
		return
	}

	var req CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	banner := &dbmodels.Banner{
		Title:       req.Title,
		Description: req.Description,
		Photo:       req.Photo,
		Link:        req.Link,
		LinkText:    req.LinkText,
		Position:    req.Position,
		Priority:    req.Priority,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		TargetRoles: req.TargetRoles,
		TargetUsers: req.TargetUsers,
		IsActive:    true,
		CreatedBy:   claims.UserID,
	}

	if err := globalStore.StStore.CreateBanner(banner); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create banner"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Banner created successfully",
		"banner":  banner,
	})
}

// GetActiveBanners godoc
// @Summary      Get active banners
// @Description  Get all active banners for current user
// @Tags         Banners
// @Accept       json
// @Produce      json
// @Param        position query string false "Filter by position"
// @Success      200 {object} map[string]interface{} "Banners list"
// @Router       /banners/active [get]
func GetActiveBanners(c *gin.Context) {
	position := c.Query("position")

	// Get user ID if authenticated
	var userID *uint
	claims, err := utils.GetclamsFromContext(c)
	if err == nil {
		userID = &claims.UserID
	}

	banners, err := globalStore.StStore.GetActiveBanners(position, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch banners"})
		return
	}

	// Track banner views
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	for _, banner := range banners {
		go globalStore.StStore.TrackBannerView(banner.ID, userID, ipAddress, userAgent)
	}

	c.JSON(http.StatusOK, gin.H{
		"banners": banners,
	})
}

// GetAllBanners godoc
// @Summary      Get all banners (Admin)
// @Description  Get all banners including inactive
// @Tags         Banners
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} map[string]interface{} "Banners list"
// @Router       /admin/banners [get]
func GetAllBanners(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	banners, total, err := globalStore.StStore.GetAllBanners(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch banners"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"banners":     banners,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// UpdateBanner godoc
// @Summary      Update banner (Admin)
// @Description  Update banner details
// @Tags         Banners
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Banner ID"
// @Param        request body CreateBannerRequest true "Updated banner details"
// @Success      200 {object} map[string]interface{} "Banner updated"
// @Router       /admin/banners/{id} [put]
func UpdateBanner(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid banner ID"})
		return
	}

	banner, err := globalStore.StStore.GetBanner(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Banner not found"})
		return
	}

	var req CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	banner.Title = req.Title
	banner.Description = req.Description
	banner.Photo = req.Photo
	banner.Link = req.Link
	banner.LinkText = req.LinkText
	banner.Position = req.Position
	banner.Priority = req.Priority
	banner.StartDate = req.StartDate
	banner.EndDate = req.EndDate
	banner.TargetRoles = req.TargetRoles
	banner.TargetUsers = req.TargetUsers

	if err := globalStore.StStore.UpdateBanner(banner); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update banner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Banner updated successfully",
		"banner":  banner,
	})
}

// DeleteBanner godoc
// @Summary      Delete banner (Admin)
// @Description  Delete a banner
// @Tags         Banners
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Banner ID"
// @Success      200 {object} map[string]interface{} "Banner deleted"
// @Router       /admin/banners/{id} [delete]
func DeleteBanner(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid banner ID"})
		return
	}

	if err := globalStore.StStore.DeleteBanner(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete banner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Banner deleted successfully",
	})
}

// TrackBannerClick godoc
// @Summary      Track banner click
// @Description  Record a banner click
// @Tags         Banners
// @Accept       json
// @Produce      json
// @Param        id path int true "Banner ID"
// @Success      200 {object} map[string]interface{} "Click tracked"
// @Router       /banners/{id}/click [post]
func TrackBannerClick(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid banner ID"})
		return
	}

	var userID *uint
	claims, err := utils.GetclamsFromContext(c)
	if err == nil {
		userID = &claims.UserID
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := globalStore.StStore.TrackBannerClick(uint(id), userID, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track click"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Click tracked successfully",
	})
}

// GetBannerAnalytics godoc
// @Summary      Get banner analytics (Admin)
// @Description  Get analytics for a specific banner
// @Tags         Banners
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Banner ID"
// @Success      200 {object} map[string]interface{} "Banner analytics"
// @Router       /admin/banners/{id}/analytics [get]
func GetBannerAnalytics(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid banner ID"})
		return
	}

	analytics, err := globalStore.StStore.GetBannerAnalytics(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics": analytics,
	})
}
