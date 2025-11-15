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

type SiteConfigRequest struct {
	SiteName string                 `json:"site_name" binding:"required" example:"my-store"`
	SiteData map[string]interface{} `json:"site_data" binding:"required" swaggertype:"object,string" example:"{\\"theme\\":\\"dark\\",\\"logo\\":\\"https://example.com/logo.png\\"}"`
}

type SiteConfigResponse struct {
	SiteName  string                 `json:"site_name" example:"my-store"`
	SiteData  map[string]interface{} `json:"site_data" swaggertype:"object,string" example:"{\\"theme\\":\\"dark\\",\\"logo\\":\\"https://example.com/logo.png\\"}"`
	UpdatedAt time.Time              `json:"updated_at" example:"2025-11-15T17:30:00Z"`
}

// GetSiteJSON retrieves site configuration by site name
// @Summary Get site configuration
// @Description Retrieves the configuration data for a specific customer website by site name
// @Tags Customer Website
// @Accept json
// @Produce json
// @Param site_name path string true "Site Name" example("my-store")
// @Success 200 {object} SiteConfigResponse "Site configuration retrieved successfully"
// @Failure 400 {object} map[string]string "Bad request - site_name is required"
// @Failure 404 {object} map[string]string "Site configuration not found"
// @Failure 500 {object} map[string]string "Failed to parse site data"
// @Router /api/customer-website/site/{site_name} [get]
func GetSiteJSON(c *gin.Context) {
	siteName := c.Param("site_name")
	if siteName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "site_name is required",
		})
		return
	}

	var siteConfig dbmodels.SiteConfig
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
// @Summary Create or update site configuration
// @Description Creates a new site configuration or updates an existing one. Users can only update their own sites unless they are admin.
// @Tags Customer Website
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer token" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param request body SiteConfigRequest true "Site Configuration Data"
// @Success 200 {object} map[string]string "Site configuration updated successfully"
// @Success 201 {object} map[string]string "Site configuration created successfully"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - cannot update other user's site"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/admin/sites [post]
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
	var existingConfig dbmodels.SiteConfig
	err = globalStore.StStore.GetSiteConfig(req.SiteName, &existingConfig)

	if err != nil {
		// Create new site config
		siteConfig := dbmodels.SiteConfig{
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

type CartItemResponse struct {
	ID        uint              `json:"id" example:"1"`
	UserID    *uint             `json:"user_id,omitempty" example:"5"`
	SessionID string            `json:"session_id" example:"guest-session-123"`
	ProductID uint              `json:"product_id" example:"10"`
	Quantity  int               `json:"quantity" example:"2"`
	Price     float64           `json:"price" example:"29.99"`
	CreatedAt time.Time         `json:"created_at" example:"2025-11-15T17:30:00Z"`
	UpdatedAt time.Time         `json:"updated_at" example:"2025-11-15T17:30:00Z"`
	Product   *dbmodels.Product `json:"product,omitempty"`
}

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required" example:"10"`
	Quantity  int  `json:"quantity" binding:"required,min=1" example:"2"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0" example:"3"`
}

type CartResponse struct {
	CartItems []CartItemResponse `json:"cart_items"`
	Subtotal  float64            `json:"subtotal" example:"89.97"`
	ItemCount int                `json:"item_count" example:"3"`
}

type CheckoutRequest struct {
	ClientID uint   `json:"client_id" binding:"required" example:"5"`
	Address  string `json:"address" example:"123 Main St, City, Country"`
	Phone    string `json:"phone" example:"+201234567890"`
	Notes    string `json:"notes" example:"Please deliver between 2-5 PM"`
}

type OrderResponse struct {
	Message string         `json:"message" example:"Order created successfully"`
	Order   dbmodels.Order `json:"order"`
}

// AddToCart adds a product to the shopping cart
// @Summary Add item to cart
// @Description Adds a product to the shopping cart. For authenticated users, uses user ID. For guests, requires X-Session-ID header.
// @Tags Shopping Cart
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer token (optional for guests)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param X-Session-ID header string false "Session ID for guest users" example("guest-session-abc123")
// @Param request body AddToCartRequest true "Product and Quantity"
// @Success 200 {object} map[string]interface{} "Cart updated successfully (item already existed)"
// @Success 201 {object} map[string]interface{} "Item added to cart successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - validation error or insufficient stock"
// @Failure 404 {object} map[string]string "Product not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/customer-website/cart/add [post]
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
	cartItem := &dbmodels.CartItem{
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
// @Summary Get shopping cart
// @Description Retrieves all items in the current user's shopping cart with calculated subtotal
// @Tags Shopping Cart
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer token (optional for guests)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param X-Session-ID header string false "Session ID for guest users" example("guest-session-abc123")
// @Success 200 {object} CartResponse "Cart retrieved successfully"
// @Failure 400 {object} map[string]string "Bad request - Session ID required for guests"
// @Failure 500 {object} map[string]string "Failed to fetch cart"
// @Router /api/customer-website/cart [get]
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
// @Summary Update cart item quantity
// @Description Updates the quantity of a specific cart item. Setting quantity to 0 removes the item.
// @Tags Shopping Cart
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer token (optional for guests)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param X-Session-ID header string false "Session ID for guest users" example("guest-session-abc123")
// @Param id path int true "Cart Item ID" example(1)
// @Param request body UpdateCartRequest true "New Quantity"
// @Success 200 {object} map[string]interface{} "Cart updated successfully or item removed"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid ID or insufficient stock"
// @Failure 404 {object} map[string]string "Cart item not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/customer-website/cart/{id} [put]
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
// @Summary Remove item from cart
// @Description Completely removes a specific item from the shopping cart
// @Tags Shopping Cart
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer token (optional for guests)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param X-Session-ID header string false "Session ID for guest users" example("guest-session-abc123")
// @Param id path int true "Cart Item ID" example(1)
// @Success 200 {object} map[string]string "Item removed from cart successfully"
// @Failure 400 {object} map[string]string "Invalid cart item ID"
// @Failure 404 {object} map[string]string "Cart item not found"
// @Failure 500 {object} map[string]string "Failed to remove item from cart"
// @Router /api/customer-website/cart/{id} [delete]
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
// @Summary Clear entire cart
// @Description Removes all items from the shopping cart
// @Tags Shopping Cart
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer token (optional for guests)" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param X-Session-ID header string false "Session ID for guest users" example("guest-session-abc123")
// @Success 200 {object} map[string]string "Cart cleared successfully"
// @Failure 400 {object} map[string]string "Session ID required for guest users"
// @Failure 500 {object} map[string]string "Failed to clear cart"
// @Router /api/customer-website/cart/clear [delete]
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
// @Summary Checkout - Create order from cart
// @Description Creates an order from all items in the shopping cart. Requires authentication. Automatically updates inventory and clears cart.
// @Tags Shopping Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer token" example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Param X-Session-ID header string false "Session ID (optional, used to migrate guest cart)" example("guest-session-abc123")
// @Param request body CheckoutRequest true "Order Details"
// @Success 201 {object} OrderResponse "Order created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - cart is empty or validation error"
// @Failure 401 {object} map[string]string "Authentication required to create order"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/customer-website/checkout [post]
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
	var req CheckoutRequest

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
