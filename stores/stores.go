package stores

import (
	"errors"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	tools "github.com/mohammedrefaat/hamber/Tools"
	"gorm.io/gorm"
)

type DbStore struct {
	db *gorm.DB
	//Router *gin.Engine
}

func NewDbStore(db *gorm.DB) (*DbStore, error) {
	err := dbmodels.Migrator(db)
	if err != nil {
		return nil, err
	}
	return &DbStore{db: db}, nil
}

// CreateUser inserts a new user into the database
func (store *DbStore) CreateUser(user *dbmodels.User) error {
	// Validate email
	if !tools.ValidateEmail(&user.Email) {
		return errors.New(" البريد الالكتروني غير صحيح")

	}

	// Check if email already exists
	var existingUser dbmodels.User
	if err := store.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return errors.New("email already exists")
	}

	// Check if username already exists
	if err := store.db.Where("name = ?", user.Name).First(&existingUser).Error; err == nil {
		return errors.New("username already exists")
	}

	return store.db.Create(user).Error
}

// GetUser retrieves a user by ID
func (store *DbStore) GetUser(id uint) (*dbmodels.User, error) {
	var user dbmodels.User
	if err := store.db.First(&user, id).Error; err != nil {
		return nil, err
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
		return nil, errors.New("invalid email or password")
	}

	// Here you should hash and compare the password; this is a simple comparison
	if user.Password != password {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}
