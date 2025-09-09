package utils

import (
	"errors"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	config "github.com/mohammedrefaat/hamber/Config"
	models "github.com/mohammedrefaat/hamber/DB_models"
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
