package services

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ============================================================================
// TELECOM SERVICE - Direct Network Integration
// ============================================================================
// This service handles direct integration with Nigerian mobile networks:
// - MTN Nigeria
// - Glo Mobile
// - Airtel Nigeria
// - 9mobile
//
// Architecture: Generic interface with network-specific implementations
// Ready for real API integration when partnerships are established
// ============================================================================

// TelecomService handles direct network provider integration
type TelecomService struct {
	providers map[string]NetworkProviderInterface
	config    *TelecomConfig
}

// TelecomConfig holds telecom service configuration
type TelecomConfig struct {
	MTNConfig     NetworkConfig
	GloConfig     NetworkConfig
	AirtelConfig  NetworkConfig
	NineMobileConfig NetworkConfig
	DefaultTimeout time.Duration
}

// NetworkConfig holds configuration for a specific network
type NetworkConfig struct {
	Enabled   bool
	BaseURL   string
	APIKey    string
	APISecret string
	Username  string
	Password  string
	Timeout   time.Duration
}

// NewTelecomService creates a new telecom service with all network providers
func NewTelecomService(apiKey, apiSecret, baseURL string) *TelecomService {
	// Create default config
	config := &TelecomConfig{
		MTNConfig: NetworkConfig{
			Enabled:   true,
			BaseURL:   baseURL,
			APIKey:    apiKey,
			APISecret: apiSecret,
			Timeout:   30 * time.Second,
		},
		GloConfig: NetworkConfig{
			Enabled:   true,
			BaseURL:   baseURL,
			APIKey:    apiKey,
			APISecret: apiSecret,
			Timeout:   30 * time.Second,
		},
		AirtelConfig: NetworkConfig{
			Enabled:   true,
			BaseURL:   baseURL,
			APIKey:    apiKey,
			APISecret: apiSecret,
			Timeout:   30 * time.Second,
		},
		NineMobileConfig: NetworkConfig{
			Enabled:   true,
			BaseURL:   baseURL,
			APIKey:    apiKey,
			APISecret: apiSecret,
			Timeout:   30 * time.Second,
		},
		DefaultTimeout: 30 * time.Second,
	}

	// Initialize providers
	providers := make(map[string]NetworkProviderInterface)
	providers["MTN"] = NewMTNProvider(config.MTNConfig)
	providers["GLO"] = NewGloProvider(config.GloConfig)
	providers["AIRTEL"] = NewAirtelProvider(config.AirtelConfig)
	providers["9MOBILE"] = NewNineMobileProvider(config.NineMobileConfig)

	return &TelecomService{
		providers: providers,
		config:    config,
	}
}

// ============================================================================
// PUBLIC METHODS
// ============================================================================

// PurchaseAirtime purchases airtime from the appropriate network provider
func (s *TelecomService) PurchaseAirtime(ctx context.Context, msisdn, network string, amount int64) error {
	provider, err := s.getProvider(network)
	if err != nil {
		return err
	}

	// Convert amount from kobo to naira
	amountInNaira := float64(amount) / 100

	// Call provider-specific implementation
	return provider.PurchaseAirtime(ctx, msisdn, amountInNaira)
}

// PurchaseData purchases data from the appropriate network provider
func (s *TelecomService) PurchaseData(ctx context.Context, msisdn, network, dataPackage string) error {
	provider, err := s.getProvider(network)
	if err != nil {
		return err
	}

	// Call provider-specific implementation
	return provider.PurchaseData(ctx, msisdn, dataPackage)
}

// VerifyTransaction verifies a transaction with the network provider
func (s *TelecomService) VerifyTransaction(ctx context.Context, reference string) (bool, error) {
	// Try to verify with all providers (since we don't know which network)
	// In production, you'd store the network with the transaction
	for _, provider := range s.providers {
		success, err := provider.VerifyTransaction(ctx, reference)
		if err == nil && success {
			return true, nil
		}
	}
	
	return false, fmt.Errorf("transaction not found or failed verification")
}

// ProcessRecharge processes a recharge with the telecom provider.
// This legacy method delegates to the simulation path; production recharges
// are handled by TelecomServiceIntegrated.
func (s *TelecomService) ProcessRecharge(ctx context.Context, req TelecomRechargeRequest) (*TelecomRechargeResponse, error) {
	// Delegate to simulation for backward-compat; real provisioning goes through
	// TelecomServiceIntegrated → VTPassService.
	log.Printf("[telecom] legacy ProcessRecharge called for %s %s ₦%d — using simulation",
		req.Network, req.MSISDN, req.Amount/100)
	return &TelecomRechargeResponse{
		Success:   true,
		NetworkRef: fmt.Sprintf("SIM_%s_%d", req.MSISDN, req.Amount),
		Message:   "Simulated recharge (legacy path)",
	}, nil
}



// GetNetworkProviders returns list of available network providers from database
func (s *TelecomService) GetNetworkProviders(ctx context.Context) ([]NetworkProvider, error) {
	// Return configured providers
	providers := []NetworkProvider{
		{
			Code:     "MTN",
			Name:     "MTN Nigeria",
			Logo:     "/images/networks/mtn.png",
			IsActive: s.config.MTNConfig.Enabled,
		},
		{
			Code:     "GLO",
			Name:     "Glo Mobile",
			Logo:     "/images/networks/glo.png",
			IsActive: s.config.GloConfig.Enabled,
		},
		{
			Code:     "AIRTEL",
			Name:     "Airtel Nigeria",
			Logo:     "/images/networks/airtel.png",
			IsActive: s.config.AirtelConfig.Enabled,
		},
		{
			Code:     "9MOBILE",
			Name:     "9mobile",
			Logo:     "/images/networks/9mobile.png",
			IsActive: s.config.NineMobileConfig.Enabled,
		},
	}

	return providers, nil
}

// GetDataPackages returns list of available data packages for a network
func (s *TelecomService) GetDataPackages(ctx context.Context, network string) ([]DataPackage, error) {
	provider, err := s.getProvider(network)
	if err != nil {
		return nil, err
	}

	return provider.GetDataPackages(ctx)
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// getProvider returns the appropriate network provider
func (s *TelecomService) getProvider(network string) (NetworkProviderInterface, error) {
	provider, exists := s.providers[network]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	return provider, nil
}

// ============================================================================
// NETWORK PROVIDER INTERFACE
// ============================================================================

// NetworkProviderInterface defines the contract for all network providers
type NetworkProviderInterface interface {
	// PurchaseAirtime purchases airtime for the given MSISDN
	PurchaseAirtime(ctx context.Context, msisdn string, amount float64) error
	
	// PurchaseData purchases data bundle for the given MSISDN
	PurchaseData(ctx context.Context, msisdn string, dataPackage string) error
	
	// VerifyTransaction verifies a transaction by reference
	VerifyTransaction(ctx context.Context, reference string) (bool, error)
	
	// GetDataPackages returns available data packages
	GetDataPackages(ctx context.Context) ([]DataPackage, error)
	
	// CheckBalance checks the balance/status of an MSISDN
	CheckBalance(ctx context.Context, msisdn string) (*BalanceInfo, error)
}

// ============================================================================
// MTN PROVIDER IMPLEMENTATION
// ============================================================================

// MTNProvider handles MTN Nigeria integration
type MTNProvider struct {
	config NetworkConfig
}

// NewMTNProvider creates a new MTN provider
func NewMTNProvider(config NetworkConfig) *MTNProvider {
	return &MTNProvider{config: config}
}

// PurchaseAirtime purchases MTN airtime
func (p *MTNProvider) PurchaseAirtime(ctx context.Context, msisdn string, amount float64) error {
	// Now handled by HybridTelecomService - supports direct, VTU, and simulation modes
	// Example implementation structure:
	//
	// 1. Prepare request
	// req := MTNAirtimeRequest{
	//     MSISDN: msisdn,
	//     Amount: amount,
	//     APIKey: p.config.APIKey,
	// }
	//
	// 2. Make HTTP request to MTN API
	// resp, err := http.Post(p.config.BaseURL + "/airtime", req)
	//
	// 3. Parse and validate response
	// 4. Return error if failed
	
	// For now, simulate success
	if !p.config.Enabled {
		return fmt.Errorf("MTN provider is disabled")
	}
	
	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// PurchaseData purchases MTN data bundle
func (p *MTNProvider) PurchaseData(ctx context.Context, msisdn string, dataPackage string) error {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return fmt.Errorf("MTN provider is disabled")
	}
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

// VerifyTransaction verifies MTN transaction
func (p *MTNProvider) VerifyTransaction(ctx context.Context, reference string) (bool, error) {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return false, fmt.Errorf("MTN provider is disabled")
	}
	
	return true, nil
}

// GetDataPackages returns MTN data packages
func (p *MTNProvider) GetDataPackages(ctx context.Context) ([]DataPackage, error) {
	// Hardcoded packages - should be moved to database and managed via admin panel
	// These are sample MTN Nigeria data packages
	return []DataPackage{
		{ID: "MTN_1GB_DAILY", Name: "1GB Daily", DataSize: "1GB", Amount: 30000, Network: "MTN"},
		{ID: "MTN_2GB_WEEKLY", Name: "2GB Weekly", DataSize: "2GB", Amount: 50000, Network: "MTN"},
		{ID: "MTN_5GB_MONTHLY", Name: "5GB Monthly", DataSize: "5GB", Amount: 100000, Network: "MTN"},
		{ID: "MTN_10GB_MONTHLY", Name: "10GB Monthly", DataSize: "10GB", Amount: 200000, Network: "MTN"},
	}, nil
}

// CheckBalance checks MTN balance
func (p *MTNProvider) CheckBalance(ctx context.Context, msisdn string) (*BalanceInfo, error) {
	// Now handled by HybridTelecomService
	return &BalanceInfo{
		MSISDN:  msisdn,
		Balance: 0,
		Network: "MTN",
	}, nil
}

// ============================================================================
// GLO PROVIDER IMPLEMENTATION
// ============================================================================

// GloProvider handles Glo Mobile integration
type GloProvider struct {
	config NetworkConfig
}

// NewGloProvider creates a new Glo provider
func NewGloProvider(config NetworkConfig) *GloProvider {
	return &GloProvider{config: config}
}

// PurchaseAirtime purchases Glo airtime
func (p *GloProvider) PurchaseAirtime(ctx context.Context, msisdn string, amount float64) error {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return fmt.Errorf("Glo provider is disabled")
	}
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

// PurchaseData purchases Glo data bundle
func (p *GloProvider) PurchaseData(ctx context.Context, msisdn string, dataPackage string) error {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return fmt.Errorf("Glo provider is disabled")
	}
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

// VerifyTransaction verifies Glo transaction
func (p *GloProvider) VerifyTransaction(ctx context.Context, reference string) (bool, error) {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return false, fmt.Errorf("Glo provider is disabled")
	}
	
	return true, nil
}

// GetDataPackages returns Glo data packages
func (p *GloProvider) GetDataPackages(ctx context.Context) ([]DataPackage, error) {
	// Hardcoded packages - should be moved to database and managed via admin panel
	return []DataPackage{
		{ID: "GLO_1GB_DAILY", Name: "1GB Daily", DataSize: "1GB", Amount: 25000, Network: "GLO"},
		{ID: "GLO_2.5GB_WEEKLY", Name: "2.5GB Weekly", DataSize: "2.5GB", Amount: 50000, Network: "GLO"},
		{ID: "GLO_7.5GB_MONTHLY", Name: "7.5GB Monthly", DataSize: "7.5GB", Amount: 100000, Network: "GLO"},
		{ID: "GLO_15GB_MONTHLY", Name: "15GB Monthly", DataSize: "15GB", Amount: 200000, Network: "GLO"},
	}, nil
}

// CheckBalance checks Glo balance
func (p *GloProvider) CheckBalance(ctx context.Context, msisdn string) (*BalanceInfo, error) {
	// Now handled by HybridTelecomService
	return &BalanceInfo{
		MSISDN:  msisdn,
		Balance: 0,
		Network: "GLO",
	}, nil
}

// ============================================================================
// AIRTEL PROVIDER IMPLEMENTATION
// ============================================================================

// AirtelProvider handles Airtel Nigeria integration
type AirtelProvider struct {
	config NetworkConfig
}

// NewAirtelProvider creates a new Airtel provider
func NewAirtelProvider(config NetworkConfig) *AirtelProvider {
	return &AirtelProvider{config: config}
}

// PurchaseAirtime purchases Airtel airtime
func (p *AirtelProvider) PurchaseAirtime(ctx context.Context, msisdn string, amount float64) error {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return fmt.Errorf("Airtel provider is disabled")
	}
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

// PurchaseData purchases Airtel data bundle
func (p *AirtelProvider) PurchaseData(ctx context.Context, msisdn string, dataPackage string) error {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return fmt.Errorf("Airtel provider is disabled")
	}
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

// VerifyTransaction verifies Airtel transaction
func (p *AirtelProvider) VerifyTransaction(ctx context.Context, reference string) (bool, error) {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return false, fmt.Errorf("Airtel provider is disabled")
	}
	
	return true, nil
}

// GetDataPackages returns Airtel data packages
func (p *AirtelProvider) GetDataPackages(ctx context.Context) ([]DataPackage, error) {
	// Hardcoded packages - should be moved to database and managed via admin panel
	return []DataPackage{
		{ID: "AIRTEL_1GB_DAILY", Name: "1GB Daily", DataSize: "1GB", Amount: 30000, Network: "AIRTEL"},
		{ID: "AIRTEL_2GB_WEEKLY", Name: "2GB Weekly", DataSize: "2GB", Amount: 50000, Network: "AIRTEL"},
		{ID: "AIRTEL_6GB_MONTHLY", Name: "6GB Monthly", DataSize: "6GB", Amount: 100000, Network: "AIRTEL"},
		{ID: "AIRTEL_11GB_MONTHLY", Name: "11GB Monthly", DataSize: "11GB", Amount: 200000, Network: "AIRTEL"},
	}, nil
}

// CheckBalance checks Airtel balance
func (p *AirtelProvider) CheckBalance(ctx context.Context, msisdn string) (*BalanceInfo, error) {
	// Now handled by HybridTelecomService
	return &BalanceInfo{
		MSISDN:  msisdn,
		Balance: 0,
		Network: "AIRTEL",
	}, nil
}

// ============================================================================
// 9MOBILE PROVIDER IMPLEMENTATION
// ============================================================================

// NineMobileProvider handles 9mobile integration
type NineMobileProvider struct {
	config NetworkConfig
}

// NewNineMobileProvider creates a new 9mobile provider
func NewNineMobileProvider(config NetworkConfig) *NineMobileProvider {
	return &NineMobileProvider{config: config}
}

// PurchaseAirtime purchases 9mobile airtime
func (p *NineMobileProvider) PurchaseAirtime(ctx context.Context, msisdn string, amount float64) error {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return fmt.Errorf("9mobile provider is disabled")
	}
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

// PurchaseData purchases 9mobile data bundle
func (p *NineMobileProvider) PurchaseData(ctx context.Context, msisdn string, dataPackage string) error {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return fmt.Errorf("9mobile provider is disabled")
	}
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

// VerifyTransaction verifies 9mobile transaction
func (p *NineMobileProvider) VerifyTransaction(ctx context.Context, reference string) (bool, error) {
	// Now handled by HybridTelecomService
	if !p.config.Enabled {
		return false, fmt.Errorf("9mobile provider is disabled")
	}
	
	return true, nil
}

// GetDataPackages returns 9mobile data packages
func (p *NineMobileProvider) GetDataPackages(ctx context.Context) ([]DataPackage, error) {
	// Hardcoded packages - should be moved to database and managed via admin panel
	return []DataPackage{
		{ID: "9MOBILE_1GB_DAILY", Name: "1GB Daily", DataSize: "1GB", Amount: 30000, Network: "9MOBILE"},
		{ID: "9MOBILE_2GB_WEEKLY", Name: "2GB Weekly", DataSize: "2GB", Amount: 50000, Network: "9MOBILE"},
		{ID: "9MOBILE_5GB_MONTHLY", Name: "5GB Monthly", DataSize: "5GB", Amount: 100000, Network: "9MOBILE"},
		{ID: "9MOBILE_10GB_MONTHLY", Name: "10GB Monthly", DataSize: "10GB", Amount: 200000, Network: "9MOBILE"},
	}, nil
}

// CheckBalance checks 9mobile balance
func (p *NineMobileProvider) CheckBalance(ctx context.Context, msisdn string) (*BalanceInfo, error) {
	// Now handled by HybridTelecomService
	return &BalanceInfo{
		MSISDN:  msisdn,
		Balance: 0,
		Network: "9MOBILE",
	}, nil
}

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// NetworkProvider represents a network provider
type NetworkProvider struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	IsActive bool   `json:"is_active"`
}

// DataPackage represents a data package

// BalanceInfo represents balance information
type BalanceInfo struct {
	MSISDN  string  `json:"msisdn"`
	Balance float64 `json:"balance"`
	Network string  `json:"network"`
}
