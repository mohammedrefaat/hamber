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

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} map[string]interface{} "List of packages"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /packages [get]
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

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user account with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration details"
// @Success      201 {object} AuthResponse "User created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      409 {object} map[string]interface{} "Email or username already exists"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /auth/register [post]
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

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Refresh token"
// @Success      200 {object} map[string]interface{} "New tokens generated"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Invalid refresh token"
// @Router       /auth/refresh [post]
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

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get current user's profile information
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} dbmodels.User "User profile"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /profile [get]
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

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update current user's profile information
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body map[string]interface{} true "Profile update data"
// @Success      200 {object} dbmodels.User "Updated profile"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /profile [put]
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

// GetUserPermissions godoc
// @Summary      Get user permissions
// @Description  Get current user's permissions from JWT context
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "User permissions"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /permissions [get]
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

// GetAllUsers godoc
// @Summary      Get all users (Admin)
// @Description  Get paginated list of all users - Admin only
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} map[string]interface{} "Users list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden - Admin only"
// @Router       /admin/users [get]
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

// DeleteUser godoc
// @Summary      Delete user (Admin)
// @Description  Delete a user account - Admin only
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "User ID"
// @Success      200 {object} map[string]interface{} "User deleted"
// @Failure      400 {object} map[string]interface{} "Invalid user ID"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /admin/users/{id} [delete]
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

// AssignRole godoc
// @Summary      Assign role to user (Admin)
// @Description  Assign a role to a user - Admin only
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "User ID"
// @Param        request body map[string]interface{} true "Role assignment (role_id)"
// @Success      200 {object} map[string]interface{} "Role assigned"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /admin/users/{id}/roles [post]
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

// RemoveRole godoc
// @Summary      Remove role from user (Admin)
// @Description  Remove a role from a user - Admin only
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "User ID"
// @Param        request body map[string]interface{} true "Role removal (role_id)"
// @Success      200 {object} map[string]interface{} "Role removed"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /admin/users/{id}/roles [delete]
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

// GetAllRoles godoc
// @Summary      Get all roles (Admin)
// @Description  Get list of all available roles - Admin only
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Roles list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /admin/roles [get]
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

// GetAllPermissions godoc
// @Summary      Get all permissions (Admin)
// @Description  Get list of all available permissions - Admin only
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Permissions list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /admin/permissions [get]
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
