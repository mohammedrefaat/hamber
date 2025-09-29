package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	tools "github.com/mohammedrefaat/hamber/Tools"
)

// Newsletter subscription request
type NewsletterSubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Newsletter unsubscribe request
type NewsletterUnsubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Contact form request
type ContactFormRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Message string `json:"message" binding:"required"`
}

// Subscribe to newsletter
func SubscribeNewsletter(c *gin.Context) {
	var req NewsletterSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate email format
	if !tools.ValidateEmail(&req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	// Check if email already subscribed
	existing, err := globalStore.StStore.GetNewsletterByEmail(req.Email)
	if err == nil && existing != nil {
		if existing.IsActive {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email is already subscribed to newsletter",
			})
			return
		} else {
			// Reactivate subscription
			existing.IsActive = true
			existing.UnsubscribedAt = nil
			if err := globalStore.StStore.UpdateNewsletter(existing); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to reactivate newsletter subscription",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Newsletter subscription reactivated successfully",
				"data":    existing,
			})
			return
		}
	}

	// Create new subscription
	newsletter := dbmodels.Newsletter{
		Email:    req.Email,
		IsActive: true,
	}

	if err := globalStore.StStore.CreateNewsletter(&newsletter); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to subscribe to newsletter",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully subscribed to newsletter",
		"data":    newsletter,
	})
}

// Unsubscribe from newsletter
func UnsubscribeNewsletter(c *gin.Context) {
	var req NewsletterUnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	newsletter, err := globalStore.StStore.GetNewsletterByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Email not found in newsletter subscriptions",
		})
		return
	}

	if !newsletter.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email is already unsubscribed",
		})
		return
	}

	// Deactivate subscription
	if err := globalStore.StStore.UnsubscribeNewsletter(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unsubscribe from newsletter",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully unsubscribed from newsletter",
	})
}

// Submit contact form
func SubmitContactForm(c *gin.Context) {
	var req ContactFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate email format
	if !tools.ValidateEmail(&req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	// Create contact record
	contact := dbmodels.Contact{
		Name:    req.Name,
		Email:   req.Email,
		Message: req.Message,
		IsRead:  false,
		Replied: false,
	}

	if err := globalStore.StStore.CreateContact(&contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to submit contact form",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Contact form submitted successfully",
		"data":    contact,
	})
}

// Admin endpoints for managing newsletter and contacts

// Get all newsletter subscriptions (Admin only)
func GetAllNewsletterSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	active := c.DefaultQuery("active", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var isActive *bool
	if active == "true" {
		val := true
		isActive = &val
	} else if active == "false" {
		val := false
		isActive = &val
	}

	subscriptions, total, err := globalStore.StStore.GetNewsletterSubscriptions(page, limit, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch newsletter subscriptions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// Get all contact form submissions (Admin only)
func GetAllContacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	unreadOnly := c.DefaultQuery("unread", "false") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	contacts, total, err := globalStore.StStore.GetContacts(page, limit, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch contacts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contacts": contacts,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// Mark contact as read (Admin only)
func MarkContactAsRead(c *gin.Context) {
	contactID := c.Param("id")
	id, err := strconv.ParseUint(contactID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid contact ID",
		})
		return
	}

	if err := globalStore.StStore.MarkContactAsRead(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark contact as read",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contact marked as read successfully",
	})
}

// Mark contact as replied (Admin only)
func MarkContactAsReplied(c *gin.Context) {
	contactID := c.Param("id")
	id, err := strconv.ParseUint(contactID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid contact ID",
		})
		return
	}

	if err := globalStore.StStore.MarkContactAsReplied(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark contact as replied",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contact marked as replied successfully",
	})
}

// Delete contact (Admin only)
func DeleteContact(c *gin.Context) {
	contactID := c.Param("id")
	id, err := strconv.ParseUint(contactID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid contact ID",
		})
		return
	}

	if err := globalStore.StStore.DeleteContact(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete contact",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contact deleted successfully",
	})
}

// Get newsletter statistics (Admin only)
func GetNewsletterStats(c *gin.Context) {
	stats, err := globalStore.StStore.GetNewsletterStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch newsletter statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Get contact statistics (Admin only)
func GetContactStats(c *gin.Context) {
	stats, err := globalStore.StStore.GetContactStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch contact statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
