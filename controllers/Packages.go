package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllPackages godoc
// @Summary      Get all packages
// @Description  Get list of all available packages
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      500 {object} map[string]interface{}

func GetAllPackages(c *gin.Context) {
	packages, err := globalStore.StStore.GetAllPackages()
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

// GetPackage godoc
// @Summary      Get package by ID
// @Description  Get details of a specific package
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Param        id path int true "Package ID"
// @Success      200 {object} dbmodels.Package "Package details"
// @Failure      404 {object} map[string]interface{} "Package not found"
// @Router       /packages/{id} [get]
func GetPackage(c *gin.Context) {
	packageID := c.Param("id")
	id, err := strconv.ParseUint(packageID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid package ID",
		})
		return
	}

	pkg, err := globalStore.StStore.GetPackage(uint(id))
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
