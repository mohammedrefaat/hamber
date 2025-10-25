package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== ADD-ON MANAGEMENT ==========

func (store *DbStore) CreateAddon(addon *dbmodels.Addon) error {
	return store.db.Create(addon).Error
}

func (store *DbStore) GetAddons(page, limit int, category string, isActive *bool) ([]dbmodels.Addon, int64, error) {
	var addons []dbmodels.Addon
	var total int64

	query := store.db.Model(&dbmodels.Addon{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count add-ons",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&addons).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch add-ons",
			Code:    http.StatusInternalServerError,
		}
	}

	return addons, total, nil
}

func (store *DbStore) GetAddon(id uint) (*dbmodels.Addon, error) {
	var addon dbmodels.Addon
	if err := store.db.First(&addon, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Add-on not found",
			Code:    http.StatusNotFound,
		}
	}
	return &addon, nil
}

func (store *DbStore) UpdateAddon(addon *dbmodels.Addon) error {
	return store.db.Save(addon).Error
}

func (store *DbStore) DeleteAddon(id uint) error {
	return store.db.Model(&dbmodels.Addon{}).Where("id = ?", id).Update("is_active", false).Error
}

// ========== PRICING TIER MANAGEMENT ==========

func (store *DbStore) CreatePricingTier(tier *dbmodels.AddonPricingTier) error {
	return store.db.Create(tier).Error
}

func (store *DbStore) GetPricingTiers(addonID uint) ([]dbmodels.AddonPricingTier, error) {
	var tiers []dbmodels.AddonPricingTier
	if err := store.db.Where("addon_id = ?", addonID).Order("min_quantity ASC").Find(&tiers).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch pricing tiers",
			Code:    http.StatusInternalServerError,
		}
	}
	return tiers, nil
}

func (store *DbStore) UpdatePricingTier(tier *dbmodels.AddonPricingTier) error {
	return store.db.Save(tier).Error
}

func (store *DbStore) DeletePricingTier(id uint) error {
	return store.db.Delete(&dbmodels.AddonPricingTier{}, id).Error
}

// ========== USER ADDON SUBSCRIPTION ==========

func (store *DbStore) CreateAddonSubscription(subscription *dbmodels.UserAddonSubscription) error {
	return store.db.Create(subscription).Error
}

func (store *DbStore) GetUserAddonSubscriptions(userID uint, status *dbmodels.AddonSubscriptionStatus) ([]dbmodels.UserAddonSubscription, error) {
	var subscriptions []dbmodels.UserAddonSubscription
	query := store.db.Preload("Addon").Preload("PricingTier").Where("user_id = ?", userID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("created_at DESC").Find(&subscriptions).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}
	return subscriptions, nil
}

func (store *DbStore) GetAddonSubscription(id uint) (*dbmodels.UserAddonSubscription, error) {
	var subscription dbmodels.UserAddonSubscription
	if err := store.db.Preload("Addon").Preload("PricingTier").Preload("Payment").First(&subscription, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Subscription not found",
			Code:    http.StatusNotFound,
		}
	}
	return &subscription, nil
}

func (store *DbStore) UpdateAddonSubscription(subscription *dbmodels.UserAddonSubscription) error {
	return store.db.Save(subscription).Error
}

func (store *DbStore) CancelAddonSubscription(id uint) error {
	return store.db.Model(&dbmodels.UserAddonSubscription{}).
		Where("id = ?", id).
		Update("status", dbmodels.AddonSubscriptionStatus_CANCELLED).Error
}

func (store *DbStore) CheckExpiredSubscriptions() error {
	now := time.Now()
	return store.db.Model(&dbmodels.UserAddonSubscription{}).
		Where("status = ? AND end_date IS NOT NULL AND end_date < ?",
			dbmodels.AddonSubscriptionStatus_ACTIVE, now).
		Update("status", dbmodels.AddonSubscriptionStatus_EXPIRED).Error
}

// ========== ADDON USAGE TRACKING ==========

func (store *DbStore) LogAddonUsage(log *dbmodels.AddonUsageLog) error {
	// Start transaction
	tx := store.db.Begin()

	// Create usage log
	if err := tx.Create(log).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update subscription usage count
	if err := tx.Model(&dbmodels.UserAddonSubscription{}).
		Where("id = ?", log.SubscriptionID).
		UpdateColumn("usage_count", tx.Raw("usage_count + ?", log.UsageAmount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (store *DbStore) GetAddonUsageLogs(subscriptionID uint, page, limit int) ([]dbmodels.AddonUsageLog, int64, error) {
	var logs []dbmodels.AddonUsageLog
	var total int64

	query := store.db.Model(&dbmodels.AddonUsageLog{}).Where("subscription_id = ?", subscriptionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count usage logs",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch usage logs",
			Code:    http.StatusInternalServerError,
		}
	}

	return logs, total, nil
}

func (store *DbStore) GetAddonSubscriptionsByAddon(addonID uint, page, limit int) ([]dbmodels.UserAddonSubscription, int64, error) {
	var subscriptions []dbmodels.UserAddonSubscription
	var total int64

	query := store.db.Model(&dbmodels.UserAddonSubscription{}).Where("addon_id = ?", addonID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("PricingTier").
		Offset(offset).Limit(limit).Order("created_at DESC").
		Find(&subscriptions).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch subscriptions",
			Code:    http.StatusInternalServerError,
		}
	}

	return subscriptions, total, nil
}
