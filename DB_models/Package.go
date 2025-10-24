package dbmodels

import "time"

// Updated Package model with benefits stored as JSON
type Package struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Price          float64   `gorm:"not null" json:"price"`
	Duration       int       `gorm:"not null" json:"duration"`  // In days or months
	Benefits       string    `gorm:"type:text" json:"benefits"` // JSON string for benefits
	Description    string    `gorm:"type:text" json:"description"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	PricePerClient bool      `json:"price_per_client"`
}
