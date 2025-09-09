package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
	"golang.org/x/oauth2"
)

var oauthConfig *config.OAuthConfig

// Initialize OAuth configuration
func InitOAuth() {
	oauthConfig = config.InitOAuthConfig()
}

// GoogleUserInfo represents user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// FacebookUserInfo represents user info from Facebook
type FacebookUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// AppleUserInfo represents user info from Apple
type AppleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
}

// Generate state for OAuth security
func generateStateOAuthCookie() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// Google OAuth Login
func GoogleLogin(c *gin.Context) {
	state := generateStateOAuthCookie()

	// Store state in session or database for validation
	c.SetCookie("oauthstate", state, 3600, "/", "", false, true)

	url := oauthConfig.Google.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Google OAuth Callback
func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	// Validate state
	oauthState, err := c.Cookie("oauthstate")
	if err != nil || state != oauthState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid OAuth state",
		})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauthstate", "", -1, "/", "", false, true)

	// Exchange code for token
	token, err := oauthConfig.Google.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to exchange token: " + err.Error(),
		})
		return
	}

	// Get user info from Google
	client := oauthConfig.Google.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get user info",
		})
		return
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse user info",
		})
		return
	}

	// Handle user authentication/registration
	authResponse, err := handleOAuthUser("google", userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture, token.AccessToken, token.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// Facebook OAuth Login
func FacebookLogin(c *gin.Context) {
	state := generateStateOAuthCookie()
	c.SetCookie("oauthstate", state, 3600, "/", "", false, true)

	url := oauthConfig.Facebook.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Facebook OAuth Callback
func FacebookCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	oauthState, err := c.Cookie("oauthstate")
	if err != nil || state != oauthState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid OAuth state",
		})
		return
	}

	c.SetCookie("oauthstate", "", -1, "/", "", false, true)

	token, err := oauthConfig.Facebook.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to exchange token: " + err.Error(),
		})
		return
	}

	client := oauthConfig.Facebook.Client(context.Background(), token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get user info",
		})
		return
	}
	defer resp.Body.Close()

	var userInfo FacebookUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse user info",
		})
		return
	}

	picture := ""
	if userInfo.Picture.Data.URL != "" {
		picture = userInfo.Picture.Data.URL
	}

	authResponse, err := handleOAuthUser("facebook", userInfo.ID, userInfo.Email, userInfo.Name, picture, token.AccessToken, token.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

/*
// Apple OAuth Login
func AppleLogin(c *gin.Context) {
	state := generateStateOAuthCookie()
	c.SetCookie("oauthstate", state, 3600, "/", "", false, true)

	url := oauthConfig.Apple.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Apple OAuth Callback
func AppleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	oauthState, err := c.Cookie("oauthstate")
	if err != nil || state != oauthState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid OAuth state",
		})
		return
	}

	c.SetCookie("oauthstate", "", -1, "/", "", false, true)

	token, err := oauthConfig.Apple.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to exchange token: " + err.Error(),
		})
		return
	}

	// Apple returns user info in ID token (JWT)
	// You'll need to decode the JWT to get user info
	// This is a simplified version - in production, properly validate the JWT
	userInfo := AppleUserInfo{
		Sub:   "apple_user_id",    // Extract from JWT
		Email: "user@example.com", // Extract from JWT
	}

	fullName := fmt.Sprintf("%s %s", userInfo.Name.FirstName, userInfo.Name.LastName)

	authResponse, err := handleOAuthUser("apple", userInfo.Sub, userInfo.Email, fullName, "", token.AccessToken, token.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}
*/
// Handle OAuth user authentication/registration
func handleOAuthUser(provider, providerID, email, name, picture, accessToken, refreshToken string) (*AuthResponse, error) {
	// Check if OAuth profile exists
	oauthProfile, err := globalStore.GetOAuthProfile(provider, providerID)
	if err == nil && oauthProfile != nil {
		// User exists, update tokens and login
		oauthProfile.AccessToken = accessToken
		oauthProfile.RefreshToken = refreshToken
		globalStore.UpdateOAuthProfile(oauthProfile)

		user, err := globalStore.GetUser(oauthProfile.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %v", err)
		}

		// Generate JWT tokens
		jwtToken, err := utils.GenerateJWT(user)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT: %v", err)
		}

		jwtRefreshToken, err := utils.GenerateRefreshToken(user)
		if err != nil {
			return nil, fmt.Errorf("failed to generate refresh token: %v", err)
		}

		return &AuthResponse{
			AccessToken:  jwtToken,
			RefreshToken: jwtRefreshToken,
			User:         *user,
		}, nil
	}

	// Check if user exists by email
	existingUser, err := globalStore.GetUserByEmail(email)
	if err != nil {
		// User doesn't exist, create new user
		newUser := dbmodels.User{
			Name:              name,
			Email:             email,
			Password:          "", // OAuth users don't need password
			Subdomain:         generateSubdomain(name),
			RoleID:            1, // default role
			PackageID:         1, // default package
			IS_ACTIVE:         true,
			IS_EMAIL_VERIFIED: true, // OAuth emails are considered verified
			Avatar:            picture,
		}

		if err := globalStore.CreateUser(&newUser); err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}

		// Create OAuth profile
		profile := dbmodels.OAuthProfile{
			UserID:       newUser.ID,
			Provider:     provider,
			ProviderID:   providerID,
			Email:        email,
			Name:         name,
			Picture:      picture,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		if err := globalStore.CreateOAuthProfile(&profile); err != nil {
			return nil, fmt.Errorf("failed to create OAuth profile: %v", err)
		}

		// Generate JWT tokens
		jwtToken, err := utils.GenerateJWT(&newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT: %v", err)
		}

		jwtRefreshToken, err := utils.GenerateRefreshToken(&newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to generate refresh token: %v", err)
		}

		return &AuthResponse{
			AccessToken:  jwtToken,
			RefreshToken: jwtRefreshToken,
			User:         newUser,
		}, nil
	}

	// User exists, link OAuth profile
	profile := dbmodels.OAuthProfile{
		UserID:       existingUser.ID,
		Provider:     provider,
		ProviderID:   providerID,
		Email:        email,
		Name:         name,
		Picture:      picture,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	if err := globalStore.CreateOAuthProfile(&profile); err != nil {
		return nil, fmt.Errorf("failed to create OAuth profile: %v", err)
	}

	// Update user avatar if not set
	if existingUser.Avatar == "" && picture != "" {
		existingUser.Avatar = picture
		globalStore.UpdateUser(existingUser)
	}

	// Generate JWT tokens
	jwtToken, err := utils.GenerateJWT(existingUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %v", err)
	}

	jwtRefreshToken, err := utils.GenerateRefreshToken(existingUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return &AuthResponse{
		AccessToken:  jwtToken,
		RefreshToken: jwtRefreshToken,
		User:         *existingUser,
	}, nil
}

// Generate a unique subdomain from name
func generateSubdomain(name string) string {
	// Simple implementation - in production, make it more robust
	subdomain := fmt.Sprintf("%s_%d", name, time.Now().Unix())
	return subdomain
}
