package dbmodels

import "time"

// ========== ADD-ONS SYSTEM ==========

// Addon represents an add-on service/feature that users can subscribe to
type Addon struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Title        string    `gorm:"size:255;not null" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	Logo         string    `gorm:"size:500" json:"logo"`                 // URL to logo image
	Photo        string    `gorm:"size:500" json:"photo"`                // Main photo URL
	Category     string    `gorm:"size:100" json:"category"`             // e.g., "AI", "Integration", "Marketing"
	PricingType  string    `gorm:"size:50;not null" json:"pricing_type"` // "time" or "usage"
	BasePrice    float64   `gorm:"not null" json:"base_price"`           // Base price
	Currency     string    `gorm:"size:10;default:'EGP'" json:"currency"`
	BillingCycle int       `gorm:"default:30" json:"billing_cycle"` // Days for time-based
	UsageUnit    string    `gorm:"size:50" json:"usage_unit"`       // e.g., "requests", "messages", "credits"
	Features     string    `gorm:"type:text" json:"features"`       // JSON array of features
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AddonPricingTier represents volume/duration discounts
type AddonPricingTier struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AddonID       uint      `gorm:"not null" json:"addon_id"`
	Addon         Addon     `gorm:"foreignKey:AddonID" json:"addon,omitempty"`
	MinQuantity   int       `gorm:"not null" json:"min_quantity"`          // Min units/months
	MaxQuantity   int       `json:"max_quantity"`                          // Max units/months (0 = unlimited)
	DiscountType  string    `gorm:"size:20;not null" json:"discount_type"` // "percentage" or "fixed"
	DiscountValue float64   `gorm:"not null" json:"discount_value"`
	FinalPrice    float64   `gorm:"not null" json:"final_price"` // Calculated price after discount
	Description   string    `gorm:"size:255" json:"description"` // e.g., "3 months - 10% off"
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserAddonSubscription represents a user's subscription to an add-on
type UserAddonSubscription struct {
	ID            uint                    `gorm:"primaryKey" json:"id"`
	UserID        uint                    `gorm:"not null" json:"user_id"`
	User          User                    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	AddonID       uint                    `gorm:"not null" json:"addon_id"`
	Addon         Addon                   `gorm:"foreignKey:AddonID" json:"addon,omitempty"`
	PricingTierID *uint                   `json:"pricing_tier_id,omitempty"`
	PricingTier   *AddonPricingTier       `gorm:"foreignKey:PricingTierID" json:"pricing_tier,omitempty"`
	Status        AddonSubscriptionStatus `gorm:"not null;default:0" json:"status"`
	Quantity      int                     `gorm:"not null" json:"quantity"` // Months or units purchased
	TotalPrice    float64                 `gorm:"not null" json:"total_price"`
	StartDate     time.Time               `gorm:"not null" json:"start_date"`
	EndDate       *time.Time              `json:"end_date,omitempty"`    // For time-based
	UsageLimit    *int                    `json:"usage_limit,omitempty"` // For usage-based
	UsageCount    int                     `gorm:"default:0" json:"usage_count"`
	AutoRenew     bool                    `gorm:"default:false" json:"auto_renew"`
	PaymentID     *uint                   `json:"payment_id,omitempty"`
	Payment       *Payment                `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

type AddonSubscriptionStatus int32

const (
	AddonSubscriptionStatus_PENDING   AddonSubscriptionStatus = 0
	AddonSubscriptionStatus_ACTIVE    AddonSubscriptionStatus = 1
	AddonSubscriptionStatus_EXPIRED   AddonSubscriptionStatus = 2
	AddonSubscriptionStatus_CANCELLED AddonSubscriptionStatus = 3
	AddonSubscriptionStatus_SUSPENDED AddonSubscriptionStatus = 4
)

var AddonSubscriptionStatus_name = map[int32]string{
	0: "PENDING",
	1: "ACTIVE",
	2: "EXPIRED",
	3: "CANCELLED",
	4: "SUSPENDED",
}

func (x AddonSubscriptionStatus) String() string {
	return AddonSubscriptionStatus_name[int32(x)]
}

// AddonUsageLog tracks usage for usage-based add-ons
type AddonUsageLog struct {
	ID             uint                  `gorm:"primaryKey" json:"id"`
	SubscriptionID uint                  `gorm:"not null" json:"subscription_id"`
	Subscription   UserAddonSubscription `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
	UsageAmount    int                   `gorm:"not null" json:"usage_amount"`
	Description    string                `gorm:"size:500" json:"description"`
	Metadata       string                `gorm:"type:text" json:"metadata"` // JSON for additional data
	CreatedAt      time.Time             `json:"created_at"`
}
