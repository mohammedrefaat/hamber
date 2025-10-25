package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== CALENDAR EVENT MANAGEMENT ==========

func (store *DbStore) CreateCalendarEvent(event *dbmodels.CalendarEvent) error {
	return store.db.Create(event).Error
}

func (store *DbStore) GetUserCalendarEvents(userID uint, startDate, endDate time.Time, includePublic bool) ([]dbmodels.CalendarEvent, error) {
	var events []dbmodels.CalendarEvent

	query := store.db.Where("start_time >= ? AND start_time <= ?", startDate, endDate)

	if includePublic {
		query = query.Where("user_id = ? OR is_public = ?", userID, true)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Order("start_time ASC").Find(&events).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch calendar events",
			Code:    http.StatusInternalServerError,
		}
	}

	return events, nil
}

func (store *DbStore) GetPublicEvents(startDate, endDate time.Time) ([]dbmodels.CalendarEvent, error) {
	var events []dbmodels.CalendarEvent

	if err := store.db.Where("is_public = ? AND start_time >= ? AND start_time <= ?",
		true, startDate, endDate).
		Preload("User").
		Order("start_time ASC").
		Find(&events).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch public events",
			Code:    http.StatusInternalServerError,
		}
	}

	return events, nil
}

func (store *DbStore) GetCalendarEvent(id uint, userID uint) (*dbmodels.CalendarEvent, error) {
	var event dbmodels.CalendarEvent
	query := store.db.Preload("User")

	// Allow access to own events or public events
	query = query.Where("id = ? AND (user_id = ? OR is_public = ?)", id, userID, true)

	if err := query.First(&event).Error; err != nil {
		return nil, &CustomError{
			Message: "Event not found",
			Code:    http.StatusNotFound,
		}
	}

	return &event, nil
}

func (store *DbStore) UpdateCalendarEvent(event *dbmodels.CalendarEvent) error {
	return store.db.Save(event).Error
}

func (store *DbStore) DeleteCalendarEvent(id uint, userID uint) error {
	return store.db.Where("id = ? AND user_id = ?", id, userID).Delete(&dbmodels.CalendarEvent{}).Error
}

// ========== EVENT ATTENDEES ==========

func (store *DbStore) AddEventAttendee(attendee *dbmodels.EventAttendee) error {
	return store.db.Create(attendee).Error
}

func (store *DbStore) GetEventAttendees(eventID uint) ([]dbmodels.EventAttendee, error) {
	var attendees []dbmodels.EventAttendee
	if err := store.db.Preload("User").Where("event_id = ?", eventID).Find(&attendees).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch attendees",
			Code:    http.StatusInternalServerError,
		}
	}
	return attendees, nil
}

func (store *DbStore) UpdateAttendeeResponse(attendeeID uint, status string) error {
	return store.db.Model(&dbmodels.EventAttendee{}).
		Where("id = ?", attendeeID).
		Update("response_status", status).Error
}

// ========== ADMIN CALENDAR FUNCTIONS ==========

func (store *DbStore) GetAllCalendarEvents(page, limit int, eventType, status string) ([]dbmodels.CalendarEvent, int64, error) {
	var events []dbmodels.CalendarEvent
	var total int64

	query := store.db.Model(&dbmodels.CalendarEvent{})

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if status != "" {
		statusMap := map[string]dbmodels.EventStatus{
			"SCHEDULED": dbmodels.EventStatus_SCHEDULED,
			"ONGOING":   dbmodels.EventStatus_ONGOING,
			"COMPLETED": dbmodels.EventStatus_COMPLETED,
			"CANCELLED": dbmodels.EventStatus_CANCELLED,
			"POSTPONED": dbmodels.EventStatus_POSTPONED,
		}
		if s, ok := statusMap[status]; ok {
			query = query.Where("status = ?", s)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count events",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").
		Offset(offset).Limit(limit).
		Order("start_time DESC").
		Find(&events).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch events",
			Code:    http.StatusInternalServerError,
		}
	}

	return events, total, nil
}

func (store *DbStore) GetCalendarStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total events
	var totalEvents int64
	store.db.Model(&dbmodels.CalendarEvent{}).Count(&totalEvents)
	stats["total_events"] = totalEvents

	// Public events
	var publicEvents int64
	store.db.Model(&dbmodels.CalendarEvent{}).Where("is_public = ?", true).Count(&publicEvents)
	stats["public_events"] = publicEvents

	// Events by status
	statusCounts := make(map[string]int64)
	statuses := []string{"SCHEDULED", "ONGOING", "COMPLETED", "CANCELLED", "POSTPONED"}
	for _, status := range statuses {
		var count int64
		statusMap := map[string]dbmodels.EventStatus{
			"SCHEDULED": dbmodels.EventStatus_SCHEDULED,
			"ONGOING":   dbmodels.EventStatus_ONGOING,
			"COMPLETED": dbmodels.EventStatus_COMPLETED,
			"CANCELLED": dbmodels.EventStatus_CANCELLED,
			"POSTPONED": dbmodels.EventStatus_POSTPONED,
		}
		store.db.Model(&dbmodels.CalendarEvent{}).Where("status = ?", statusMap[status]).Count(&count)
		statusCounts[status] = count
	}
	stats["by_status"] = statusCounts

	// Events by type
	var typeCounts []map[string]interface{}
	store.db.Model(&dbmodels.CalendarEvent{}).
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Scan(&typeCounts)
	stats["by_type"] = typeCounts

	// Upcoming events (next 7 days)
	var upcomingEvents int64
	now := time.Now()
	nextWeek := now.AddDate(0, 0, 7)
	store.db.Model(&dbmodels.CalendarEvent{}).
		Where("start_time >= ? AND start_time <= ? AND status = ?", now, nextWeek, dbmodels.EventStatus_SCHEDULED).
		Count(&upcomingEvents)
	stats["upcoming_events"] = upcomingEvents

	return stats, nil
}
