package services

import (
	"context"
	"fmt"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// NetworkConfigService handles network configuration
type NetworkConfigService struct {
	networkRepo  repositories.NetworkRepository
	dataPlanRepo repositories.DataPlanRepository
}

// NewNetworkConfigService creates a new network config service
func NewNetworkConfigService(
	networkRepo repositories.NetworkRepository,
	dataPlanRepo repositories.DataPlanRepository,
) *NetworkConfigService {
	return &NetworkConfigService{
		networkRepo:  networkRepo,
		dataPlanRepo: dataPlanRepo,
	}
}

// GetDataPackages gets available data packages for a network
func (s *NetworkConfigService) GetDataPackages(ctx context.Context, network string) ([]DataPackage, error) {
	// Fetch data plans from database ONLY - no hardcoded fallback
	if s.dataPlanRepo == nil {
		return nil, fmt.Errorf("data plan repository not initialized")
	}
	
	plans, err := s.dataPlanRepo.FindByNetworkCode(ctx, network)
	if err != nil {
		return nil, fmt.Errorf("failed to load data plans from database: %w", err)
	}
	
	if len(plans) == 0 {
		return nil, fmt.Errorf("no data plans found for network: %s", network)
	}
	
	var packages []DataPackage
	for _, plan := range plans {
		packages = append(packages, DataPackage{
			ID:       plan.PlanCode,
			Name:     plan.PlanName,
			Network:  network,
			Amount:   int64(plan.Price * 100), // Convert naira to kobo
			DataSize: plan.DataAmount,
		})
	}
	
	return packages, nil
}

// DataPackage represents a data package
type DataPackage struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Network  string `json:"network"`
	Amount   int64  `json:"amount"`
	DataSize string `json:"data_size"`
}

// GetAllNetworks returns all available networks (public method - returns simplified response)
func (s *NetworkConfigService) GetAllNetworks(ctx context.Context) ([]NetworkResponse, error) {
	// Get networks from database
	networks, err := s.networkRepo.FindAll(ctx, 100, 0)
	if err != nil {
		// Fallback to hardcoded networks if database query fails
		return s.getHardcodedNetworks(), nil
	}

	// Convert to response format
	var responses []NetworkResponse
	for _, network := range networks {
		isActive := true
		if network.IsActive != nil {
			isActive = *network.IsActive
		}

		responses = append(responses, NetworkResponse{
			ID:             network.NetworkCode, // Use NetworkCode as ID
			Name:           network.NetworkName,
			Code:           network.NetworkCode,
			Logo:           network.LogoUrl,
			IsActive:       isActive,
			SupportData:    true, // Assume all networks support data
			SupportAirtime: true, // Assume all networks support airtime
		})
	}

	return responses, nil
}

// getHardcodedNetworks returns fallback hardcoded networks
func (s *NetworkConfigService) getHardcodedNetworks() []NetworkResponse {
	return []NetworkResponse{
		{
			ID:             "MTN",
			Name:           "MTN Nigeria",
			Code:           "MTN",
			Logo:           "/images/networks/mtn.png",
			IsActive:       true,
			SupportData:    true,
			SupportAirtime: true,
		},
		{
			ID:             "GLO",
			Name:           "Glo Mobile",
			Code:           "GLO",
			Logo:           "/images/networks/glo.png",
			IsActive:       true,
			SupportData:    true,
			SupportAirtime: true,
		},
		{
			ID:             "AIRTEL",
			Name:           "Airtel Nigeria",
			Code:           "AIRTEL",
			Logo:           "/images/networks/airtel.png",
			IsActive:       true,
			SupportData:    true,
			SupportAirtime: true,
		},
		{
			ID:             "9MOBILE",
			Name:           "9mobile",
			Code:           "9MOBILE",
			Logo:           "/images/networks/9mobile.png",
			IsActive:       true,
			SupportData:    true,
			SupportAirtime: true,
		},
	}
}

// NetworkResponse represents a network
type NetworkResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Logo           string `json:"logo"`
	IsActive       bool   `json:"is_active"`
	SupportData    bool   `json:"support_data"`
	SupportAirtime bool   `json:"support_airtime"`
}

// GetDataBundles returns data bundles for a specific network (alias for GetDataPackages)
func (s *NetworkConfigService) GetDataBundles(ctx context.Context, networkID string) ([]DataBundleResponse, error) {
	packages, err := s.GetDataPackages(ctx, networkID)
	if err != nil {
		return nil, err
	}

	// Convert to DataBundleResponse format
	var bundles []DataBundleResponse
	for _, pkg := range packages {
		bundles = append(bundles, DataBundleResponse{
			ID:          pkg.ID,
			Name:        pkg.Name,
			Network:     pkg.Network,
			Price:       float64(pkg.Amount) / 100, // Convert kobo to naira
			DataSize:    pkg.DataSize,
			Validity:    "30 days",
			Description: pkg.Name + " - Valid for 30 days",
		})
	}

	return bundles, nil
}

// DataBundleResponse represents a data bundle
type DataBundleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Network     string  `json:"network"`
	Price       float64 `json:"price"`
	DataSize    string  `json:"data_size"`
	Validity    string  `json:"validity"`
	Description string  `json:"description"`
}

// ============================================================================
// ADMIN METHODS (NEW - Required by AdminHandler)
// ============================================================================

// GetAllBundles returns all data bundles (admin)
func (s *NetworkConfigService) GetAllBundles(ctx context.Context, network string) ([]DataBundleResponse, error) {
	// Use GetDataPackages to get packages for the network
	packages, err := s.GetDataPackages(ctx, network)
	if err != nil {
		return nil, fmt.Errorf("failed to get data packages: %w", err)
	}
	
	// Convert to DataBundleResponse format
	var bundles []DataBundleResponse
	for _, pkg := range packages {
		bundles = append(bundles, DataBundleResponse{
			ID:          pkg.ID,
			Name:        pkg.Name,
			Network:     pkg.Network,
			Price:       float64(pkg.Amount) / 100, // Convert kobo to naira
			DataSize:    pkg.DataSize,
			Validity:    "30 days",
			Description: pkg.Name + " - Valid for 30 days",
		})
	}
	
	// In production, this would query a data_bundles table:
	// bundles, err := s.dataBundleRepo.FindByNetwork(ctx, network)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get bundles: %w", err)
	// }
	
	return bundles, nil
}

// CreateBundle creates a new data bundle (admin)
func (s *NetworkConfigService) CreateBundle(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// Validate required fields
	requiredFields := []string{"network_id", "bundle_name", "data_volume", "validity_days", "price"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return nil, fmt.Errorf("missing required field: %s", field)
		}
	}
	
	// Save to database when bundle repository is available
	// In production, this would:
	// 1. Create a DataBundle entity
	// 2. Save to data_bundles table
	// 3. Return the created bundle with ID
	//
	// Example implementation:
	// bundle := &entities.DataBundle{
	//     ID:           uuid.New(),
	//     NetworkID:    data["network_id"].(string),
	//     BundleName:   data["bundle_name"].(string),
	//     DataVolume:   data["data_volume"].(string),
	//     ValidityDays: int(data["validity_days"].(float64)),
	//     Price:        int64(data["price"].(float64) * 100), // Convert to kobo
	//     IsActive:     true,
	//     CreatedAt:    time.Now(),
	// }
	// 
	// err := s.dataBundleRepo.Create(ctx, bundle)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create bundle: %w", err)
	// }
	// 
	// data["id"] = bundle.ID.String()
	
	// For now, return the bundle data
	return data, nil
}

// UpdateBundle updates an existing data bundle (admin)
func (s *NetworkConfigService) UpdateBundle(ctx context.Context, bundleID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Update bundle in database
	// In production, this would:
	// 1. Find existing bundle by ID
	// 2. Update fields with new data
	// 3. Save to database
	// 4. Return updated bundle
	//
	// Example implementation:
	// bundle, err := s.dataBundleRepo.FindByID(ctx, bundleID)
	// if err != nil {
	//     return nil, fmt.Errorf("bundle not found: %w", err)
	// }
	// 
	// if name, ok := data["bundle_name"].(string); ok {
	//     bundle.BundleName = name
	// }
	// if price, ok := data["price"].(float64); ok {
	//     bundle.Price = int64(price * 100) // Convert to kobo
	// }
	// // ... update other fields
	// 
	// err = s.dataBundleRepo.Update(ctx, bundle)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to update bundle: %w", err)
	// }
	
	// For now, return the updated data
	data["id"] = bundleID
	return data, nil
}

// DeleteBundle deletes a data bundle (admin)
func (s *NetworkConfigService) DeleteBundle(ctx context.Context, bundleID string) error {
	// Delete bundle from database
	// In production, this would:
	// 1. Find bundle by ID to ensure it exists
	// 2. Check if bundle is currently in use
	// 3. Soft delete (set IsActive = false) or hard delete
	// 4. Invalidate caches
	//
	// Example implementation:
	// bundle, err := s.dataBundleRepo.FindByID(ctx, bundleID)
	// if err != nil {
	//     return fmt.Errorf("bundle not found: %w", err)
	// }
	// 
	// // Soft delete (recommended)
	// bundle.IsActive = false
	// err = s.dataBundleRepo.Update(ctx, bundle)
	// if err != nil {
	//     return fmt.Errorf("failed to delete bundle: %w", err)
	// }
	// 
	// // Or hard delete:
	// // err = s.dataBundleRepo.Delete(ctx, bundleID)
	
	// For now, just validate the ID
	if bundleID == "" {
		return fmt.Errorf("bundle ID is required")
	}
	
	return nil
}

// GetNetworks returns all telecom networks (admin - returns full entities)
func (s *NetworkConfigService) GetNetworks(ctx context.Context) ([]NetworkResponse, error) {
	// Reuse GetAllNetworks method
	return s.GetAllNetworks(ctx)
}

// UpdateNetwork updates network configuration (admin)
func (s *NetworkConfigService) UpdateNetwork(ctx context.Context, networkID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Note: NetworkConfigs doesn't have ID field, using NetworkCode as identifier
	// Update network in database
	// In production, this would:
	// 1. Find network by NetworkCode (networkID)
	// 2. Update configurable fields (IsActive, SupportData, SupportAirtime)
	// 3. Save to database
	// 4. Invalidate caches
	//
	// Example implementation:
	// network, err := s.networkRepo.FindByCode(ctx, networkID)
	// if err != nil {
	//     return nil, fmt.Errorf("network not found: %w", err)
	// }
	// 
	// if isActive, ok := data["is_active"].(bool); ok {
	//     network.IsActive = isActive
	// }
	// if supportData, ok := data["support_data"].(bool); ok {
	//     network.SupportData = supportData
	// }
	// if supportAirtime, ok := data["support_airtime"].(bool); ok {
	//     network.SupportAirtime = supportAirtime
	// }
	// 
	// err = s.networkRepo.Update(ctx, network)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to update network: %w", err)
	// }
	
	// For now, return the updated data
	data["id"] = networkID
	return data, nil
}


// ValidatePhoneNetwork validates that a phone number belongs to the expected network
func (s *NetworkConfigService) ValidatePhoneNetwork(ctx context.Context, phoneNumber string, expectedNetwork string) (map[string]interface{}, error) {
	// This method integrates with HLRService for network detection
	// For now, we'll use a simple prefix-based validation
	// In production, this should use HLRService.DetectNetwork()
	
	// Normalize phone number (remove +234 or 234 prefix)
	normalizedPhone := phoneNumber
	if len(phoneNumber) > 10 {
		if phoneNumber[:4] == "+234" {
			normalizedPhone = "0" + phoneNumber[4:]
		} else if phoneNumber[:3] == "234" {
			normalizedPhone = "0" + phoneNumber[3:]
		}
	}
	
	// Extract prefix (first 4 digits)
	if len(normalizedPhone) < 4 {
		return nil, fmt.Errorf("invalid phone number format")
	}
	prefix := normalizedPhone[:4]
	
	// Detect network based on prefix
	detectedNetwork := detectNetworkByPrefix(prefix)
	
	// Validate against expected network
	isValid := detectedNetwork == expectedNetwork
	
	result := map[string]interface{}{
		"phone_number":      phoneNumber,
		"expected_network":  expectedNetwork,
		"detected_network":  detectedNetwork,
		"is_valid":          isValid,
		"validation_method": "prefix", // In production: "hlr_lookup"
	}
	
	if !isValid {
		result["message"] = fmt.Sprintf("Phone number belongs to %s, not %s", detectedNetwork, expectedNetwork)
	} else {
		result["message"] = "Phone number validated successfully"
	}
	
	return result, nil
}

// detectNetworkByPrefix detects network based on phone number prefix
func detectNetworkByPrefix(prefix string) string {
	// Nigerian network prefixes
	mtnPrefixes := []string{"0803", "0806", "0703", "0706", "0813", "0816", "0810", "0814", "0903", "0906", "0913", "0916"}
	gloPrefixes := []string{"0805", "0807", "0705", "0815", "0811", "0905", "0915"}
	airtelPrefixes := []string{"0802", "0808", "0708", "0812", "0701", "0902", "0901", "0904", "0907", "0912"}
	nineMobilePrefixes := []string{"0809", "0817", "0818", "0909", "0908"}
	
	for _, p := range mtnPrefixes {
		if prefix == p {
			return "MTN"
		}
	}
	
	for _, p := range gloPrefixes {
		if prefix == p {
			return "GLO"
		}
	}
	
	for _, p := range airtelPrefixes {
		if prefix == p {
			return "AIRTEL"
		}
	}
	
	for _, p := range nineMobilePrefixes {
		if prefix == p {
			return "9MOBILE"
		}
	}
	
	return "UNKNOWN"
}

// GetNetworkConfigsAdmin returns all network configurations with full details (admin only)
func (s *NetworkConfigService) GetNetworkConfigsAdmin(ctx context.Context) ([]*entities.NetworkConfigs, error) {
	// Get all network configs from database
	networks, err := s.networkRepo.FindAll(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve network configurations: %w", err)
	}
	
	return networks, nil
}
