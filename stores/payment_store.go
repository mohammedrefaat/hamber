package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== PAYMENT MANAGEMENT ==========

func (store *DbStore) CreatePayment(payment *dbmodels.Payment) error {
	return store.db.Create(payment).Error
}

func (store *DbStore) GetPayment(id uint) (*dbmodels.Payment, error) {
	var payment dbmodels.Payment
	if err := store.db.Preload("User").Preload("Package").First(&payment, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Payment not found",
			Code:    http.StatusNotFound,
		}
	}
	return &payment, nil
}

func (store *DbStore) GetPaymentByReference(reference string) (*dbmodels.Payment, error) {
	var payment dbmodels.Payment
	if err := store.db.Preload("User").Preload("Package").
		Where("reference_number = ?", reference).First(&payment).Error; err != nil {
		return nil, &CustomError{
			Message: "Payment not found",
			Code:    http.StatusNotFound,
		}
	}
	return &payment, nil
}

func (store *DbStore) GetPaymentByTransactionID(transactionID string) (*dbmodels.Payment, error) {
	var payment dbmodels.Payment
	if err := store.db.Preload("User").Preload("Package").
		Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
		return nil, &CustomError{
			Message: "Payment not found",
			Code:    http.StatusNotFound,
		}
	}
	return &payment, nil
}

func (store *DbStore) UpdatePaymentStatus(id uint, status dbmodels.PaymentStatus, transactionID string) error {
	updates := map[string]interface{}{
		"payment_status": status,
	}

	if transactionID != "" {
		updates["transaction_id"] = transactionID
	}

	if status == dbmodels.PaymentStatus_PAID {
		now := time.Now()
		updates["paid_at"] = &now
	}

	return store.db.Model(&dbmodels.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (store *DbStore) GetUserPayments(userID uint, page, limit int) ([]dbmodels.Payment, int64, error) {
	var payments []dbmodels.Payment
	var total int64

	query := store.db.Model(&dbmodels.Payment{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count payments",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("Package").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch payments",
			Code:    http.StatusInternalServerError,
		}
	}

	return payments, total, nil
}

// ========== PACKAGE CHANGE MANAGEMENT ==========

func (store *DbStore) CreatePackageChange(change *dbmodels.PackageChange) error {
	return store.db.Create(change).Error
}

func (store *DbStore) GetPackageChange(id uint) (*dbmodels.PackageChange, error) {
	var change dbmodels.PackageChange
	if err := store.db.Preload("User").
		Preload("OldPackage").
		Preload("NewPackage").
		Preload("Payment").
		First(&change, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Package change not found",
			Code:    http.StatusNotFound,
		}
	}
	return &change, nil
}

func (store *DbStore) UpdatePackageChangeStatus(id uint, status dbmodels.ChangeStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == dbmodels.ChangeStatus_APPROVED || status == dbmodels.ChangeStatus_COMPLETED {
		now := time.Now()
		updates["approved_at"] = &now
	}

	return store.db.Model(&dbmodels.PackageChange{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (store *DbStore) CompletePackageChange(changeID uint, newPackageID uint) error {
	// Start transaction
	tx := store.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get package change
	var change dbmodels.PackageChange
	if err := tx.First(&change, changeID).Error; err != nil {
		tx.Rollback()
		return &CustomError{
			Message: "Package change not found",
			Code:    http.StatusNotFound,
		}
	}

	// Update user's package
	if err := tx.Model(&dbmodels.User{}).
		Where("id = ?", change.UserID).
		Update("package_id", newPackageID).Error; err != nil {
		tx.Rollback()
		return &CustomError{
			Message: "Failed to update user package",
			Code:    http.StatusInternalServerError,
		}
	}

	// Update package change status
	now := time.Now()
	if err := tx.Model(&dbmodels.PackageChange{}).
		Where("id = ?", changeID).
		Updates(map[string]interface{}{
			"status":      dbmodels.ChangeStatus_COMPLETED,
			"approved_at": &now,
		}).Error; err != nil {
		tx.Rollback()
		return &CustomError{
			Message: "Failed to update package change status",
			Code:    http.StatusInternalServerError,
		}
	}

	// Create subscription record
	pkg, err := store.GetPackage(newPackageID)
	if err != nil {
		tx.Rollback()
		return err
	}

	subscription := dbmodels.Subscription{
		UserID:    change.UserID,
		PackageID: newPackageID,
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, pkg.Duration),
		Price:     pkg.Price,
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		return &CustomError{
			Message: "Failed to create subscription",
			Code:    http.StatusInternalServerError,
		}
	}

	return tx.Commit().Error
}

func (store *DbStore) GetUserPackageChanges(userID uint, page, limit int) ([]dbmodels.PackageChange, int64, error) {
	var changes []dbmodels.PackageChange
	var total int64

	query := store.db.Model(&dbmodels.PackageChange{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count package changes",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("OldPackage").
		Preload("NewPackage").
		Preload("Payment").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&changes).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch package changes",
			Code:    http.StatusInternalServerError,
		}
	}

	return changes, total, nil
}
