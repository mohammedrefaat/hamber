package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== INTERNAL MESSAGING SYSTEM ==========

type SendMessageRequest struct {
	ReceiverID  uint   `json:"receiver_id" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
	Body        string `json:"body" binding:"required"`
	Attachments string `json:"attachments"` // JSON array of file paths
}

type MessageResponse struct {
	ID           uint       `json:"id"`
	SenderID     uint       `json:"sender_id"`
	SenderName   string     `json:"sender_name"`
	ReceiverID   uint       `json:"receiver_id"`
	ReceiverName string     `json:"receiver_name"`
	Subject      string     `json:"subject"`
	Body         string     `json:"body"`
	IsRead       bool       `json:"is_read"`
	ReadAt       *time.Time `json:"read_at,omitempty"`
	IsStarred    bool       `json:"is_starred"`
	IsArchived   bool       `json:"is_archived"`
	Attachments  string     `json:"attachments"`
	CreatedAt    time.Time  `json:"created_at"`
}

// SendMessage godoc
// @Summary      Send internal message
// @Description  Send a message to another user (internal email)
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body SendMessageRequest true "Message details"
// @Success      201 {object} map[string]interface{} "Message sent"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /messages/send [post]
func SendMessage(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if receiver exists
	receiver, err := globalStore.StStore.GetUser(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		return
	}

	message := &dbmodels.Message{
		SenderID:    claims.UserID,
		ReceiverID:  req.ReceiverID,
		Subject:     req.Subject,
		Body:        req.Body,
		Attachments: req.Attachments,
		IsRead:      false,
		IsDraft:     false,
	}

	if err := globalStore.StStore.CreateMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Send notification to receiver
	if globalStore.NotifService != nil {
		go globalStore.NotifService.NotifyNewMessage(req.ReceiverID, receiver.Name, req.Subject)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message sent successfully",
		"data":    message,
	})
}

// GetInbox godoc
// @Summary      Get inbox messages
// @Description  Get all received messages
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Param        unread query boolean false "Show only unread messages"
// @Success      200 {object} map[string]interface{} "Messages list"
// @Router       /messages/inbox [get]
func GetInbox(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	unreadOnly := c.DefaultQuery("unread", "false") == "true"

	messages, total, err := globalStore.StStore.GetInboxMessages(claims.UserID, page, limit, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":    messages,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// GetSentMessages godoc
// @Summary      Get sent messages
// @Description  Get all messages sent by user
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} map[string]interface{} "Messages list"
// @Router       /messages/sent [get]
func GetSentMessages(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	messages, total, err := globalStore.StStore.GetSentMessages(claims.UserID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":    messages,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// GetMessage godoc
// @Summary      Get single message
// @Description  Get message details by ID
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Message ID"
// @Success      200 {object} map[string]interface{} "Message details"
// @Router       /messages/{id} [get]
func GetMessage(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	message, err := globalStore.StStore.GetMessage(uint(id), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Mark as read if receiver is viewing
	if message.ReceiverID == claims.UserID && !message.IsRead {
		globalStore.StStore.MarkMessageAsRead(uint(id))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// DeleteMessage godoc
// @Summary      Delete message
// @Description  Delete a message (soft delete)
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Message ID"
// @Success      200 {object} map[string]interface{} "Message deleted"
// @Router       /messages/{id} [delete]
func DeleteMessage(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := globalStore.StStore.DeleteMessage(uint(id), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
	})
}

// StarMessage godoc
// @Summary      Star/unstar message
// @Description  Toggle star status on a message
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Message ID"
// @Success      200 {object} map[string]interface{} "Star toggled"
// @Router       /messages/{id}/star [patch]
func StarMessage(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := globalStore.StStore.ToggleStarMessage(uint(id), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle star"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Star toggled successfully",
	})
}

// ArchiveMessage godoc
// @Summary      Archive message
// @Description  Move message to archive
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Message ID"
// @Success      200 {object} map[string]interface{} "Message archived"
// @Router       /messages/{id}/archive [patch]
func ArchiveMessage(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := globalStore.StStore.ArchiveMessage(uint(id), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message archived successfully",
	})
}

// GetMessageStats godoc
// @Summary      Get message statistics
// @Description  Get inbox statistics
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Message stats"
// @Router       /messages/stats [get]
func GetMessageStats(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stats, err := globalStore.StStore.GetMessageStats(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
