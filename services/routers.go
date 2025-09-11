package services

import (
	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	middleware "github.com/mohammedrefaat/hamber/Middleware"
	"github.com/mohammedrefaat/hamber/controllers"
)

func GetRouter(cfg *config.Config) (*gin.Engine, error) {
	router := gin.Default()

	// Initialize OAuth configuration with loaded config
	controllers.InitOAuth()

	// Add CORS and Language middleware
	router.Use(middleware.CORS())
	router.Use(middleware.LanguageMiddleware())

	// Public routes (no authentication required)
	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// Package routes (public)
		packages := api.Group("/packages")
		{
			packages.GET("/", controllers.GetAllPackages)
			packages.GET("/:id", controllers.GetPackage)
		}

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.Login)
			auth.POST("/register", controllers.Register)
			auth.POST("/refresh", controllers.RefreshToken)
			auth.POST("/forgot-password", controllers.ForgotPassword)
			auth.POST("/reset-password", controllers.ResetPassword)

			// OAuth routes
			oauth := auth.Group("/oauth")
			{
				// Google OAuth
				oauth.GET("/google", controllers.GoogleLogin)
				oauth.GET("/google/callback", controllers.GoogleCallback)

				// Facebook OAuth
				oauth.GET("/facebook", controllers.FacebookLogin)
				oauth.GET("/facebook/callback", controllers.FacebookCallback)

				// Apple OAuth
				oauth.GET("/apple", controllers.AppleLogin)
				oauth.GET("/apple/callback", controllers.AppleCallback)
			}
		}

		// Email verification routes
		verify := api.Group("/verify")
		{
			verify.POST("/send-email", controllers.SendEmailVerification)
			verify.POST("/email", controllers.VerifyEmail)
		}

		// Public blog routes
		blogs := api.Group("/blogs")
		{
			blogs.GET("/", controllers.GetBlogs) // Public blogs (published only)
			blogs.GET("/:id", controllers.GetBlog)
		}

		// Newsletter routes (public)
		newsletter := api.Group("/newsletter")
		{
			newsletter.POST("/subscribe", controllers.SubscribeNewsletter)
			newsletter.POST("/unsubscribe", controllers.UnsubscribeNewsletter)
		}

		// Contact form route (public)
		api.POST("/contact", controllers.SubmitContactForm)
	}

	// Protected routes (authentication required)
	protected := api.Group("/")
	protected.Use(middleware.JWTMiddleware())
	{
		// User profile routes
		protected.GET("/profile", controllers.GetProfile)
		protected.PUT("/profile", controllers.UpdateProfile)

		// Protected blog routes
		protectedBlogs := protected.Group("/blogs")
		{
			protectedBlogs.POST("/", controllers.CreateBlog)
			protectedBlogs.PUT("/:id", controllers.UpdateBlog)
			protectedBlogs.DELETE("/:id", controllers.DeleteBlog)
			protectedBlogs.POST("/:id/photos", controllers.UploadBlogPhoto)
		}

		// Admin only routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			// User management
			admin.GET("/users", controllers.GetAllUsers)
			admin.DELETE("/users/:id", controllers.DeleteUser)

			// Blog management
			admin.GET("/blogs", controllers.GetAllBlogsAdmin) // All blogs including unpublished
			admin.GET("/blogs/analytics", controllers.GetBlogAnalytics)

			// Newsletter management
			adminNewsletter := admin.Group("/newsletter")
			{
				adminNewsletter.GET("/subscriptions", controllers.GetAllNewsletterSubscriptions)
				adminNewsletter.GET("/stats", controllers.GetNewsletterStats)
			}

			// Contact management
			adminContact := admin.Group("/contacts")
			{
				adminContact.GET("/", controllers.GetAllContacts)
				adminContact.PUT("/:id/read", controllers.MarkContactAsRead)
				adminContact.PUT("/:id/replied", controllers.MarkContactAsReplied)
				adminContact.DELETE("/:id", controllers.DeleteContact)
				adminContact.GET("/stats", controllers.GetContactStats)
			}
		}
	}

	return router, nil
}
