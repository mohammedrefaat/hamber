package services

import (
	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	middleware "github.com/mohammedrefaat/hamber/Middleware"
	"github.com/mohammedrefaat/hamber/controllers"
)

func GetRouter(cfg *config.Config) (*gin.Engine, error) {
	router := gin.Default()

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

		// Photo routes
		photos := protected.Group("/photos")
		{
			// Avatar upload
			photos.POST("/avatar", controllers.UploadAvatarPhoto)

			// Get presigned URL for private photos
			photos.GET("/presigned-url", controllers.GetPhotoPresignedURL)
		}

		// Protected blog routes
		protectedBlogs := protected.Group("/blogs")
		{
			protectedBlogs.POST("/", controllers.CreateBlogWithPhotos)
			protectedBlogs.PUT("/:id", controllers.UpdateBlog)
			protectedBlogs.DELETE("/:id", controllers.DeleteBlog)

			// Photo management for blogs
			protectedBlogs.POST("/:id/photos", controllers.UploadBlogPhotoV2)
			protectedBlogs.DELETE("/:id/photos", controllers.DeleteBlogPhoto)
		}

		// User permissions
		protected.GET("/permissions", controllers.GetUserPermissions)

		// Admin only routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			// User management
			admin.GET("/users", controllers.GetAllUsers)
			admin.DELETE("/users/:id", controllers.DeleteUser)

			// Role management
			admin.GET("/roles", controllers.GetAllRoles)
			admin.GET("/permissions", controllers.GetAllPermissions)
			admin.POST("/users/:id/roles", controllers.AssignRole)
			admin.DELETE("/users/:id/roles", controllers.RemoveRole)

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

			// Photo statistics
			admin.GET("/photos/stats", controllers.GetPhotoStats)
		}
	}
	/*// User routes (protected)
	user := protected.Group("/user")
	{
		user.GET("/me", controllers.GetCurrentUser)
		user.GET("/profile", controllers.GetUserProfile)
	}*/

	// Product routes (protected)
	products := protected.Group("/products")
	{
		products.POST("/", controllers.CreateProduct)
		products.GET("/", controllers.GetProducts)
		products.GET("/:id", controllers.GetProduct)
		products.PUT("/:id", controllers.UpdateProduct)
		products.DELETE("/:id", controllers.DeleteProduct)
		products.PATCH("/:id/quantity", controllers.UpdateProductQuantity)
	}

	// Order routes (protected)
	orders := protected.Group("/orders")
	{
		orders.POST("/", controllers.CreateOrder)
		orders.GET("/", controllers.GetOrders)
		orders.GET("/:id", controllers.GetOrder)
		orders.PATCH("/:id/status", controllers.UpdateOrderStatus)
		orders.PATCH("/:id/cancel", controllers.CancelOrder)
	}

	// To do routes (protected)
	todos := protected.Group("/todos")
	{
		todos.POST("/", controllers.CreateTodo)
		todos.GET("/", controllers.GetTodos)
		todos.GET("/:id", controllers.GetTodo)
		todos.PUT("/:id", controllers.UpdateTodo)
		todos.DELETE("/:id", controllers.DeleteTodo)
		todos.PATCH("/:id/toggle", controllers.ToggleTodoComplete)
	}

	return router, nil
}
