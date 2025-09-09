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
	if err := store.db.Preload("Role").Preload("Role.Permissions").First(&user, id).Error; err != nil {
		return nil, &CustomError{
			Message: "user not found",
			Code:    http.StatusNotFound,
		}
	}
	return &user, nil
}

// NEW: GetAllUsers - MISSING IMPLEMENTATION
func (store *DbStore) GetAllUsers(page, limit int) ([]dbmodels.User, int64, error) {
	var users []dbmodels.User
	var total int64

	// Get total count
	if err := store.db.Model(&dbmodels.User{}).Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count users",
			Code:    http.StatusInternalServerError,
		}
	}

	// Get paginated results with role information
	offset := (page - 1) * limit
	if err := store.db.Preload("Role").Preload("Package").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch users",
			Code:    http.StatusInternalServerError,
		}
	}

	return users, total, nil
}

// NEW: GetUserWithRole - Get user with role preloaded for JWT generation
func (store *DbStore) GetUserWithRole(id uint) (*dbmodels.User, error) {
	var user dbmodels.User
	if err := store.db.Preload("Role").Preload("Role.Permissions").First(&user, id).Error; err != nil {
		return nil, &CustomError{
			Message: "user not found",
			Code:    http.StatusNotFound,
		}
	}
	return &user, nil
}

// NEW: GetUserPermissions - Get user permissions by user ID
func (store *DbStore) GetUserPermissions(userID uint) ([]dbmodels.Permission, error) {
	var permissions []dbmodels.Permission

	// Get user with roles and permissions
	var user dbmodels.User
	if err := store.db.Preload("Role").Preload("Role.Permissions").First(&user, userID).Error; err != nil {
		return nil, &CustomError{
			Message: "User not found",
			Code:    http.StatusNotFound,
		}
	}

	// Collect all permissions from all roles
	permissionMap := make(map[uint]dbmodels.Permission)
	for _, role := range user.Role {
		for _, permission := range role.Permissions {
			permissionMap[permission.ID] = permission
		}
	}

	// Convert map to slice
	for _, permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// NEW: Role and Permission management methods
func (store *DbStore) CreateRole(role *dbmodels.Role) error {
	return store.db.Create(role).Error
}

func (store *DbStore) GetRole(id uint) (*dbmodels.Role, error) {
	var role dbmodels.Role
	if err := store.db.Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Role not found",
			Code:    http.StatusNotFound,
		}
	}
	return &role, nil
}

func (store *DbStore) GetAllRoles() ([]dbmodels.Role, error) {
	var roles []dbmodels.Role
	if err := store.db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch roles",
			Code:    http.StatusInternalServerError,
		}
	}
	return roles, nil
}

func (store *DbStore) CreatePermission(permission *dbmodels.Permission) error {
	return store.db.Create(permission).Error
}

func (store *DbStore) GetAllPermissions() ([]dbmodels.Permission, error) {
	var permissions []dbmodels.Permission
	if err := store.db.Find(&permissions).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch permissions",
			Code:    http.StatusInternalServerError,
		}
	}
	return permissions, nil
}

func (store *DbStore) AssignRoleToUser(userID, roleID uint) error {
	// Check if user exists
	var user dbmodels.User
	if err := store.db.First(&user, userID).Error; err != nil {
		return &CustomError{
			Message: "User not found",
			Code:    http.StatusNotFound,
		}
	}

	// Check if role exists
	var role dbmodels.Role
	if err := store.db.First(&role, roleID).Error; err != nil {
		return &CustomError{
			Message: "Role not found",
			Code:    http.StatusNotFound,
		}
	}

	// Add role to user
	return store.db.Model(&user).Association("Role").Append(&role)
}

func (store *DbStore) RemoveRoleFromUser(userID, roleID uint) error {
	var user dbmodels.User
	if err := store.db.First(&user, userID).Error; err != nil {
		return &CustomError{
			Message: "User not found",
			Code:    http.StatusNotFound,
		}
	}

	var role dbmodels.Role
	if err := store.db.First(&role, roleID).Error; err != nil {
		return &CustomError{
			Message: "Role not found",
			Code:    http.StatusNotFound,
		}
	}

	return store.db.Model(&user).Association("Role").Delete(&role)
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
	if err := store.db.Preload("Role").Preload("Role.Permissions").Where("email = ?", email).First(&user).Error; err != nil {
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
	if err := store.db.Preload("Role").Preload("Role.Permissions").Where("email = ?", email).First(&user).Error; err != nil {
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

// ... (rest of the original methods remain the same)
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

// Blog methods (keeping all original methods)
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

// Newsletter methods (keeping all original methods)
func (store *DbStore) CreateNewsletter(newsletter *dbmodels.Newsletter) error {
	newsletter.SubscribedAt = time.Now()
	return store.db.Create(newsletter).Error
}

func (store *DbStore) GetNewsletterByEmail(email string) (*dbmodels.Newsletter, error) {
	var newsletter dbmodels.Newsletter
	if err := store.db.Where("email = ?", email).First(&newsletter).Error; err != nil {
		return nil, err
	}
	return &newsletter, nil
}

func (store *DbStore) UpdateNewsletter(newsletter *dbmodels.Newsletter) error {
	return store.db.Save(newsletter).Error
}

func (store *DbStore) UnsubscribeNewsletter(email string) error {
	now := time.Now()
	return store.db.Model(&dbmodels.Newsletter{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"is_active":       false,
			"unsubscribed_at": &now,
		}).Error
}

func (store *DbStore) GetNewsletterSubscriptions(page, limit int, isActive *bool) ([]dbmodels.Newsletter, int64, error) {
	var newsletters []dbmodels.Newsletter
	var total int64

	query := store.db.Model(&dbmodels.Newsletter{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count newsletter subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&newsletters).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch newsletter subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	return newsletters, total, nil
}

// Contact methods (keeping all original methods)
func (store *DbStore) CreateContact(contact *dbmodels.Contact) error {
	return store.db.Create(contact).Error
}

func (store *DbStore) GetContacts(page, limit int, unreadOnly bool) ([]dbmodels.Contact, int64, error) {
	var contacts []dbmodels.Contact
	var total int64

	query := store.db.Model(&dbmodels.Contact{})
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&contacts).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	return contacts, total, nil
}

func (store *DbStore) GetContact(id uint) (*dbmodels.Contact, error) {
	var contact dbmodels.Contact
	if err := store.db.First(&contact, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Contact not found",
			Code:    http.StatusNotFound,
		}
	}
	return &contact, nil
}

func (store *DbStore) MarkContactAsRead(id uint) error {
	return store.db.Model(&dbmodels.Contact{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

func (store *DbStore) MarkContactAsReplied(id uint) error {
	return store.db.Model(&dbmodels.Contact{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"replied": true,
			"is_read": true,
		}).Error
}

func (store *DbStore) DeleteContact(id uint) error {
	return store.db.Delete(&dbmodels.Contact{}, id).Error
}

// OAuth methods (keeping all original methods)
func (store *DbStore) CreateOAuthProfile(profile *dbmodels.OAuthProfile) error {
	return store.db.Create(profile).Error
}

func (store *DbStore) GetOAuthProfile(provider, providerID string) (*dbmodels.OAuthProfile, error) {
	var profile dbmodels.OAuthProfile
	if err := store.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (store *DbStore) UpdateOAuthProfile(profile *dbmodels.OAuthProfile) error {
	return store.db.Save(profile).Error
}

func (store *DbStore) GetOAuthProfilesByUser(userID uint) ([]dbmodels.OAuthProfile, error) {
	var profiles []dbmodels.OAuthProfile
	if err := store.db.Where("user_id = ?", userID).Find(&profiles).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch OAuth profiles",
			Code:    http.StatusInternalServerError,
		}
	}
	return profiles, nil
}

func (store *DbStore) DeleteOAuthProfile(id uint) error {
	return store.db.Delete(&dbmodels.OAuthProfile{}, id).Error
}

// Statistics methods (keeping all original methods)
type NewsletterStats struct {
	TotalSubscriptions     int64 `json:"total_subscriptions"`
	ActiveSubscriptions    int64 `json:"active_subscriptions"`
	InactiveSubscriptions  int64 `json:"inactive_subscriptions"`
	SubscriptionsToday     int64 `json:"subscriptions_today"`
	SubscriptionsThisWeek  int64 `json:"subscriptions_this_week"`
	SubscriptionsThisMonth int64 `json:"subscriptions_this_month"`
}

func (store *DbStore) GetNewsletterStats() (*NewsletterStats, error) {
	var stats NewsletterStats

	// Total subscriptions
	if err := store.db.Model(&dbmodels.Newsletter{}).Count(&stats.TotalSubscriptions).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count total subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	// Active subscriptions
	if err := store.db.Model(&dbmodels.Newsletter{}).Where("is_active = ?", true).Count(&stats.ActiveSubscriptions).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count active subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	// Inactive subscriptions
	stats.InactiveSubscriptions = stats.TotalSubscriptions - stats.ActiveSubscriptions

	// Subscriptions today
	today := time.Now().Truncate(24 * time.Hour)
	if err := store.db.Model(&dbmodels.Newsletter{}).Where("subscribed_at >= ?", today).Count(&stats.SubscriptionsToday).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count today's subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	// Subscriptions this week
	startOfWeek := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	startOfWeek = startOfWeek.Truncate(24 * time.Hour)
	if err := store.db.Model(&dbmodels.Newsletter{}).Where("subscribed_at >= ?", startOfWeek).Count(&stats.SubscriptionsThisWeek).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count this week's subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	// Subscriptions this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if err := store.db.Model(&dbmodels.Newsletter{}).Where("subscribed_at >= ?", startOfMonth).Count(&stats.SubscriptionsThisMonth).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count this month's subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	return &stats, nil
}

type ContactStats struct {
	TotalContacts     int64 `json:"total_contacts"`
	UnreadContacts    int64 `json:"unread_contacts"`
	RepliedContacts   int64 `json:"replied_contacts"`
	ContactsToday     int64 `json:"contacts_today"`
	ContactsThisWeek  int64 `json:"contacts_this_week"`
	ContactsThisMonth int64 `json:"contacts_this_month"`
}

func (store *DbStore) GetContactStats() (*ContactStats, error) {
	var stats ContactStats

	// Total contacts
	if err := store.db.Model(&dbmodels.Contact{}).Count(&stats.TotalContacts).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count total contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	// Unread contacts
	if err := store.db.Model(&dbmodels.Contact{}).Where("is_read = ?", false).Count(&stats.UnreadContacts).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count unread contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	// Replied contacts
	if err := store.db.Model(&dbmodels.Contact{}).Where("replied = ?", true).Count(&stats.RepliedContacts).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count replied contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	// Contacts today
	today := time.Now().Truncate(24 * time.Hour)
	if err := store.db.Model(&dbmodels.Contact{}).Where("created_at >= ?", today).Count(&stats.ContactsToday).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count today's contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	// Contacts this week
	startOfWeek := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	startOfWeek = startOfWeek.Truncate(24 * time.Hour)
	if err := store.db.Model(&dbmodels.Contact{}).Where("created_at >= ?", startOfWeek).Count(&stats.ContactsThisWeek).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count this week's contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	// Contacts this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if err := store.db.Model(&dbmodels.Contact{}).Where("created_at >= ?", startOfMonth).Count(&stats.ContactsThisMonth).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to count this month's contacts",
			Code:    http.StatusInternalServerError,
		}
	}

	return &stats, nil
}
