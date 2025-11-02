package dbmodels

import "time"

// ========== INTERNAL MESSAGING SYSTEM (Like Email) ==========

// Message represents an internal message between users
type Message struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	SenderID   uint       `gorm:"not null" json:"sender_id"`
	Sender     User       `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReceiverID uint       `gorm:"not null" json:"receiver_id"`
	Receiver   User       `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	Subject    string     `gorm:"size:500;not null" json:"subject"`
	Body       string     `gorm:"type:text;not null" json:"body"`
	IsRead     bool       `gorm:"default:false" json:"is_read"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	IsStarred  bool       `gorm:"default:false" json:"is_starred"`
	IsArchived bool       `gorm:"default:false" json:"is_archived"`
	IsDraft    bool       `gorm:"default:false" json:"is_draft"`

	// For deleted messages (soft delete)
	DeletedBySender   bool `gorm:"default:false" json:"deleted_by_sender"`
	DeletedByReceiver bool `gorm:"default:false" json:"deleted_by_receiver"`

	// Attachments stored as JSON array of file paths
	Attachments string    `gorm:"type:text" json:"attachments"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MessageFolder represents custom folders for organizing messages
type MessageFolder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Color     string    `gorm:"size:20" json:"color"`
	Icon      string    `gorm:"size:50" json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageLabel represents labels/tags for messages
type MessageLabel struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	MessageID uint          `gorm:"not null" json:"message_id"`
	Message   Message       `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	FolderID  uint          `gorm:"not null" json:"folder_id"`
	Folder    MessageFolder `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}
