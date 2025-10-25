package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== NOTIFICATION MANAGEMENT ==========

func (store *DbStore) CreateNotification(notification *dbmodels.Notification) error {
	return store.db.Create(notification).Error
}

func (store *DbStore) GetUserNotifications(userID uint, page, limit int, unreadOnly bool) ([]dbmodels.Notification, int64, error) {
	var notifications []dbmodels.Notification
	var total int64

	query := store.db.Model(&dbmodels.Notification{}).Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count notifications",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch notifications",
			Code:    http.StatusInternalServerError,
		}
	}

	return notifications, total, nil
}

func (store *DbStore) GetNotification(id uint) (*dbmodels.Notification, error) {
	var notification dbmodels.Notification
	if err := store.db.First(&notification, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Notification not found",
			Code:    http.StatusNotFound,
		}
	}
	return &notification, nil
}

func (store *DbStore) MarkNotificationAsRead(id uint) error {
	now := time.Now()
	return store.db.Model(&dbmodels.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (store *DbStore) MarkAllNotificationsAsRead(userID uint) error {
	now := time.Now()
	return store.db.Model(&dbmodels.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (store *DbStore) DeleteNotification(id uint) error {
	return store.db.Delete(&dbmodels.Notification{}, id).Error
}

func (store *DbStore) GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	if err := store.db.Model(&dbmodels.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, &CustomError{
			Message: "Failed to count unread notifications",
			Code:    http.StatusInternalServerError,
		}
	}
	return count, nil
}

func (store *DbStore) GetAllUsersSimple() ([]dbmodels.User, error) {
	var users []dbmodels.User
	if err := store.db.Select("id").Where("is_active = ?", true).Find(&users).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch users",
			Code:    http.StatusInternalServerError,
		}
	}
	return users, nil
}

// DeleteOldNotifications deletes notifications older than specified days
func (store *DbStore) DeleteOldNotifications(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return store.db.Where("created_at < ? AND is_read = ?", cutoffDate, true).
		Delete(&dbmodels.Notification{}).Error
}
