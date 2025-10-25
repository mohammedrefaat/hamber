package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== NOTIFICATION CONTROLLERS ==========

type CreateNotificationRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Type    string `json:"type"` // info, warning, error, success
	Link    string `json:"link"`
}

// CreateNotification creates a new notification (admin only)
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
// @Description  Get paginated list of user notifications
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Param        unread_only query boolean false "Show only unread" default(false)
// @Success      200 {object} map[string]interface{} "Notifications list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
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
// @Description  Mark a specific notification as read
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Notification ID"
// @Success      200 {object} map[string]interface{} "Notification marked as read"
// @Failure      404 {object} map[string]interface{} "Notification not found"
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

	// Verify ownership
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

// MarkAllNotificationsAsRead marks all notifications as read for the current user
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

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}

// DeleteNotification deletes a notification
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

	// Verify ownership
	if notification.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := globalStore.StStore.DeleteNotification(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification deleted successfully",
	})
}

// GetUnreadCount godoc
// @Summary      Get unread notifications count
// @Description  Get count of unread notifications for current user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Unread count"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
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

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// BroadcastNotification sends a notification to all users (admin only)
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

	var req struct {
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Type    string `json:"type"`
		Link    string `json:"link"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get all users
	users, err := globalStore.StStore.GetAllUsersSimple()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	// Create notification for each user
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
