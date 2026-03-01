package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rechargemax/internal/utils"
)

// NetworkValidationResult contains the result of network validation
type NetworkValidationResult struct {
	MSISDN          string    `json:"msisdn"`
	SelectedNetwork string    `json:"selected_network"`
	ActualNetwork   string    `json:"actual_network"`
	IsValid         bool      `json:"is_valid"`
	ValidationSource string   `json:"validation_source"` // 'hlr_api', 'prefix', 'cache'
	Confidence      string    `json:"confidence"`        // 'high', 'medium', 'low'
	Message         string    `json:"message"`
	CachedNetwork   *string   `json:"cached_network,omitempty"`
	LastRecharged   *time.Time `json:"last_recharged,omitempty"`
}

// GetCachedNetworkForUser retrieves the network from recent successful recharges (last 7-30 days)
// This is used to auto-suggest network for returning users
func (s *HLRService) GetCachedNetworkForUser(ctx context.Context, msisdn string) (*NetworkValidationResult, error) {
	// Normalize phone to international format for cache lookup
	normalizedMSISDN := normalizeToInternational(msisdn)
	
	// Look for valid cache entry from last 30 days
	cache, err := s.networkCacheRepo.FindValidCache(ctx, normalizedMSISDN)
	if err != nil {
		return nil, err // No cache found
	}
	
	// Check if cache is from a successful recharge (not just prefix detection)
	if cache.LookupSource == "prefix_fallback" {
		return nil, fmt.Errorf("only prefix-based cache available, not reliable")
	}
	
	// Check if cache is recent (last 30 days)
	daysSinceLastRecharge := time.Since(cache.LastVerified).Hours() / 24
	if daysSinceLastRecharge > 30 {
		return nil, fmt.Errorf("cache too old: %d days", int(daysSinceLastRecharge))
	}
	
	return &NetworkValidationResult{
		MSISDN:           msisdn,
		ActualNetwork:    cache.Network,
		ValidationSource: "cache",
		Confidence:       s.getConfidenceLevel(cache.LookupSource),
		CachedNetwork:    &cache.Network,
		LastRecharged:    &cache.LastVerified,
		Message:          fmt.Sprintf("Last recharged on %s (%d days ago)", cache.Network, int(daysSinceLastRecharge)),
	}, nil
}

// ValidateNetworkSelection validates that the selected network matches the actual network
// This is called BEFORE showing data plans or processing payment
func (s *HLRService) ValidateNetworkSelection(ctx context.Context, msisdn string, selectedNetwork string) (*NetworkValidationResult, error) {
	// Normalise MSISDN to canonical international format (234...) at the service boundary
	if normalized, err := utils.NormalizeMSISDN(msisdn); err == nil {
		msisdn = normalized
	}
	result := &NetworkValidationResult{
		MSISDN:          msisdn,
		SelectedNetwork: selectedNetwork,
		IsValid:         false,
	}
	
	// Step 1: Try HLR API validation (highest confidence)
	hlrResult, err := s.lookupViaHLR(ctx, msisdn)
	if err == nil && hlrResult != nil {
		result.ActualNetwork = hlrResult.Network
		result.ValidationSource = "hlr_api"
		result.Confidence = "high"
		
		if strings.EqualFold(hlrResult.Network, selectedNetwork) {
			result.IsValid = true
			result.Message = fmt.Sprintf("Network validated: %s", selectedNetwork)
			
			// Save successful validation to cache
			s.saveHLRResult(ctx, msisdn, hlrResult.Network, "termii", hlrResult)
		} else {
			result.IsValid = false
			result.Message = fmt.Sprintf("Network mismatch: Selected %s but number belongs to %s", selectedNetwork, hlrResult.Network)
		}
		
		return result, nil
	}
	
	// Step 2: Fallback to prefix validation if HLR API fails
	prefixResult := s.detectByPrefix(msisdn)
	if prefixResult != nil {
		result.ActualNetwork = prefixResult.Network
		result.ValidationSource = "prefix"
		result.Confidence = "medium"
		
		if strings.EqualFold(prefixResult.Network, selectedNetwork) {
			result.IsValid = true
			result.Message = fmt.Sprintf("Network validated via prefix: %s", selectedNetwork)
			
			// Save prefix validation to cache (shorter TTL)
			s.savePrefixDetection(ctx, msisdn, prefixResult.Network)
		} else {
			result.IsValid = false
			result.Message = fmt.Sprintf("Prefix mismatch: Selected %s but prefix suggests %s", selectedNetwork, prefixResult.Network)
		}
		
		return result, nil
	}
	
	// Step 3: Accept user selection when both HLR and prefix validation unavailable
	// This handles cases where:
	// - HLR API is down/unavailable
	// - Number has been ported to different network (prefix no longer reliable)
	// - New number prefixes not yet in our database
	result.ActualNetwork = selectedNetwork
	result.ValidationSource = "user_selection"
	result.Confidence = "low"
	result.IsValid = true
	result.Message = fmt.Sprintf("Network accepted based on user selection: %s (HLR API unavailable, prefix validation unavailable)", selectedNetwork)
	
	// Save user selection to cache with short TTL (7 days)
	// Will be updated if recharge succeeds
	s.savePrefixDetection(ctx, msisdn, selectedNetwork)
	
	return result, nil
}

// ValidateAndDetectNetwork combines cache lookup and validation
// Used in the recharge flow:
// 1. Check cache for auto-suggestion
// 2. If user selects network, validate it
// 3. Return validated network or error
func (s *HLRService) ValidateAndDetectNetwork(ctx context.Context, msisdn string, userSelectedNetwork *string) (*NetworkValidationResult, error) {
	// Normalise MSISDN to canonical international format (234...) at the service boundary
	if normalized, err := utils.NormalizeMSISDN(msisdn); err == nil {
		msisdn = normalized
	}
	// If user selected a network, validate it
	if userSelectedNetwork != nil && *userSelectedNetwork != "" {
		return s.ValidateNetworkSelection(ctx, msisdn, *userSelectedNetwork)
	}
	
	// No user selection - try to get from cache
	cachedResult, err := s.GetCachedNetworkForUser(ctx, msisdn)
	if err == nil && cachedResult != nil {
		// Return cached network as suggestion
		cachedResult.SelectedNetwork = cachedResult.ActualNetwork
		cachedResult.IsValid = true
		return cachedResult, nil
	}
	
	// No cache and no user selection - require user to select
	return nil, fmt.Errorf("network selection required: no recent recharge history found")
}
