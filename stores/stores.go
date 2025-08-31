package stores

import (
	"net/http"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	tools "github.com/mohammedrefaat/hamber/Tools"
	"gorm.io/gorm"
)

// CustomError struct
type CustomError struct {
	Message string
	Code    int
}

func (e *CustomError) Error() string {
	return e.Message
}

// DbStore struct
type DbStore struct {
	db *gorm.DB
}

// NewDbStore initializes a new DbStore
func NewDbStore(db *gorm.DB) (*DbStore, error) {
	err := dbmodels.Migrator(db)
	return nil, err
}

// CreateUser inserts a new user into the database
func (store *DbStore) CreateUser(user *dbmodels.User) error {
	// Validate email
	if !tools.ValidateEmail(&user.Email) {
		return &CustomError{
			Message: "البريد الالكتروني غير صحيح",
			Code:    http.StatusBadRequest,
		}
	}

	// Check if email already exists
	var existingUser dbmodels.User
	if err := store.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return &CustomError{
			Message: "email already exists",
			Code:    http.StatusConflict, // 409 Conflict
		}
	}

	// Check if username already exists
	if err := store.db.Where("name = ?", user.Name).First(&existingUser).Error; err == nil {
		return &CustomError{
			Message: "username already exists",
			Code:    http.StatusConflict, // 409 Conflict
		}
	}

	return store.db.Create(user).Error
}

// GetUser retrieves a user by ID
func (store *DbStore) GetUser(id uint) (*dbmodels.User, error) {
	var user dbmodels.User
	if err := store.db.First(&user, id).Error; err != nil {
		return nil, &CustomError{
			Message: "user not found",
			Code:    http.StatusNotFound,
		}
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (store *DbStore) UpdateUser(user *dbmodels.User) error {
	return store.db.Save(user).Error
}

// DeleteUser deletes a user by ID
func (store *DbStore) DeleteUser(id uint) error {
	return store.db.Delete(&dbmodels.User{}, id).Error
}

// Login validates the user's credentials
func (store *DbStore) Login(email, password string) (*dbmodels.User, error) {
	var user dbmodels.User
	if err := store.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, &CustomError{
			Message: "invalid email or password",
			Code:    http.StatusUnauthorized, // 401 Unauthorized
		}
	}

	// Here you should hash and compare the password; this is a simple comparison
	if user.Password != password {
		return nil, &CustomError{
			Message: "invalid email or password",
			Code:    http.StatusUnauthorized, // 401 Unauthorized
		}
	}

	return &user, nil
}
