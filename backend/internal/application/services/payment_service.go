package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"rechargemax/internal/domain/repositories"
)

// PaymentService handles payment gateway integrations
type PaymentService struct {
	paystackSecretKey    string
	flutterwaveSecretKey string
	client               *http.Client
	paymentRepo          repositories.PaymentLogRepository
}

// PaymentRequest represents a payment initialization request
type PaymentRequest struct {
	Amount      int64                  `json:"amount"`
	Email       string                 `json:"email"`
	Reference   string                 `json:"reference"`
	CallbackURL string                 `json:"callback_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PaymentResponse represents payment initialization response
type PaymentResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	AuthorizationURL string `json:"authorization_url"`
	AccessCode       string `json:"access_code"`
	Reference        string `json:"reference"`
}

// PaystackInitRequest represents Paystack initialization request
type PaystackInitRequest struct {
	Amount      int64                  `json:"amount"`
	Email       string                 `json:"email"`
	Reference   string                 `json:"reference"`
	CallbackURL string                 `json:"callback_url"`
	Metadata    map[string]interface{} `json:"metadata"`
	Currency    string                 `json:"currency"`
}

// PaystackInitResponse represents Paystack initialization response
type PaystackInitResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// FlutterwaveInitRequest represents Flutterwave initialization request
type FlutterwaveInitRequest struct {
	TxRef       string                 `json:"tx_ref"`
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency"`
	RedirectURL string                 `json:"redirect_url"`
	Customer    FlutterwaveCustomer    `json:"customer"`
	Meta        map[string]interface{} `json:"meta"`
}

// FlutterwaveCustomer represents Flutterwave customer data
type FlutterwaveCustomer struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// FlutterwaveInitResponse represents Flutterwave initialization response
type FlutterwaveInitResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Link string `json:"link"`
	} `json:"data"`
}

// PaymentWebhookPayload represents payment webhook payload
type PaymentWebhookPayload struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

// NewPaymentService creates a new payment service
func NewPaymentService(paystackSecretKey, flutterwaveSecretKey string, paymentRepo repositories.PaymentLogRepository) *PaymentService {
	return &PaymentService{
		paystackSecretKey:    paystackSecretKey,
		flutterwaveSecretKey: flutterwaveSecretKey,
		paymentRepo:          paymentRepo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitializePayment initializes a payment with the specified gateway
func (s *PaymentService) InitializePayment(ctx context.Context, req PaymentRequest) (string, error) {
	// Default to Paystack if no gateway specified
	gateway := "paystack"
	if gatewayMeta, ok := req.Metadata["gateway"]; ok {
		if gw, ok := gatewayMeta.(string); ok {
			gateway = gw
		}
	}

	switch gateway {
	case "paystack":
		return s.initializePaystack(ctx, req)
	case "flutterwave":
		return s.initializeFlutterwave(ctx, req)
	default:
		return "", fmt.Errorf("unsupported payment gateway: %s", gateway)
	}
}

// initializePaystack initializes payment with Paystack
func (s *PaymentService) initializePaystack(ctx context.Context, req PaymentRequest) (string, error) {
	paystackReq := PaystackInitRequest{
		Amount:      req.Amount * 100, // Convert naira to kobo for Paystack
		Email:       req.Email,
		Reference:   req.Reference,
		CallbackURL: req.CallbackURL,
		Metadata:    req.Metadata,
		Currency:    "NGN",
	}

	jsonData, err := json.Marshal(paystackReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Paystack request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.paystack.co/transaction/initialize", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create Paystack request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("Paystack request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Paystack response: %w", err)
	}

	var paystackResp PaystackInitResponse
	if err := json.Unmarshal(body, &paystackResp); err != nil {
		return "", fmt.Errorf("failed to parse Paystack response: %w", err)
	}

	if !paystackResp.Status {
		return "", fmt.Errorf("Paystack initialization failed: %s", paystackResp.Message)
	}

	return paystackResp.Data.AuthorizationURL, nil
}

// initializeFlutterwave initializes payment with Flutterwave
func (s *PaymentService) initializeFlutterwave(ctx context.Context, req PaymentRequest) (string, error) {
	flwReq := FlutterwaveInitRequest{
		TxRef:       req.Reference,
		Amount:      req.Amount / 100, // Flutterwave expects amount in naira
		Currency:    "NGN",
		RedirectURL: req.CallbackURL,
		Customer: FlutterwaveCustomer{
			Email: req.Email,
			Name:  "RechargeMax User",
		},
		Meta: req.Metadata,
	}

	jsonData, err := json.Marshal(flwReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Flutterwave request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.flutterwave.com/v3/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create Flutterwave request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.flutterwaveSecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("Flutterwave request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Flutterwave response: %w", err)
	}

	var flwResp FlutterwaveInitResponse
	if err := json.Unmarshal(body, &flwResp); err != nil {
		return "", fmt.Errorf("failed to parse Flutterwave response: %w", err)
	}

	if flwResp.Status != "success" {
		return "", fmt.Errorf("Flutterwave initialization failed: %s", flwResp.Message)
	}

	return flwResp.Data.Link, nil
}

// VerifyPayment verifies a payment with the gateway
func (s *PaymentService) VerifyPayment(ctx context.Context, reference, gateway string) (bool, map[string]interface{}, error) {
	switch gateway {
	case "paystack":
		return s.verifyPaystack(ctx, reference)
	case "flutterwave":
		return s.verifyFlutterwave(ctx, reference)
	default:
		return false, nil, fmt.Errorf("unsupported payment gateway: %s", gateway)
	}
}

// verifyPaystack verifies payment with Paystack
func (s *PaymentService) verifyPaystack(ctx context.Context, reference string) (bool, map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", reference)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to create Paystack verify request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("Paystack verify request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read Paystack verify response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, nil, fmt.Errorf("failed to parse Paystack verify response: %w", err)
	}

	status, ok := result["status"].(bool)
	if !ok || !status {
		return false, result, nil
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return false, result, nil
	}

	txStatus, ok := data["status"].(string)
	if !ok {
		return false, result, nil
	}

	return txStatus == "success", result, nil
}

// verifyFlutterwave verifies payment with Flutterwave
func (s *PaymentService) verifyFlutterwave(ctx context.Context, reference string) (bool, map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.flutterwave.com/v3/transactions/verify_by_reference?tx_ref=%s", reference)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to create Flutterwave verify request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.flutterwaveSecretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("Flutterwave verify request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read Flutterwave verify response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, nil, fmt.Errorf("failed to parse Flutterwave verify response: %w", err)
	}

	status, ok := result["status"].(string)
	if !ok || status != "success" {
		return false, result, nil
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return false, result, nil
	}

	txStatus, ok := data["status"].(string)
	if !ok {
		return false, result, nil
	}

	return txStatus == "successful", result, nil
}

// ProcessWebhook processes payment webhooks
func (s *PaymentService) ProcessWebhook(payload []byte, signature, gateway string) (string, string, error) {
	// Verify webhook signature
	if !s.verifyWebhookSignature(payload, signature, gateway) {
		return "", "", fmt.Errorf("invalid webhook signature")
	}

	var webhook PaymentWebhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return "", "", fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Extract transaction reference
	var reference string
	if data, ok := webhook.Data["reference"]; ok {
		reference, _ = data.(string)
	} else if data, ok := webhook.Data["tx_ref"]; ok {
		reference, _ = data.(string)
	}

	if reference == "" {
		return "", "", fmt.Errorf("no transaction reference found in webhook")
	}

	// Determine transaction status
	var status string
	switch webhook.Event {
	case "charge.success", "transaction.successful":
		status = "completed"
	case "charge.failed", "transaction.failed":
		status = "failed"
	default:
		// Ignore other events
		return "", "", nil
	}

	return reference, status, nil
}

// verifyWebhookSignature verifies webhook signatures
func (s *PaymentService) verifyWebhookSignature(payload []byte, signature, gateway string) bool {
	var secret string
	switch gateway {
	case "paystack":
		secret = s.paystackSecretKey
	case "flutterwave":
		secret = s.flutterwaveSecretKey
	default:
		return false
	}

	// Paystack uses SHA512 HMAC
	if gateway == "paystack" {
		// Debug logging
		fmt.Printf("DEBUG: Payload length: %d\n", len(payload))
		if len(payload) > 100 {
			fmt.Printf("DEBUG: Payload (first 100 chars): %s\n", string(payload[:100]))
		} else {
			fmt.Printf("DEBUG: Payload: %s\n", string(payload))
		}
		fmt.Printf("DEBUG: Secret key: %s\n", secret)
		fmt.Printf("DEBUG: Received signature: %s\n", signature)
		
		h := hmac.New(sha512.New, []byte(secret))
		h.Write(payload)
		expectedSignature := hex.EncodeToString(h.Sum(nil))
		fmt.Printf("DEBUG: Expected signature: %s\n", expectedSignature)
		fmt.Printf("DEBUG: Signatures match: %v\n", hmac.Equal([]byte(signature), []byte(expectedSignature)))
		return hmac.Equal([]byte(signature), []byte(expectedSignature))
	}

	// Flutterwave uses different signature method
	if gateway == "flutterwave" {
		// Flutterwave sends the secret key as the signature
		return signature == secret
	}

	return false
}

// GetPaymentMethods returns available payment methods
func (s *PaymentService) GetPaymentMethods() []map[string]interface{} {
	methods := []map[string]interface{}{
		{
			"code":        "paystack",
			"name":        "Paystack",
			"description": "Pay with card, bank transfer, or USSD",
			"logo":        "/static/images/payment/paystack.png",
			"is_active":   s.paystackSecretKey != "",
		},
		{
			"code":        "flutterwave",
			"name":        "Flutterwave",
			"description": "Pay with card, bank transfer, or mobile money",
			"logo":        "/static/images/payment/flutterwave.png",
			"is_active":   s.flutterwaveSecretKey != "",
		},
	}

	return methods
}

// RefundPayment processes payment refunds
func (s *PaymentService) RefundPayment(ctx context.Context, reference, gateway string, amount int64, reason string) error {
	switch gateway {
	case "paystack":
		return s.refundPaystack(ctx, reference, amount, reason)
	case "flutterwave":
		return s.refundFlutterwave(ctx, reference, amount, reason)
	default:
		return fmt.Errorf("unsupported payment gateway: %s", gateway)
	}
}

// refundPaystack processes Paystack refunds
func (s *PaymentService) refundPaystack(ctx context.Context, reference string, amount int64, reason string) error {
	refundReq := map[string]interface{}{
		"transaction": reference,
		"amount":      amount,
		"currency":    "NGN",
		"reason":      reason,
	}

	jsonData, err := json.Marshal(refundReq)
	if err != nil {
		return fmt.Errorf("failed to marshal refund request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.paystack.co/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create refund request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("refund request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("refund failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// refundFlutterwave processes Flutterwave refunds
func (s *PaymentService) refundFlutterwave(ctx context.Context, reference string, amount int64, reason string) error {
	// First, get transaction ID from reference
	// Then process refund using Flutterwave API
	// Implementation depends on Flutterwave's refund API structure
	return fmt.Errorf("Flutterwave refunds not implemented yet")
}

// ProcessTransfer processes bank transfer for affiliate payouts
func (s *PaymentService) ProcessTransfer(ctx context.Context, transferRequest map[string]interface{}) (map[string]interface{}, error) {
	// Paystack Transfer API
	url := "https://api.paystack.co/transfer"

	// Prepare transfer data
	transferData := map[string]interface{}{
		"source":    "balance",
		"amount":    transferRequest["amount"],
		"recipient": transferRequest["recipient_code"], // This should be created first
		"reason":    transferRequest["narration"],
		"reference": transferRequest["reference"],
	}

	// If recipient_code not provided, create recipient first
	if _, exists := transferRequest["recipient_code"]; !exists {
		recipientCode, err := s.createTransferRecipient(ctx, transferRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to create recipient: %v", err)
		}
		transferData["recipient"] = recipientCode
	}

	jsonData, err := json.Marshal(transferData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transfer data: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transfer failed: %s", response["message"])
	}

	return response, nil
}

// createTransferRecipient creates a transfer recipient for bank transfers
func (s *PaymentService) createTransferRecipient(ctx context.Context, transferRequest map[string]interface{}) (string, error) {
	url := "https://api.paystack.co/transferrecipient"

	recipientData := map[string]interface{}{
		"type":           "nuban",
		"name":           transferRequest["account_name"],
		"account_number": transferRequest["account_number"],
		"bank_code":      transferRequest["bank_code"],
		"currency":       "NGN",
	}

	jsonData, err := json.Marshal(recipientData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal recipient data: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("recipient creation failed: %s", response["message"])
	}

	// Extract recipient code
	if data, ok := response["data"].(map[string]interface{}); ok {
		if recipientCode, ok := data["recipient_code"].(string); ok {
			return recipientCode, nil
		}
	}

	return "", fmt.Errorf("failed to get recipient code from response")
}

// VerifyBankAccount verifies bank account details before transfer
func (s *PaymentService) VerifyBankAccount(ctx context.Context, accountNumber, bankCode string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.paystack.co/bank/resolve?account_number=%s&bank_code=%s", accountNumber, bankCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account verification failed: %s", response["message"])
	}

	return response, nil
}

// GetBankList retrieves list of supported banks
func (s *PaymentService) GetBankList(ctx context.Context) ([]map[string]interface{}, error) {
	url := "https://api.paystack.co/bank"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get bank list: %s", response["message"])
	}

	// Extract bank list from response
	if data, ok := response["data"].([]interface{}); ok {
		banks := make([]map[string]interface{}, len(data))
		for i, bank := range data {
			if bankMap, ok := bank.(map[string]interface{}); ok {
				banks[i] = bankMap
			}
		}
		return banks, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

// IsPaymentProcessed checks if a payment reference has already been processed
// This is critical for idempotency - prevents duplicate processing of webhook events
func (s *PaymentService) IsPaymentProcessed(ctx context.Context, reference string) bool {
	if reference == "" {
		return false
	}
	
	// Check if payment log exists with this reference and status "completed"
	paymentLog, err := s.paymentRepo.GetByReference(ctx, reference)
	if err != nil {
		// If error or not found, assume not processed
		return false
	}
	
	// Check if payment is in a final state (completed or failed)
	// If completed, it's been processed
	// If failed, we should not process again
	if paymentLog.EventType == "completed" || paymentLog.EventType == "failed" {
		return true
	}
	
	return false
}
