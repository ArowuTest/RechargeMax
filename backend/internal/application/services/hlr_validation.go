package services

import (
	"context"
	"fmt"
	"time"

	"rechargemax/internal/utils"
)

// NetworkValidationResult contains the result of network validation
type NetworkValidationResult struct {
	MSISDN           string     `json:"msisdn"`
	SelectedNetwork  string     `json:"selected_network"`
	ActualNetwork    string     `json:"actual_network"`
	IsValid          bool       `json:"is_valid"`
	ValidationSource string     `json:"validation_source"` // 'hlr_api', 'user_selection', 'cache'
	Confidence       string     `json:"confidence"`        // 'high', 'medium', 'low'
	Message          string     `json:"message"`
	CachedNetwork    *string    `json:"cached_network,omitempty"`
	LastRecharged    *time.Time `json:"last_recharged,omitempty"`
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

	// Only trust hlr_api and user_selection sourced entries
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

// ValidateNetworkSelection validates the user's network selection.
//
// Strategy (in order of priority):
//  1. If Termii HLR API is available and working, use it to validate the selection.
//  2. Otherwise, trust the user's explicit selection completely.
//     Prefix-based detection is NOT used — it is unreliable for ported numbers.
//
// This means: if the user selects MTN for their number, we accept MTN.
// When Termii is live, we will cross-check and warn if there's a mismatch.
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

	// Step 1: Try HLR API validation (highest confidence) — only when Termii is configured and reachable
	hlrResult, err := s.lookupViaHLR(ctx, msisdn)
	if err == nil && hlrResult != nil {
		result.ActualNetwork = hlrResult.Network
		result.ValidationSource = "hlr_api"
		result.Confidence = "high"

		if hlrResult.Network == selectedNetwork {
			result.IsValid = true
			result.Message = fmt.Sprintf("Network validated via HLR: %s", selectedNetwork)
			// Cache the validated result
			s.saveHLRResult(ctx, msisdn, hlrResult.Network, "termii", hlrResult)
		} else {
			result.IsValid = false
			result.Message = fmt.Sprintf("Network mismatch: selected %s but number belongs to %s", selectedNetwork, hlrResult.Network)
		}
		return result, nil
	}

	// Step 2: HLR unavailable — trust the user's explicit network selection.
	// Do NOT use prefix-based detection: Nigerian number portability makes prefixes unreliable.
	result.ActualNetwork = selectedNetwork
	result.ValidationSource = "user_selection"
	result.Confidence = "medium"
	result.IsValid = true
	result.Message = fmt.Sprintf("Network accepted: %s (HLR API unavailable — user selection trusted)", selectedNetwork)

	// Persist the user selection to cache so future requests are faster
	s.saveUserSelection(ctx, msisdn, selectedNetwork)

	return result, nil
}

// ValidateAndDetectNetwork combines cache lookup and validation.
// Used in the recharge flow:
//  1. If user selects a network, validate/trust it.
//  2. If no selection, check cache for a previous trusted result.
//  3. If no cache, require the user to select a network.
func (s *HLRService) ValidateAndDetectNetwork(ctx context.Context, msisdn string, userSelectedNetwork *string) (*NetworkValidationResult, error) {
	// Normalise MSISDN to canonical international format (234...) at the service boundary
	if normalized, err := utils.NormalizeMSISDN(msisdn); err == nil {
		msisdn = normalized
	}

	// If user selected a network, validate (or trust) it
	if userSelectedNetwork != nil && *userSelectedNetwork != "" {
		return s.ValidateNetworkSelection(ctx, msisdn, *userSelectedNetwork)
	}

	// No user selection — try to get from cache
	cachedResult, err := s.GetCachedNetworkForUser(ctx, msisdn)
	if err == nil && cachedResult != nil {
		cachedResult.SelectedNetwork = cachedResult.ActualNetwork
		cachedResult.IsValid = true
		return cachedResult, nil
	}

	// No cache and no user selection — require user to select
	return nil, fmt.Errorf("network selection required: please select your network provider")
}
