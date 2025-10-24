package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== ADD-ON CONTROLLERS ==========

type CreateAddonRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description"`
	Logo         string   `json:"logo"`
	Photo        string   `json:"photo"`
	Category     string   `json:"category"`
	PricingType  string   `json:"pricing_type" binding:"required,oneof=time usage"`
	BasePrice    float64  `json:"base_price" binding:"required"`
	Currency     string   `json:"currency"`
	BillingCycle int      `json:"billing_cycle"`
	UsageUnit    string   `json:"usage_unit"`
	Features     []string `json:"features"`
}

func CreateAddon(c *gin.Context) {
	user, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create add-ons"})
		return
	}

	var req CreateAddonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	featuresJSON, _ := json.Marshal(req.Features)

	addon := dbmodels.Addon{
		Title:        req.Title,
		Description:  req.Description,
		Logo:         req.Logo,
		Photo:        req.Photo,
		Category:     req.Category,
		PricingType:  req.PricingType,
		BasePrice:    req.BasePrice,
		Currency:     req.Currency,
		BillingCycle: req.BillingCycle,
		UsageUnit:    req.UsageUnit,
		Features:     string(featuresJSON),
		IsActive:     true,
	}

	if err := globalStore.StStore.CreateAddon(&addon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create add-on"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"addon":   addon,
		"message": "Add-on created successfully",
	})
}

func GetAddons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	activeStr := c.Query("active")

	var isActive *bool
	if activeStr != "" {
		val := activeStr == "true"
		isActive = &val
	}

	addons, total, err := globalStore.StStore.GetAddons(page, limit, category, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch add-ons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"addons":      addons,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

func GetAddon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid add-on ID"})
		return
	}

	addon, err := globalStore.StStore.GetAddon(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Add-on not found"})
		return
	}

	// Get pricing tiers
	tiers, _ := globalStore.StStore.GetPricingTiers(addon.ID)

	c.JSON(http.StatusOK, gin.H{
		"addon": addon,
		"tiers": tiers,
	})
}

// ========== PRICING TIER CONTROLLERS ==========

type CreatePricingTierRequest struct {
	AddonID       uint    `json:"addon_id" binding:"required"`
	MinQuantity   int     `json:"min_quantity" binding:"required"`
	MaxQuantity   int     `json:"max_quantity"`
	DiscountType  string  `json:"discount_type" binding:"required,oneof=percentage fixed"`
	DiscountValue float64 `json:"discount_value" binding:"required"`
	Description   string  `json:"description"`
}

func CreatePricingTier(c *gin.Context) {
	user, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create add-ons"})
		return
	}
	var req CreatePricingTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get addon to calculate final price
	addon, err := globalStore.StStore.GetAddon(req.AddonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Add-on not found"})
		return
	}

	// Calculate final price
	finalPrice := addon.BasePrice
	if req.DiscountType == "percentage" {
		finalPrice = addon.BasePrice * (1 - req.DiscountValue/100)
	} else {
		finalPrice = addon.BasePrice - req.DiscountValue
	}

	tier := dbmodels.AddonPricingTier{
		AddonID:       req.AddonID,
		MinQuantity:   req.MinQuantity,
		MaxQuantity:   req.MaxQuantity,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		FinalPrice:    finalPrice,
		Description:   req.Description,
	}

	if err := globalStore.StStore.CreatePricingTier(&tier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pricing tier"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tier":    tier,
		"message": "Pricing tier created successfully",
	})
}
