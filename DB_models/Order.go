package dbmodels

import "time"

type Order struct {
	ID                uint        `gorm:"primaryKey"`
	ClientID          uint        `gorm:"not null"`
	Client            Client      `gorm:"foreignKey:ClientID"`
	UserID            uint        `gorm:"not null"`
	User              User        `gorm:"foreignKey:UserID"`
	Total             float64     `gorm:"not null"`
	Status            OrderStatus `gorm:"not null"` // Enum as int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Items             []OrderItem `gorm:"foreignKey:OrderID"`
	Address           string      `gorm:"type:text"` // Shipping address
	Phone             string      `gorm:"size:20"`   // Shipping phone
	Notes             string      `gorm:"type:text"` // Order notes
	PaymentStatus     string      `gorm:"size:50"`   // Payment status
	PaymentAmount     float64     `gorm:"default:0"` // Payment amount
	PaymentMethodId   int64       `gorm:"default:0"` // Payment method identifier
	PaymentMethodDesc string      `gorm:"type:text"` // Payment method description
	PaymentDate       *time.Time  // Date of payment
	PaymentRef        string      `gorm:"size:255"` // Payment reference number

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
