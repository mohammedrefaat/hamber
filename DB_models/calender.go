package dbmodels

import "time"

// ========== CALENDAR EVENTS SYSTEM ==========

// CalendarEvent represents an event in the calendar
type CalendarEvent struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	UserID         uint        `gorm:"not null" json:"user_id"`
	User           User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title          string      `gorm:"size:255;not null" json:"title"`
	Description    string      `gorm:"type:text" json:"description"`
	Location       string      `gorm:"size:500" json:"location"`
	StartTime      time.Time   `gorm:"not null" json:"start_time"`
	EndTime        time.Time   `gorm:"not null" json:"end_time"`
	AllDay         bool        `gorm:"default:false" json:"all_day"`
	EventType      string      `gorm:"size:50" json:"event_type"`      // "meeting", "task", "reminder", "public"
	Color          string      `gorm:"size:20" json:"color"`           // Hex color for UI
	IsPublic       bool        `gorm:"default:false" json:"is_public"` // Public events visible to all
	Recurring      bool        `gorm:"default:false" json:"recurring"`
	RecurrenceRule string      `gorm:"size:500" json:"recurrence_rule"` // RRULE format
	RemindBefore   int         `gorm:"default:0" json:"remind_before"`  // Minutes before event
	Status         EventStatus `gorm:"default:0" json:"status"`
	Metadata       string      `gorm:"type:text" json:"metadata"` // JSON for extra data
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type EventStatus int32

const (
	EventStatus_SCHEDULED EventStatus = 0
	EventStatus_ONGOING   EventStatus = 1
	EventStatus_COMPLETED EventStatus = 2
	EventStatus_CANCELLED EventStatus = 3
	EventStatus_POSTPONED EventStatus = 4
)

var EventStatus_name = map[int32]string{
	0: "SCHEDULED",
	1: "ONGOING",
	2: "COMPLETED",
	3: "CANCELLED",
	4: "POSTPONED",
}

func (x EventStatus) String() string {
	return EventStatus_name[int32(x)]
}

// EventAttendee represents users invited to an event
type EventAttendee struct {
	ID               uint          `gorm:"primaryKey" json:"id"`
	EventID          uint          `gorm:"not null" json:"event_id"`
	Event            CalendarEvent `gorm:"foreignKey:EventID" json:"event,omitempty"`
	UserID           *uint         `json:"user_id,omitempty"` // Null for external attendees
	User             *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Email            string        `gorm:"size:255" json:"email"` // For external attendees
	Name             string        `gorm:"size:255" json:"name"`
	ResponseStatus   string        `gorm:"size:20;default:'pending'" json:"response_status"` // pending, accepted, declined
	NotificationSent bool          `gorm:"default:false" json:"notification_sent"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}
