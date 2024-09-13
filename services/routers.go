package services

import "github.com/gin-gonic/gin"

func GetRouter() (*gin.Engine, error) {
	router := gin.Default()
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
