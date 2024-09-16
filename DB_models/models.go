package dbmodels

import (
	"time"

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
	}
}

// Modify Migrator to use the correct type
func Migrator(db *gorm.DB) error {
	// Now modelsToMigrate is a slice of interfaces, so we can range over it
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

type User struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"size:255;not null"`
	Email     string  `gorm:"size:255;unique;not null"`
	Password  string  `gorm:"not null"`
	Subdomain string  `gorm:"size:255;unique;not null"`
	RoleID    uint    `gorm:"not null"`              // Foreign key to the Role table
	Role      []Role  `gorm:"many2many:user_roles;"` // Many-to-many relationship between users and roles
	PackageID uint    `gorm:"not null"`
	Package   Package `gorm:"foreignKey:PackageID"`
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

type Package struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"size:255;not null"`
	Price     float64 `gorm:"not null"`
	Duration  int     `gorm:"not null"` // In days or months
	CreatedAt time.Time
	UpdatedAt time.Time
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
