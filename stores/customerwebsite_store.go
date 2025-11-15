package stores

import (
	"net/http"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"gorm.io/gorm"
)

// ========== SITE CONFIGURATION METHODS ==========

// GetSiteConfig retrieves a site configuration by site name
func (store *DbStore) GetSiteConfig(siteName string, config *dbmodels.SiteConfig) error {
	if err := store.db.Where("site_name = ?", siteName).First(config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &CustomError{
				Message: "Site configuration not found",
				Code:    http.StatusNotFound,
			}
		}
		return &CustomError{
			Message: "Failed to fetch site configuration",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// CreateSiteConfig creates a new site configuration
func (store *DbStore) CreateSiteConfig(config *dbmodels.SiteConfig) error {
	// Check if site name already exists
	var existing dbmodels.SiteConfig
	if err := store.db.Where("site_name = ?", config.SiteName).First(&existing).Error; err == nil {
		return &CustomError{
			Message: "Site configuration with this name already exists",
			Code:    http.StatusConflict,
		}
	}

	if err := store.db.Create(config).Error; err != nil {
		return &CustomError{
			Message: "Failed to create site configuration",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// UpdateSiteConfig updates an existing site configuration
func (store *DbStore) UpdateSiteConfig(config *dbmodels.SiteConfig) error {
	if err := store.db.Save(config).Error; err != nil {
		return &CustomError{
			Message: "Failed to update site configuration",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// DeleteSiteConfig deletes a site configuration by ID
func (store *DbStore) DeleteSiteConfig(id uint) error {
	if err := store.db.Delete(&dbmodels.SiteConfig{}, id).Error; err != nil {
		return &CustomError{
			Message: "Failed to delete site configuration",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// GetAllSiteConfigs retrieves all site configurations for a user
func (store *DbStore) GetAllSiteConfigs(userID uint) ([]dbmodels.SiteConfig, error) {
	var configs []dbmodels.SiteConfig
	if err := store.db.Where("user_id = ?", userID).Find(&configs).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch site configurations",
			Code:    http.StatusInternalServerError,
		}
	}
	return configs, nil
}

// ========== SHOPPING CART METHODS ==========

// AddToCart adds a product to the shopping cart
func (store *DbStore) AddToCart(cartItem *dbmodels.CartItem) error {
	if err := store.db.Create(cartItem).Error; err != nil {
		return &CustomError{
			Message: "Failed to add item to cart",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// GetCart retrieves all cart items for a user or session
func (store *DbStore) GetCart(userID *uint, sessionID string) ([]*dbmodels.CartItem, error) {
	var cartItems []*dbmodels.CartItem
	query := store.db.Preload("Product")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if sessionID != "" {
		query = query.Where("session_id = ? AND user_id IS NULL", sessionID)
	} else {
		return nil, &CustomError{
			Message: "Either user ID or session ID must be provided",
			Code:    http.StatusBadRequest,
		}
	}

	if err := query.Find(&cartItems).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch cart items",
			Code:    http.StatusInternalServerError,
		}
	}
	return cartItems, nil
}

// GetCartItem retrieves a specific cart item by user/session and product
func (store *DbStore) GetCartItem(userID *uint, sessionID string, productID uint) (*dbmodels.CartItem, error) {
	var cartItem dbmodels.CartItem
	query := store.db

	if userID != nil {
		query = query.Where("user_id = ? AND product_id = ?", *userID, productID)
	} else if sessionID != "" {
		query = query.Where("session_id = ? AND product_id = ? AND user_id IS NULL", sessionID, productID)
	} else {
		return nil, &CustomError{
			Message: "Either user ID or session ID must be provided",
			Code:    http.StatusBadRequest,
		}
	}

	if err := query.First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, &CustomError{
			Message: "Failed to fetch cart item",
			Code:    http.StatusInternalServerError,
		}
	}
	return &cartItem, nil
}

// GetCartItemByID retrieves a cart item by its ID
func (store *DbStore) GetCartItemByID(id uint, userID *uint, sessionID string) (*dbmodels.CartItem, error) {
	var cartItem dbmodels.CartItem
	query := store.db.Where("id = ?", id)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if sessionID != "" {
		query = query.Where("session_id = ? AND user_id IS NULL", sessionID)
	}

	if err := query.First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &CustomError{
				Message: "Cart item not found",
				Code:    http.StatusNotFound,
			}
		}
		return nil, &CustomError{
			Message: "Failed to fetch cart item",
			Code:    http.StatusInternalServerError,
		}
	}
	return &cartItem, nil
}

// UpdateCartItem updates a cart item's quantity
func (store *DbStore) UpdateCartItem(cartItem *dbmodels.CartItem) error {
	if err := store.db.Save(cartItem).Error; err != nil {
		return &CustomError{
			Message: "Failed to update cart item",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// DeleteCartItem removes an item from the cart
func (store *DbStore) DeleteCartItem(id uint) error {
	if err := store.db.Delete(&dbmodels.CartItem{}, id).Error; err != nil {
		return &CustomError{
			Message: "Failed to delete cart item",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// ClearCart removes all items from a user's or session's cart
func (store *DbStore) ClearCart(userID *uint, sessionID string) error {
	query := store.db

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if sessionID != "" {
		query = query.Where("session_id = ? AND user_id IS NULL", sessionID)
	} else {
		return &CustomError{
			Message: "Either user ID or session ID must be provided",
			Code:    http.StatusBadRequest,
		}
	}

	if err := query.Delete(&dbmodels.CartItem{}).Error; err != nil {
		return &CustomError{
			Message: "Failed to clear cart",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// MigrateGuestCart migrates cart items from a guest session to a logged-in user
func (store *DbStore) MigrateGuestCart(sessionID string, userID uint) error {
	// Update all cart items with the session ID to the user ID
	if err := store.db.Model(&dbmodels.CartItem{}).
		Where("session_id = ? AND user_id IS NULL", sessionID).
		Update("user_id", userID).Error; err != nil {
		return &CustomError{
			Message: "Failed to migrate guest cart",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// GetCartItemCount returns the total number of items in the cart
func (store *DbStore) GetCartItemCount(userID *uint, sessionID string) (int64, error) {
	var count int64
	query := store.db.Model(&dbmodels.CartItem{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if sessionID != "" {
		query = query.Where("session_id = ? AND user_id IS NULL", sessionID)
	} else {
		return 0, &CustomError{
			Message: "Either user ID or session ID must be provided",
			Code:    http.StatusBadRequest,
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, &CustomError{
			Message: "Failed to count cart items",
			Code:    http.StatusInternalServerError,
		}
	}
	return count, nil
}

// GetCartTotal calculates the total price of items in the cart
func (store *DbStore) GetCartTotal(userID *uint, sessionID string) (float64, error) {
	var total float64
	cartItems, err := store.GetCart(userID, sessionID)
	if err != nil {
		return 0, err
	}

	for _, item := range cartItems {
		total += item.Price * float64(item.Quantity)
	}
	return total, nil
}

// ========== ORDER METHODS ==========

// CreateOrderItem creates a new order item
func (store *DbStore) CreateOrderItem(orderItem *dbmodels.OrderItem) error {
	if err := store.db.Create(orderItem).Error; err != nil {
		return &CustomError{
			Message: "Failed to create order item",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// DeleteOrder deletes an order (for rollback purposes)
func (store *DbStore) DeleteOrder(orderID uint) error {
	return store.db.Delete(&dbmodels.Order{}, orderID).Error
}
