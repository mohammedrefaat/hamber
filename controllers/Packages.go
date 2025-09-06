package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllPackages returns all active packages
func GetAllPackages(c *gin.Context) {
	packages, err := globalStore.GetAllPackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch packages",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"packages": packages,
	})
}

// GetPackage returns a specific package by ID
func GetPackage(c *gin.Context) {
	packageID := c.Param("id")
	id, err := strconv.ParseUint(packageID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid package ID",
		})
		return
	}

	pkg, err := globalStore.GetPackage(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Package not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"package": pkg,
	})
}
