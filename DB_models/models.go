package dbmodels

import (
	"time"

	"gorm.io/gorm"
)

type Models struct {
	User                  User
	Role                  Role
	Permission            Permission
	Admin                 Admin
	Client                Client
	Package               Package
	Order                 Order
	Subscription          Subscription
	RolePermission        RolePermission
	UserRole              UserRole
	EmailVerification     EmailVerification
	PasswordReset         PasswordReset
	Blog                  Blog
	Newsletter            Newsletter
	Contact               Contact
	OAuthProfile          OAuthProfile
	Product               Product
	OrderItem             OrderItem
	Todo                  Todo
	Payment               Payment
	PackageChange         PackageChange
	Addon                 Addon
	AddonPricingTier      AddonPricingTier
	UserAddonSubscription UserAddonSubscription
	OrderReceipt          OrderReceipt
	CalendarEvent         CalendarEvent
	EventAttendee         EventAttendee
	AddonUsageLog         AddonUsageLog
	Notification          Notification
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

// Payment model for tracking payment transactions
type Payment struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"not null" json:"user_id"`
	User            User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PackageID       uint          `gorm:"not null" json:"package_id"`
	Package         Package       `gorm:"foreignKey:PackageID" json:"package,omitempty"`
	Amount          float64       `gorm:"not null" json:"amount"`
	Currency        string        `gorm:"size:10;default:'EGP'" json:"currency"`
	PaymentMethod   string        `gorm:"size:50;not null" json:"payment_method"` // 'fawry', 'paymob'
	PaymentStatus   PaymentStatus `gorm:"not null;default:0" json:"payment_status"`
	TransactionID   string        `gorm:"size:255" json:"transaction_id"`
	ReferenceNumber string        `gorm:"size:255;unique" json:"reference_number"` // Fawry reference
	PaymobOrderID   string        `gorm:"size:255" json:"paymob_order_id"`
	PaymentData     string        `gorm:"type:text" json:"payment_data"` // JSON for additional data
	ExpiresAt       *time.Time    `json:"expires_at,omitempty"`
	PaidAt          *time.Time    `json:"paid_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type PaymentStatus int32

const (
	PaymentStatus_PENDING   PaymentStatus = 0
	PaymentStatus_PAID      PaymentStatus = 1
	PaymentStatus_FAILED    PaymentStatus = 2
	PaymentStatus_CANCELLED PaymentStatus = 3
	PaymentStatus_EXPIRED   PaymentStatus = 4
	PaymentStatus_REFUNDED  PaymentStatus = 5
)

var (
	PaymentStatus_name = map[int32]string{
		0: "PENDING",
		1: "PAID",
		2: "FAILED",
		3: "CANCELLED",
		4: "EXPIRED",
		5: "REFUNDED",
	}
	PaymentStatus_value = map[string]int32{
		"PENDING":   0,
		"PAID":      1,
		"FAILED":    2,
		"CANCELLED": 3,
		"EXPIRED":   4,
		"REFUNDED":  5,
	}
)

func (x PaymentStatus) String() string {
	return PaymentStatus_name[int32(x)]
}

// PackageChange model to track package upgrade/downgrade requests
type PackageChange struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	UserID       uint         `gorm:"not null" json:"user_id"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	OldPackageID uint         `gorm:"not null" json:"old_package_id"`
	OldPackage   Package      `gorm:"foreignKey:OldPackageID" json:"old_package,omitempty"`
	NewPackageID uint         `gorm:"not null" json:"new_package_id"`
	NewPackage   Package      `gorm:"foreignKey:NewPackageID" json:"new_package,omitempty"`
	PaymentID    *uint        `json:"payment_id,omitempty"`
	Payment      *Payment     `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
	Status       ChangeStatus `gorm:"not null;default:0" json:"status"`
	ChangeReason string       `gorm:"type:text" json:"change_reason,omitempty"`
	ApprovedAt   *time.Time   `json:"approved_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type ChangeStatus int32

const (
	ChangeStatus_PENDING   ChangeStatus = 0
	ChangeStatus_APPROVED  ChangeStatus = 1
	ChangeStatus_REJECTED  ChangeStatus = 2
	ChangeStatus_COMPLETED ChangeStatus = 3
)

var (
	ChangeStatus_name = map[int32]string{
		0: "PENDING",
		1: "APPROVED",
		2: "REJECTED",
		3: "COMPLETED",
	}
)

func (x ChangeStatus) String() string {
	return ChangeStatus_name[int32(x)]
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
		&Payment{},
		&PackageChange{},
		&Addon{},
		&AddonPricingTier{},
		&UserAddonSubscription{},
		&AddonUsageLog{},
		&CalendarEvent{},
		&EventAttendee{},
		&OrderReceipt{},
		&Notification{},
		&Message{},
		&MessageFolder{},
		&MessageLabel{},
		&Banner{},
		&BannerView{},
		&BannerClick{},
	}
}
