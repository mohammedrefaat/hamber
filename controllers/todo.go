package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

type CreateTodoRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
}

// CreateTodo godoc
// @Summary      Create a new todo
// @Description  Create a new todo task
// @Tags         Todos
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateTodoRequest true "Todo details"
// @Success      201 {object} map[string]interface{} "Todo created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /todos [post]
func CreateTodo(c *gin.Context) {
	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	todo := dbmodels.Todo{
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		DueDate:     req.DueDate,
		UserID:      userID,
		IsCompleted: false,
	}

	if err := globalStore.StStore.CreateTodo(&todo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"todo":    todo,
		"message": "Todo created successfully",
	})
}

// GetTodos godoc
// @Summary      Get todos list
// @Description  Get paginated list of user todos
// @Tags         Todos
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Param        completed query boolean false "Filter by completion status"
// @Success      200 {object} map[string]interface{} "Todos list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /todos [get]
func GetTodos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var isCompleted *bool
	if completedStr := c.Query("completed"); completedStr != "" {
		completed := completedStr == "true"
		isCompleted = &completed
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	todos, total, err := globalStore.StStore.GetTodos(page, limit, userID, isCompleted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"todos":       todos,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// GetTodo retrieves a single todo
func GetTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	todo, err := globalStore.StStore.GetTodo(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"todo": todo})
}

// UpdateTodo updates an existing todo
func UpdateTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	todo, err := globalStore.StStore.GetTodo(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo.Title = req.Title
	todo.Description = req.Description
	todo.Priority = req.Priority
	todo.DueDate = req.DueDate

	if err := globalStore.StStore.UpdateTodo(todo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"todo":    todo,
		"message": "Todo updated successfully",
	})
}

// DeleteTodo deletes a todo
func DeleteTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := globalStore.StStore.DeleteTodo(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

// ToggleTodoComplete toggles todo completion status
func ToggleTodoComplete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	todo, err := globalStore.StStore.GetTodo(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	if todo.IsCompleted {
		err = globalStore.StStore.MarkTodoIncomplete(uint(id), userID)
	} else {
		err = globalStore.StStore.MarkTodoCompleted(uint(id), userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo status updated successfully"})
}
