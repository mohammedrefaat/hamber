package dbmodels

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255;not null"`
	Email     string `gorm:"size:255;unique;not null"`
	Password  string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Enhanced User model with phone support for future use
type User struct {
	ID                          uint      `gorm:"primaryKey" json:"ID,omitempty"`
	Name                        string    `gorm:"size:255;not null" json:"Name,omitempty"`
	Email                       string    `gorm:"size:255;unique;not null" json:"Email,omitempty"`
	Password                    string    `gorm:"not null" json:"Password,omitempty"`
	Phone                       string    `gorm:"size:20" json:"Phone,omitempty"` // Added phone field
	Subdomain                   string    `gorm:"size:255;unique;not null" json:"Subdomain,omitempty"`
	RoleID                      uint      `gorm:"not null" json:"RoleID,omitempty"`            // Foreign key to the Role table
	Role                        []Role    `gorm:"many2many:user_roles;" json:"Role,omitempty"` // Many-to-many relationship between users and roles
	PackageID                   uint      `gorm:"not null" json:"PackageID,omitempty"`
	Package                     Package   `gorm:"foreignKey:PackageID" json:"Package,omitempty"`
	CreatedAt                   time.Time `json:"CreatedAt,omitempty"`
	UpdatedAt                   time.Time `json:"UpdatedAt,omitempty"`
	IS_ACTIVE                   bool      `gorm:"default:true" json:"IS_ACTIVE,omitempty"`
	ACTIVATION_CODE             string    `gorm:"size:255" json:"ACTIVATION_CODE,omitempty"`
	DEVICE_TOKEN                string    `gorm:"size:255" json:"DEVICE_TOKEN,omitempty"`
	IS_BLOCKED                  bool      `gorm:"default:false" json:"IS_BLOCKED,omitempty"`
	COUNT_SEND_ACTIVATION_EMAIL int       `gorm:"default:0" json:"COUNT_SEND_ACTIVATION_EMAIL,omitempty"`
	NID                         string    `gorm:"size:255" json:"NID,omitempty"`
	RESET_CODE                  string    `gorm:"size:255" json:"RESET_CODE,omitempty"`
	EXPIRESAT                   time.Time `json:"EXPIRESAT,omitempty"`
	LAST_LOGIN_IP               string    `gorm:"size:255" json:"LAST_LOGIN_IP,omitempty"`
	IS_EMAIL_VERIFIED           bool      `gorm:"default:false" json:"IS_EMAIL_VERIFIED,omitempty"`
	IS_MOBILE_VERIFIED          bool      `gorm:"default:false" json:"IS_MOBILE_VERIFIED,omitempty"`

	// New fields for future phone verification
	PHONE_VERIFICATION_CODE string     `gorm:"size:10" json:"PHONE_VERIFICATION_CODE,omitempty"`
	PHONE_CODE_EXPIRES_AT   time.Time  `json:"PHONE_CODE_EXPIRES_AT,omitempty"`
	PHONE_VERIFIED_AT       *time.Time `json:"PHONE_VERIFIED_AT,omitempty"`

	// Profile enhancement
	Avatar   string `gorm:"size:500" json:"Avatar,omitempty"`
	Bio      string `gorm:"type:text" json:"Bio,omitempty"`
	Website  string `gorm:"size:500" json:"Website,omitempty"`
	Location string `gorm:"size:255" json:"Location,omitempty"`
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Phone verification model for future SMS verification
type PhoneVerification struct {
	ID        uint      `gorm:"primaryKey"`
	Phone     string    `gorm:"size:20;not null"`
	Code      string    `gorm:"size:10;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	UserID    *uint     `gorm:"index"` // Optional, for linking to user
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Role struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:100;unique;not null"` // Role name
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Permissions []Permission `gorm:"many2many:role_permissions;"` // Role-Permissions relationship
}

type Permission struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;unique;not null"` // Permission name (e.g., "CREATE_ORDER")
	CreatedAt time.Time
	UpdatedAt time.Time
}
type Client struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255;not null"`
	Email     string `gorm:"size:255;not null"`
	UserID    uint   `gorm:"not null"`
	User      User   `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (u User) HasPermission(permissionName string) bool {
	for _, role := range u.Role {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				return true
			}
		}
	}
	return false
}

// EmailVerification model for email verification codes
type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"size:255;not null"`
	Code      string    `gorm:"size:10;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PasswordReset model for password reset codes
type PasswordReset struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"size:255;not null"`
	Code      string    `gorm:"size:10;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
