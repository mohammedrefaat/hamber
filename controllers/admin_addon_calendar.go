package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== ADMIN ADDON CONTROLLERS ==========

// UpdateAddon updates an existing addon (admin only)
func UpdateAddon(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update add-ons"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid addon ID"})
		return
	}

	addon, err := globalStore.StStore.GetAddon(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Add-on not found"})
		return
	}

	var req CreateAddonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	addon.Title = req.Title
	addon.Description = req.Description
	addon.Logo = req.Logo
	addon.Photo = req.Photo
	addon.Category = req.Category
	addon.PricingType = req.PricingType
	addon.BasePrice = req.BasePrice
	if req.Currency != "" {
		addon.Currency = req.Currency
	}
	addon.BillingCycle = req.BillingCycle
	addon.UsageUnit = req.UsageUnit

	if len(req.Features) > 0 {
		featuresJSON, _ := json.Marshal(req.Features)
		addon.Features = string(featuresJSON)
	}

	if err := globalStore.StStore.UpdateAddon(addon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update add-on"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"addon":   addon,
		"message": "Add-on updated successfully",
	})
}

// DeleteAddon soft deletes an addon (admin only)
func DeleteAddon(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can delete add-ons"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid addon ID"})
		return
	}

	if err := globalStore.StStore.DeleteAddon(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete add-on"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Add-on deleted successfully",
	})
}

// GetAddonSubscriptions returns all subscriptions for an addon (admin only)
func GetAddonSubscriptions(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid addon ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	subscriptions, total, err := globalStore.StStore.GetAddonSubscriptionsByAddon(uint(id), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"total_pages":   (int(total) + limit - 1) / limit,
	})
}

// ========== ADMIN CALENDAR CONTROLLERS ==========

// GetAllEvents returns all calendar events (admin only)
func GetAllEvents(c *gin.Context) {
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	eventType := c.Query("type")
	status := c.Query("status")

	events, total, err := globalStore.StStore.GetAllCalendarEvents(page, limit, eventType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":      events,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// GetCalendarStats returns calendar statistics (admin only)
func GetCalendarStats(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	stats, err := globalStore.StStore.GetCalendarStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
