package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/datatypes"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// HLRService handles network detection via HLR lookup
type HLRService struct {
	networkCacheRepo repositories.NetworkCacheRepository
	termiiAPIKey     string
	cacheTTLDays     int
	httpClient       *http.Client
}

// NewHLRService creates a new HLR service instance
func NewHLRService(
	networkCacheRepo repositories.NetworkCacheRepository,
	termiiAPIKey string,
) *HLRService {
	return &HLRService{
		networkCacheRepo: networkCacheRepo,
		termiiAPIKey:     termiiAPIKey,
		cacheTTLDays:     60, // 60-day cache TTL
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NetworkDetectionResult contains the result of network detection
type NetworkDetectionResult struct {
	MSISDN       string
	Network      string
	Source       string // 'hlr_api', 'cache', 'user_selection', 'prefix_fallback'
	Confidence   string // 'high', 'medium', 'low'
	CachedUntil  *time.Time
	ErrorMessage string
}

// DetectNetwork detects the network for a given MSISDN
func (s *HLRService) DetectNetwork(ctx context.Context, msisdn string, userSelectedNetwork *string) (*NetworkDetectionResult, error) {
	// 1. Check cache first
	cachedResult, err := s.getCachedNetwork(ctx, msisdn)
	if err == nil && cachedResult != nil {
		return cachedResult, nil
	}

	// 2. If user provided network selection, use it
	if userSelectedNetwork != nil && *userSelectedNetwork != "" {
		return s.saveUserSelection(ctx, msisdn, *userSelectedNetwork)
	}

	// 3. Try HLR API lookup
	hlrResult, err := s.lookupViaHLR(ctx, msisdn)
	if err == nil && hlrResult != nil {
		return hlrResult, nil
	}

	// 4. Fallback to prefix-based detection
	prefixResult := s.detectByPrefix(msisdn)
	if prefixResult != nil {
		return s.savePrefixDetection(ctx, msisdn, prefixResult.Network)
	}

	return nil, errors.New("unable to detect network for MSISDN")
}

// getCachedNetwork retrieves network from cache if valid
func (s *HLRService) getCachedNetwork(ctx context.Context, msisdn string) (*NetworkDetectionResult, error) {
	// Normalize phone to international format for cache lookup
	normalizedMSISDN := normalizeToInternational(msisdn)
	cache, err := s.networkCacheRepo.FindValidCache(ctx, normalizedMSISDN)
	if err != nil {
		return nil, err
	}

	return &NetworkDetectionResult{
		MSISDN:      msisdn,
		Network:     cache.Network,
		Source:      "cache",
		Confidence:  s.getConfidenceLevel(cache.LookupSource),
		CachedUntil: &cache.CacheExpires,
	}, nil
}

// lookupViaHLR performs HLR lookup via Termii API
func (s *HLRService) lookupViaHLR(ctx context.Context, msisdn string) (*NetworkDetectionResult, error) {
	if s.termiiAPIKey == "" {
		return nil, errors.New("Termii API key not configured")
	}

	// Termii HLR Lookup API endpoint
	url := fmt.Sprintf("https://api.ng.termii.com/api/check/dnd?api_key=%s&phone_number=%s", s.termiiAPIKey, msisdn)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HLR request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HLR API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HLR API returned status %d: %s", resp.StatusCode, string(body))
	}

	var hlrResponse TermiiHLRResponse
	if err := json.NewDecoder(resp.Body).Decode(&hlrResponse); err != nil {
		return nil, fmt.Errorf("failed to decode HLR response: %w", err)
	}

	// Map Termii network name to our standard format
	network := s.normalizeNetworkName(hlrResponse.Network)
	if network == "" {
		return nil, errors.New("invalid network returned from HLR API")
	}

	// Save to cache
	return s.saveHLRResult(ctx, msisdn, network, "termii", hlrResponse)
}

// saveHLRResult saves HLR lookup result to cache
func (s *HLRService) saveHLRResult(ctx context.Context, msisdn, network, provider string, response interface{}) (*NetworkDetectionResult, error) {
	// Normalize phone to international format (234...) for database storage
	normalizedMSISDN := normalizeToInternational(msisdn)
	
	responseJSON, _ := json.Marshal(response)
	
	now := time.Now()
	cacheExpires := now.AddDate(0, 0, s.cacheTTLDays)

	cache := &entities.NetworkCache{
		MSISDN:       normalizedMSISDN,
		Network:      network,
		LastVerified: now,
		CacheExpires: cacheExpires,
		LookupSource: "hlr_api",
		HLRProvider:  &provider,
		HLRResponse:  datatypes.JSON(responseJSON),
		IsValid:      true,
	}

	// Try to find existing cache entry
	existing, err := s.networkCacheRepo.FindByMSISDN(ctx, normalizedMSISDN)
	if err == nil && existing != nil {
		// Update existing
		cache.ID = existing.ID
		err = s.networkCacheRepo.Update(ctx, cache)
	} else {
		// Create new
		err = s.networkCacheRepo.Create(ctx, cache)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save HLR result to cache: %w", err)
	}

	return &NetworkDetectionResult{
		MSISDN:      msisdn,
		Network:     network,
		Source:      "hlr_api",
		Confidence:  "high",
		CachedUntil: &cacheExpires,
	}, nil
}

// saveUserSelection saves user-selected network to cache
func (s *HLRService) saveUserSelection(ctx context.Context, msisdn, network string) (*NetworkDetectionResult, error) {
	// Normalize phone to international format (234...) for database storage
	normalizedMSISDN := normalizeToInternational(msisdn)
	
	now := time.Now()
	cacheExpires := now.AddDate(0, 0, s.cacheTTLDays)

	cache := &entities.NetworkCache{
		MSISDN:       normalizedMSISDN,
		Network:      network,
		LastVerified: now,
		CacheExpires: cacheExpires,
		LookupSource: "user_selection",
		IsValid:      true,
	}

	existing, err := s.networkCacheRepo.FindByMSISDN(ctx, normalizedMSISDN)
	if err == nil && existing != nil {
		cache.ID = existing.ID
		err = s.networkCacheRepo.Update(ctx, cache)
	} else {
		err = s.networkCacheRepo.Create(ctx, cache)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save user selection: %w", err)
	}

	return &NetworkDetectionResult{
		MSISDN:      msisdn,
		Network:     network,
		Source:      "user_selection",
		Confidence:  "medium",
		CachedUntil: &cacheExpires,
	}, nil
}

// savePrefixDetection saves prefix-based detection to cache
func (s *HLRService) savePrefixDetection(ctx context.Context, msisdn, network string) (*NetworkDetectionResult, error) {
	// Normalize phone to international format (234...) for database storage
	normalizedMSISDN := normalizeToInternational(msisdn)
	
	now := time.Now()
	cacheExpires := now.AddDate(0, 0, 7) // Only 7 days for prefix-based (less reliable)

	cache := &entities.NetworkCache{
		MSISDN:       normalizedMSISDN,
		Network:      network,
		LastVerified: now,
		CacheExpires: cacheExpires,
		LookupSource: "prefix_fallback",
		IsValid:      true,
	}

	existing, err := s.networkCacheRepo.FindByMSISDN(ctx, normalizedMSISDN)
	if err == nil && existing != nil {
		cache.ID = existing.ID
		err = s.networkCacheRepo.Update(ctx, cache)
	} else {
		err = s.networkCacheRepo.Create(ctx, cache)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save prefix detection: %w", err)
	}

	return &NetworkDetectionResult{
		MSISDN:      msisdn,
		Network:     network,
		Source:      "prefix_fallback",
		Confidence:  "low",
		CachedUntil: &cacheExpires,
	}, nil
}

// InvalidateCache invalidates cached network for an MSISDN (called when recharge fails)
func (s *HLRService) InvalidateCache(ctx context.Context, msisdn, reason string) error {
	// Normalize phone to international format for cache lookup
	normalizedMSISDN := normalizeToInternational(msisdn)
	return s.networkCacheRepo.Invalidate(ctx, normalizedMSISDN, reason)
}

// detectByPrefix performs prefix-based network detection (fallback)
func (s *HLRService) detectByPrefix(msisdn string) *NetworkDetectionResult {
	if len(msisdn) < 4 {
		return nil
	}

	prefix := msisdn[:4]
	network := ""

	// MTN prefixes
	mtnPrefixes := []string{"0803", "0806", "0703", "0706", "0813", "0816", "0810", "0814", "0903", "0906", "0913", "0916"}
	for _, p := range mtnPrefixes {
		if prefix == p {
			network = "MTN"
			break
		}
	}

	// Airtel prefixes
	if network == "" {
		airtelPrefixes := []string{"0802", "0808", "0708", "0812", "0701", "0902", "0907", "0901", "0904", "0912"}
		for _, p := range airtelPrefixes {
			if prefix == p {
				network = "Airtel"
				break
			}
		}
	}

	// Glo prefixes
	if network == "" {
		gloPrefixes := []string{"0805", "0807", "0705", "0815", "0811", "0905", "0915"}
		for _, p := range gloPrefixes {
			if prefix == p {
				network = "Glo"
				break
			}
		}
	}

	// 9mobile prefixes
	if network == "" {
		nineMobilePrefixes := []string{"0809", "0817", "0818", "0909", "0908"}
		for _, p := range nineMobilePrefixes {
			if prefix == p {
				network = "9mobile"
				break
			}
		}
	}

	if network == "" {
		return nil
	}

	return &NetworkDetectionResult{
		MSISDN:     msisdn,
		Network:    network,
		Source:     "prefix_fallback",
		Confidence: "low",
	}
}

// normalizeNetworkName converts various network name formats to standard format
func (s *HLRService) normalizeNetworkName(name string) string {
	switch name {
	case "MTN", "mtn", "MTN Nigeria":
		return "MTN"
	case "Airtel", "airtel", "Airtel Nigeria":
		return "Airtel"
	case "Glo", "glo", "Globacom":
		return "Glo"
	case "9mobile", "9Mobile", "Etisalat":
		return "9mobile"
	default:
		return ""
	}
}

// getConfidenceLevel returns confidence level based on lookup source
func (s *HLRService) getConfidenceLevel(source string) string {
	switch source {
	case "hlr_api":
		return "high"
	case "user_selection":
		return "medium"
	case "prefix_fallback":
		return "low"
	default:
		return "unknown"
	}
}

// TermiiHLRResponse represents the response from Termii HLR API
type TermiiHLRResponse struct {
	Number      string `json:"number"`
	Status      string `json:"status"`
	Network     string `json:"network"`
	NetworkCode string `json:"network_code"`
}

// normalizeToInternational converts phone number to international format (234...)
// Accepts: 08031234567 or 2348031234567
// Returns: 2348031234567
func normalizeToInternational(phone string) string {
	// Remove all non-digit characters
	digitsOnly := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digitsOnly += string(char)
		}
	}
	
	// If starts with 0 (local format), replace with 234
	if len(digitsOnly) == 11 && digitsOnly[0] == '0' {
		return "234" + digitsOnly[1:]
	}
	
	// If already in international format, return as-is
	if len(digitsOnly) == 13 && digitsOnly[:3] == "234" {
		return digitsOnly
	}
	
	// Fallback: return as-is (will fail validation)
	return digitsOnly
}
