package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== RECEIPT GENERATION CONTROLLERS ==========

type CompanyInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
	Logo    string `json:"logo"`
	TaxID   string `json:"tax_id"`
}

// GenerateOrderReceipt godoc
// @Summary      Generate order receipt
// @Description  Generate PDF receipt for an order
// @Tags         Receipts
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        order_id path int true "Order ID"
// @Param        request body map[string]interface{} false "Company info"
// @Success      201 {object} map[string]interface{} "Receipt generated"
// @Failure      404 {object} map[string]interface{} "Order not found"
// @Router       /receipts/order/{order_id} [post]
func GenerateOrderReceipt(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Get order details
	order, err := globalStore.StStore.GetOrderByID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Verify ownership
	if order.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check if receipt already exists
	existingReceipt, err := globalStore.StStore.GetOrderReceipt(uint(orderID))
	if err == nil && existingReceipt != nil {
		// Return existing receipt
		c.JSON(http.StatusOK, gin.H{
			"receipt": existingReceipt,
			"message": "Receipt already exists",
		})
		return
	}

	// Get company info from request or use defaults
	var companyInfo CompanyInfo
	if err := c.ShouldBindJSON(&companyInfo); err != nil {
		// Use default company info
		companyInfo = CompanyInfo{
			Name:    "Hamber Platform",
			Address: "123 Business St, Cairo, Egypt",
			Phone:   "+20 123 456 7890",
			Email:   "info@hamber.local",
			Website: "www.hamber.local",
			TaxID:   "TAX-123456",
		}
	}

	// Generate receipt number
	receiptNumber := fmt.Sprintf("RCP-%d-%d", orderID, time.Now().Unix())

	// Generate PDF
	pdfPath, err := generateReceiptPDF(order, &companyInfo, receiptNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	// Save receipt record
	companyInfoJSON, _ := json.Marshal(companyInfo)
	receipt := &dbmodels.OrderReceipt{
		OrderID:         uint(orderID),
		ReceiptNumber:   receiptNumber,
		PDFPath:         pdfPath,
		TemplateVersion: "v1",
		CompanyInfo:     string(companyInfoJSON),
		GeneratedAt:     &time.Time{},
	}
	*receipt.GeneratedAt = time.Now()

	if err := globalStore.StStore.CreateOrderReceipt(receipt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save receipt"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"receipt": receipt,
		"message": "Receipt generated successfully",
	})
}

// GetOrderReceipt retrieves an existing receipt
func GetOrderReceipt(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	receipt, err := globalStore.StStore.GetOrderReceipt(uint(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	// Verify ownership
	if receipt.Order.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"receipt": receipt,
	})
}

// DownloadReceipt godoc
// @Summary      Download receipt PDF
// @Description  Download receipt as PDF file
// @Tags         Receipts
// @Accept       json
// @Produce      application/pdf
// @Security     Bearer
// @Param        order_id path int true "Order ID"
// @Success      200 {file} file "PDF file"
// @Failure      404 {object} map[string]interface{} "Receipt not found"
// @Router       /receipts/order/{order_id}/download [get]{object} AuthResponse "Login successful"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Invalid credentials"
// @Router       /auth/login [post]
func DownloadReceipt(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	receipt, err := globalStore.StStore.GetOrderReceipt(uint(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	// Verify ownership
	if receipt.Order.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Serve the PDF file
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", receipt.ReceiptNumber))
	c.File(receipt.PDFPath)
}

// generateReceiptPDF creates a PDF receipt
func generateReceiptPDF(order *dbmodels.Order, company *CompanyInfo, receiptNumber string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)

	// Company Header
	pdf.Cell(190, 10, company.Name)
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 5, company.Address)
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("Phone: %s | Email: %s", company.Phone, company.Email))
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("Website: %s | Tax ID: %s", company.Website, company.TaxID))
	pdf.Ln(10)

	// Receipt Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, "RECEIPT")
	pdf.Ln(10)

	// Receipt Details
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 6, fmt.Sprintf("Receipt Number: %s", receiptNumber))
	pdf.Cell(95, 6, fmt.Sprintf("Date: %s", time.Now().Format("2006-01-02")))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Order ID: #%d", order.ID))
	pdf.Cell(95, 6, fmt.Sprintf("Order Date: %s", order.CreatedAt.Format("2006-01-02")))
	pdf.Ln(10)

	// Customer Details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Customer Details")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("Name: %s", order.Client.Name))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Email: %s", order.Client.Email))
	pdf.Ln(6)
	if order.Phone != "" {
		pdf.Cell(190, 6, fmt.Sprintf("Phone: %s", order.Phone))
		pdf.Ln(6)
	}
	if order.Address != "" {
		pdf.Cell(190, 6, fmt.Sprintf("Address: %s", order.Address))
		pdf.Ln(6)
	}
	pdf.Ln(5)

	// Order Items Table
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Order Items")
	pdf.Ln(8)

	// Table Header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(80, 7, "Product", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 7, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 7, "Total", "1", 0, "R", true, 0, "")
	pdf.Ln(7)

	// Table Body
	pdf.SetFont("Arial", "", 10)
	for _, item := range order.Items {
		pdf.CellFormat(80, 7, item.Product.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 7, strconv.Itoa(item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%.2f EGP", item.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%.2f EGP", item.Price*float64(item.Quantity)), "1", 0, "R", false, 0, "")
		pdf.Ln(7)
	}

	// Total
	pdf.Ln(3)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(150, 8, "Total Amount:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f EGP", order.Total), "1", 0, "R", false, 0, "")
	pdf.Ln(10)

	// Payment Information
	if order.PaymentStatus != "" {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "Payment Information")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(190, 6, fmt.Sprintf("Payment Status: %s", order.PaymentStatus))
		pdf.Ln(6)
		if order.PaymentMethodDesc != "" {
			pdf.Cell(190, 6, fmt.Sprintf("Payment Method: %s", order.PaymentMethodDesc))
			pdf.Ln(6)
		}
		if order.PaymentDate != nil {
			pdf.Cell(190, 6, fmt.Sprintf("Payment Date: %s", order.PaymentDate.Format("2006-01-02")))
			pdf.Ln(6)
		}
		if order.PaymentRef != "" {
			pdf.Cell(190, 6, fmt.Sprintf("Reference: %s", order.PaymentRef))
			pdf.Ln(6)
		}
	}

	// Notes
	if order.Notes != "" {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "Notes")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(190, 5, order.Notes, "", "", false)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 5, "Thank you for your business!")
	pdf.Ln(5)
	pdf.Cell(190, 5, "This is a computer-generated receipt and does not require a signature.")

	// Save PDF
	pdfPath := fmt.Sprintf("./uploads/receipts/%s.pdf", receiptNumber)
	if err := pdf.OutputFileAndClose(pdfPath); err != nil {
		return "", err
	}

	return pdfPath, nil
}

// GetReceiptHTML generates HTML view of receipt
func GetReceiptHTML(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := globalStore.StStore.GetOrderByID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Verify ownership
	if order.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	receipt, _ := globalStore.StStore.GetOrderReceipt(uint(orderID))

	var company CompanyInfo
	if receipt != nil {
		json.Unmarshal([]byte(receipt.CompanyInfo), &company)
	} else {
		company = CompanyInfo{
			Name:    "Hamber Platform",
			Address: "123 Business St, Cairo, Egypt",
			Phone:   "+20 123 456 7890",
			Email:   "info@hamber.local",
			Website: "www.hamber.local",
			TaxID:   "TAX-123456",
		}
	}

	// Generate HTML
	tmpl := template.Must(template.New("receipt").Parse(receiptHTMLTemplate))

	var buf bytes.Buffer
	data := map[string]interface{}{
		"Company": company,
		"Order":   order,
		"Receipt": receipt,
		"Date":    time.Now().Format("2006-01-02"),
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate HTML"})
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, buf.String())
}

const receiptHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Receipt</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            text-align: center;
            border-bottom: 2px solid #333;
            padding-bottom: 20px;
            margin-bottom: 20px;
        }
        .company-name {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 10px;
        }
        .receipt-title {
            font-size: 20px;
            font-weight: bold;
            text-align: center;
            margin: 20px 0;
        }
        .section {
            margin: 20px 0;
        }
        .section-title {
            font-size: 16px;
            font-weight: bold;
            margin-bottom: 10px;
            border-bottom: 1px solid #ccc;
            padding-bottom: 5px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 10px 0;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
        .total {
            text-align: right;
            font-size: 18px;
            font-weight: bold;
            margin-top: 20px;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ccc;
            font-size: 12px;
            color: #666;
        }
        @media print {
            .no-print {
                display: none;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="company-name">{{.Company.Name}}</div>
        <div>{{.Company.Address}}</div>
        <div>Phone: {{.Company.Phone}} | Email: {{.Company.Email}}</div>
        <div>Website: {{.Company.Website}} | Tax ID: {{.Company.TaxID}}</div>
    </div>

    <div class="receipt-title">RECEIPT</div>

    <div class="section">
        <div><strong>Receipt Number:</strong> {{if .Receipt}}{{.Receipt.ReceiptNumber}}{{else}}N/A{{end}}</div>
        <div><strong>Date:</strong> {{.Date}}</div>
        <div><strong>Order ID:</strong> #{{.Order.ID}}</div>
    </div>

    <div class="section">
        <div class="section-title">Customer Details</div>
        <div><strong>Name:</strong> {{.Order.Client.Name}}</div>
        <div><strong>Email:</strong> {{.Order.Client.Email}}</div>
        {{if .Order.Phone}}<div><strong>Phone:</strong> {{.Order.Phone}}</div>{{end}}
        {{if .Order.Address}}<div><strong>Address:</strong> {{.Order.Address}}</div>{{end}}
    </div>

    <div class="section">
        <div class="section-title">Order Items</div>
        <table>
            <thead>
                <tr>
                    <th>Product</th>
                    <th>Quantity</th>
                    <th>Price</th>
                    <th>Total</th>
                </tr>
            </thead>
            <tbody>
                {{range .Order.Items}}
                <tr>
                    <td>{{.Product.Name}}</td>
                    <td>{{.Quantity}}</td>
                    <td>{{printf "%.2f" .Price}} EGP</td>
                    <td>{{printf "%.2f" (multiply .Price .Quantity)}} EGP</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        <div class="total">Total Amount: {{printf "%.2f" .Order.Total}} EGP</div>
    </div>

    {{if .Order.PaymentStatus}}
    <div class="section">
        <div class="section-title">Payment Information</div>
        <div><strong>Payment Status:</strong> {{.Order.PaymentStatus}}</div>
        {{if .Order.PaymentMethodDesc}}<div><strong>Payment Method:</strong> {{.Order.PaymentMethodDesc}}</div>{{end}}
        {{if .Order.PaymentRef}}<div><strong>Reference:</strong> {{.Order.PaymentRef}}</div>{{end}}
    </div>
    {{end}}

    {{if .Order.Notes}}
    <div class="section">
        <div class="section-title">Notes</div>
        <div>{{.Order.Notes}}</div>
    </div>
    {{end}}

    <div class="footer">
        <p>Thank you for your business!</p>
        <p>This is a computer-generated receipt.</p>
    </div>

    <div class="no-print" style="text-align: center; margin-top: 20px;">
        <button onclick="window.print()">Print Receipt</button>
    </div>
</body>
</html>
`
