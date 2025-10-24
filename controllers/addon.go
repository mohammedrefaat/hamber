package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

// ========== USER ADDON SUBSCRIPTION CONTROLLERS ==========

type SubscribeAddonRequest struct {
	AddonID       uint   `json:"addon_id" binding:"required"`
	PricingTierID *uint  `json:"pricing_tier_id"` // Optional, for discounted pricing
	Quantity      int    `json:"quantity" binding:"required,min=1"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=fawry paymob"`
}

// SubscribeToAddon allows users to subscribe to an add-on
func SubscribeToAddon(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req SubscribeAddonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get addon details
	addon, err := globalStore.StStore.GetAddon(req.AddonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Add-on not found"})
		return
	}

	if !addon.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Add-on is not active"})
		return
	}

	// Calculate total price
	var totalPrice float64
	var pricingTier *dbmodels.AddonPricingTier

	if req.PricingTierID != nil {
		// Get pricing tier
		tiers, _ := globalStore.StStore.GetPricingTiers(addon.ID)
		for _, tier := range tiers {
			if tier.ID == *req.PricingTierID {
				pricingTier = &tier
				break
			}
		}

		if pricingTier == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pricing tier"})
			return
		}

		// Check if quantity matches tier
		if req.Quantity < pricingTier.MinQuantity ||
			(pricingTier.MaxQuantity > 0 && req.Quantity > pricingTier.MaxQuantity) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity doesn't match pricing tier"})
			return
		}

		totalPrice = pricingTier.FinalPrice * float64(req.Quantity)
	} else {
		totalPrice = addon.BasePrice * float64(req.Quantity)
	}

	// Create payment record
	payment := &dbmodels.Payment{
		UserID:        claims.UserID,
		PackageID:     1, // Use a default package or create addon-specific payment type
		Amount:        totalPrice,
		Currency:      addon.Currency,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: dbmodels.PaymentStatus_PENDING,
	}

	if err := globalStore.StStore.CreatePayment(payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	// Create subscription
	subscription := &dbmodels.UserAddonSubscription{
		UserID:        claims.UserID,
		AddonID:       req.AddonID,
		PricingTierID: req.PricingTierID,
		Status:        dbmodels.AddonSubscriptionStatus_PENDING,
		Quantity:      req.Quantity,
		TotalPrice:    totalPrice,
		StartDate:     time.Now(),
		PaymentID:     &payment.ID,
	}

	// Set end date or usage limit based on addon type
	if addon.PricingType == "time" {
		endDate := time.Now().AddDate(0, 0, addon.BillingCycle*req.Quantity)
		subscription.EndDate = &endDate
	} else if addon.PricingType == "usage" {
		usageLimit := req.Quantity
		subscription.UsageLimit = &usageLimit
	}

	if err := globalStore.StStore.CreateAddonSubscription(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"subscription": subscription,
		"payment":      payment,
		"message":      "Subscription created. Please complete payment.",
	})
}

// GetUserSubscriptions returns all subscriptions for the current user
func GetUserSubscriptions(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	statusStr := c.Query("status")
	var status *dbmodels.AddonSubscriptionStatus
	if statusStr != "" {
		statusMap := map[string]dbmodels.AddonSubscriptionStatus{
			"PENDING":   dbmodels.AddonSubscriptionStatus_PENDING,
			"ACTIVE":    dbmodels.AddonSubscriptionStatus_ACTIVE,
			"EXPIRED":   dbmodels.AddonSubscriptionStatus_EXPIRED,
			"CANCELLED": dbmodels.AddonSubscriptionStatus_CANCELLED,
			"SUSPENDED": dbmodels.AddonSubscriptionStatus_SUSPENDED,
		}
		if s, ok := statusMap[statusStr]; ok {
			status = &s
		}
	}

	subscriptions, err := globalStore.StStore.GetUserAddonSubscriptions(claims.UserID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
	})
}

// GetSubscription returns a specific subscription
func GetSubscription(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	subscription, err := globalStore.StStore.GetAddonSubscription(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Verify ownership
	if subscription.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
	})
}

// CancelSubscription cancels a user's subscription
func CancelSubscription(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	subscription, err := globalStore.StStore.GetAddonSubscription(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Verify ownership
	if subscription.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := globalStore.StStore.CancelAddonSubscription(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription cancelled successfully",
	})
}

// LogUsage logs usage for usage-based addons
func LogUsage(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var req struct {
		UsageAmount int                    `json:"usage_amount" binding:"required,min=1"`
		Description string                 `json:"description"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := globalStore.StStore.GetAddonSubscription(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Verify ownership
	if subscription.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check if subscription is active
	if subscription.Status != dbmodels.AddonSubscriptionStatus_ACTIVE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription is not active"})
		return
	}

	// Check usage limit
	if subscription.UsageLimit != nil {
		if subscription.UsageCount+req.UsageAmount > *subscription.UsageLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Usage limit exceeded"})
			return
		}
	}

	// Create usage log
	metadataJSON, _ := json.Marshal(req.Metadata)
	usageLog := &dbmodels.AddonUsageLog{
		SubscriptionID: subscription.ID,
		UsageAmount:    req.UsageAmount,
		Description:    req.Description,
		Metadata:       string(metadataJSON),
	}

	if err := globalStore.StStore.LogAddonUsage(usageLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log usage"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"usage_log": usageLog,
		"message":   "Usage logged successfully",
	})
}

// GetUsageLogs returns usage logs for a subscription
func GetUsageLogs(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	subscription, err := globalStore.StStore.GetAddonSubscription(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Verify ownership
	if subscription.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := globalStore.StStore.GetAddonUsageLogs(uint(id), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch usage logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}
