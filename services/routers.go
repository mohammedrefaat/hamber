package services

import (
	"github.com/gin-gonic/gin"
	middleware "github.com/mohammedrefaat/hamber/Middleware"
	"github.com/mohammedrefaat/hamber/controllers"
)

func GetRouter() (*gin.Engine, error) {
	router := gin.Default()
	router.Use(middleware.LanguageMiddleware())

	// Public routes (no authentication required)
	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.Login)
			auth.POST("/register", controllers.Register)
			auth.POST("/refresh", controllers.RefreshToken)
		}
	}

	// Protected routes (authentication required)
	protected := api.Group("/")
	protected.Use(middleware.JWTMiddleware())
	{
		protected.GET("/profile", controllers.GetProfile)
		protected.PUT("/profile", controllers.UpdateProfile)

		// Admin only routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", controllers.GetAllUsers)
			admin.DELETE("/users/:id", controllers.DeleteUser)
		}
	}

	return router, nil
}
