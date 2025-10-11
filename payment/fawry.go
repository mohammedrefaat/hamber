package payment

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	config "github.com/mohammedrefaat/hamber/Config"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== FAWRY PAYMENT SERVICE ==========

type FawryService struct {
	config config.FawryConfig
}

func NewFawryService(cfg config.FawryConfig) *FawryService {
	return &FawryService{config: cfg}
}

type FawryPaymentRequest struct {
	MerchantCode   string            `json:"merchantCode"`
	MerchantRefNum string            `json:"merchantRefNum"`
	CustomerName   string            `json:"customerName"`
	CustomerMobile string            `json:"customerMobile"`
	CustomerEmail  string            `json:"customerEmail"`
	PaymentAmount  float64           `json:"paymentAmount"`
	CurrencyCode   string            `json:"currencyCode"`
	PaymentMethod  string            `json:"paymentMethod"`
	Description    string            `json:"description"`
	ChargeItems    []FawryChargeItem `json:"chargeItems"`
	Signature      string            `json:"signature"`
	PaymentExpiry  int64             `json:"paymentExpiry"`
}

type FawryChargeItem struct {
	ItemID      string  `json:"itemId"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

type FawryPaymentResponse struct {
	Type            string  `json:"type"`
	ReferenceNumber string  `json:"referenceNumber"`
	MerchantRefNum  string  `json:"merchantRefNum"`
	OrderAmount     float64 `json:"orderAmount"`
	PaymentAmount   float64 `json:"paymentAmount"`
	FawryFees       float64 `json:"fawryFees"`
	OrderStatus     string  `json:"orderStatus"`
	PaymentMethod   string  `json:"paymentMethod"`
	ExpirationTime  int64   `json:"expirationTime"`
}

// Generate Fawry signature
func (s *FawryService) generateSignature(merchantRefNum string, amount float64) string {
	// Signature = SHA256(merchantCode + merchantRefNum + customerEmail + paymentAmount + securityKey)
	data := fmt.Sprintf("%s%s%.2f%s",
		s.config.MerchantCode,
		merchantRefNum,
		amount,
		s.config.SecurityKey)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *FawryService) InitiatePayment(payment *dbmodels.Payment, user *dbmodels.User, pkg *dbmodels.Package) (*FawryPaymentResponse, error) {
	// Generate unique reference number
	refNum := fmt.Sprintf("PKG-%d-%d-%d", user.ID, pkg.ID, time.Now().Unix())

	// Generate signature
	signature := s.generateSignature(refNum, payment.Amount)

	// Prepare request
	request := FawryPaymentRequest{
		MerchantCode:   s.config.MerchantCode,
		MerchantRefNum: refNum,
		CustomerName:   user.Name,
		CustomerMobile: user.Phone,
		CustomerEmail:  user.Email,
		PaymentAmount:  payment.Amount,
		CurrencyCode:   "EGP",
		PaymentMethod:  "PAYATFAWRY", // Or "CARD" for card payments
		Description:    fmt.Sprintf("Package: %s", pkg.Name),
		ChargeItems: []FawryChargeItem{
			{
				ItemID:      fmt.Sprintf("PKG-%d", pkg.ID),
				Description: pkg.Name,
				Price:       payment.Amount,
				Quantity:    1,
			},
		},
		Signature:     signature,
		PaymentExpiry: time.Now().Add(24*time.Hour).Unix() * 1000, // 24 hours
	}

	// Make API request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/ECommerceWeb/Fawry/payments/charge", s.config.APIURL)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fawry API error: %s", string(body))
	}

	var fawryResp FawryPaymentResponse
	if err := json.Unmarshal(body, &fawryResp); err != nil {
		return nil, err
	}

	return &fawryResp, nil
}

func (s *FawryService) VerifyCallback(signature string, refNum string, amount float64, orderStatus string) bool {
	// Verify signature: SHA256(merchantCode + referenceNumber + paymentAmount + orderStatus + securityKey)
	expectedSig := fmt.Sprintf("%s%s%.2f%s%s",
		s.config.MerchantCode,
		refNum,
		amount,
		orderStatus,
		s.config.SecurityKey)

	hash := sha256.Sum256([]byte(expectedSig))
	calculatedSig := hex.EncodeToString(hash[:])

	return signature == calculatedSig
}
