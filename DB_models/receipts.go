package dbmodels

import "time"

// ========== ORDER RECEIPT SYSTEM ==========

// OrderReceipt stores receipt metadata for PDF generation
type OrderReceipt struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	OrderID         uint       `gorm:"not null;unique" json:"order_id"`
	Order           Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	ReceiptNumber   string     `gorm:"size:100;unique;not null" json:"receipt_number"`
	PDFPath         string     `gorm:"size:500" json:"pdf_path"` // Path to generated PDF
	GeneratedAt     *time.Time `json:"generated_at,omitempty"`
	TemplateVersion string     `gorm:"size:50;default:'v1'" json:"template_version"`
	CompanyInfo     string     `gorm:"type:text" json:"company_info"` // JSON with company details
	Notes           string     `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
