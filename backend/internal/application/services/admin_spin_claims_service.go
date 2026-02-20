package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// AdminSpinClaimService handles admin operations for spin prize claims
type AdminSpinClaimService struct {
	spinRepo repositories.SpinRepository
	userRepo repositories.UserRepository
	db       *gorm.DB
}

// NewAdminSpinClaimService creates a new admin spin claim service
func NewAdminSpinClaimService(
	spinRepo repositories.SpinRepository,
	userRepo repositories.UserRepository,
	db *gorm.DB,
) *AdminSpinClaimService {
	return &AdminSpinClaimService{
		spinRepo: spinRepo,
		userRepo: userRepo,
		db:       db,
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// ClaimFilters represents filters for listing claims
type ClaimFilters struct {
	Status     string    // PENDING, CLAIMED, PENDING_ADMIN_REVIEW, APPROVED, REJECTED, EXPIRED
	PrizeType  string    // AIRTIME, DATA, CASH
	FromDate   time.Time
	ToDate     time.Time
	MSISDN     string
	SearchTerm string // Search in user name, spin code, etc.
}

// Pagination represents pagination parameters
type Pagination struct {
	Page    int
	Limit   int
	SortBy  string // created_at, claimed_at, prize_value
	Order   string // asc, desc
}

// ClaimListItem represents a claim in the list
type ClaimListItem struct {
	ID              uuid.UUID              `json:"id"`
	SpinCode        string                 `json:"spin_code"`
	MSISDN          string                 `json:"msisdn"`
	UserID          *uuid.UUID             `json:"user_id"`
	UserName        string                 `json:"user_name,omitempty"`
	PrizeType       string                 `json:"prize_type"`
	PrizeName       string                 `json:"prize_name"`
	PrizeValue      int64                  `json:"prize_value"`
	ClaimStatus     string                 `json:"claim_status"`
	CreatedAt       time.Time              `json:"created_at"`
	ClaimDate       *time.Time             `json:"claim_date,omitempty"`
	BankDetails     *BankDetailsResponse   `json:"bank_details,omitempty"`
	ReviewedBy      *uuid.UUID             `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time             `json:"reviewed_at,omitempty"`
	RejectionReason *string                `json:"rejection_reason,omitempty"`
}

// BankDetailsResponse represents bank details in response
type BankDetailsResponse struct {
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	BankName      string `json:"bank_name"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	CurrentPage  int   `json:"current_page"`
	TotalPages   int   `json:"total_pages"`
	TotalItems   int64 `json:"total_items"`
	ItemsPerPage int   `json:"items_per_page"`
}

// ClaimListResponse represents the response for list claims
type ClaimListResponse struct {
	Claims     []ClaimListItem    `json:"claims"`
	Pagination PaginationResponse `json:"pagination"`
}

// UserDetailsResponse represents user details in claim details
type UserDetailsResponse struct {
	ID          uuid.UUID `json:"id"`
	MSISDN      string    `json:"msisdn"`
	Email       *string   `json:"email,omitempty"`
	FullName    *string   `json:"full_name,omitempty"`
	TotalPoints int64     `json:"total_points"`
	LoyaltyTier string    `json:"loyalty_tier"`
}

// ClaimDetailsResponse represents detailed claim information
type ClaimDetailsResponse struct {
	ID               uuid.UUID            `json:"id"`
	SpinCode         string               `json:"spin_code"`
	MSISDN           string               `json:"msisdn"`
	UserID           *uuid.UUID           `json:"user_id"`
	PrizeType        string               `json:"prize_type"`
	PrizeName        string               `json:"prize_name"`
	PrizeValue       int64                `json:"prize_value"`
	ClaimStatus      string               `json:"claim_status"`
	CreatedAt        time.Time            `json:"created_at"`
	ClaimDate        *time.Time           `json:"claim_date,omitempty"`
	BankDetails      *BankDetailsResponse `json:"bank_details,omitempty"`
	ReviewedBy       *uuid.UUID           `json:"reviewed_by,omitempty"`
	ReviewedAt       *time.Time           `json:"reviewed_at,omitempty"`
	RejectionReason  string               `json:"rejection_reason,omitempty"`
	AdminNotes       string               `json:"admin_notes,omitempty"`
	PaymentReference string               `json:"payment_reference,omitempty"`
	UserDetails      *UserDetailsResponse `json:"user_details,omitempty"`
}

// ApproveClaimRequest represents request to approve a claim
type ApproveClaimRequest struct {
	AdminNotes       string `json:"admin_notes"`
	PaymentReference string `json:"payment_reference" validate:"required"`
}

// RejectClaimRequest represents request to reject a claim
type RejectClaimRequest struct {
	RejectionReason string `json:"rejection_reason" validate:"required"`
	AdminNotes      string `json:"admin_notes"`
}

// ClaimStatistics represents statistics for claims
type ClaimStatistics struct {
	TotalClaims          int64                       `json:"total_claims"`
	PendingClaims        int64                       `json:"pending_claims"`
	ApprovedClaims       int64                       `json:"approved_claims"`
	RejectedClaims       int64                       `json:"rejected_claims"`
	ClaimedPrizes        int64                       `json:"claimed_prizes"`
	ExpiredPrizes        int64                       `json:"expired_prizes"`
	TotalPrizeValue      int64                       `json:"total_prize_value"`
	ClaimsByType         map[string]int64            `json:"claims_by_type"`
	ClaimsByStatus       map[string]int64            `json:"claims_by_status"`
	RecentClaims         []ClaimListItem             `json:"recent_claims"`
}

// ============================================================================
// Service Methods
// ============================================================================

// ListClaims returns a paginated list of claims with filters
func (s *AdminSpinClaimService) ListClaims(ctx context.Context, filters ClaimFilters, pagination Pagination) (*ClaimListResponse, error) {
	// Build query
	query := s.db.WithContext(ctx).Model(&entities.SpinResults{})

	// Apply filters
	if filters.Status != "" {
		query = query.Where("claim_status = ?", filters.Status)
	}
	if filters.PrizeType != "" {
		query = query.Where("prize_type = ?", filters.PrizeType)
	}
	if !filters.FromDate.IsZero() {
		query = query.Where("created_at >= ?", filters.FromDate)
	}
	if !filters.ToDate.IsZero() {
		query = query.Where("created_at <= ?", filters.ToDate)
	}
	if filters.MSISDN != "" {
		query = query.Where("msisdn = ?", filters.MSISDN)
	}
	if filters.SearchTerm != "" {
		query = query.Where("spin_code LIKE ? OR msisdn LIKE ?", 
			"%"+filters.SearchTerm+"%", "%"+filters.SearchTerm+"%")
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count claims: %w", err)
	}

	// Apply pagination
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 20
	}
	if pagination.Limit > 100 {
		pagination.Limit = 100
	}

	offset := (pagination.Page - 1) * pagination.Limit

	// Apply sorting
	sortBy := "created_at"
	if pagination.SortBy != "" {
		sortBy = pagination.SortBy
	}
	order := "DESC"
	if pagination.Order != "" {
		order = strings.ToUpper(pagination.Order)
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, order)).
		Limit(pagination.Limit).
		Offset(offset)

	// Fetch results with user join
	type Result struct {
		entities.SpinResults
		UserName string
	}

	var results []Result
	err := query.
		Select("spin_results.*, users.full_name as user_name").
		Joins("LEFT JOIN users ON users.id = spin_results.user_id").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch claims: %w", err)
	}

	// Convert to response
	claims := make([]ClaimListItem, len(results))
	for i, result := range results {
		claims[i] = ClaimListItem{
			ID:          result.ID,
			SpinCode:    result.SpinCode,
			MSISDN:      result.Msisdn,
			UserID:      result.UserID,
			UserName:    result.UserName,
			PrizeType:   result.PrizeType,
			PrizeName:   result.PrizeName,
			PrizeValue:  result.PrizeValue,
			ClaimStatus: result.ClaimStatus,
			CreatedAt:   result.CreatedAt,
			ClaimDate:   result.ClaimedAt,
			ReviewedBy:  result.ReviewedBy,
			ReviewedAt:  result.ReviewedAt,
		}

		// Add rejection reason if present
		if result.RejectionReason != "" {
			claims[i].RejectionReason = &result.RejectionReason
		}

		// Add bank details if present
		if result.BankAccountNumber != "" && result.BankAccountName != "" && result.BankName != "" {
			claims[i].BankDetails = &BankDetailsResponse{
				AccountNumber: result.BankAccountNumber,
				AccountName:   result.BankAccountName,
				BankName:      result.BankName,
			}
		}
	}

	totalPages := int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))

	return &ClaimListResponse{
		Claims: claims,
		Pagination: PaginationResponse{
			CurrentPage:  pagination.Page,
			TotalPages:   totalPages,
			TotalItems:   total,
			ItemsPerPage: pagination.Limit,
		},
	}, nil
}

// GetPendingClaims returns claims pending admin review
func (s *AdminSpinClaimService) GetPendingClaims(ctx context.Context) (*ClaimListResponse, error) {
	filters := ClaimFilters{
		Status: "PENDING_ADMIN_REVIEW",
	}
	pagination := Pagination{
		Page:   1,
		Limit:  100,
		SortBy: "created_at",
		Order:  "ASC",
	}
	return s.ListClaims(ctx, filters, pagination)
}

// GetClaimDetails returns detailed information about a specific claim
func (s *AdminSpinClaimService) GetClaimDetails(ctx context.Context, claimID string) (*ClaimDetailsResponse, error) {
	// Parse claim ID
	id, err := uuid.Parse(claimID)
	if err != nil {
		return nil, fmt.Errorf("invalid claim ID format: %w", err)
	}

	// Fetch claim
	var claim entities.SpinResults
	err = s.db.WithContext(ctx).Where("id = ?", id).First(&claim).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("claim not found")
		}
		return nil, fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Build response
	response := &ClaimDetailsResponse{
		ID:               claim.ID,
		SpinCode:         claim.SpinCode,
		MSISDN:           claim.Msisdn,
		UserID:           claim.UserID,
		PrizeType:        claim.PrizeType,
		PrizeName:        claim.PrizeName,
		PrizeValue:       claim.PrizeValue,
		ClaimStatus:      claim.ClaimStatus,
		CreatedAt:        claim.CreatedAt,
		ClaimDate:        claim.ClaimedAt,
		ReviewedBy:       claim.ReviewedBy,
		ReviewedAt:       claim.ReviewedAt,
		RejectionReason:  claim.RejectionReason,
		AdminNotes:       claim.AdminNotes,
		PaymentReference: claim.PaymentReference,
	}

	// Add user details if available
	if claim.UserID != nil {
		var user entities.Users
		err = s.db.WithContext(ctx).Where("id = ?", claim.UserID).First(&user).Error
		if err == nil {
			response.UserDetails = &UserDetailsResponse{
				ID:          user.ID,
				MSISDN:      user.MSISDN,  // Uppercase MSISDN
				Email:       &user.Email,  // Convert to pointer
				FullName:    &user.FullName,  // Convert to pointer
				TotalPoints: int64(user.TotalPoints), // Convert int to int64
				LoyaltyTier: user.LoyaltyTier,
			}
		}
	}

	// Add bank details if present
	if claim.BankAccountNumber != "" && claim.BankAccountName != "" && claim.BankName != "" {
		response.BankDetails = &BankDetailsResponse{
			AccountNumber: claim.BankAccountNumber,
			AccountName:   claim.BankAccountName,
			BankName:      claim.BankName,
		}
	}

	return response, nil
}

// ApproveClaim approves a cash prize claim
func (s *AdminSpinClaimService) ApproveClaim(ctx context.Context, claimID string, adminID string, request ApproveClaimRequest) error {
	// Parse IDs
	id, err := uuid.Parse(claimID)
	if err != nil {
		return fmt.Errorf("invalid claim ID format: %w", err)
	}

	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return fmt.Errorf("invalid admin ID format: %w", err)
	}

	// Fetch claim
	var claim entities.SpinResults
	err = s.db.WithContext(ctx).Where("id = ?", id).First(&claim).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("claim not found")
		}
		return fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Validate claim status
	if claim.ClaimStatus != "PENDING_ADMIN_REVIEW" {
		return fmt.Errorf("claim is not pending review, current status: %s", claim.ClaimStatus)
	}

	// Update claim
	now := time.Now()
	updates := map[string]interface{}{
		"claim_status":      "APPROVED",
		"reviewed_by":       adminUUID,
		"reviewed_at":       now,
		"admin_notes":       request.AdminNotes,
		"payment_reference": request.PaymentReference,
	}

	err = s.db.WithContext(ctx).Model(&claim).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("failed to approve claim: %w", err)
	}

	return nil
}

// RejectClaim rejects a cash prize claim
func (s *AdminSpinClaimService) RejectClaim(ctx context.Context, claimID string, adminID string, request RejectClaimRequest) error {
	// Parse IDs
	id, err := uuid.Parse(claimID)
	if err != nil {
		return fmt.Errorf("invalid claim ID format: %w", err)
	}

	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return fmt.Errorf("invalid admin ID format: %w", err)
	}

	// Fetch claim
	var claim entities.SpinResults
	err = s.db.WithContext(ctx).Where("id = ?", id).First(&claim).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("claim not found")
		}
		return fmt.Errorf("failed to fetch claim: %w", err)
	}

	// Validate claim status
	if claim.ClaimStatus != "PENDING_ADMIN_REVIEW" {
		return fmt.Errorf("claim is not pending review, current status: %s", claim.ClaimStatus)
	}

	// Update claim
	now := time.Now()
	updates := map[string]interface{}{
		"claim_status":     "REJECTED",
		"reviewed_by":      adminUUID,
		"reviewed_at":      now,
		"rejection_reason": request.RejectionReason,
		"admin_notes":      request.AdminNotes,
	}

	err = s.db.WithContext(ctx).Model(&claim).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("failed to reject claim: %w", err)
	}

	return nil
}

// GetStatistics returns comprehensive statistics for claims
func (s *AdminSpinClaimService) GetStatistics(ctx context.Context) (*ClaimStatistics, error) {
	stats := &ClaimStatistics{
		ClaimsByType:   make(map[string]int64),
		ClaimsByStatus: make(map[string]int64),
	}

	// Total claims
	s.db.WithContext(ctx).Model(&entities.SpinResults{}).Count(&stats.TotalClaims)

	// Claims by status
	var statusCounts []struct {
		ClaimStatus string
		Count       int64
	}
	s.db.WithContext(ctx).Model(&entities.SpinResults{}).
		Select("claim_status, COUNT(*) as count").
		Group("claim_status").
		Find(&statusCounts)

	for _, sc := range statusCounts {
		stats.ClaimsByStatus[sc.ClaimStatus] = sc.Count
		switch sc.ClaimStatus {
		case "PENDING_ADMIN_REVIEW":
			stats.PendingClaims = sc.Count
		case "APPROVED":
			stats.ApprovedClaims = sc.Count
		case "REJECTED":
			stats.RejectedClaims = sc.Count
		case "CLAIMED":
			stats.ClaimedPrizes = sc.Count
		case "EXPIRED":
			stats.ExpiredPrizes = sc.Count
		}
	}

	// Claims by type
	var typeCounts []struct {
		PrizeType string
		Count     int64
	}
	s.db.WithContext(ctx).Model(&entities.SpinResults{}).
		Select("prize_type, COUNT(*) as count").
		Group("prize_type").
		Find(&typeCounts)

	for _, tc := range typeCounts {
		stats.ClaimsByType[tc.PrizeType] = tc.Count
	}

	// Total prize value
	s.db.WithContext(ctx).Model(&entities.SpinResults{}).
		Select("COALESCE(SUM(prize_value), 0)").
		Row().Scan(&stats.TotalPrizeValue)

	// Recent claims
	recentResponse, _ := s.ListClaims(ctx, ClaimFilters{}, Pagination{
		Page:   1,
		Limit:  10,
		SortBy: "created_at",
		Order:  "DESC",
	})
	if recentResponse != nil {
		stats.RecentClaims = recentResponse.Claims
	}

	return stats, nil
}

// ExportClaims exports claims to CSV format
func (s *AdminSpinClaimService) ExportClaims(ctx context.Context, filters ClaimFilters) ([]byte, error) {
	// Fetch all claims matching filters (no pagination)
	response, err := s.ListClaims(ctx, filters, Pagination{
		Page:   1,
		Limit:  10000, // Large limit for export
		SortBy: "created_at",
		Order:  "DESC",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch claims for export: %w", err)
	}

	// Create CSV
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ID", "Spin Code", "MSISDN", "User Name", "Prize Type", "Prize Name",
		"Prize Value", "Claim Status", "Created At", "Claim Date",
		"Bank Account Number", "Bank Account Name", "Bank Name",
		"Reviewed By", "Reviewed At", "Rejection Reason",
	}
	writer.Write(header)

	// Write rows
	for _, claim := range response.Claims {
		row := []string{
			claim.ID.String(),
			claim.SpinCode,
			claim.MSISDN,
			claim.UserName,
			claim.PrizeType,
			claim.PrizeName,
			fmt.Sprintf("%d", claim.PrizeValue),
			claim.ClaimStatus,
			claim.CreatedAt.Format("2006-01-02 15:04:05"),
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		}

		if claim.ClaimDate != nil {
			row[9] = claim.ClaimDate.Format("2006-01-02 15:04:05")
		}

		if claim.BankDetails != nil {
			row[10] = claim.BankDetails.AccountNumber
			row[11] = claim.BankDetails.AccountName
			row[12] = claim.BankDetails.BankName
		}

		if claim.ReviewedBy != nil {
			row[13] = claim.ReviewedBy.String()
		}

		if claim.ReviewedAt != nil {
			row[14] = claim.ReviewedAt.Format("2006-01-02 15:04:05")
		}

		if claim.RejectionReason != nil {
			row[15] = *claim.RejectionReason
		}

		writer.Write(row)
	}

	writer.Flush()
	return []byte(buf.String()), nil
}
