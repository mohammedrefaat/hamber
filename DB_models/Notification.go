package dbmodels

import "time"

type Notification struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title     string     `gorm:"size:255;not null" json:"title"`
	Message   string     `gorm:"type:text;not null" json:"message"`
	Type      string     `gorm:"size:50" json:"type"` // e.g., 'info', 'warning', 'error', 'success'
	IsRead    bool       `gorm:"default:false" json:"is_read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	Link      string     `gorm:"size:500" json:"link"` // Optional link for the notification
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
