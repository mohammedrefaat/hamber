package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== SITE CONFIGURATION ==========

type SiteConfig struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	SiteName  string    `json:"site_name"`
	SiteData  string    `json:"site_data"` // JSON string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SiteConfigRequest struct {
	SiteName string                 `json:"site_name" binding:"required"`
	SiteData map[string]interface{} `json:"site_data" binding:"required"`
}

// GetSiteJSON retrieves site configuration by site name
func GetSiteJSON(c *gin.Context) {
	siteName := c.Param("site_name")
	if siteName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "site_name is required",
		})
		return
	}

	var siteConfig SiteConfig
	if err := globalStore.StStore.GetSiteConfig(siteName, &siteConfig); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Site configuration not found",
		})
		return
	}

	// Parse JSON data
	var siteData map[string]interface{}
	if err := json.Unmarshal([]byte(siteConfig.SiteData), &siteData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse site data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"site_name":  siteConfig.SiteName,
		"site_data":  siteData,
		"updated_at": siteConfig.UpdatedAt,
	})
}

// CreateOrUpdateSiteJSON creates or updates site configuration
func CreateOrUpdateSiteJSON(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req SiteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert site data to JSON string
	siteDataJSON, err := json.Marshal(req.SiteData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to serialize site data",
		})
		return
	}

	// Check if site config exists
	var existingConfig SiteConfig
	err = globalStore.StStore.GetSiteConfig(req.SiteName, &existingConfig)

	if err != nil {
		// Create new site config
		siteConfig := SiteConfig{
			UserID:   claims.UserID,
			SiteName: req.SiteName,
			SiteData: string(siteDataJSON),
		}

		if err := globalStore.StStore.CreateSiteConfig(&siteConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create site configuration",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":   "Site configuration created successfully",
			"site_name": siteConfig.SiteName,
		})
		return
	}

	// Update existing site config
	if existingConfig.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You can only update your own site configuration",
		})
		return
	}

	existingConfig.SiteData = string(siteDataJSON)
	if err := globalStore.StStore.UpdateSiteConfig(&existingConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update site configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Site configuration updated successfully",
		"site_name": existingConfig.SiteName,
	})
}

// ========== SHOPPING CART ==========

type CartItem struct {
	ID        uint              `json:"id"`
	UserID    *uint             `json:"user_id,omitempty"` // null for guest carts
	SessionID string            `json:"session_id"`
	ProductID uint              `json:"product_id"`
	Quantity  int               `json:"quantity"`
	Price     float64           `json:"price"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Product   *dbmodels.Product `json:"product,omitempty"`
}

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

// AddToCart adds a product to the shopping cart
func AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID if authenticated, otherwise use session ID
	var userID *uint
	claims, err := utils.GetclamsFromContext(c)
	if err == nil {
		userID = &claims.UserID
	}

	// Get or create session ID for guest users
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" && userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID required for guest users",
		})
		return
	}

	// Get product details
	product, err := globalStore.StStore.GetProduct(req.ProductID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	// Check stock availability
	if product.Quantity < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Insufficient stock",
			"available": product.Quantity,
		})
		return
	}

	// Check if item already exists in cart
	existingItem, err := globalStore.StStore.GetCartItem(userID, sessionID, req.ProductID)
	if err == nil && existingItem != nil {
		// Update quantity
		existingItem.Quantity += req.Quantity
		if err := globalStore.StStore.UpdateCartItem(existingItem); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update cart",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Cart updated successfully",
			"cart_item": existingItem,
		})
		return
	}

	// Add new item to cart
	cartItem := &CartItem{
		UserID:    userID,
		SessionID: sessionID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     product.Price,
	}

	if err := globalStore.StStore.AddToCart(cartItem); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add to cart",
		})
		return
	}

	// Load product details
	cartItem.Product = product

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Item added to cart successfully",
		"cart_item": cartItem,
	})
}

// GetCart retrieves the current user's cart
func GetCart(c *gin.Context) {
	var userID *uint
	claims, err := utils.GetclamsFromContext(c)
	if err == nil {
		userID = &claims.UserID
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" && userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID required for guest users",
		})
		return
	}

	cartItems, err := globalStore.StStore.GetCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cart",
		})
		return
	}

	// Calculate totals
	var subtotal float64
	for _, item := range cartItems {
		subtotal += item.Price * float64(item.Quantity)
	}

	c.JSON(http.StatusOK, gin.H{
		"cart_items": cartItems,
		"subtotal":   subtotal,
		"item_count": len(cartItems),
	})
}

// UpdateCartItem updates the quantity of a cart item
func UpdateCartItem(c *gin.Context) {
	cartItemID := c.Param("id")
	id, err := utils.ParseUint(cartItemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cart item ID",
		})
		return
	}

	var req UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uint
	claims, err := utils.GetclamsFromContext(c)
	if err == nil {
		userID = &claims.UserID
	}

	sessionID := c.GetHeader("X-Session-ID")

	// Get cart item
	cartItem, err := globalStore.StStore.GetCartItemByID(id, userID, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cart item not found",
		})
		return
	}

	// If quantity is 0, delete the item
	if req.Quantity == 0 {
		if err := globalStore.StStore.DeleteCartItem(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to remove item from cart",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Item removed from cart",
		})
		return
	}

	// Check stock availability
	product, err := globalStore.StStore.GetProduct(cartItem.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify product availability",
		})
		return
	}

	if product.Quantity < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Insufficient stock",
			"available": product.Quantity,
		})
		return
	}

	// Update quantity
	cartItem.Quantity = req.Quantity
	if err := globalStore.StStore.UpdateCartItem(cartItem); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cart updated successfully",
		"cart_item": cartItem,
	})
}

// RemoveFromCart removes an item from the cart
func RemoveFromCart(c *gin.Context) {
	cartItemID := c.Param("id")
	id, err := utils.ParseUint(cartItemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cart item ID",
		})
		return
	}

	var userID *uint
	claims, err := utils.GetclamsFromContext(c)
	if err == nil {
		userID = &claims.UserID
	}

	sessionID := c.GetHeader("X-Session-ID")

	// Verify ownership
	_, err = globalStore.StStore.GetCartItemByID(id, userID, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cart item not found",
		})
		return
	}

	if err := globalStore.StStore.DeleteCartItem(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove item from cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item removed from cart successfully",
	})
}

// ClearCart removes all items from the cart
func ClearCart(c *gin.Context) {
	var userID *uint
	claims, err := utils.GetclamsFromContext(c)
	if err == nil {
		userID = &claims.UserID
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" && userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID required for guest users",
		})
		return
	}

	if err := globalStore.StStore.ClearCart(userID, sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
	})
}

// CreateOrderFromCart creates an order from the current cart
func CreateOrderFromCart(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to create order"})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")

	// Get cart items
	cartItems, err := globalStore.StStore.GetCart(&claims.UserID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cart",
		})
		return
	}

	if len(cartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cart is empty",
		})
		return
	}

	// Get client ID from request
	var req struct {
		ClientID uint   `json:"client_id" binding:"required"`
		Address  string `json:"address"`
		Phone    string `json:"phone"`
		Notes    string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate total
	var total float64
	for _, item := range cartItems {
		total += item.Price * float64(item.Quantity)
	}

	// Create order
	order := dbmodels.Order{
		ClientID: req.ClientID,
		UserID:   claims.UserID,
		Total:    total,
		Status:   dbmodels.OrderStatus_PENDING,
		Address:  req.Address,
		Phone:    req.Phone,
		Notes:    req.Notes,
	}

	if err := globalStore.StStore.CreateOrder(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create order",
		})
		return
	}

	// Create order items
	for _, cartItem := range cartItems {
		orderItem := dbmodels.OrderItem{
			OrderID:   order.ID,
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
		}

		if err := globalStore.StStore.CreateOrderItem(&orderItem); err != nil {
			// Rollback order if order items fail
			globalStore.StStore.DeleteOrder(order.ID)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create order items",
			})
			return
		}

		// Update product quantity
		product, _ := globalStore.StStore.GetProduct(cartItem.ProductID)
		if product != nil {
			newQuantity := product.Quantity - cartItem.Quantity
			globalStore.StStore.UpdateProductQuantity(product.ID, newQuantity)
		}
	}

	// Clear cart after order creation
	globalStore.StStore.ClearCart(&claims.UserID, sessionID)

	// Send notification
	if globalStore.NotifService != nil {
		go globalStore.NotifService.NotifyNewOrder(claims.UserID, order.ID, order.Total)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
	})
}
