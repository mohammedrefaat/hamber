package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== ORDER MANAGEMENT ==========

func (store *DbStore) CreateOrder(order *dbmodels.Order) error {
	return store.db.Create(order).Error
}

func (store *DbStore) GetOrders(page, limit int, userID uint, status *dbmodels.OrderStatus) ([]dbmodels.Order, int64, error) {
	var orders []dbmodels.Order
	var total int64

	query := store.db.Model(&dbmodels.Order{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to count orders",
			Code:    http.StatusInternalServerError,
		}
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("Client").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, 0, &CustomError{
			Message: "Failed to fetch orders",
			Code:    http.StatusInternalServerError,
		}
	}

	return orders, total, nil
}

func (store *DbStore) GetOrderByID(id uint) (*dbmodels.Order, error) {
	var order dbmodels.Order
	if err := store.db.Preload("User").Preload("Client").First(&order, id).Error; err != nil {
		return nil, &CustomError{
			Message: "Order not found",
			Code:    http.StatusNotFound,
		}
	}
	return &order, nil
}

func (store *DbStore) UpdateOrderStatus(id uint, status dbmodels.OrderStatus) error {
	return store.db.Model(&dbmodels.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (store *DbStore) CancelOrder(id uint) error {
	return store.db.Model(&dbmodels.Order{}).
		Where("id = ?", id).
		Update("status", dbmodels.OrderStatus_CANCELED).Error
}

type PaymentUpdate struct {
	PaymentStatus string     `json:"payment_status" binding:"required"`
	Amount        float64    `json:"amount" binding:"required"`
	PaymentRef    string     `json:"payment_ref"`
	PaymentDate   *time.Time `json:"payment_date"`
	PaymentMethod int64      `json:"payment_method"`
	PaymentDesc   string     `json:"payment_desc"`
}

func (store *DbStore) UpdateOrderPayment(id uint, paymentUpdate PaymentUpdate) error {
	return store.db.Model(&dbmodels.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"payment_status": paymentUpdate.PaymentStatus,
			"payment_amount": paymentUpdate.Amount,
			"payment_date":   paymentUpdate.PaymentDate,
			"payment_method": paymentUpdate.PaymentMethod,
			"payment_desc":   paymentUpdate.PaymentDesc,
			"payment_ref":    paymentUpdate.PaymentRef,
		}).Error
}
