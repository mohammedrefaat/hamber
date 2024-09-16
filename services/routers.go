package services

import (
	"github.com/gin-gonic/gin"
	middleware "github.com/mohammedrefaat/hamber/Middleware"
)

func GetRouter() (*gin.Engine, error) {
	router := gin.Default()
	router.Use(middleware.LanguageMiddleware())
	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	return router, nil
}
