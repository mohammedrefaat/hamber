package dbmodels

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Models struct {
	User           User
	Role           Role
	Permission     Permission
	Admin          Admin
	Client         Client
	Package        Package
	Order          Order
	Subscription   Subscription
	RolePermission RolePermission
	UserRole       UserRole
}

// Migrator runs auto-migration for all models
func Migrator(db *gorm.DB) error {
	modelsToMigrate := GetMod()

	// Loop through the models and auto-migrate each one
	for _, model := range modelsToMigrate {
		err := db.AutoMigrate(model)
		if err != nil {
			return err
		}
	}
	return nil
}

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
type Order struct {
	ID        uint        `gorm:"primaryKey"`
	ClientID  uint        `gorm:"not null"`
	Client    Client      `gorm:"foreignKey:ClientID"`
	UserID    uint        `gorm:"not null"`
	User      User        `gorm:"foreignKey:UserID"`
	Total     float64     `gorm:"not null"`
	Status    OrderStatus `gorm:"not null"` // Enum as int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Subscription struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	PackageID uint      `gorm:"not null"`
	Package   Package   `gorm:"foreignKey:PackageID"`
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Price     float64 `gorm:"not null"`
}

type OrderStatus int32

const (
	OrderStatus_PENDING   OrderStatus = 0
	OrderStatus_SHIPPED   OrderStatus = 1
	OrderStatus_DELIVERED OrderStatus = 2
	OrderStatus_CANCELED  OrderStatus = 3
)

// Enum value maps for OrderStatus.
var (
	OrderStatus_name = map[int32]string{
		0: "PENDING",
		1: "SHIPPED",
		2: "DELIVERED",
		3: "CANCELED",
	}
	OrderStatus_value = map[string]int32{
		"PENDING":   0,
		"SHIPPED":   1,
		"DELIVERED": 2,
		"CANCELED":  3,
	}
)

func (x OrderStatus) Enum() *OrderStatus {
	p := new(OrderStatus)
	*p = x
	return p
}

func (x OrderStatus) String() string {
	return OrderStatus_name[int32(x)]
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

// Blog model for blog posts
type Blog struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"size:500;not null" json:"title"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	Summary     string     `gorm:"type:text" json:"summary"`
	Slug        string     `gorm:"size:500;unique;not null" json:"slug"`
	AuthorID    uint       `gorm:"not null" json:"author_id"`
	Author      User       `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Photos      string     `gorm:"type:text" json:"photos"` // JSON array of photo URLs
	IsPublished bool       `gorm:"default:false" json:"is_published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Newsletter subscription model
type Newsletter struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Email          string     `gorm:"size:255;unique;not null" json:"email"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	SubscribedAt   time.Time  `json:"subscribed_at"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Contact form submission model
type Contact struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Email     string    `gorm:"size:255;not null" json:"email"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	Replied   bool      `gorm:"default:false" json:"replied"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OAuth user profile model for storing OAuth user data
type OAuthProfile struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Provider     string    `gorm:"size:50;not null" json:"provider"` // 'google', 'facebook', 'apple'
	ProviderID   string    `gorm:"size:255;not null" json:"provider_id"`
	Email        string    `gorm:"size:255" json:"email"`
	Name         string    `gorm:"size:255" json:"name"`
	Picture      string    `gorm:"size:500" json:"picture"`
	AccessToken  string    `gorm:"type:text" json:"access_token"`
	RefreshToken string    `gorm:"type:text" json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Product model for e-commerce products
type Product struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	Price         float64   `gorm:"not null" json:"price"`
	DiscountPrice float64   `json:"discount_price,omitempty"`
	Quantity      int       `gorm:"not null;default:0" json:"quantity"`
	SKU           string    `gorm:"size:100;unique;not null" json:"sku"`
	Category      string    `gorm:"size:255" json:"category"`
	Brand         string    `gorm:"size:255" json:"brand"`
	Images        string    `gorm:"type:text" json:"images"` // JSON array of image URLs
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	Weight        float64   `gorm:"default:0" json:"weight"`
	Tags          string    `gorm:"type:text" json:"tags"` // JSON array of tags
	UserID        uint      `gorm:"not null" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Favorite      bool      `json:"favorite,omitempty"`
}

// Enhanced Order model
type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"not null" json:"order_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"` // Price at time of order
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// To do model for task management
type Todo struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"size:500;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	IsCompleted bool       `gorm:"default:false" json:"is_completed"`
	Priority    string     `gorm:"size:50;default:'medium'" json:"priority"` // low, medium, high, urgent
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UserID      uint       `gorm:"not null" json:"user_id"`
	User        User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func GetMod() []interface{} {
	return []interface{}{
		&User{},
		&Role{},
		&Permission{},
		&Admin{},
		&Client{},
		&Package{},
		&Order{},
		&Subscription{},
		&RolePermission{},
		&UserRole{},
		&EmailVerification{},
		&PasswordReset{},
		&Blog{},
		&Newsletter{},
		&Contact{},
		&OAuthProfile{},
		&Product{},
		&OrderItem{},
		&Todo{},
	}
}
