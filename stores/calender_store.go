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
