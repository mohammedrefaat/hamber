// Create a new file: payment/paymob.go
package payment

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	config "github.com/mohammedrefaat/hamber/Config"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== PAYMOB PAYMENT SERVICE ==========

type PaymobService struct {
	config config.PaymobConfig
}

func NewPaymobService(cfg config.PaymobConfig) *PaymobService {
	return &PaymobService{config: cfg}
}

type PaymobAuthResponse struct {
	Token string `json:"token"`
}

type PaymobOrderRequest struct {
	AuthToken       string                   `json:"auth_token"`
	DeliveryNeeded  string                   `json:"delivery_needed"`
	AmountCents     int                      `json:"amount_cents"`
	Currency        string                   `json:"currency"`
	MerchantOrderID string                   `json:"merchant_order_id"`
	Items           []map[string]interface{} `json:"items"`
}

type PaymobOrderResponse struct {
	ID int `json:"id"`
}

type PaymobPaymentKeyRequest struct {
	AuthToken     string            `json:"auth_token"`
	AmountCents   int               `json:"amount_cents"`
	Expiration    int               `json:"expiration"`
	OrderID       string            `json:"order_id"`
	BillingData   PaymobBillingData `json:"billing_data"`
	Currency      string            `json:"currency"`
	IntegrationID int               `json:"integration_id"`
}

type PaymobBillingData struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	PhoneNumber    string `json:"phone_number"`
	Apartment      string `json:"apartment"`
	Floor          string `json:"floor"`
	Street         string `json:"street"`
	Building       string `json:"building"`
	ShippingMethod string `json:"shipping_method"`
	PostalCode     string `json:"postal_code"`
	City           string `json:"city"`
	Country        string `json:"country"`
	State          string `json:"state"`
}

type PaymobPaymentKeyResponse struct {
	Token string `json:"token"`
}

func (s *PaymobService) Authenticate() (string, error) {
	reqBody := map[string]string{
		"api_key": s.config.APIKey,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(
		fmt.Sprintf("%s/auth/tokens", s.config.APIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var authResp PaymobAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	return authResp.Token, nil
}

func (s *PaymobService) CreateOrder(authToken string, payment *dbmodels.Payment, pkg *dbmodels.Package) (int, error) {
	orderReq := PaymobOrderRequest{
		AuthToken:       authToken,
		DeliveryNeeded:  "false",
		AmountCents:     int(payment.Amount * 100), // Convert to cents
		Currency:        "EGP",
		MerchantOrderID: fmt.Sprintf("PKG-%d-%d", payment.UserID, time.Now().Unix()),
		Items: []map[string]interface{}{
			{
				"name":         pkg.Name,
				"amount_cents": int(payment.Amount * 100),
				"description":  pkg.Description,
				"quantity":     1,
			},
		},
	}

	jsonData, _ := json.Marshal(orderReq)
	resp, err := http.Post(
		fmt.Sprintf("%s/ecommerce/orders", s.config.APIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var orderResp PaymobOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return 0, err
	}

	return orderResp.ID, nil
}

func (s *PaymobService) GetPaymentKey(authToken string, orderID int, payment *dbmodels.Payment, user *dbmodels.User) (string, error) {
	integrationID, _ := strconv.Atoi(s.config.IntegrationID)

	billingData := PaymobBillingData{
		FirstName:      user.Name,
		LastName:       user.Name,
		Email:          user.Email,
		PhoneNumber:    user.Phone,
		Apartment:      "NA",
		Floor:          "NA",
		Street:         "NA",
		Building:       "NA",
		ShippingMethod: "NA",
		PostalCode:     "NA",
		City:           "NA",
		Country:        "EG",
		State:          "NA",
	}

	paymentKeyReq := PaymobPaymentKeyRequest{
		AuthToken:     authToken,
		AmountCents:   int(payment.Amount * 100),
		Expiration:    3600, // 1 hour
		OrderID:       strconv.Itoa(orderID),
		BillingData:   billingData,
		Currency:      "EGP",
		IntegrationID: integrationID,
	}

	jsonData, _ := json.Marshal(paymentKeyReq)
	resp, err := http.Post(
		fmt.Sprintf("%s/acceptance/payment_keys", s.config.APIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var keyResp PaymobPaymentKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&keyResp); err != nil {
		return "", err
	}

	return keyResp.Token, nil
}

func (s *PaymobService) InitiatePayment(payment *dbmodels.Payment, user *dbmodels.User, pkg *dbmodels.Package) (string, error) {
	// Step 1: Authenticate
	authToken, err := s.Authenticate()
	if err != nil {
		return "", errors.New("failed to authenticate with Paymob")
	}

	// Step 2: Create order
	orderID, err := s.CreateOrder(authToken, payment, pkg)
	if err != nil {
		return "", errors.New("failed to create Paymob order")
	}

	// Step 3: Get payment key
	paymentKey, err := s.GetPaymentKey(authToken, orderID, payment, user)
	if err != nil {
		return "", errors.New("failed to get payment key")
	}

	// Return iframe URL with payment key
	iframeURL := fmt.Sprintf("https://accept.paymob.com/api/acceptance/iframes/%s?payment_token=%s",
		s.config.IframeID, paymentKey)

	return iframeURL, nil
}

func (s *PaymobService) VerifyCallback(hmacFromCallback, amountCents, currency, success, orderId, merchantOrderId string) bool {
	// Concatenate the callback data
	concatenatedString := fmt.Sprintf("%s%s%s%s%s",
		amountCents,
		currency,
		success,
		orderId,
		merchantOrderId)

	// Calculate HMAC
	hash := sha256.New()
	hash.Write([]byte(concatenatedString))
	hash.Write([]byte(s.config.HMACSecret))
	calculatedHMAC := hex.EncodeToString(hash.Sum(nil))

	return hmacFromCallback == calculatedHMAC
}
