package services

import (
	"github.com/gin-gonic/gin"
	middleware "github.com/mohammedrefaat/hamber/Middleware"
	"github.com/mohammedrefaat/hamber/controllers"
)

func GetRouter() (*gin.Engine, error) {
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
			admin.GET("/users", controllers.GetAllUsers)
			admin.DELETE("/users/:id", controllers.DeleteUser)

			// Admin blog management
			admin.GET("/blogs", controllers.GetAllBlogsAdmin) // All blogs including unpublished
		}
	}

	return router, nil
}
