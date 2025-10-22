package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/stores"
	"github.com/mohammedrefaat/hamber/utils"
)

type CreateOrderRequest struct {
	ClientID uint    `json:"client_id" binding:"required"`
	Total    float64 `json:"total" binding:"required"`
}

// CreateOrder creates a new order
func CreateOrder(c *gin.Context) {

	//todo check if client exists
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	order := dbmodels.Order{
		ClientID: req.ClientID,
		UserID:   userID,
		Total:    req.Total,
		Status:   dbmodels.OrderStatus_PENDING,
	}

	if err := globalStore.StStore.CreateOrder(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order":   order,
		"message": "Order created successfully",
	})
}

// GetOrders retrieves orders with pagination
func GetOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userID, _ := utils.GetUserIDFromContext(c)

	orders, total, err := globalStore.StStore.GetOrders(page, limit, userID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":      orders,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// UpdateOrderStatus updates order status
func UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string status to enum
	var status dbmodels.OrderStatus
	switch req.Status {
	case "PENDING":
		status = dbmodels.OrderStatus_PENDING
	case "SHIPPED":
		status = dbmodels.OrderStatus_SHIPPED
	case "DELIVERED":
		status = dbmodels.OrderStatus_DELIVERED
	case "CANCELED":
		status = dbmodels.OrderStatus_CANCELED
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	if err := globalStore.StStore.UpdateOrderStatus(uint(id), status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}

// CancelOrder cancels an order
func CancelOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := globalStore.StStore.CancelOrder(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

// GetOrder retrieves a single order by ID
func GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	order, err := globalStore.StStore.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Check if user owns this order or is admin
	if order.UserID != userID {
		userRole, _ := c.Get("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You can only view your own orders",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}
func UpdateOrderPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		PaymentStatus string     `json:"payment_status" binding:"required"`
		Amount        float64    `json:"amount" binding:"required"`
		PaymentRef    string     `json:"payment_ref"`
		PaymentDate   *time.Time `json:"payment_date"`
		PaymentMethod int64      `json:"payment_method"`
		PaymentDesc   string     `json:"payment_desc"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update order payment details
	if err := globalStore.StStore.UpdateOrderPayment(uint(id), stores.PaymentUpdate{
		PaymentStatus: req.PaymentStatus,
		Amount:        req.Amount,
		PaymentRef:    req.PaymentRef,
		PaymentDate:   req.PaymentDate,
		PaymentMethod: req.PaymentMethod,
		PaymentDesc:   req.PaymentDesc,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order payment updated successfully"})
}
