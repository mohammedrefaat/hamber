package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllBlogsAdmin returns all blogs (including unpublished) for admin users
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
		blogs, total, err = globalStore.GetBlogsByAuthor(uint(id), page, limit)
	} else {
		// Get all blogs (including unpublished)
		blogs, total, err = globalStore.GetBlogs(page, limit, false)
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

// GetBlogAnalytics returns blog statistics for admin dashboard
func GetBlogAnalytics(c *gin.Context) {
	analytics, err := globalStore.GetBlogAnalytics()
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
