package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	db "github.com/mohammedrefaat/hamber/Db"
	"github.com/mohammedrefaat/hamber/notification"
	"github.com/mohammedrefaat/hamber/stores"
	"github.com/mohammedrefaat/hamber/utils"
)

// Global store variable - this should be properly initialized
var globalStore *GlobalService

type GlobalService struct {
	StStore      *stores.DbStore
	Config       *config.Config
	PhotoSrv     *db.PhotoSrv
	NotifService *notification.NotificationService
}

// SetStore initializes the global store
func SetStore(Service *GlobalService) {
	globalStore = Service
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	MobileNo  string `json:"mobile_no"`
	Password  string `json:"password" binding:"required,min=6"`
	Name      string `json:"name" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
	PackageID uint   `json:"package_id"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         dbmodels.User `json:"user"`
}

// NEW: Permission response for the new permission endpoint
type PermissionResponse struct {
	UserID      uint                  `json:"user_id"`
	Email       string                `json:"email"`
	Role        string                `json:"role"`
	Permissions []dbmodels.Permission `json:"permissions"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := globalStore.StStore.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	accessToken, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	})
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set default package if not provided
	packageID := req.PackageID
	if packageID == 0 {
		packageID = 1 // Default to free plan
	}

	user := dbmodels.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		Subdomain: req.Subdomain,
		RoleID:    1, // default role ID
		PackageID: packageID,
		IS_ACTIVE: true,
	}

	if err := globalStore.StStore.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user: " + err.Error(),
		})
		return
	}

	// Get user with role for JWT generation
	userWithRole, err := globalStore.StStore.GetUserWithRole(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user details",
		})
		return
	}
	//  Send welcome notification
	if globalStore.NotifService != nil {
		go globalStore.NotifService.NotifyWelcome(user.ID, user.Name)
	}
	accessToken, err := utils.GenerateJWT(userWithRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(userWithRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *userWithRole,
	})
}

func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	claims, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}

	user, err := globalStore.StStore.GetUserWithRole(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	newAccessToken, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate new access token",
		})
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate new refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := globalStore.StStore.GetUser(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := globalStore.StStore.GetUser(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	var updateData struct {
		Name     string `json:"name"`
		Bio      string `json:"bio"`
		Website  string `json:"website"`
		Location string `json:"location"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update fields
	if updateData.Name != "" {
		user.Name = updateData.Name
	}
	if updateData.Bio != "" {
		user.Bio = updateData.Bio
	}
	if updateData.Website != "" {
		user.Website = updateData.Website
	}
	if updateData.Location != "" {
		user.Location = updateData.Location
	}

	if err := globalStore.StStore.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// NEW: Permission endpoint - Gets user permissions from JWT token context
func GetUserPermissions(c *gin.Context) {
	// Get user information from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userEmail, _ := c.Get("user_email")
	userRole, _ := c.Get("user_role")
	claims, _ := c.Get("claims")

	// Get user permissions from database
	permissions, err := globalStore.StStore.GetUserPermissions(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user permissions",
		})
		return
	}

	// Also get permissions from JWT claims (if available)
	var jwtPermissions []string
	if jwtClaims, ok := claims.(*utils.JWTClaim); ok {
		jwtPermissions = utils.GetUserPermissions(jwtClaims)
	}

	response := PermissionResponse{
		UserID:      userID.(uint),
		Email:       userEmail.(string),
		Role:        userRole.(string),
		Permissions: permissions,
	}

	c.JSON(http.StatusOK, gin.H{
		"user_permissions": response,
		"jwt_permissions":  jwtPermissions,
		"message":          "Permissions retrieved successfully",
	})
}

// FIXED: GetAllUsers - Now properly implemented
func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := globalStore.StStore.GetAllUsers(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	if err := globalStore.StStore.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// NEW: Role management endpoints

// AssignRole assigns a role to a user (Admin only)
func AssignRole(c *gin.Context) {
	userID := c.Param("id")
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req struct {
		RoleID uint `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := globalStore.StStore.AssignRoleToUser(uint(id), req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role to user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role assigned successfully",
	})
}

// RemoveRole removes a role from a user (Admin only)
func RemoveRole(c *gin.Context) {
	userID := c.Param("id")
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req struct {
		RoleID uint `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := globalStore.StStore.RemoveRoleFromUser(uint(id), req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role from user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role removed successfully",
	})
}

// GetAllRoles returns all available roles (Admin only)
func GetAllRoles(c *gin.Context) {
	roles, err := globalStore.StStore.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

// GetAllPermissions returns all available permissions (Admin only)
func GetAllPermissions(c *gin.Context) {
	permissions, err := globalStore.StStore.GetAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
	})
}
