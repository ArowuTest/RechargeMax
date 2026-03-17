package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// TelecomServiceIntegrated handles VTU operations with dynamic provider switching
type TelecomServiceIntegrated struct {
	db            *sql.DB
	vtpassService *VTPassService
}

// ProviderConfig represents a provider configuration from database
type ProviderConfig struct {
	ID           int64
	Network      string
	ServiceType  string
	ProviderMode string
	ProviderName string
	Priority     int
	Config       map[string]interface{}
}

// VTUResponse represents a unified response from any VTU provider
type VTUResponse struct {
	Success           bool
	Status            string // COMPLETED, PROCESSING, FAILED, REVERSED
	ProviderReference string
	VTPassRequestID   string // The request_id we sent to VTPass — used for requery when ProviderReference is empty
	Message           string
	Commission        float64
	RawResponse       map[string]interface{}
	ProviderMode      string
	ProviderName      string
	ResponseTimeMs    int64
}

// NewTelecomServiceIntegrated creates a new integrated telecom service
func NewTelecomServiceIntegrated(db *sql.DB) *TelecomServiceIntegrated {
	return &TelecomServiceIntegrated{
		db: db,
	}
}

// PurchaseAirtime purchases airtime using the configured provider
func (s *TelecomServiceIntegrated) PurchaseAirtime(ctx context.Context, network, phone string, amountKobo int) (*VTUResponse, error) {
	// Get active provider configuration
	providerConfig, err := s.getActiveProvider(ctx, network, "AIRTIME")
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config: %w", err)
	}

	// Convert kobo to naira for provider
	amountNaira := amountKobo / 100

	// Route to appropriate provider based on mode
	startTime := time.Now()
	var response *VTUResponse

	switch providerConfig.ProviderMode {
	case "VTU":
		response, err = s.purchaseAirtimeVTU(ctx, providerConfig, network, phone, amountNaira)
	case "DIRECT":
		response, err = s.purchaseAirtimeDirect(ctx, providerConfig, network, phone, amountNaira)
	case "SIMULATION":
		response, err = s.purchaseAirtimeSimulation(ctx, providerConfig, network, phone, amountNaira)
	default:
		return nil, fmt.Errorf("unsupported provider mode: %s", providerConfig.ProviderMode)
	}

	if err != nil {
		return nil, err
	}

	// Calculate response time
	response.ResponseTimeMs = time.Since(startTime).Milliseconds()
	response.ProviderMode = providerConfig.ProviderMode
	response.ProviderName = providerConfig.ProviderName

	return response, nil
}

// PurchaseData purchases data using the configured provider
func (s *TelecomServiceIntegrated) PurchaseData(ctx context.Context, network, phone, variationCode string, amountKobo int) (*VTUResponse, error) {
	// Get active provider configuration
	providerConfig, err := s.getActiveProvider(ctx, network, "DATA")
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config: %w", err)
	}

	// Convert kobo to naira for provider
	amountNaira := amountKobo / 100

	// Route to appropriate provider based on mode
	startTime := time.Now()
	var response *VTUResponse

	switch providerConfig.ProviderMode {
	case "VTU":
		response, err = s.purchaseDataVTU(ctx, providerConfig, network, phone, variationCode, amountNaira)
	case "DIRECT":
		response, err = s.purchaseDataDirect(ctx, providerConfig, network, phone, variationCode, amountNaira)
	case "SIMULATION":
		response, err = s.purchaseDataSimulation(ctx, providerConfig, network, phone, variationCode, amountNaira)
	default:
		return nil, fmt.Errorf("unsupported provider mode: %s", providerConfig.ProviderMode)
	}

	if err != nil {
		return nil, err
	}

	// Calculate response time
	response.ResponseTimeMs = time.Since(startTime).Milliseconds()
	response.ProviderMode = providerConfig.ProviderMode
	response.ProviderName = providerConfig.ProviderName

	return response, nil
}

// getActiveProvider retrieves the active provider configuration for a network and service type
func (s *TelecomServiceIntegrated) getActiveProvider(ctx context.Context, network, serviceType string) (*ProviderConfig, error) {
	// Try to get from database first
	query := `
		SELECT id, network, service_type, provider_mode, provider_name, priority, config
		FROM get_active_provider($1, $2)
	`

	var config ProviderConfig
	var configJSON []byte

	err := s.db.QueryRowContext(ctx, query, network, serviceType).Scan(
		&config.ID,
		&config.Network,
		&config.ServiceType,
		&config.ProviderMode,
		&config.ProviderName,
		&config.Priority,
		&configJSON,
	)

	if err == sql.ErrNoRows {
		// FALLBACK: No provider configured in database, use environment-based VTPass
		log.Printf("⚠️  No provider configured for %s/%s, using environment fallback\n", network, serviceType)
		return s.envFallbackProvider(network, serviceType), nil
	}

	if err != nil {
		// FALLBACK: DB function may not exist yet (migrations pending) or any other DB error.
		// Fall back to environment-based VTPass so recharges still work.
		errMsg := err.Error()
		if strings.Contains(errMsg, "does not exist") ||
			strings.Contains(errMsg, "42883") ||
			strings.Contains(errMsg, "42P01") ||
			strings.Contains(errMsg, "no rows") {
			log.Printf("⚠️  get_active_provider DB error (%s) for %s/%s, using environment fallback\n", errMsg, network, serviceType)
			return s.envFallbackProvider(network, serviceType), nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Parse config JSON
	if err := json.Unmarshal(configJSON, &config.Config); err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	log.Printf("✅ Using provider: %s (mode: %s) for %s/%s\n", config.ProviderName, config.ProviderMode, network, serviceType)
	return &config, nil
}

// purchaseAirtimeVTU handles airtime purchase via VTU aggregator (VTPass)
func (s *TelecomServiceIntegrated) purchaseAirtimeVTU(ctx context.Context, providerConfig *ProviderConfig, network, phone string, amountNaira int) (*VTUResponse, error) {
	// Always re-initialize from latest env vars (credentials may have been updated without restart).
	// This is cheap — it just reads env vars and builds a struct.
	s.vtpassService = s.initializeVTPassService(providerConfig.Config)

	// Normalize network to uppercase for VTPass
	networkUpper := strings.ToUpper(network)
	
	// Call VTPass API
	vtpassResp, err := s.vtpassService.PurchaseAirtime(ctx, networkUpper, phone, amountNaira)
	if err != nil {
		return &VTUResponse{
			Success: false,
			Status:  "FAILED",
			Message: fmt.Sprintf("VTPass API error: %v", err),
		}, nil // Don't return error, return failed response
	}

	// Convert VTPass response to unified response
	return &VTUResponse{
		Success:           vtpassResp.IsSuccessful(),
		Status:            vtpassResp.GetStatus(),
		ProviderReference: vtpassResp.GetProviderReference(),
		VTPassRequestID:   vtpassResp.RequestID, // echo'd back by VTPass; used for requery
		Message:           vtpassResp.ResponseDescription,
		Commission:        vtpassResp.GetCommission(),
		RawResponse:       vtpassResp.RawResponse,
	}, nil
}

// purchaseDataVTU handles data purchase via VTU aggregator (VTPass)
func (s *TelecomServiceIntegrated) purchaseDataVTU(ctx context.Context, providerConfig *ProviderConfig, network, phone, variationCode string, amountNaira int) (*VTUResponse, error) {
	// Always re-initialize from latest env vars.
	s.vtpassService = s.initializeVTPassService(providerConfig.Config)

	// Normalize network to uppercase for VTPass
	networkUpper := strings.ToUpper(network)
	
	// Call VTPass API
	vtpassResp, err := s.vtpassService.PurchaseData(ctx, networkUpper, phone, variationCode, amountNaira)
	if err != nil {
		return &VTUResponse{
			Success: false,
			Status:  "FAILED",
			Message: fmt.Sprintf("VTPass API error: %v", err),
		}, nil
	}

	// Convert VTPass response to unified response
	return &VTUResponse{
		Success:           vtpassResp.IsSuccessful(),
		Status:            vtpassResp.GetStatus(),
		ProviderReference: vtpassResp.GetProviderReference(),
		VTPassRequestID:   vtpassResp.RequestID, // echo'd back by VTPass; used for requery
		Message:           vtpassResp.ResponseDescription,
		Commission:        vtpassResp.GetCommission(),
		RawResponse:       vtpassResp.RawResponse,
	}, nil
}

// purchaseAirtimeDirect handles airtime purchase via direct network API.
// Requires a signed carrier partnership and API credentials loaded in providerConfig.Config.
// Until partnerships are active, this will always return an error causing PurchaseAirtime
// to fall back to the VTPass path.
func (s *TelecomServiceIntegrated) purchaseAirtimeDirect(ctx context.Context, providerConfig *ProviderConfig, network, phone string, amountNaira int) (*VTUResponse, error) {
	return nil, fmt.Errorf("direct %s integration requires a signed carrier partnership — use VTPass path instead", network)
}

// purchaseDataDirect handles data purchase via direct network API.
// Requires a signed carrier partnership — falls back to VTPass until then.
func (s *TelecomServiceIntegrated) purchaseDataDirect(ctx context.Context, providerConfig *ProviderConfig, network, phone, variationCode string, amountNaira int) (*VTUResponse, error) {
	return nil, fmt.Errorf("direct %s data integration requires a signed carrier partnership — use VTPass path instead", network)
}

// purchaseAirtimeSimulation handles airtime purchase in simulation mode
func (s *TelecomServiceIntegrated) purchaseAirtimeSimulation(ctx context.Context, providerConfig *ProviderConfig, network, phone string, amountNaira int) (*VTUResponse, error) {
	// Simulate API delay
	time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)

	// Get success rate from config (default 95%)
	successRate := 0.95
	if rate, ok := providerConfig.Config["success_rate"].(float64); ok {
		successRate = rate
	}

	// Simulate success/failure
	success := rand.Float64() < successRate

	if success {
		return &VTUResponse{
			Success:           true,
			Status:            "SUCCESS",
			ProviderReference: fmt.Sprintf("SIM-%d", time.Now().Unix()),
			Message:           "Simulation: Transaction successful",
			Commission:        float64(amountNaira) * 0.035, // 3.5% commission
			RawResponse: map[string]interface{}{
				"simulation": true,
				"network":    network,
				"phone":      phone,
				"amount":     amountNaira,
			},
		}, nil
	}

	return &VTUResponse{
		Success:           false,
		Status:            "FAILED",
		ProviderReference: fmt.Sprintf("SIM-%d", time.Now().Unix()),
		Message:           "Simulation: Transaction failed",
		RawResponse: map[string]interface{}{
			"simulation": true,
			"error":      "Simulated failure",
		},
	}, nil
}

// purchaseDataSimulation handles data purchase in simulation mode
func (s *TelecomServiceIntegrated) purchaseDataSimulation(ctx context.Context, providerConfig *ProviderConfig, network, phone, variationCode string, amountNaira int) (*VTUResponse, error) {
	// Simulate API delay
	time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)

	// Get success rate from config (default 95%)
	successRate := 0.95
	if rate, ok := providerConfig.Config["success_rate"].(float64); ok {
		successRate = rate
	}

	// Simulate success/failure
	success := rand.Float64() < successRate

	if success {
		return &VTUResponse{
			Success:           true,
			Status:            "SUCCESS",
			ProviderReference: fmt.Sprintf("SIM-%d", time.Now().Unix()),
			Message:           "Simulation: Data purchase successful",
			Commission:        float64(amountNaira) * 0.03, // 3% commission for data
			RawResponse: map[string]interface{}{
				"simulation":     true,
				"network":        network,
				"phone":          phone,
				"variation_code": variationCode,
				"amount":         amountNaira,
			},
		}, nil
	}

	return &VTUResponse{
		Success:           false,
		Status:            "FAILED",
		ProviderReference: fmt.Sprintf("SIM-%d", time.Now().Unix()),
		Message:           "Simulation: Data purchase failed",
		RawResponse: map[string]interface{}{
			"simulation": true,
			"error":      "Simulated failure",
		},
	}, nil
}

// envFallbackProvider returns a default VTPass provider config using environment variables.
// Used when the database get_active_provider function is missing or returns no rows.
func (s *TelecomServiceIntegrated) envFallbackProvider(network, serviceType string) *ProviderConfig {
	return &ProviderConfig{
		ID:           0,
		Network:      network,
		ServiceType:  serviceType,
		ProviderMode: "VTU",
		ProviderName: "VTPass",
		Priority:     1,
		Config: map[string]interface{}{
			"mode": "sandbox",
		},
	}
}

// initializeVTPassService creates a VTPass service from config
func (s *TelecomServiceIntegrated) initializeVTPassService(config map[string]interface{}) *VTPassService {
	// Try to get from config first, fallback to environment variables
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" {
		apiKey = os.Getenv("VTPASS_API_KEY")
	}
	
	publicKey, _ := config["public_key"].(string)
	if publicKey == "" {
		publicKey = os.Getenv("VTPASS_PUBLIC_KEY")
	}
	
	secretKey, _ := config["secret_key"].(string)
	if secretKey == "" {
		secretKey = os.Getenv("VTPASS_SECRET_KEY")
	}
	
	isSandbox, _ := config["is_sandbox"].(bool)
	if !isSandbox {
		// Check all possible sandbox env var names (order of precedence):
		// 1. VTPASS_SANDBOX_MODE=true  (Render/render.yaml style)
		// 2. VTPASS_MODE=sandbox        (legacy style)
		// 3. config["mode"]=="sandbox"  (DB provider_configs style)
		sandboxMode := os.Getenv("VTPASS_SANDBOX_MODE")
		mode := os.Getenv("VTPASS_MODE")
		configMode, _ := config["mode"].(string)
		isSandbox = sandboxMode == "true" || sandboxMode == "1" ||
			mode == "sandbox" ||
			configMode == "sandbox"
	}

	return NewVTPassService(VTPassConfig{
		APIKey:    apiKey,
		PublicKey: publicKey,
		SecretKey: secretKey,
		IsSandbox: isSandbox,
	})
}

// LogProviderTransaction logs a provider transaction to the database
func (s *TelecomServiceIntegrated) LogProviderTransaction(ctx context.Context, transactionID int64, providerConfigID int64, response *VTUResponse, requestPayload map[string]interface{}) error {
	requestJSON, _ := json.Marshal(requestPayload)
	responseJSON, _ := json.Marshal(response.RawResponse)

	query := `
		SELECT log_provider_transaction($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	var logID int64
	err := s.db.QueryRowContext(
		ctx,
		query,
		transactionID,
		providerConfigID,
		response.ProviderMode,
		response.ProviderName,
		requestJSON,
		responseJSON,
		response.Status,
		response.Message,
		response.ResponseTimeMs,
	).Scan(&logID)

	if err != nil {
		return fmt.Errorf("failed to log provider transaction: %w", err)
	}

	return nil
}

// QueryTransactionStatus re-checks a pending VTPass transaction by provider reference.
// Returns a normalised status string: "SUCCESS", "FAILED", "PENDING", or "PROCESSING".
func (s *TelecomServiceIntegrated) QueryTransactionStatus(ctx context.Context, providerRef string) (string, error) {
	if s.vtpassService == nil {
		return "PENDING", nil // nothing to query
	}
	status, err := s.vtpassService.QueryTransaction(ctx, providerRef)
	if err != nil {
		return "PENDING", fmt.Errorf("VTPass query: %w", err)
	}
	return status, nil
}
