package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== INTERNAL MESSAGING STORE ==========

func (store *DbStore) CreateMessage(message *dbmodels.Message) error {
	return store.db.Create(message).Error
}

func (store *DbStore) GetInboxMessages(userID uint, page, limit int, unreadOnly bool) ([]dbmodels.Message, int64, error) {
	var messages []dbmodels.Message
	var total int64

	query := store.db.Model(&dbmodels.Message{}).Where("receiver_id = ? AND deleted_by_receiver = ?", userID, false)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	query = query.Where("is_archived = ?", false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count messages",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("Sender").Offset(offset).Limit(limit).Order("created_at DESC").Find(&messages).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch messages",
			Code:    http.StatusInternalServerError,
		}
	}

	return messages, total, nil
}

func (store *DbStore) GetSentMessages(userID uint, page, limit int) ([]dbmodels.Message, int64, error) {
	var messages []dbmodels.Message
	var total int64

	query := store.db.Model(&dbmodels.Message{}).Where("sender_id = ? AND deleted_by_sender = ?", userID, false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count messages",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("Receiver").Offset(offset).Limit(limit).Order("created_at DESC").Find(&messages).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch messages",
			Code:    http.StatusInternalServerError,
		}
	}

	return messages, total, nil
}

func (store *DbStore) GetMessage(id, userID uint) (*dbmodels.Message, error) {
	var message dbmodels.Message
	if err := store.db.Preload("Sender").Preload("Receiver").
		Where("id = ? AND (sender_id = ? OR receiver_id = ?)", id, userID, userID).
		First(&message).Error; err != nil {
		return nil, &CustomError{
			Message: "Message not found",
			Code:    http.StatusNotFound,
		}
	}
	return &message, nil
}

func (store *DbStore) MarkMessageAsRead(id uint) error {
	now := time.Now()
	return store.db.Model(&dbmodels.Message{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

func (store *DbStore) DeleteMessage(id, userID uint) error {
	message, err := store.GetMessage(id, userID)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if message.SenderID == userID {
		updates["deleted_by_sender"] = true
	}
	if message.ReceiverID == userID {
		updates["deleted_by_receiver"] = true
	}

	return store.db.Model(&dbmodels.Message{}).Where("id = ?", id).Updates(updates).Error
}

func (store *DbStore) ToggleStarMessage(id, userID uint) error {
	var message dbmodels.Message
	if err := store.db.Where("id = ? AND receiver_id = ?", id, userID).First(&message).Error; err != nil {
		return &CustomError{
			Message: "Message not found",
			Code:    http.StatusNotFound,
		}
	}

	return store.db.Model(&dbmodels.Message{}).Where("id = ?", id).Update("is_starred", !message.IsStarred).Error
}

func (store *DbStore) ArchiveMessage(id, userID uint) error {
	return store.db.Model(&dbmodels.Message{}).
		Where("id = ? AND receiver_id = ?", id, userID).
		Update("is_archived", true).Error
}

func (store *DbStore) GetMessageStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total inbox messages
	var inboxCount int64
	store.db.Model(&dbmodels.Message{}).Where("receiver_id = ? AND deleted_by_receiver = ?", userID, false).Count(&inboxCount)
	stats["inbox_count"] = inboxCount

	// Unread messages
	var unreadCount int64
	store.db.Model(&dbmodels.Message{}).Where("receiver_id = ? AND is_read = ? AND deleted_by_receiver = ?", userID, false, false).Count(&unreadCount)
	stats["unread_count"] = unreadCount

	// Sent messages
	var sentCount int64
	store.db.Model(&dbmodels.Message{}).Where("sender_id = ? AND deleted_by_sender = ?", userID, false).Count(&sentCount)
	stats["sent_count"] = sentCount

	// Starred messages
	var starredCount int64
	store.db.Model(&dbmodels.Message{}).Where("receiver_id = ? AND is_starred = ? AND deleted_by_receiver = ?", userID, true, false).Count(&starredCount)
	stats["starred_count"] = starredCount

	// Archived messages
	var archivedCount int64
	store.db.Model(&dbmodels.Message{}).Where("receiver_id = ? AND is_archived = ? AND deleted_by_receiver = ?", userID, true, false).Count(&archivedCount)
	stats["archived_count"] = archivedCount

	return stats, nil
}

// ========== BANNER STORE ==========

func (store *DbStore) CreateBanner(banner *dbmodels.Banner) error {
	return store.db.Create(banner).Error
}

func (store *DbStore) GetActiveBanners(position string, userID *uint) ([]dbmodels.Banner, error) {
	var banners []dbmodels.Banner
	now := time.Now()

	query := store.db.Where("is_active = ?", true).
		Where("(start_date IS NULL OR start_date <= ?)", now).
		Where("(end_date IS NULL OR end_date >= ?)", now)

	if position != "" {
		query = query.Where("position = ?", position)
	}

	// TODO: Add targeting logic based on userID and roles

	if err := query.Order("priority DESC, created_at DESC").Find(&banners).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch banners",
			Code:    http.StatusInternalServerError,
		}
	}

	return banners, nil
}

func (store *DbStore) GetAllBanners(page, limit int) ([]dbmodels.Banner, int64, error) {
	var banners []dbmodels.Banner
	var total int64

	query := store.db.Model(&dbmodels.Banner{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count banners",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("Creator").Offset(offset).Limit(limit).Order("created_at DESC").Find(&banners).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch banners",
			Code:    http.StatusInternalServerError,
		}
	}

	return banners, total, nil
}

func (store *DbStore) GetBanner(id uint) (*dbmodels.Banner, error) {
	var banner dbmodels.Banner
	if err := store.db.Preload("Creator").First(&banner, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Banner not found",
			Code:    http.StatusNotFound,
		}
	}
	return &banner, nil
}

func (store *DbStore) UpdateBanner(banner *dbmodels.Banner) error {
	return store.db.Save(banner).Error
}

func (store *DbStore) DeleteBanner(id uint) error {
	return store.db.Delete(&dbmodels.Banner{}, id).Error
}

func (store *DbStore) TrackBannerView(bannerID uint, userID *uint, ipAddress, userAgent string) error {
	view := dbmodels.BannerView{
		BannerID:  bannerID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := store.db.Create(&view).Error; err != nil {
		return err
	}

	// Increment view count
	return store.db.Model(&dbmodels.Banner{}).Where("id = ?", bannerID).UpdateColumn("view_count", store.db.Raw("view_count + 1")).Error
}

func (store *DbStore) TrackBannerClick(bannerID uint, userID *uint, ipAddress, userAgent string) error {
	click := dbmodels.BannerClick{
		BannerID:  bannerID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := store.db.Create(&click).Error; err != nil {
		return err
	}

	// Increment click count
	return store.db.Model(&dbmodels.Banner{}).Where("id = ?", bannerID).UpdateColumn("click_count", store.db.Raw("click_count + 1")).Error
}

func (store *DbStore) GetBannerAnalytics(bannerID uint) (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

	banner, err := store.GetBanner(bannerID)
	if err != nil {
		return nil, err
	}

	analytics["banner_id"] = banner.ID
	analytics["title"] = banner.Title
	analytics["view_count"] = banner.ViewCount
	analytics["click_count"] = banner.ClickCount

	// Calculate CTR (Click Through Rate)
	if banner.ViewCount > 0 {
		ctr := float64(banner.ClickCount) / float64(banner.ViewCount) * 100
		analytics["ctr"] = ctr
	} else {
		analytics["ctr"] = 0.0
	}

	// Views by day (last 7 days)
	var dailyViews []map[string]interface{}
	store.db.Model(&dbmodels.BannerView{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("banner_id = ? AND created_at >= ?", bannerID, time.Now().AddDate(0, 0, -7)).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&dailyViews)
	analytics["daily_views"] = dailyViews

	// Clicks by day (last 7 days)
	var dailyClicks []map[string]interface{}
	store.db.Model(&dbmodels.BannerClick{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("banner_id = ? AND created_at >= ?", bannerID, time.Now().AddDate(0, 0, -7)).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&dailyClicks)
	analytics["daily_clicks"] = dailyClicks

	return analytics, nil
}
