package stores

import (
	"net/http"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== PRODUCT MANAGEMENT ==========

func (store *DbStore) CreateProduct(product *dbmodels.Product) error {
	var existingProduct dbmodels.Product
	if err := store.db.Where("sku = ?", product.SKU).First(&existingProduct).Error; err == nil {
		return &CustomError{
			Message: "Product with this SKU already exists",
			Code:    http.StatusConflict,
		}
	}

	return store.db.Create(product).Error
}

func (store *DbStore) GetProducts(page, limit int, userID uint, category string, isActive *bool) ([]dbmodels.Product, int64, error) {
	var products []dbmodels.Product
	var total int64

	query := store.db.Model(&dbmodels.Product{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count products",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&products).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch products",
			Code:    http.StatusInternalServerError,
		}
	}

	return products, total, nil
}

func (store *DbStore) GetProduct(id uint) (*dbmodels.Product, error) {
	var product dbmodels.Product
	if err := store.db.Preload("User").First(&product, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Product not found",
			Code:    http.StatusNotFound,
		}
	}
	return &product, nil
}

func (store *DbStore) UpdateProduct(product *dbmodels.Product) error {
	return store.db.Save(product).Error
}

func (store *DbStore) DeleteProduct(id uint) error {
	return store.db.Model(&dbmodels.Product{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (store *DbStore) UpdateProductQuantity(id uint, quantity int) error {
	if quantity < 0 {
		return &CustomError{
			Message: "Quantity cannot be negative",
			Code:    http.StatusBadRequest,
		}
	}

	return store.db.Model(&dbmodels.Product{}).
		Where("id = ?", id).
		Update("quantity", quantity).Error
}
