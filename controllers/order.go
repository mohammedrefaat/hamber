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

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Create a new order for products
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateOrderRequest true "Order details"
// @Success      201 {object} map[string]interface{} "Order created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /orders [post]
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

	//   Send notification
	if globalStore.NotifService != nil {
		go globalStore.NotifService.NotifyNewOrder(userID, order.ID, order.Total)
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

// GetOrders godoc
// @Summary      Get orders list
// @Description  Get paginated list of user orders
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200 {object} map[string]interface{} "Orders list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /orders [get]
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

// UpdateOrderStatus godoc
// @Summary      Update order status
// @Description  Update the status of an order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Order ID"
// @Param        request body map[string]string true "Status update"
// @Success      200 {object} map[string]interface{} "Status updated"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /orders/{id}/status [patch]
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

	// Get order to get user ID
	order, err := globalStore.StStore.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if err := globalStore.StStore.UpdateOrderStatus(uint(id), status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// NEW: Send notification
	if globalStore.NotifService != nil {
		go globalStore.NotifService.NotifyOrderStatusChange(order.UserID, order.ID, req.Status)
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

// GetOrder godoc
// @Summary      Get order by ID
// @Description  Get details of a specific order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Order ID"
// @Success      200 {object} map[string]interface{} "Order details"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Order not found"
// @Router       /orders/{id} [get]
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
