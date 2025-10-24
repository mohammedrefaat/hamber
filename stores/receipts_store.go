package stores

import (
	"net/http"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== ORDER RECEIPT MANAGEMENT ==========

func (store *DbStore) CreateOrderReceipt(receipt *dbmodels.OrderReceipt) error {
	return store.db.Create(receipt).Error
}

func (store *DbStore) GetOrderReceipt(orderID uint) (*dbmodels.OrderReceipt, error) {
	var receipt dbmodels.OrderReceipt
	if err := store.db.Preload("Order").Preload("Order.Items").Preload("Order.Items.Product").
		Preload("Order.Client").Preload("Order.User").
		Where("order_id = ?", orderID).First(&receipt).Error; err != nil {
		return nil, &CustomError{
			Message: "Receipt not found",
			Code:    http.StatusNotFound,
		}
	}
	return &receipt, nil
}

func (store *DbStore) UpdateOrderReceipt(receipt *dbmodels.OrderReceipt) error {
	return store.db.Save(receipt).Error
}
