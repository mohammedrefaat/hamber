package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ===================== NOTIFICATION CONTROLLERS =====================

// CreateNotificationRequest represents the body for creating a notification
type CreateNotificationRequest struct {
	UserID  uint   `json:"user_id" binding:"required" example:"123"`
	Title   string `json:"title" binding:"required" example:"Order Update"`
	Message string `json:"message" binding:"required" example:"Your order has been shipped"`
	Type    string `json:"type" example:"info"` // info, warning, error, success
	Link    string `json:"link" example:"https://hamber-hub.com/orders/123"`
}

// CreateNotification godoc
// @Summary      Create a new notification (Admin only)
// @Description  Allows an admin to create a notification for a specific user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateNotificationRequest true "Notification payload"
// @Success      201 {object} map[string]interface{} "Notification created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      500 {object} map[string]interface{} "Failed to create notification"
// @Router       /notifications [post]
func CreateNotification(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create notifications"})
		return
	}

	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification := &dbmodels.Notification{
		UserID:  req.UserID,
		Title:   req.Title,
		Message: req.Message,
		Type:    req.Type,
		Link:    req.Link,
		IsRead:  false,
	}

	if err := globalStore.StStore.CreateNotification(notification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"notification": notification,
		"message":      "Notification created successfully",
	})
}

// GetUserNotifications godoc
// @Summary      Get user notifications
// @Description  Returns a paginated list of notifications for the logged-in user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Param        unread_only query boolean false "Show only unread notifications" default(false)
// @Success      200 {object} map[string]interface{} "Notifications list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Failed to fetch notifications"
// @Router       /notifications [get]
func GetUserNotifications(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	unreadOnly := c.DefaultQuery("unread_only", "false") == "true"

	notifications, total, err := globalStore.StStore.GetUserNotifications(claims.UserID, page, limit, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"total_pages":   (int(total) + limit - 1) / limit,
	})
}

// MarkNotificationAsRead godoc
// @Summary      Mark notification as read
// @Description  Marks a specific notification as read for the logged-in user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Notification ID"
// @Success      200 {object} map[string]interface{} "Notification marked as read"
// @Failure      400 {object} map[string]interface{} "Invalid notification ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Access denied"
// @Failure      404 {object} map[string]interface{} "Notification not found"
// @Failure      500 {object} map[string]interface{} "Failed to mark as read"
// @Router       /notifications/{id}/read [patch]
func MarkNotificationAsRead(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	notification, err := globalStore.StStore.GetNotification(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if notification.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := globalStore.StStore.MarkNotificationAsRead(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// MarkAllNotificationsAsRead godoc
// @Summary      Mark all notifications as read
// @Description  Marks all notifications for the logged-in user as read
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "All notifications marked as read"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Failed to mark notifications as read"
// @Router       /notifications/read-all [patch]
func MarkAllNotificationsAsRead(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := globalStore.StStore.MarkAllNotificationsAsRead(claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification godoc
// @Summary      Delete a notification
// @Description  Deletes a specific notification belonging to the user or admin
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Notification ID"
// @Success      200 {object} map[string]interface{} "Notification deleted successfully"
// @Failure      400 {object} map[string]interface{} "Invalid notification ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Access denied"
// @Failure      404 {object} map[string]interface{} "Notification not found"
// @Failure      500 {object} map[string]interface{} "Failed to delete notification"
// @Router       /notifications/{id} [delete]
func DeleteNotification(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	notification, err := globalStore.StStore.GetNotification(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if notification.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := globalStore.StStore.DeleteNotification(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// GetUnreadCount godoc
// @Summary      Get unread notifications count
// @Description  Returns the number of unread notifications for the logged-in user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Unread count"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Failed to get unread count"
// @Router       /notifications/unread-count [get]
func GetUnreadCount(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	count, err := globalStore.StStore.GetUnreadNotificationCount(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// BroadcastNotification godoc
// @Summary      Broadcast notification (Admin only)
// @Description  Sends a notification to all users
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateNotificationRequest true "Broadcast payload"
// @Success      201 {object} map[string]interface{} "Broadcast sent successfully"
// @Failure      400 {object} map[string]interface{} "Invalid input"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      500 {object} map[string]interface{} "Failed to send broadcast"
// @Router       /notifications/broadcast [post]
func BroadcastNotification(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can broadcast notifications"})
		return
	}

	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := globalStore.StStore.GetAllUsersSimple()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	count := 0
	for _, user := range users {
		notification := &dbmodels.Notification{
			UserID:  user.ID,
			Title:   req.Title,
			Message: req.Message,
			Type:    req.Type,
			Link:    req.Link,
			IsRead:  false,
		}
		if err := globalStore.StStore.CreateNotification(notification); err == nil {
			count++
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Broadcast notification sent",
		"users_notified": count,
	})
}
