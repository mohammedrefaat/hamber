package dbmodels

import "time"

// ========== BANNER MANAGEMENT SYSTEM ==========

// Banner represents promotional banners on the platform
type Banner struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"size:255;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Photo       string `gorm:"size:500;not null" json:"photo"` // Photo URL or base64
	Link        string `gorm:"size:500" json:"link"`           // Where banner links to
	LinkText    string `gorm:"size:100" json:"link_text"`      // Button text
	Position    string `gorm:"size:50" json:"position"`        // top, middle, bottom, sidebar
	Priority    int    `gorm:"default:0" json:"priority"`      // Higher = shown first
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// Scheduling
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`

	// Targeting
	TargetRoles string `gorm:"type:text" json:"target_roles"` // JSON array of role names
	TargetUsers string `gorm:"type:text" json:"target_users"` // JSON array of user IDs

	// Analytics
	ViewCount  int `gorm:"default:0" json:"view_count"`
	ClickCount int `gorm:"default:0" json:"click_count"`

	// Creator
	CreatedBy uint `gorm:"not null" json:"created_by"`
	Creator   User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BannerView tracks banner views
type BannerView struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BannerID  uint      `gorm:"not null" json:"banner_id"`
	Banner    Banner    `gorm:"foreignKey:BannerID" json:"banner,omitempty"`
	UserID    *uint     `json:"user_id,omitempty"` // null for anonymous
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IPAddress string    `gorm:"size:50" json:"ip_address"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// BannerClick tracks banner clicks
type BannerClick struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BannerID  uint      `gorm:"not null" json:"banner_id"`
	Banner    Banner    `gorm:"foreignKey:BannerID" json:"banner,omitempty"`
	UserID    *uint     `json:"user_id,omitempty"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IPAddress string    `gorm:"size:50" json:"ip_address"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}
