package dbmodels

import "time"

// SiteConfig model for storing customer website configurations
type SiteConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SiteName  string    `gorm:"size:255;unique;not null" json:"site_name"`
	SiteData  string    `gorm:"type:text;not null" json:"site_data"` // JSON string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartItem model for shopping cart functionality
type CartItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `json:"user_id,omitempty"` // null for guest carts
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SessionID string    `gorm:"size:255" json:"session_id"` // For guest users
	ProductID uint      `gorm:"not null" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"` // Price snapshot at time of adding
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
