package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	config "github.com/mohammedrefaat/hamber/Config"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	models "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/stores"
)

// Global config variable - will be set from main
var globalConfig *config.Config

// SetConfig sets the global config for JWT utilities
func SetConfig(cfg *config.Config) {
	globalConfig = cfg
}

type JWTClaim struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GetJWTSecret() string {
	cfg := config.GetConfig()
	if cfg != nil {
		return cfg.GetJWTSecret() // This will decrypt automatically
	}
	return "fallback-secret-key-change-this"
}

func GetJWTExpirationHours() int {
	cfg := config.GetConfig()
	if cfg != nil {
		return cfg.GetJWTExpirationHours()
	}
	return 24
}

// Generate JWT with role information
func GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(GetJWTExpirationHours()) * time.Hour)

	// Get user role name
	roleName := "user" // default role
	if len(user.Role) > 0 {
		roleName = user.Role[0].Name
	}

	claims := &JWTClaim{
		UserID: user.ID,
		Email:  user.Email,
		Role:   roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.Itoa(int(user.ID)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetJWTSecret()))
}

func ValidateJWT(tokenString string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func GenerateRefreshToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 days

	// Get user role name
	roleName := "user" // default role
	if len(user.Role) > 0 {
		roleName = user.Role[0].Name
	}

	claims := &JWTClaim{
		UserID: user.ID,
		Email:  user.Email,
		Role:   roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.Itoa(int(user.ID)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetJWTSecret() + "_refresh"))
}

func ValidateRefreshToken(tokenString string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret() + "_refresh"), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	return claims, nil
}

// Get user permissions from JWT
func GetUserPermissions(claims *JWTClaim) []string {
	switch claims.Role {
	case "admin":
		return []string{"CREATE_USER", "DELETE_USER", "UPDATE_USER", "VIEW_ALL_USERS", "MANAGE_BLOG", "MANAGE_NEWSLETTER", "MANAGE_CONTACTS", "SYSTEM_CONFIG"}
	case "moderator":
		return []string{"UPDATE_USER", "VIEW_ALL_USERS", "MANAGE_BLOG", "VIEW_BLOG_ANALYTICS"}
	case "user":
		return []string{"UPDATE_PROFILE", "CREATE_BLOG", "VIEW_OWN_BLOG", "UPLOAD_PHOTOS"}
	default:
		return []string{}
	}
}

// GetUserFromJWT extracts user data from JWT token
func GetUserFromJWT(c *gin.Context, store *stores.DbStore) (*dbmodels.User, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("Authorization header missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return nil, errors.New("Invalid authorization format")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil // استخدم المفتاح السري الخاص بك
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaim); ok && token.Valid {
		return store.GetUserWithRole(claims.UserID)
	}

	return nil, errors.New("Invalid token")
}

// GetUserIDFromContext gets user ID from gin context
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("User ID not found in context")
	}

	if id, ok := userID.(uint); ok {
		return id, nil
	}

	return 0, errors.New("Invalid user ID type")
}

// Middleware to extract user from JWT
func JWTAuthMiddleware(store *stores.DbStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetUserFromJWT(c, store)
		if err != nil {
			c.JSON(401, gin.H{"error": "Unauthorized", "message": err.Error()})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user_name", user.Name)

		c.Next()
	}
}
