package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== TODO MANAGEMENT ==========

func (store *DbStore) CreateTodo(todo *dbmodels.Todo) error {
	return store.db.Create(todo).Error
}

func (store *DbStore) GetTodos(page, limit int, userID uint, isCompleted *bool) ([]dbmodels.Todo, int64, error) {
	var todos []dbmodels.Todo
	var total int64

	query := store.db.Model(&dbmodels.Todo{}).Where("user_id = ?", userID)

	if isCompleted != nil {
		query = query.Where("is_completed = ?", *isCompleted)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count todos",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&todos).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch todos",
			Code:    http.StatusInternalServerError,
		}
	}

	return todos, total, nil
}

func (store *DbStore) GetTodo(id uint, userID uint) (*dbmodels.Todo, error) {
	var todo dbmodels.Todo
	if err := store.db.Where("id = ? AND user_id = ?", id, userID).First(&todo).Error; err != nil {
		return nil, &CustomError{
			Message: "Todo not found",
			Code:    http.StatusNotFound,
		}
	}
	return &todo, nil
}

func (store *DbStore) UpdateTodo(todo *dbmodels.Todo) error {
	return store.db.Save(todo).Error
}

func (store *DbStore) DeleteTodo(id uint, userID uint) error {
	return store.db.Where("id = ? AND user_id = ?", id, userID).Delete(&dbmodels.Todo{}).Error
}

func (store *DbStore) MarkTodoCompleted(id uint, userID uint) error {
	now := time.Now()
	return store.db.Model(&dbmodels.Todo{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": &now,
		}).Error
}

func (store *DbStore) MarkTodoIncomplete(id uint, userID uint) error {
	return store.db.Model(&dbmodels.Todo{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_completed": false,
			"completed_at": nil,
		}).Error
}
