package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllBlogsAdmin godoc
// @Summary      Get all blogs (Admin)
// @Description  Get all blogs including unpublished - Admin only
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Param        author_id query int false "Filter by author ID"
// @Success      200 {object} map[string]interface{} "Blogs list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /admin/blogs [get]
func GetAllBlogsAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	authorID := c.Query("author_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var blogs interface{}
	var total int64
	var err error

	if authorID != "" {
		// Get blogs by specific author
		id, parseErr := strconv.ParseUint(authorID, 10, 32)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid author ID",
			})
			return
		}
		blogs, total, err = globalStore.StStore.GetBlogsByAuthor(uint(id), page, limit)
	} else {
		// Get all blogs (including unpublished)
		blogs, total, err = globalStore.StStore.GetBlogs(page, limit, false)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch blogs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blogs": blogs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetBlogAnalytics godoc
// @Summary      Get blog analytics (Admin)
// @Description  Get blog statistics for admin dashboard
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Blog analytics"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /admin/blogs/analytics [get]
func GetBlogAnalytics(c *gin.Context) {
	analytics, err := globalStore.StStore.GetBlogAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch blog analytics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics": analytics,
	})
}
