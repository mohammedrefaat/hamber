package services

import (
	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	middleware "github.com/mohammedrefaat/hamber/Middleware"
	"github.com/mohammedrefaat/hamber/controllers"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	// Import your generated docs
	//_ "github.com/mohammedrefaat/hamber/docs"
)

func GetRouter(cfg *config.Config) (*gin.Engine, error) {
	router := gin.Default()
	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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

		// Add-on routes (public - viewing only)
		addons := api.Group("/addons")
		{
			addons.GET("/", controllers.GetAddons)
			addons.GET("/:id", controllers.GetAddon)
		}

		// Public calendar events
		api.GET("/calendar/public", controllers.GetPublicEvents)

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

		// Payment callback routes (public - called by payment gateways)
		paymentCallbacks := api.Group("/payment")
		{
			paymentCallbacks.POST("/fawry/callback", controllers.FawryCallback)
			paymentCallbacks.POST("/paymob/callback", controllers.PaymobCallback)
		}
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
			photos.POST("/avatar", controllers.UploadAvatarPhoto)
			photos.GET("/presigned-url", controllers.GetPhotoPresignedURL)
		}

		// Protected blog routes
		protectedBlogs := protected.Group("/blogs")
		{
			protectedBlogs.POST("/", controllers.CreateBlogWithPhotos)
			protectedBlogs.PUT("/:id", controllers.UpdateBlog)
			protectedBlogs.DELETE("/:id", controllers.DeleteBlog)
			protectedBlogs.POST("/:id/photos", controllers.UploadBlogPhotoV2)
			protectedBlogs.DELETE("/:id/photos", controllers.DeleteBlogPhoto)
		}

		// User permissions
		protected.GET("/permissions", controllers.GetUserPermissions)

		// Product routes (protected)
		products := protected.Group("/products")
		{
			products.POST("/", controllers.CreateProduct)
			products.GET("/", controllers.GetProducts)
			products.GET("/:id", controllers.GetProduct)
			products.PUT("/:id", controllers.UpdateProduct)
			products.DELETE("/:id", controllers.DeleteProduct)
			products.PATCH("/:id/quantity", controllers.UpdateProductQuantity)
			products.GET("/categories", controllers.GetProductCategories)
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

		// Receipt routes (protected)
		receipts := protected.Group("/receipts")
		{
			receipts.POST("/order/:order_id", controllers.GenerateOrderReceipt)
			receipts.GET("/order/:order_id", controllers.GetOrderReceipt)
			receipts.GET("/order/:order_id/download", controllers.DownloadReceipt)
			receipts.GET("/order/:order_id/html", controllers.GetReceiptHTML)
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

		// Calendar routes (protected)
		calendar := protected.Group("/calendar")
		{
			calendar.POST("/events", controllers.CreateCalendarEvent)
			calendar.GET("/events", controllers.GetUserEvents)
			calendar.GET("/events/:id", controllers.GetCalendarEvent)
			calendar.PUT("/events/:id", controllers.UpdateCalendarEvent)
			calendar.DELETE("/events/:id", controllers.DeleteCalendarEvent)
			calendar.PATCH("/events/:id/status", controllers.UpdateEventStatus)
			calendar.PATCH("/attendees/:attendee_id/respond", controllers.RespondToInvitation)
		}

		// Add-on subscription routes (protected)
		addonSubscriptions := protected.Group("/subscriptions")
		{
			addonSubscriptions.POST("/", controllers.SubscribeToAddon)
			addonSubscriptions.GET("/", controllers.GetUserSubscriptions)
			addonSubscriptions.GET("/:id", controllers.GetSubscription)
			addonSubscriptions.DELETE("/:id/cancel", controllers.CancelSubscription)
			addonSubscriptions.POST("/:id/usage", controllers.LogUsage)
			addonSubscriptions.GET("/:id/usage", controllers.GetUsageLogs)
		}

		// Notification routes (protected)
		notifications := protected.Group("/notifications")
		{
			notifications.GET("/", controllers.GetUserNotifications)
			notifications.GET("/unread-count", controllers.GetUnreadCount)
			notifications.PATCH("/:id/read", controllers.MarkNotificationAsRead)
			notifications.PATCH("/read-all", controllers.MarkAllNotificationsAsRead)
			notifications.DELETE("/:id", controllers.DeleteNotification)
		}

		// Payment routes (protected)
		payment := protected.Group("/payment")
		{
			payment.POST("/change-package", controllers.RequestPackageChange)
			payment.GET("/status/:id", controllers.GetPaymentStatus)
			payment.GET("/history", controllers.GetUserPayments)
			payment.GET("/package-changes", controllers.GetPackageChangeHistory)
		}

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
			admin.GET("/blogs", controllers.GetAllBlogsAdmin)
			admin.GET("/blogs/analytics", controllers.GetBlogAnalytics)

			// Newsletter management
			adminNewsletter := admin.Group("/newsletter")
			{
				adminNewsletter.GET("/subscriptions", controllers.GetAllNewsletterSubscriptions)
				adminNewsletter.GET("/stats", controllers.GetNewsletterStats)
			}

			// Payment management
			adminPayment := admin.Group("/payments")
			{
				adminPayment.GET("/", controllers.GetAllPayments)
				adminPayment.GET("/:id", controllers.GetPaymentStatus)
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

			// Add-on management (admin)
			adminAddons := admin.Group("/addons")
			{
				adminAddons.POST("/", controllers.CreateAddon)
				adminAddons.PUT("/:id", controllers.UpdateAddon)
				adminAddons.DELETE("/:id", controllers.DeleteAddon)
				adminAddons.POST("/pricing-tiers", controllers.CreatePricingTier)
			}

			// Calendar management (admin)
			adminCalendar := admin.Group("/calendar")
			{
				adminCalendar.GET("/all-events", controllers.GetAllEvents)
				adminCalendar.GET("/stats", controllers.GetCalendarStats)
			}

			// Photo statistics
			admin.GET("/photos/stats", controllers.GetPhotoStats)
		}
	}

	return router, nil
}
