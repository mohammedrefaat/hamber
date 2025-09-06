package stores

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

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
	if err != nil {
		return nil, err
	}
	return &DbStore{db: db}, nil
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
	if err := user.HashPassword(user.Password); err != nil {
		return &CustomError{
			Message: "Failed to hash password",
			Code:    http.StatusInternalServerError,
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
	if !user.CheckPassword(password) {
		return nil, &CustomError{
			Message: "invalid email or password",
			Code:    http.StatusUnauthorized, // 401 Unauthorized
		}
	}

	return &user, nil
}

// Package related methods
func (store *DbStore) GetAllPackages() ([]dbmodels.Package, error) {
	var packages []dbmodels.Package
	if err := store.db.Where("is_active = ?", true).Find(&packages).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch packages",
			Code:    http.StatusInternalServerError,
		}
	}
	return packages, nil
}

func (store *DbStore) GetPackage(id uint) (*dbmodels.Package, error) {
	var pkg dbmodels.Package
	if err := store.db.Where("id = ? AND is_active = ?", id, true).First(&pkg).Error; err != nil {
		return nil, &CustomError{
			Message: "Package not found",
			Code:    http.StatusNotFound,
		}
	}
	return &pkg, nil
}

// User related methods
func (store *DbStore) GetUserByEmail(email string) (*dbmodels.User, error) {
	var user dbmodels.User
	if err := store.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, &CustomError{
			Message: "User not found",
			Code:    http.StatusNotFound,
		}
	}
	return &user, nil
}

func (store *DbStore) MarkEmailAsVerified(email string) error {
	return store.db.Model(&dbmodels.User{}).
		Where("email = ?", email).
		Update("is_email_verified", true).Error
}

func (store *DbStore) ResetPassword(email, newPassword string) error {
	var user dbmodels.User
	if err := store.db.Where("email = ?", email).First(&user).Error; err != nil {
		return &CustomError{
			Message: "User not found",
			Code:    http.StatusNotFound,
		}
	}

	if err := user.HashPassword(newPassword); err != nil {
		return &CustomError{
			Message: "Failed to hash password",
			Code:    http.StatusInternalServerError,
		}
	}

	return store.db.Save(&user).Error
}

// Email verification methods
func (store *DbStore) CreateEmailVerification(email string) error {
	// Generate 6-digit code
	code, err := generateVerificationCode()
	if err != nil {
		return &CustomError{
			Message: "Failed to generate verification code",
			Code:    http.StatusInternalServerError,
		}
	}

	// Delete any existing verification codes for this email
	store.db.Where("email = ?", email).Delete(&dbmodels.EmailVerification{})

	verification := dbmodels.EmailVerification{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 minutes expiry
		Used:      false,
	}

	if err := store.db.Create(&verification).Error; err != nil {
		return &CustomError{
			Message: "Failed to create verification record",
			Code:    http.StatusInternalServerError,
		}
	}

	// TODO: Send actual email here
	// For now, just log the code (remove in production)
	fmt.Printf("Verification code for %s: %s\n", email, code)

	return nil
}

func (store *DbStore) VerifyEmailCode(email, code string) (bool, error) {
	var verification dbmodels.EmailVerification
	err := store.db.Where("email = ? AND code = ? AND used = ? AND expires_at > ?",
		email, code, false, time.Now()).First(&verification).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, &CustomError{
			Message: "Database error",
			Code:    http.StatusInternalServerError,
		}
	}

	// Mark as used
	verification.Used = true
	store.db.Save(&verification)

	return true, nil
}

// Password reset methods
func (store *DbStore) CreatePasswordReset(email string) error {
	// Generate 6-digit code
	code, err := generateVerificationCode()
	if err != nil {
		return &CustomError{
			Message: "Failed to generate reset code",
			Code:    http.StatusInternalServerError,
		}
	}

	// Delete any existing reset codes for this email
	store.db.Where("email = ?", email).Delete(&dbmodels.PasswordReset{})

	reset := dbmodels.PasswordReset{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 minutes expiry
		Used:      false,
	}

	if err := store.db.Create(&reset).Error; err != nil {
		return &CustomError{
			Message: "Failed to create reset record",
			Code:    http.StatusInternalServerError,
		}
	}

	// TODO: Send actual email here
	// For now, just log the code (remove in production)
	fmt.Printf("Password reset code for %s: %s\n", email, code)

	return nil
}

func (store *DbStore) VerifyPasswordResetCode(email, code string) (bool, error) {
	var reset dbmodels.PasswordReset
	err := store.db.Where("email = ? AND code = ? AND used = ? AND expires_at > ?",
		email, code, false, time.Now()).First(&reset).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, &CustomError{
			Message: "Database error",
			Code:    http.StatusInternalServerError,
		}
	}

	// Mark as used
	reset.Used = true
	store.db.Save(&reset)

	return true, nil
}

// Blog methods
func (store *DbStore) CreateBlog(blog *dbmodels.Blog) error {
	// Check if slug already exists
	var existingBlog dbmodels.Blog
	if err := store.db.Where("slug = ?", blog.Slug).First(&existingBlog).Error; err == nil {
		return &CustomError{
			Message: "Blog with this slug already exists",
			Code:    http.StatusConflict,
		}
	}

	return store.db.Create(blog).Error
}

func (store *DbStore) GetBlogs(page, limit int, publishedOnly bool) ([]dbmodels.Blog, int64, error) {
	var blogs []dbmodels.Blog
	var total int64

	query := store.db.Model(&dbmodels.Blog{})
	if publishedOnly {
		query = query.Where("is_published = ?", true)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	// Get paginated results with author information
	offset := (page - 1) * limit
	if err := query.Preload("Author").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&blogs).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	return blogs, total, nil
}

func (store *DbStore) GetBlog(id uint) (*dbmodels.Blog, error) {
	var blog dbmodels.Blog
	if err := store.db.Preload("Author").First(&blog, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Blog not found",
			Code:    http.StatusNotFound,
		}
	}
	return &blog, nil
}

func (store *DbStore) GetBlogBySlug(slug string) (*dbmodels.Blog, error) {
	var blog dbmodels.Blog
	if err := store.db.Preload("Author").Where("slug = ?", slug).First(&blog).Error; err != nil {
		return nil, &CustomError{
			Message: "Blog not found",
			Code:    http.StatusNotFound,
		}
	}
	return &blog, nil
}

func (store *DbStore) UpdateBlog(blog *dbmodels.Blog) error {
	// Check if slug already exists for other blogs
	var existingBlog dbmodels.Blog
	if err := store.db.Where("slug = ? AND id != ?", blog.Slug, blog.ID).First(&existingBlog).Error; err == nil {
		return &CustomError{
			Message: "Another blog with this slug already exists",
			Code:    http.StatusConflict,
		}
	}

	return store.db.Save(blog).Error
}

func (store *DbStore) DeleteBlog(id uint) error {
	return store.db.Delete(&dbmodels.Blog{}, id).Error
}

func (store *DbStore) GetBlogsByAuthor(authorID uint, page, limit int) ([]dbmodels.Blog, int64, error) {
	var blogs []dbmodels.Blog
	var total int64

	query := store.db.Model(&dbmodels.Blog{}).Where("author_id = ?", authorID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Preload("Author").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&blogs).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	return blogs, total, nil
}

// Helper function to generate verification code
func generateVerificationCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

type BlogAnalytics struct {
	TotalBlogs       int64 `json:"total_blogs"`
	PublishedBlogs   int64 `json:"published_blogs"`
	UnpublishedBlogs int64 `json:"unpublished_blogs"`
	TotalAuthors     int64 `json:"total_authors"`
	BlogsThisMonth   int64 `json:"blogs_this_month"`
	BlogsThisWeek    int64 `json:"blogs_this_week"`
}

func (store *DbStore) GetBlogAnalytics() (*BlogAnalytics, error) {
	var analytics BlogAnalytics

	// Total blogs
	if err := store.db.Model(&dbmodels.Blog{}).Count(&analytics.TotalBlogs).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count total blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	// Published blogs
	if err := store.db.Model(&dbmodels.Blog{}).Where("is_published = ?", true).Count(&analytics.PublishedBlogs).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count published blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	// Unpublished blogs
	analytics.UnpublishedBlogs = analytics.TotalBlogs - analytics.PublishedBlogs

	// Total unique authors
	if err := store.db.Model(&dbmodels.Blog{}).Distinct("author_id").Count(&analytics.TotalAuthors).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count authors",
			Code:    http.StatusInternalServerError,
		}
	}

	// Blogs this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if err := store.db.Model(&dbmodels.Blog{}).Where("created_at >= ?", startOfMonth).Count(&analytics.BlogsThisMonth).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count monthly blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	// Blogs this week
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	if err := store.db.Model(&dbmodels.Blog{}).Where("created_at >= ?", startOfWeek).Count(&analytics.BlogsThisWeek).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count weekly blogs",
			Code:    http.StatusInternalServerError,
		}
	}

	return &analytics, nil
}
