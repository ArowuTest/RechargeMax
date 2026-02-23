package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// VTPassService handles all VTPass API integrations
type VTPassService struct {
	apiKey    string
	publicKey string
	secretKey string
	baseURL   string
	client    *http.Client
	isSandbox bool
}

// VTPassConfig holds VTPass configuration
type VTPassConfig struct {
	APIKey    string
	PublicKey string
	SecretKey string
	IsSandbox bool
}

// VTPassPurchaseRequest represents a purchase request to VTPass
type VTPassPurchaseRequest struct {
	RequestID     string `json:"request_id"`
	ServiceID     string `json:"serviceID"`
	Amount        int    `json:"amount"`
	Phone         string `json:"phone"`
	BillersCode   string `json:"billersCode,omitempty"`
	VariationCode string `json:"variation_code,omitempty"`
}

// VTPassRequeryRequest represents a requery request
type VTPassRequeryRequest struct {
	RequestID string `json:"request_id"`
}

// VTPassResponse represents the standard VTPass API response
type VTPassResponse struct {
	Code                string                 `json:"code"`
	ResponseDescription string                 `json:"response_description"`
	RequestID           string                 `json:"requestId"`
	Amount              float64                `json:"amount"`
	TransactionDate     string                 `json:"transaction_date"`
	PurchasedCode       string                 `json:"purchased_code"`
	Content             VTPassResponseContent  `json:"content"`
	RawResponse         map[string]interface{} `json:"-"`
}

// VTPassResponseContent contains transaction details
type VTPassResponseContent struct {
	Transactions VTPassTransaction `json:"transactions"`
}

// VTPassTransaction contains detailed transaction information
type VTPassTransaction struct {
	Status              string                  `json:"status"`
	ProductName         string                  `json:"product_name"`
	UniqueElement       string                  `json:"unique_element"`
	UnitPrice           interface{}             `json:"unit_price"` // Can be string or number
	Quantity            int                     `json:"quantity"`
	ServiceVerification interface{}             `json:"service_verification"`
	Channel             string                  `json:"channel"`
	Commission          float64                 `json:"commission"`
	TotalAmount         float64                 `json:"total_amount"`
	Discount            interface{}             `json:"discount"`
	Type                string                  `json:"type"`
	Email               string                  `json:"email"`
	Phone               string                  `json:"phone"`
	Name                interface{}             `json:"name"`
	ConvenienceFee      float64                 `json:"convinience_fee"`
	Amount              interface{}             `json:"amount"` // Can be string or number
	Platform            string                  `json:"platform"`
	Method              string                  `json:"method"`
	TransactionID       string                  `json:"transactionId"`
	CommissionDetails   *VTPassCommissionDetail `json:"commission_details,omitempty"`
}

// VTPassCommissionDetail contains commission breakdown
type VTPassCommissionDetail struct {
	Amount          float64 `json:"amount"`
	Rate            string  `json:"rate"`
	RateType        string  `json:"rate_type"`
	ComputationType string  `json:"computation_type"`
}

// VTPassVariation represents a service variation (data plan)
type VTPassVariation struct {
	VariationCode   string `json:"variation_code"`
	Name            string `json:"name"`
	VariationAmount string `json:"variation_amount"`
	FixedPrice      string `json:"fixedPrice"`
}

// VTPassVariationsResponse represents the variations API response
type VTPassVariationsResponse struct {
	Code     string `json:"code"`
	Content  struct {
		Variations []VTPassVariation `json:"varations"` // Note: VTPass has typo in their API
	} `json:"content"`
}

// Network service IDs for VTPass
const (
	VTPassServiceMTNAirtime     = "mtn"
	VTPassServiceGloAirtime     = "glo"
	VTPassServiceAirtelAirtime  = "airtel"
	VTPassService9mobileAirtime = "etisalat"
	VTPassServiceMTNData        = "mtn-data"
	VTPassServiceGloData        = "glo-data"
	VTPassServiceAirtelData     = "airtel-data"
	VTPassService9mobileData    = "etisalat-data"
)

// Response codes
const (
	VTPassCodeSuccess  = "000"
	VTPassCodePending  = "011"
	VTPassCodeFailed   = "015"
	VTPassCodeReversed = "016"
	VTPassCodeInvalid  = "099"
)

// NewVTPassService creates a new VTPass service instance
func NewVTPassService(config VTPassConfig) *VTPassService {
	baseURL := "https://vtpass.com/api"
	if config.IsSandbox {
		baseURL = "https://sandbox.vtpass.com/api"
	}

	fmt.Printf("🔧 Initializing VTPassService with:\n")
	fmt.Printf("   API Key: %s\n", config.APIKey)
	fmt.Printf("   Public Key: %s\n", config.PublicKey)
	// Only show first 20 chars of secret if it's long enough
	if len(config.SecretKey) > 20 {
		fmt.Printf("   Secret Key: %s...\n", config.SecretKey[:20])
	} else {
		fmt.Printf("   Secret Key: %s\n", config.SecretKey)
	}
	fmt.Printf("   Base URL: %s\n", baseURL)
	fmt.Printf("   Is Sandbox: %v\n", config.IsSandbox)
	
	return &VTPassService{
		apiKey:    config.APIKey,
		publicKey: config.PublicKey,
		secretKey: config.SecretKey,
		baseURL:   baseURL,
		isSandbox: config.IsSandbox,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// PurchaseAirtime purchases airtime through VTPass
func (s *VTPassService) PurchaseAirtime(ctx context.Context, network, phone string, amount int) (*VTPassResponse, error) {
	serviceID := s.getAirtimeServiceID(network)
	if serviceID == "" {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	// Convert phone to local format for VTPass sandbox (234XXXXXXXXXX -> 0XXXXXXXXXX)
	// VTPass sandbox expects local format for test numbers
	phoneForVTPass := s.formatPhoneForVTPass(phone)

	requestID := s.generateRequestID()

	request := VTPassPurchaseRequest{
		RequestID: requestID,
		ServiceID: serviceID,
		Amount:    amount,
		Phone:     phoneForVTPass,
	}

	return s.purchase(ctx, request)
}

// PurchaseData purchases data bundle through VTPass
func (s *VTPassService) PurchaseData(ctx context.Context, network, phone, variationCode string, amount int) (*VTPassResponse, error) {
	serviceID := s.getDataServiceID(network)
	if serviceID == "" {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	// Convert phone to local format for VTPass sandbox
	phoneForVTPass := s.formatPhoneForVTPass(phone)

	requestID := s.generateRequestID()

	request := VTPassPurchaseRequest{
		RequestID:     requestID,
		ServiceID:     serviceID,
		Amount:        amount,
		Phone:         phoneForVTPass,
		BillersCode:   phoneForVTPass,
		VariationCode: variationCode,
	}

	return s.purchase(ctx, request)
}

// GetDataVariations retrieves available data plans for a network
func (s *VTPassService) GetDataVariations(ctx context.Context, network string) ([]VTPassVariation, error) {
	serviceID := s.getDataServiceID(network)
	if serviceID == "" {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	url := fmt.Sprintf("%s/service-variations?serviceID=%s", s.baseURL, serviceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for GET request
	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("public-key", s.publicKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var variationsResp VTPassVariationsResponse
	if err := json.Unmarshal(body, &variationsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if variationsResp.Code != VTPassCodeSuccess {
		return nil, fmt.Errorf("VTPass API error: code %s", variationsResp.Code)
	}

	return variationsResp.Content.Variations, nil
}

// RequeryTransaction queries the status of a transaction
func (s *VTPassService) RequeryTransaction(ctx context.Context, requestID string) (*VTPassResponse, error) {
	url := fmt.Sprintf("%s/requery", s.baseURL)

	request := VTPassRequeryRequest{
		RequestID: requestID,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for POST request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("secret-key", s.secretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log raw response for debugging
	fmt.Printf("VTPass Raw Response (Status %d): %s\n", resp.StatusCode, string(body))

	var vtpassResp VTPassResponse
	if err := json.Unmarshal(body, &vtpassResp); err != nil {
		return nil, fmt.Errorf("failed to parse response (raw: %s): %w", string(body), err)
	}

	// Store raw response for debugging
	var rawResp map[string]interface{}
	json.Unmarshal(body, &rawResp)
	vtpassResp.RawResponse = rawResp

	return &vtpassResp, nil
}

// purchase executes a purchase request
func (s *VTPassService) purchase(ctx context.Context, request VTPassPurchaseRequest) (*VTPassResponse, error) {
	url := fmt.Sprintf("%s/pay", s.baseURL)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for POST request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("secret-key", s.secretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log raw response for debugging
	fmt.Printf("VTPass Raw Response (Status %d): %s\n", resp.StatusCode, string(body))

	var vtpassResp VTPassResponse
	if err := json.Unmarshal(body, &vtpassResp); err != nil {
		return nil, fmt.Errorf("failed to parse response (raw: %s): %w", string(body), err)
	}

	// Store raw response for debugging
	var rawResp map[string]interface{}
	json.Unmarshal(body, &rawResp)
	vtpassResp.RawResponse = rawResp

	return &vtpassResp, nil
}

// getAirtimeServiceID returns the VTPass service ID for airtime
func (s *VTPassService) getAirtimeServiceID(network string) string {
	switch network {
	case "MTN":
		return VTPassServiceMTNAirtime
	case "GLO":
		return VTPassServiceGloAirtime
	case "AIRTEL":
		return VTPassServiceAirtelAirtime
	case "9MOBILE", "NINE_MOBILE":
		return VTPassService9mobileAirtime
	default:
		return ""
	}
}

// getDataServiceID returns the VTPass service ID for data
func (s *VTPassService) getDataServiceID(network string) string {
	switch network {
	case "MTN":
		return VTPassServiceMTNData
	case "GLO":
		return VTPassServiceGloData
	case "AIRTEL":
		return VTPassServiceAirtelData
	case "9MOBILE", "NINE_MOBILE":
		return VTPassService9mobileData
	default:
		return ""
	}
}

// generateRequestID generates a unique request ID
func (s *VTPassService) generateRequestID() string {
	// Format: YYYYMMDDHHMMSS + random UUID suffix
	timestamp := time.Now().Format("20060102150405")
	uuid := uuid.New().String()[:8]
	return timestamp + uuid
}

// IsSuccessful checks if the response indicates a successful transaction
func (r *VTPassResponse) IsSuccessful() bool {
	return r.Code == VTPassCodeSuccess && r.Content.Transactions.Status == "delivered"
}

// IsPending checks if the response indicates a pending transaction
func (r *VTPassResponse) IsPending() bool {
	return r.Code == VTPassCodePending || r.Content.Transactions.Status == "pending"
}

// IsFailed checks if the response indicates a failed transaction
func (r *VTPassResponse) IsFailed() bool {
	return r.Code == VTPassCodeFailed || r.Content.Transactions.Status == "failed"
}

// IsReversed checks if the response indicates a reversed transaction
func (r *VTPassResponse) IsReversed() bool {
	return r.Code == VTPassCodeReversed || r.Content.Transactions.Status == "reversed"
}

// GetStatus returns the transaction status
func (r *VTPassResponse) GetStatus() string {
	if r.IsSuccessful() {
		return "SUCCESS"
	}
	if r.IsPending() {
		return "PROCESSING"
	}
	if r.IsFailed() {
		return "FAILED"
	}
	if r.IsReversed() {
		return "CANCELLED"
	}
	return "FAILED"
}

// GetProviderReference returns the VTPass transaction ID
func (r *VTPassResponse) GetProviderReference() string {
	return r.Content.Transactions.TransactionID
}

// GetCommission returns the commission amount
func (r *VTPassResponse) GetCommission() float64 {
	return r.Content.Transactions.Commission
}

// GetErrorMessage returns a human-readable error message
func (r *VTPassResponse) GetErrorMessage() string {
	if r.IsSuccessful() {
		return ""
	}
	return r.ResponseDescription
}

// formatPhoneForVTPass converts phone number to local Nigerian format (0XXXXXXXXXX)
// VTPass accepts both formats, but we standardize on local format for consistency
// Converts: 234XXXXXXXXXX -> 0XXXXXXXXXX
// Keeps: 0XXXXXXXXXX -> 0XXXXXXXXXX
func (s *VTPassService) formatPhoneForVTPass(phone string) string {
	// Remove all non-digit characters
	digitsOnly := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digitsOnly += string(char)
		}
	}

	// If in international format (234XXXXXXXXXX), convert to local format (0XXXXXXXXXX)
	if len(digitsOnly) == 13 && digitsOnly[:3] == "234" {
		return "0" + digitsOnly[3:]
	}

	// If already in local format (0XXXXXXXXXX), return as is
	if len(digitsOnly) == 11 && digitsOnly[0] == '0' {
		return digitsOnly
	}

	// For other formats, try to add 0 prefix if it's 10 digits
	if len(digitsOnly) == 10 {
		return "0" + digitsOnly
	}

	// Return as is if we can't determine format
	return phone
}
