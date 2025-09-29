package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// SendEmailVerification sends a verification code to the user's email
func SendEmailVerification(c *gin.Context) {
	var req SendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if email exists
	user, err := globalStore.StStore.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Email not found",
		})
		return
	}

	// Generate and send verification code
	err = globalStore.StStore.CreateEmailVerification(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send verification code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code sent to your email",
		"user_id": user.ID,
	})
}

// VerifyEmail verifies the email using the provided code
func VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Verify the code
	valid, err := globalStore.StStore.VerifyEmailCode(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify code",
		})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired verification code",
		})
		return
	}

	// Update user's email verification status
	err = globalStore.StStore.MarkEmailAsVerified(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update email verification status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// ForgotPassword sends a password reset code to the user's email
func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if email exists
	_, err := globalStore.StStore.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Email not found",
		})
		return
	}

	// Generate and send reset code
	err = globalStore.StStore.CreatePasswordReset(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send reset code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset code sent to your email",
	})
}

// ResetPassword resets the password using the provided code
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Verify the reset code
	valid, err := globalStore.StStore.VerifyPasswordResetCode(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify reset code",
		})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired reset code",
		})
		return
	}

	// Reset the password
	err = globalStore.StStore.ResetPassword(req.Email, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reset password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}
