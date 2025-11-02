package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== DASHBOARD STATISTICS ==========

// GetUserDashboard godoc
// @Summary      Get user dashboard statistics
// @Description  Get comprehensive dashboard stats for current user
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Dashboard statistics"
// @Router       /dashboard/stats [get]
func GetUserDashboard(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stats, err := globalStore.StStore.GetDashboardStats(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetRevenueChart godoc
// @Summary      Get revenue chart data
// @Description  Get revenue data for charts (last 12 months)
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Revenue chart data"
// @Router       /dashboard/revenue-chart [get]
func GetRevenueChart(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chartData, err := globalStore.StStore.GetRevenueChartData(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": chartData,
	})
}

// GetOrdersChart godoc
// @Summary      Get orders chart data
// @Description  Get orders data by status for charts
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Orders chart data"
// @Router       /dashboard/orders-chart [get]
func GetOrdersChart(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chartData, err := globalStore.StStore.GetOrdersChartData(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": chartData,
	})
}

// GetTopProducts godoc
// @Summary      Get top selling products
// @Description  Get top 10 best selling products
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Top products"
// @Router       /dashboard/top-products [get]
func GetTopProducts(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	products, err := globalStore.StStore.GetTopSellingProducts(claims.UserID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch top products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}

// GetRecentOrders godoc
// @Summary      Get recent orders
// @Description  Get last 10 orders
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Recent orders"
// @Router       /dashboard/recent-orders [get]
func GetRecentOrders(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orders, err := globalStore.StStore.GetRecentOrders(claims.UserID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

// GetRecentActivities godoc
// @Summary      Get recent activities
// @Description  Get user's recent activities/actions
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Recent activities"
// @Router       /dashboard/recent-activities [get]
func GetRecentActivities(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	activities, err := globalStore.StStore.GetRecentActivities(claims.UserID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
	})
}

// GetProductStats godoc
// @Summary      Get product statistics
// @Description  Get detailed product statistics
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Product statistics"
// @Router       /dashboard/product-stats [get]
func GetProductStats(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stats, err := globalStore.StStore.GetProductStatistics(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetClientStats godoc
// @Summary      Get client statistics
// @Description  Get client analytics and statistics
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Client statistics"
// @Router       /dashboard/client-stats [get]
func GetClientStats(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stats, err := globalStore.StStore.GetClientStatistics(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch client stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// ========== ADMIN DASHBOARD ==========

// GetAdminDashboard godoc
// @Summary      Get admin dashboard (Admin)
// @Description  Get comprehensive admin dashboard with platform-wide stats
// @Tags         Admin Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Admin dashboard"
// @Router       /admin/dashboard [get]
func GetAdminDashboard(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	stats, err := globalStore.StStore.GetAdminDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admin stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetPlatformAnalytics godoc
// @Summary      Get platform analytics (Admin)
// @Description  Get detailed platform-wide analytics
// @Tags         Admin Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Platform analytics"
// @Router       /admin/analytics [get]
func GetPlatformAnalytics(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	analytics, err := globalStore.StStore.GetPlatformAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics": analytics,
	})
}

// GetUserGrowthChart godoc
// @Summary      Get user growth chart (Admin)
// @Description  Get user registration growth over time
// @Tags         Admin Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "User growth data"
// @Router       /admin/user-growth-chart [get]
func GetUserGrowthChart(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	chartData, err := globalStore.StStore.GetUserGrowthChart()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user growth data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": chartData,
	})
}

// GetRevenueBreakdown godoc
// @Summary      Get revenue breakdown (Admin)
// @Description  Get revenue breakdown by different categories
// @Tags         Admin Dashboard
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Revenue breakdown"
// @Router       /admin/revenue-breakdown [get]
func GetRevenueBreakdown(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	breakdown, err := globalStore.StStore.GetRevenueBreakdown()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue breakdown"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"breakdown": breakdown,
	})
}
