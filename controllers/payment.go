// controllers/payment.go
package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/payment"
)

// ========== PACKAGE CHANGE REQUEST ==========

type ChangePackageRequest struct {
	NewPackageID  uint   `json:"new_package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=fawry paymob"`
	Reason        string `json:"reason"`
}

type ChangePackageResponse struct {
	PackageChangeID uint       `json:"package_change_id"`
	PaymentID       uint       `json:"payment_id"`
	PaymentURL      string     `json:"payment_url,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty"`
	Message         string     `json:"message"`
	Amount          float64    `json:"amount"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

// RequestPackageChange godoc
// @Summary      Request package change
// @Description  Request to upgrade or downgrade package
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body ChangePackageRequest true "Package change request"
// @Success      200 {object} ChangePackageResponse "Change request created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /payment/change-package [post]
func RequestPackageChange(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req ChangePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get current user
	user, err := globalStore.StStore.GetUser(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Check if user is trying to change to the same package
	if user.PackageID == req.NewPackageID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You are already on this package",
		})
		return
	}

	// Get new package details
	newPackage, err := globalStore.StStore.GetPackage(req.NewPackageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Package not found",
		})
		return
	}

	// Get old package details
	//oldPackage
	_, err = globalStore.StStore.GetPackage(user.PackageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Current package not found",
		})
		return
	}

	// Calculate payment amount
	amount := newPackage.Price

	// Check if payment is required (free packages or downgrades might not require payment)
	requiresPayment := amount > 0

	// Create payment record
	expiresAt := time.Now().Add(24 * time.Hour)
	paymentdb := dbmodels.Payment{
		UserID:        user.ID,
		PackageID:     req.NewPackageID,
		Amount:        amount,
		Currency:      "EGP",
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: dbmodels.PaymentStatus_PENDING,
		ExpiresAt:     &expiresAt,
	}

	if err := globalStore.StStore.CreatePayment(&paymentdb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create payment record",
		})
		return
	}

	// Create package change record
	packageChange := dbmodels.PackageChange{
		UserID:       user.ID,
		OldPackageID: user.PackageID,
		NewPackageID: req.NewPackageID,
		PaymentID:    &paymentdb.ID,
		Status:       dbmodels.ChangeStatus_PENDING,
		ChangeReason: req.Reason,
	}

	if err := globalStore.StStore.CreatePackageChange(&packageChange); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create package change request",
		})
		return
	}

	response := ChangePackageResponse{
		PackageChangeID: packageChange.ID,
		PaymentID:       paymentdb.ID,
		Amount:          amount,
		ExpiresAt:       &expiresAt,
	}

	// If payment is required, initiate payment with chosen method
	if requiresPayment {
		switch req.PaymentMethod {
		case "fawry":
			if !globalStore.Config.IsFawryEnabled() {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Fawry payment is not enabled",
				})
				return
			}

			fawryService := payment.NewFawryService(globalStore.Config.GetFawryConfig())
			fawryResp, err := fawryService.InitiatePayment(&paymentdb, user, newPackage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to initiate Fawry payment: " + err.Error(),
				})
				return
			}

			// Update payment with reference number
			paymentdb.ReferenceNumber = fawryResp.ReferenceNumber
			globalStore.StStore.UpdatePaymentStatus(paymentdb.ID, dbmodels.PaymentStatus_PENDING, "")

			response.ReferenceNumber = fawryResp.ReferenceNumber
			response.Message = fmt.Sprintf("Please pay at any Fawry location using reference number: %s", fawryResp.ReferenceNumber)

		case "paymob":
			if !globalStore.Config.IsPaymobEnabled() {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Paymob payment is not enabled",
				})
				return
			}

			paymobService := payment.NewPaymobService(globalStore.Config.GetPaymobConfig())
			paymentURL, err := paymobService.InitiatePayment(&paymentdb, user, newPackage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to initiate Paymob payment: " + err.Error(),
				})
				return
			}

			response.PaymentURL = paymentURL
			response.Message = "Please complete payment using the provided URL"
		}
	} else {
		// No payment required, approve immediately
		globalStore.StStore.UpdatePaymentStatus(paymentdb.ID, dbmodels.PaymentStatus_PAID, "NO_PAYMENT_REQUIRED")
		globalStore.StStore.CompletePackageChange(packageChange.ID, req.NewPackageID)
		response.Message = "Package changed successfully (no payment required)"
	}

	c.JSON(http.StatusOK, response)
}

// ========== FAWRY CALLBACK ==========

type FawryCallbackRequest struct {
	RequestID        string  `json:"requestId"`
	FawryRefNumber   string  `json:"fawryRefNumber"`
	MerchantRefNum   string  `json:"merchantRefNumber"`
	OrderAmount      float64 `json:"orderAmount"`
	PaymentAmount    float64 `json:"paymentAmount"`
	OrderStatus      string  `json:"orderStatus"`
	PaymentMethod    string  `json:"paymentMethod"`
	MessageSignature string  `json:"messageSignature"`
}

func FawryCallback(c *gin.Context) {
	var req FawryCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid callback data",
		})
		return
	}

	fawryService := payment.NewFawryService(globalStore.Config.GetFawryConfig())
	if !fawryService.VerifyCallback(req.MessageSignature, req.FawryRefNumber, req.PaymentAmount, req.OrderStatus) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid signature",
		})
		return
	}

	payment, err := globalStore.StStore.GetPaymentByReference(req.FawryRefNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Payment not found",
		})
		return
	}

	var newStatus dbmodels.PaymentStatus
	switch req.OrderStatus {
	case "PAID":
		newStatus = dbmodels.PaymentStatus_PAID
		// NEW: Send success notification
		if globalStore.NotifService != nil {
			go globalStore.NotifService.NotifyPaymentSuccess(payment.UserID, payment.ID, payment.Amount)
		}
	case "FAILED":
		newStatus = dbmodels.PaymentStatus_FAILED
		// NEW: Send failure notification
		if globalStore.NotifService != nil {
			go globalStore.NotifService.NotifyPaymentFailed(payment.UserID, payment.ID, "Payment processing failed")
		}
	case "CANCELED":
		newStatus = dbmodels.PaymentStatus_CANCELLED
	case "EXPIRED":
		newStatus = dbmodels.PaymentStatus_EXPIRED
	default:
		newStatus = dbmodels.PaymentStatus_PENDING
	}

	if err := globalStore.StStore.UpdatePaymentStatus(payment.ID, newStatus, req.FawryRefNumber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update payment status",
		})
		return
	}

	if newStatus == dbmodels.PaymentStatus_PAID {
		if err := globalStore.StStore.CompletePackageChange(payment.ID, payment.PackageID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to complete package change",
			})
			return
		}

		// Get package info for notification
		pkg, _ := globalStore.StStore.GetPackage(payment.PackageID)
		if pkg != nil && globalStore.NotifService != nil {
			go globalStore.NotifService.NotifyPackageChange(payment.UserID, "old", pkg.Name)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment callback processed",
	})
}

// ========== PAYMOB CALLBACK ==========

type PaymobCallbackRequest struct {
	Type   string                 `json:"type"`
	Object map[string]interface{} `json:"obj"`
}

func PaymobCallback(c *gin.Context) {
	var req PaymobCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid callback data",
		})
		return
	}

	// Extract data from callback
	obj := req.Object
	success := obj["success"].(string)
	amountCents := fmt.Sprintf("%.0f", obj["amount_cents"].(float64))
	currency := obj["currency"].(string)
	orderId := fmt.Sprintf("%.0f", obj["order"].(map[string]interface{})["id"].(float64))
	merchantOrderId := obj["order"].(map[string]interface{})["merchant_order_id"].(string)
	hmac := obj["hmac"].(string)

	// Verify HMAC
	paymobService := payment.NewPaymobService(globalStore.Config.GetPaymobConfig())
	if !paymobService.VerifyCallback(hmac, amountCents, currency, success, orderId, merchantOrderId) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid HMAC signature",
		})
		return
	}

	// Extract transaction ID
	transactionID := fmt.Sprintf("%.0f", obj["id"].(float64))

	// Find payment by merchant order ID or transaction ID
	// Parse payment ID from merchant order ID (format: PKG-{userID}-{timestamp})
	var payment *dbmodels.Payment
	// You might need to store paymob_order_id in payment table to find it easier

	// For now, we'll try to find by transaction ID
	payment, err := globalStore.StStore.GetPaymentByTransactionID(transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Payment not found",
		})
		return
	}

	// Update payment status
	var newStatus dbmodels.PaymentStatus
	if success == "true" {
		newStatus = dbmodels.PaymentStatus_PAID
	} else {
		newStatus = dbmodels.PaymentStatus_FAILED
	}

	if err := globalStore.StStore.UpdatePaymentStatus(payment.ID, newStatus, transactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update payment status",
		})
		return
	}

	// If payment is successful, complete package change
	if newStatus == dbmodels.PaymentStatus_PAID {
		if err := globalStore.StStore.CompletePackageChange(payment.ID, payment.PackageID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to complete package change",
			})
			return
		}

		// TODO: Send confirmation email to user
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment callback processed",
	})
}

// ========== GET PAYMENT STATUS ==========

func GetPaymentStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	paymentID := c.Param("id")

	id, err := strconv.ParseUint(paymentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	payment, err := globalStore.StStore.GetPayment(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Payment not found",
		})
		return
	}

	// Verify that payment belongs to the user
	if payment.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Unauthorized access to payment",
		})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// ========== GET USER PAYMENTS ==========

func GetUserPayments(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	payments, total, err := globalStore.StStore.GetUserPayments(userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch payments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// ========== GET PACKAGE CHANGE HISTORY ==========

func GetPackageChangeHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	changes, total, err := globalStore.StStore.GetUserPackageChanges(userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch package changes",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"changes": changes,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// ========== ADMIN: GET ALL PAYMENTS ==========

func GetAllPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// This would need a new store method
	// For now, returning a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin payment list endpoint - implement store method",
		"page":    page,
		"limit":   limit,
		"status":  status,
	})
}
