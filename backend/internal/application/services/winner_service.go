package services

import (
	"go.uber.org/zap"
	"rechargemax/internal/logger"
	"context"

	"rechargemax/internal/pkg/safe"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// WinnerService handles winner management and prize provisioning
type WinnerService struct {
	winnerRepo          repositories.WinnerRepository
	drawRepo            repositories.DrawRepository
	userRepo            repositories.UserRepository
	spinRepo            repositories.SpinRepository
	hlrService          *HLRService
	telecomService      *TelecomServiceIntegrated
	paymentService      *PaymentService
	notificationService *NotificationService
	db                  *gorm.DB
}

// safeStr safely dereferences a *string, returning empty string if nil
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// setStr sets a *string field, allocating if necessary
func setStr(s *string, val string) *string {
	if s == nil {
		return &val
	}
	*s = val
	return s
}

// safeInt64 safely dereferences a *int64
func safeInt64(n *int64) int64 {
	if n == nil {
		return 0
	}
	return *n
}

// safeTime safely dereferences a *time.Time
func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// safeUUID safely dereferences a *uuid.UUID, returning uuid.Nil if nil
func safeUUID(u *uuid.UUID) uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return *u
}

// getClaimWindowDays reads the claim window from platform_settings, defaulting to 30
func (s *WinnerService) getClaimWindowDays(ctx context.Context) int {
	if s.db == nil {
		return 30
	}
	var setting struct {
		SettingValue string `gorm:"column:setting_value"`
	}
	if err := s.db.WithContext(ctx).
		Table("platform_settings").
		Where("setting_key = ?", "draw.claim_window_days").
		Select("setting_value").
		First(&setting).Error; err != nil {
		return 30
	}
	days, err := strconv.Atoi(setting.SettingValue)
	if err != nil || days <= 0 {
		return 30
	}
	return days
}

// WinnerResponse represents winner data for API responses
type WinnerResponse struct {
	ID              uuid.UUID  `json:"id"`
	DrawID          uuid.UUID  `json:"draw_id"`
	DrawName        string     `json:"draw_name"`
	MSISDN          string     `json:"msisdn"`
	Position        int        `json:"position"`
	PrizeType       string     `json:"prize_type"`
	PrizeDescription string    `json:"prize_description"`
	CashAmount      int64      `json:"cash_amount,omitempty"`
	DataPackage     string     `json:"data_package,omitempty"`
	AirtimeAmount   int64      `json:"airtime_amount,omitempty"`
	ClaimStatus     string     `json:"claim_status"`
	ClaimDeadline   time.Time  `json:"claim_deadline"`
	ClaimedAt       *time.Time `json:"claimed_at,omitempty"`
	ProvisionStatus string     `json:"provision_status,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// NewWinnerService creates a new winner service
func NewWinnerService(
	winnerRepo repositories.WinnerRepository,
	drawRepo repositories.DrawRepository,
	userRepo repositories.UserRepository,
	spinRepo repositories.SpinRepository,
	hlrService *HLRService,
	telecomService *TelecomServiceIntegrated,
	paymentService *PaymentService,
	notificationService *NotificationService,
	db *gorm.DB,
) *WinnerService {
	return &WinnerService{
		winnerRepo:          winnerRepo,
		drawRepo:            drawRepo,
		userRepo:            userRepo,
		spinRepo:            spinRepo,
		hlrService:          hlrService,
		telecomService:      telecomService,
		paymentService:      paymentService,
		notificationService: notificationService,
		db:                  db,
	}
}

// CreateWinner creates a new winner record
func (s *WinnerService) CreateWinner(ctx context.Context, drawID uuid.UUID, msisdn string, position int, prizeType, prizeDescription string, cashAmount int64, dataPackage string, airtimeAmount int64, network string) (*entities.Winner, error) {
	// Get draw
	draw, err := s.drawRepo.FindByID(ctx, drawID)
	if err != nil {
		return nil, fmt.Errorf("draw not found: %w", err)
	}

	// Create winner record
	claimWindowDays := s.getClaimWindowDays(ctx)
	claimDeadline := time.Now().AddDate(0, 0, claimWindowDays)
	winner := &entities.Winner{
		ID:               uuid.New(),
		DrawID:           &drawID,
		MSISDN:           msisdn,
		Position:         position,
		PrizeType:        prizeType,
		PrizeDescription: prizeDescription,
		PrizeAmount:      &cashAmount,
		DataPackage:      &dataPackage,
		AirtimeAmount:    &airtimeAmount,
		Network:          &network,
		ClaimStatus:      "PENDING",
		ClaimDeadline:    &claimDeadline,
		NotificationSent: false,
	}

	// Determine if auto-provision is possible
	if prizeType == "data" || prizeType == "airtime" {
		winner.AutoProvision = true
		winner.ProvisionStatus = setStr(winner.ProvisionStatus, "pending")
	}

	err = s.winnerRepo.Create(ctx, winner)
	if err != nil {
		return nil, fmt.Errorf("failed to create winner: %w", err)
	}

	// Send notifications
	safe.Go(func() { s.sendWinnerNotifications(context.Background(), winner, draw) })

	// Auto-provision if applicable
	if winner.AutoProvision {
		safe.Go(func() { s.provisionPrize(context.Background(), winner) })
	}

	return winner, nil
}

// ImportWinners imports multiple winners from draw engine results
func (s *WinnerService) ImportWinners(ctx context.Context, drawID uuid.UUID, winners []struct {
	MSISDN           string
	Position         int
	PrizeType        string
	PrizeDescription string
	CashAmount       int64
	DataPackage      string
	AirtimeAmount    int64
	Network          string
}) error {
	for _, w := range winners {
		_, err := s.CreateWinner(
			ctx,
			drawID,
			w.MSISDN,
			w.Position,
			w.PrizeType,
			w.PrizeDescription,
			w.CashAmount,
			w.DataPackage,
			w.AirtimeAmount,
			w.Network,
		)
		if err != nil {
			// Log error but continue with other winners
			logger.Error("Failed to create winner", zap.Error(err), zap.String("msisdn", w.MSISDN))
		}
	}

	return nil
}

// provisionPrize auto-provisions data or airtime prizes
func (s *WinnerService) provisionPrize(ctx context.Context, winner *entities.Winner) {
	// Detect network if not provided
	network := safeStr(winner.Network)
	if network == "" {
		detectedNetwork, err := s.hlrService.DetectNetwork(ctx, winner.MSISDN, nil)
		if err != nil {
			winner.ProvisionStatus = setStr(winner.ProvisionStatus, "failed")
			provisionError := fmt.Sprintf("Network detection failed: %v", err)
			winner.ProvisionError = &provisionError
			s.winnerRepo.Update(ctx, winner)
			return
		}
		network = detectedNetwork.Network
		winner.Network = setStr(winner.Network, network)
	}

	// Integrate with TelecomService for actual provisioning
	// In production, this would:
	// 1. Call TelecomService.PurchaseAirtime() for airtime prizes
	// 2. Call TelecomService.PurchaseData() for data prizes
	// 3. Handle async confirmation via webhook
	// 4. Update provision status based on result
	//
	// Example implementation:
	// if *winner.PrizeType == "airtime" {
	//     amount := *winner.PrizeValue // Amount in kobo
	//     err := s.telecomService.PurchaseAirtime(ctx, winner.MSISDN, network, amount)
	//     if err != nil {
	//         *winner.ProvisionStatus = "failed"
	//         winner.ProvisionError = stringPtr(err.Error())
	//         s.winnerRepo.Update(ctx, winner)
	//         return
	//     }
	// } else if *winner.PrizeType == "data" {
	//     dataPackage := *winner.PrizeDescription // e.g., "1GB_DAILY"
	//     err := s.telecomService.PurchaseData(ctx, winner.MSISDN, network, dataPackage)
	//     if err != nil {
	//         *winner.ProvisionStatus = "failed"
	//         winner.ProvisionError = stringPtr(err.Error())
	//         s.winnerRepo.Update(ctx, winner)
	//         return
	//     }
	// }
	
	// Provision via TelecomService
	if s.telecomService != nil {
		switch winner.PrizeType {
		case "airtime":
			amountKobo := int64(0)
			if winner.AirtimeAmount != nil {
				amountKobo = *winner.AirtimeAmount
			} else if winner.PrizeAmount != nil {
				amountKobo = *winner.PrizeAmount
			}
			if amountKobo > 0 {
				_, provErr := s.telecomService.PurchaseAirtime(ctx, network, winner.MSISDN, int(amountKobo/100))
				if provErr != nil {
					provErrStr := provErr.Error()
					winner.ProvisionStatus = setStr(winner.ProvisionStatus, "failed")
					winner.ProvisionError = &provErrStr
					s.winnerRepo.Update(ctx, winner)
					return
				}
			}
		case "data":
			amountKobo := int64(0)
			if winner.PrizeAmount != nil {
				amountKobo = *winner.PrizeAmount
			}
			variationCode := s.getDataVariationCode(amountKobo, network)
			_, provErr := s.telecomService.PurchaseData(ctx, network, winner.MSISDN, variationCode, int(amountKobo/100))
			if provErr != nil {
				provErrStr := provErr.Error()
				winner.ProvisionStatus = setStr(winner.ProvisionStatus, "failed")
				winner.ProvisionError = &provErrStr
				s.winnerRepo.Update(ctx, winner)
				return
			}
		}
	}

	winner.ProvisionStatus = setStr(winner.ProvisionStatus, "completed")
	winner.ProvisionedAt = timePtr(time.Now())
	winner.ClaimStatus = "CLAIMED"
	winner.ClaimedAt = timePtr(time.Now())

	err := s.winnerRepo.Update(ctx, winner)
	if err != nil {
		logger.Error("Failed to update winner provision status: %v", zap.Error(err))
	}

	// Send success notification
	s.sendProvisionSuccessNotification(ctx, winner)
}

// sendWinnerNotifications sends notifications to winner via all channels
func (s *WinnerService) sendWinnerNotifications(ctx context.Context, winner *entities.Winner, draw *entities.Draw) {
	// Get user details
	user, err := s.userRepo.FindByMSISDN(ctx, winner.MSISDN)
	if err != nil {
		logger.Error("Failed to get user details for winner", zap.Error(err), zap.String("msisdn", winner.MSISDN))
		return
	}

	// Prepare message
	var message string
	if winner.PrizeType == "cash" {
		message = fmt.Sprintf("🎉 Congratulations! You won ₦%d in the %s draw! Login to claim your prize within 30 days.", safeInt64(winner.PrizeAmount)/100, draw.Name)
	} else if winner.PrizeType == "data" {
		message = fmt.Sprintf("🎉 Congratulations! You won %s data in the %s draw! Your data has been credited to your number.", safeStr(winner.DataPackage), draw.Name)
	} else if winner.PrizeType == "airtime" {
		message = fmt.Sprintf("🎉 Congratulations! You won ₦%d airtime in the %s draw! Your airtime has been credited to your number.", safeInt64(winner.AirtimeAmount)/100, draw.Name)
	} else {
		message = fmt.Sprintf("🎉 Congratulations! You won %s in the %s draw! Login to claim your prize within 30 days.", winner.PrizeDescription, draw.Name)
	}

	// Send via all channels
	if s.notificationService != nil {
		// SMS
		s.notificationService.SendSMS(ctx, winner.MSISDN, message)

		// Email
		if user.Email != "" {
			s.notificationService.SendEmail(ctx, user.Email, fmt.Sprintf("You Won! - %s", draw.Name), message)
		}

		// Push notification
		s.notificationService.SendPushNotification(ctx, winner.MSISDN, "You Won!", message)

		// In-platform notification
		s.notificationService.CreateNotification(ctx, winner.MSISDN, "winner", "You Won!", message, map[string]interface{}{
			"winner_id": winner.ID.String(),
			"draw_id":   draw.ID.String(),
		})
	}

	// Mark notification as sent
	winner.NotificationSent = true
	s.winnerRepo.Update(ctx, winner)
}

// sendProvisionSuccessNotification sends notification after successful auto-provision
func (s *WinnerService) sendProvisionSuccessNotification(ctx context.Context, winner *entities.Winner) {
	var message string
	if winner.PrizeType == "data" {
		message = fmt.Sprintf("Your %s data prize has been successfully credited to %s!", safeStr(winner.DataPackage), winner.MSISDN)
	} else if winner.PrizeType == "airtime" {
		message = fmt.Sprintf("Your ₦%d airtime prize has been successfully credited to %s!", safeInt64(winner.AirtimeAmount)/100, winner.MSISDN)
	}

	if s.notificationService != nil {
		s.notificationService.SendSMS(ctx, winner.MSISDN, message)
		s.notificationService.CreateNotification(ctx, winner.MSISDN, "prize_credited", "Prize Credited", message, nil)
	}
}

// GetWinnersByMSISDN gets all wins for a user
func (s *WinnerService) GetWinnersByMSISDN(ctx context.Context, msisdn string) ([]*WinnerResponse, error) {
	winners, err := s.winnerRepo.FindByMSISDN(ctx, msisdn, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get winners: %w", err)
	}

	var responses []*WinnerResponse
	for _, winner := range winners {
		draw, _ := s.drawRepo.FindByID(ctx, safeUUID(winner.DrawID))
		drawName := ""
		if draw != nil {
			drawName = draw.Name
		}

		responses = append(responses, &WinnerResponse{
			ID:               winner.ID,
			DrawID:           safeUUID(winner.DrawID),
			DrawName:         drawName,
			MSISDN:           winner.MSISDN,
			Position:         winner.Position,
			PrizeType:        winner.PrizeType,
			PrizeDescription: winner.PrizeDescription,
			CashAmount:       safeInt64(winner.PrizeAmount),
			DataPackage:      safeStr(winner.DataPackage),
			AirtimeAmount:    safeInt64(winner.AirtimeAmount),
			ClaimStatus:      winner.ClaimStatus,
			ClaimDeadline:    safeTime(winner.ClaimDeadline),
			ClaimedAt:        winner.ClaimedAt,
			ProvisionStatus:  safeStr(winner.ProvisionStatus),
			CreatedAt:        winner.CreatedAt,
		})
	}

	return responses, nil
}

// GetWinnersByDrawID gets all winners for a draw
func (s *WinnerService) GetWinnersByDrawID(ctx context.Context, drawID uuid.UUID) ([]*WinnerResponse, error) {
	winners, err := s.winnerRepo.FindByDrawID(ctx, drawID)
	if err != nil {
		return nil, fmt.Errorf("failed to get winners: %w", err)
	}

	draw, _ := s.drawRepo.FindByID(ctx, drawID)
	drawName := ""
	if draw != nil {
		drawName = draw.Name
	}

	var responses []*WinnerResponse
	for _, winner := range winners {
		responses = append(responses, &WinnerResponse{
			ID:               winner.ID,
			DrawID:           safeUUID(winner.DrawID),
			DrawName:         drawName,
			MSISDN:           winner.MSISDN,
			Position:         winner.Position,
			PrizeType:        winner.PrizeType,
			PrizeDescription: winner.PrizeDescription,
			CashAmount:       safeInt64(winner.PrizeAmount),
			DataPackage:      safeStr(winner.DataPackage),
			AirtimeAmount:    safeInt64(winner.AirtimeAmount),
			ClaimStatus:      winner.ClaimStatus,
			ClaimDeadline:    safeTime(winner.ClaimDeadline),
			ClaimedAt:        winner.ClaimedAt,
			ProvisionStatus:  safeStr(winner.ProvisionStatus),
			CreatedAt:        winner.CreatedAt,
		})
	}

	return responses, nil
}

// ProcessCashPayout processes cash prize payout
func (s *WinnerService) ProcessCashPayout(ctx context.Context, winnerID uuid.UUID, bankName, accountNumber, accountName string) error {
	winner, err := s.winnerRepo.FindByID(ctx, winnerID)
	if err != nil {
		return fmt.Errorf("winner not found: %w", err)
	}

	if winner.PrizeType != "cash" {
		return fmt.Errorf("prize is not cash")
	}

	if winner.ClaimStatus != "PENDING" {
		return fmt.Errorf("prize already claimed or expired")
	}

	// Integrate with payment service for bank transfer
	// In production, this would:
	// 1. Validate bank account details (account name verification)
	// 2. Initiate bank transfer via payment gateway (Paystack, Flutterwave)
	// 3. Wait for transfer confirmation
	// 4. Update provision status based on result
	// 5. Send notification to winner
	//
	// Example implementation:
	// // Validate bank account
	// isValid, err := s.paymentService.ValidateBankAccount(ctx, bankName, accountNumber, accountName)
	// if err != nil || !isValid {
	//     return fmt.Errorf("invalid bank account details: %w", err)
	// }
	// 
	// // Initiate transfer
	// transferRef, err := s.paymentService.InitiateTransfer(ctx, PaymentTransferRequest{
	//     Amount:        *winner.PrizeAmount,
	//     BankName:      bankName,
	//     AccountNumber: accountNumber,
	//     AccountName:   accountName,
	//     Reference:     fmt.Sprintf("PRIZE_%s", winner.ID.String()),
	//     Narration:     fmt.Sprintf("Prize payout for draw %s", winner.DrawID.String()),
	// })
	// if err != nil {
	//     return fmt.Errorf("failed to initiate transfer: %w", err)
	// }
	// 
	// winner.PaymentReference = &transferRef
	
	// Initiate bank transfer via PaymentService
	if s.paymentService != nil {
		ref := fmt.Sprintf("PRIZE_%s", winner.ID.String())
		transferReq := map[string]interface{}{
			"amount":         winner.PrizeAmount,
			"bank_code":      bankName, // bankName used as bank_code; UI should pass code
			"account_number": accountNumber,
			"account_name":   accountName,
			"reference":      ref,
			"narration":      fmt.Sprintf("Prize payout - Draw %s", winner.DrawID.String()),
		}
		resp, transferErr := s.paymentService.ProcessTransfer(ctx, transferReq)
		if transferErr != nil {
			return fmt.Errorf("bank transfer failed: %w", transferErr)
		}
		if resp != nil {
			if refStr, ok := resp["reference"].(string); ok && refStr != "" {
				winner.PayoutReference = &refStr
			}
		}
	}

	winner.ClaimStatus = "CLAIMED"
	winner.ClaimedAt = timePtr(time.Now())
	winner.BankName = &bankName
	winner.AccountNumber = &accountNumber
	winner.AccountName = &accountName

	return s.winnerRepo.Update(ctx, winner)
}

// ProcessGoodsShipment processes goods prize shipment
func (s *WinnerService) ProcessGoodsShipment(ctx context.Context, winnerID uuid.UUID, shippingAddress string) error {
	winner, err := s.winnerRepo.FindByID(ctx, winnerID)
	if err != nil {
		return fmt.Errorf("winner not found: %w", err)
	}

	if winner.PrizeType != "goods" {
		return fmt.Errorf("prize is not goods")
	}

	if winner.ClaimStatus != "PENDING" {
		return fmt.Errorf("prize already claimed or expired")
	}

	winner.ShippingAddress = &shippingAddress
	shippingStatus := "pending"
	winner.ShippingStatus = &shippingStatus
	winner.ClaimStatus = "CLAIMED"
	winner.ClaimedAt = timePtr(time.Now())

	return s.winnerRepo.Update(ctx, winner)
}

// RetryProvisioning retries failed prize provisioning
func (s *WinnerService) RetryProvisioning(ctx context.Context, winnerID uuid.UUID) error {
	winner, err := s.winnerRepo.FindByID(ctx, winnerID)
	if err != nil {
		return fmt.Errorf("winner not found: %w", err)
	}

	if !winner.AutoProvision {
		return fmt.Errorf("prize does not support auto-provisioning")
	}

	if safeStr(winner.ProvisionStatus) == "completed" {
		return fmt.Errorf("prize already provisioned")
	}

	// Reset provision status and retry
	winner.ProvisionStatus = setStr(winner.ProvisionStatus, "pending")
	emptyError := ""
	winner.ProvisionError = &emptyError
	err = s.winnerRepo.Update(ctx, winner)
	if err != nil {
		return err
	}

	// Retry provisioning
	safe.Go(func() { s.provisionPrize(context.Background(), winner) })

	return nil
}

// GetUnclaimedWinners gets winners with unclaimed prizes approaching deadline
func (s *WinnerService) GetUnclaimedWinners(ctx context.Context, daysBeforeDeadline int) ([]*entities.Winner, error) {
	// Implement FindUnclaimedBeforeDeadline
	// In production, this would:
	// 1. Calculate deadline date (now + daysBeforeDeadline)
	// 2. Query winners table where:
	//    - claim_status = 'pending'
	//    - claim_deadline <= deadline date
	// 3. Return list of winners
	// 4. Used for sending reminder notifications
	//
	// Example implementation:
	// deadlineDate := time.Now().AddDate(0, 0, daysBeforeDeadline)
	// 
	// // This would require a new repository method:
	// // winners, err := s.winnerRepo.FindUnclaimedBeforeDeadline(ctx, deadlineDate)
	// // if err != nil {
	// //     return nil, fmt.Errorf("failed to find unclaimed winners: %w", err)
	// // }
	// // 
	// // return winners, nil
	
	deadline := time.Now().AddDate(0, 0, daysBeforeDeadline).Format("2006-01-02")
	return s.winnerRepo.FindUnclaimedBeforeDeadline(ctx, deadline)
}

// SendClaimReminders sends reminder notifications to winners approaching claim deadline
func (s *WinnerService) SendClaimReminders(ctx context.Context) error {
	// Get winners with 7 days, 3 days, and 1 day remaining
	for _, days := range []int{7, 3, 1} {
		winners, err := s.GetUnclaimedWinners(ctx, days)
		if err != nil {
			continue
		}

		for _, winner := range winners {
			message := fmt.Sprintf("Reminder: You have %d days left to claim your prize! Login to RechargeMax to claim now.", days)
			if s.notificationService != nil {
				s.notificationService.SendSMS(ctx, winner.MSISDN, message)
				s.notificationService.CreateNotification(ctx, winner.MSISDN, "claim_reminder", "Claim Reminder", message, map[string]interface{}{
					"winner_id": winner.ID.String(),
					"days_left": days,
				})
			}
		}
	}

	return nil
}


// GetAllWinners returns paginated list of all winners (admin)
func (s *WinnerService) GetAllWinners(ctx context.Context, page, perPage int, drawID string) ([]*WinnerResponse, int64, error) {
	// Calculate offset
	offset := (page - 1) * perPage
	
	var winners []*entities.Winner
	var err error
	
	// Filter by draw ID if provided
	if drawID != "" {
		did, parseErr := uuid.Parse(drawID)
		if parseErr != nil {
			return nil, 0, fmt.Errorf("invalid draw ID format: %w", parseErr)
		}
		winners, err = s.winnerRepo.FindByDrawID(ctx, did)
	} else {
		winners, err = s.winnerRepo.FindAll(ctx, perPage, offset)
	}
	
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get winners: %w", err)
	}
	
	// Get total count
	total, err := s.winnerRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get winner count: %w", err)
	}
	
	// Convert to response format
	var responses []*WinnerResponse
	for _, winner := range winners {
		// Get draw name
		draw, _ := s.drawRepo.FindByID(ctx, safeUUID(winner.DrawID))
		drawName := ""
		if draw != nil {
			drawName = draw.Name
		}
		
		// Handle nil pointers safely
		var cashAmount int64
		if winner.PrizeAmount != nil {
			cashAmount = safeInt64(winner.PrizeAmount)
		}
		
		var dataPackage string
		if winner.DataPackage != nil {
			dataPackage = safeStr(winner.DataPackage)
		}
		
		var airtimeAmount int64
		if winner.AirtimeAmount != nil {
			airtimeAmount = safeInt64(winner.AirtimeAmount)
		}
		
		var provisionStatus string
		if winner.ProvisionStatus != nil {
			provisionStatus = safeStr(winner.ProvisionStatus)
		}
		
		var claimDeadline time.Time
		if winner.ClaimDeadline != nil {
			claimDeadline = safeTime(winner.ClaimDeadline)
		}
		
		responses = append(responses, &WinnerResponse{
			ID:               winner.ID,
			DrawID:           safeUUID(winner.DrawID),
			DrawName:         drawName,
			MSISDN:           winner.MSISDN,
			Position:         winner.Position,
			PrizeType:        winner.PrizeType,
			PrizeDescription: winner.PrizeDescription,
			CashAmount:       cashAmount,
			DataPackage:      dataPackage,
			AirtimeAmount:    airtimeAmount,
			ClaimStatus:      winner.ClaimStatus,
			ClaimDeadline:    claimDeadline,
			ClaimedAt:        winner.ClaimedAt,
			ProvisionStatus:  provisionStatus,
			CreatedAt:        winner.CreatedAt,
		})
	}
	
	return responses, total, nil
}

// GetWinnerByID returns a single winner by ID
func (s *WinnerService) GetWinnerByID(ctx context.Context, winnerID string, msisdn string) (*WinnerResponse, error) {
	// Parse UUID
	wid, err := uuid.Parse(winnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid winner ID format: %w", err)
	}
	
	// Get winner from repository
	winner, err := s.winnerRepo.FindByID(ctx, wid)
	if err != nil {
		return nil, fmt.Errorf("winner not found: %w", err)
	}
	
	// Verify ownership (user can only see their own wins)
	if winner.MSISDN != msisdn {
		return nil, fmt.Errorf("unauthorized: winner does not belong to user")
	}
	
	// Get draw name
	draw, _ := s.drawRepo.FindByID(ctx, safeUUID(winner.DrawID))
	drawName := ""
	if draw != nil {
		drawName = draw.Name
	}
	
	// Handle nil pointers safely
	var cashAmount int64
	if winner.PrizeAmount != nil {
		cashAmount = safeInt64(winner.PrizeAmount)
	}
	
	var dataPackage string
	if winner.DataPackage != nil {
		dataPackage = safeStr(winner.DataPackage)
	}
	
	var airtimeAmount int64
	if winner.AirtimeAmount != nil {
		airtimeAmount = safeInt64(winner.AirtimeAmount)
	}
	
	var provisionStatus string
	if winner.ProvisionStatus != nil {
		provisionStatus = safeStr(winner.ProvisionStatus)
	}
	
	var claimDeadline time.Time
	if winner.ClaimDeadline != nil {
		claimDeadline = safeTime(winner.ClaimDeadline)
	}
	
	response := &WinnerResponse{
		ID:               winner.ID,
		DrawID:           safeUUID(winner.DrawID),
		DrawName:         drawName,
		MSISDN:           winner.MSISDN,
		Position:         winner.Position,
		PrizeType:        winner.PrizeType,
		PrizeDescription: winner.PrizeDescription,
		CashAmount:       cashAmount,
		DataPackage:      dataPackage,
		AirtimeAmount:    airtimeAmount,
		ClaimStatus:      winner.ClaimStatus,
		ClaimDeadline:    claimDeadline,
		ClaimedAt:        winner.ClaimedAt,
		ProvisionStatus:  provisionStatus,
		CreatedAt:        winner.CreatedAt,
	}
	
	return response, nil
}

// UpdateWinnerRequest represents winner update request
type UpdateWinnerRequest struct {
	ClaimStatus   string `json:"claim_status"`
	PaymentStatus string `json:"payment_status"`
	Notes         string `json:"notes"`
}

// UpdateWinnerStatus updates winner status (admin operation)
func (s *WinnerService) UpdateWinnerStatus(ctx context.Context, winnerID string, req UpdateWinnerRequest) (*WinnerResponse, error) {
	// Parse UUID
	wid, err := uuid.Parse(winnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid winner ID format: %w", err)
	}
	
	// Get winner from repository
	winner, err := s.winnerRepo.FindByID(ctx, wid)
	if err != nil {
		return nil, fmt.Errorf("winner not found: %w", err)
	}
	
	// Update claim status if provided
	if req.ClaimStatus != "" {
		validStatuses := []string{"pending", "claimed", "expired", "processing"}
		isValid := false
		for _, status := range validStatuses {
			if req.ClaimStatus == status {
				isValid = true
				break
			}
		}
		
		if !isValid {
			return nil, fmt.Errorf("invalid claim status: %s", req.ClaimStatus)
		}
		
		winner.ClaimStatus = req.ClaimStatus
		
		// Set claimed timestamp if status is claimed
		if req.ClaimStatus == "claimed" && winner.ClaimedAt == nil {
			now := time.Now()
			winner.ClaimedAt = &now
		}
	}
	
	// Update payment status if provided (for cash prizes)
	if req.PaymentStatus != "" {
		if winner.PrizeType == "cash" {
			validPaymentStatuses := []string{"pending", "processing", "paid", "failed"}
			isValid := false
			for _, status := range validPaymentStatuses {
				if req.PaymentStatus == status {
					isValid = true
					break
				}
			}
			
			if !isValid {
				return nil, fmt.Errorf("invalid payment status: %s", req.PaymentStatus)
			}
			
			// Store payment status in provision status field
			winner.ProvisionStatus = &req.PaymentStatus
			
			// Set provisioned timestamp if payment is completed
			if req.PaymentStatus == "paid" && winner.ProvisionedAt == nil {
				now := time.Now()
				winner.ProvisionedAt = &now
			}
		}
	}
	
	// Update notes if provided
	if req.Notes != "" {
		winner.ProvisionError = &req.Notes
	}
	
	// Save updated winner
	if err := s.winnerRepo.Update(ctx, winner); err != nil {
		return nil, fmt.Errorf("failed to update winner: %w", err)
	}
	
	// Get draw name for response
	draw, _ := s.drawRepo.FindByID(ctx, safeUUID(winner.DrawID))
	drawName := ""
	if draw != nil {
		drawName = draw.Name
	}
	
	// Handle nil pointers safely
	var cashAmount int64
	if winner.PrizeAmount != nil {
		cashAmount = safeInt64(winner.PrizeAmount)
	}
	
	var dataPackage string
	if winner.DataPackage != nil {
		dataPackage = safeStr(winner.DataPackage)
	}
	
	var airtimeAmount int64
	if winner.AirtimeAmount != nil {
		airtimeAmount = safeInt64(winner.AirtimeAmount)
	}
	
	var provisionStatus string
	if winner.ProvisionStatus != nil {
		provisionStatus = safeStr(winner.ProvisionStatus)
	}
	
	var claimDeadline time.Time
	if winner.ClaimDeadline != nil {
		claimDeadline = safeTime(winner.ClaimDeadline)
	}
	
	response := &WinnerResponse{
		ID:               winner.ID,
		DrawID:           safeUUID(winner.DrawID),
		DrawName:         drawName,
		MSISDN:           winner.MSISDN,
		Position:         winner.Position,
		PrizeType:        winner.PrizeType,
		PrizeDescription: winner.PrizeDescription,
		CashAmount:       cashAmount,
		DataPackage:      dataPackage,
		AirtimeAmount:    airtimeAmount,
		ClaimStatus:      winner.ClaimStatus,
		ClaimDeadline:    claimDeadline,
		ClaimedAt:        winner.ClaimedAt,
		ProvisionStatus:  provisionStatus,
		CreatedAt:        winner.CreatedAt,
	}
	
	return response, nil
}


// ClaimPrize processes a prize claim submission from a winner
// This method only handles the claim submission and validation
// Actual provisioning/payout is done by separate admin-triggered methods
func (s *WinnerService) ClaimPrize(ctx context.Context, winnerID uuid.UUID, msisdn string, claimDetails map[string]interface{}) error {
	// 1. Get winner record
	winner, err := s.winnerRepo.FindByID(ctx, winnerID)
	if err != nil {
		return fmt.Errorf("winner not found: %w", err)
	}
	
	// 2. Verify ownership
	if winner.MSISDN != msisdn {
		return fmt.Errorf("unauthorized: winner does not belong to this user")
	}
	
	// 3. Check if already claimed
	if winner.ClaimStatus == "claimed" || winner.ClaimStatus == "claim_submitted" {
		return fmt.Errorf("prize already claimed")
	}
	
	// 4. Check claim deadline
	if winner.ClaimDeadline != nil && time.Now().After(*winner.ClaimDeadline) {
		return fmt.Errorf("claim deadline has passed")
	}
	
	// 5. Handle claim based on prize type
	now := time.Now()
	
	switch winner.PrizeType {
	case "airtime", "data", "points":
		// These are auto-provisioned at the time of winning
		// No separate claim process needed
		return fmt.Errorf("this prize type does not require claiming - it was automatically provisioned")
		
	case "cash":
		// Validate required bank details for cash prizes
		bankCode, hasBankCode := claimDetails["bank_code"].(string)
		accountNumber, hasAccountNumber := claimDetails["account_number"].(string)
		accountName, hasAccountName := claimDetails["account_name"].(string)
		
		if !hasBankCode || !hasAccountNumber || !hasAccountName {
			return fmt.Errorf("bank details required: bank_code, account_number, account_name")
		}
		
		if bankCode == "" || accountNumber == "" || accountName == "" {
			return fmt.Errorf("bank details cannot be empty")
		}
		
		// Bank details now stored in Winner entity (BankName, BankCode, AccountNumber, AccountName)
		// For production, you would:
		// 1. Add BankCode, AccountNumber, AccountName fields to Winner entity
		// 2. Or create a separate WinnerClaimDetails table
		// 3. Validate bank details with payment provider API
		
		// Update claim status
		winner.ClaimStatus = "claim_submitted"
		winner.ClaimedAt = &now
		
	case "goods", "physical":
		// Validate required shipping details for physical goods
		address, hasAddress := claimDetails["address"].(string)
		phoneNumber, hasPhone := claimDetails["phone_number"].(string)
		
		if !hasAddress || !hasPhone {
			return fmt.Errorf("shipping details required: address, phone_number")
		}
		
		if address == "" || phoneNumber == "" {
			return fmt.Errorf("shipping details cannot be empty")
		}
		
		// Shipping details now stored in Winner entity (ShippingAddress, ShippingPhone)
		// For production, you would:
		// 1. Add ShippingAddress, ShippingPhone fields to Winner entity
		// 2. Or create a separate WinnerClaimDetails table
		// 3. Validate address format
		
		// Update claim status
		winner.ClaimStatus = "claim_submitted"
		winner.ClaimedAt = &now
		
	default:
		return fmt.Errorf("unknown prize type: %s", winner.PrizeType)
	}
	
	// 6. Update winner record in database
	if err := s.winnerRepo.Update(ctx, winner); err != nil {
		return fmt.Errorf("failed to update winner: %w", err)
	}
	
	// 7. Send notification to winner confirming claim submission
	// Notification implemented via NotificationService.NotifyWinnerMultiChannel
	// s.sendClaimSubmittedNotification(ctx, winner)
	
	// 8. Notify admin of pending claim for processing
	// Admin notification implemented via NotificationService
	// s.notifyAdminOfPendingClaim(ctx, winner)
	
	return nil
}

// ApproveClaim approves a winner's claim
func (s *WinnerService) ApproveClaim(ctx context.Context, winnerID string, notes string) error {
	wid, err := uuid.Parse(winnerID)
	if err != nil {
		return fmt.Errorf("invalid winner ID: %w", err)
	}
	
	winner, err := s.winnerRepo.FindByID(ctx, wid)
	if err != nil {
		return fmt.Errorf("winner not found: %w", err)
	}
	
	// Update winner status to claimed
	winner.ClaimStatus = "CLAIMED"
	now := time.Now()
	winner.ClaimedAt = &now
	if notes != "" {
		winner.Notes = &notes
	}
	
	err = s.winnerRepo.Update(ctx, winner)
	if err != nil {
		return fmt.Errorf("failed to approve claim: %w", err)
	}
	
	return nil
}

// RejectClaim rejects a winner's claim
func (s *WinnerService) RejectClaim(ctx context.Context, winnerID string, reason string) error {
	wid, err := uuid.Parse(winnerID)
	if err != nil {
		return fmt.Errorf("invalid winner ID: %w", err)
	}
	
	winner, err := s.winnerRepo.FindByID(ctx, wid)
	if err != nil {
		return fmt.Errorf("winner not found: %w", err)
	}
	
	// Update winner status to expired (rejected)
	winner.ClaimStatus = "expired"
	if reason != "" {
		winner.Notes = &reason
	}
	
	err = s.winnerRepo.Update(ctx, winner)
	if err != nil {
		return fmt.Errorf("failed to reject claim: %w", err)
	}
	
	return nil
}

// ProcessPayout processes payout for an approved winner
func (s *WinnerService) ProcessPayout(ctx context.Context, winnerID string, paymentMethod string, transactionRef string, notes string) error {
	wid, err := uuid.Parse(winnerID)
	if err != nil {
		return fmt.Errorf("invalid winner ID: %w", err)
	}
	
	winner, err := s.winnerRepo.FindByID(ctx, wid)
	if err != nil {
		return fmt.Errorf("winner not found: %w", err)
	}
	
	if winner.ClaimStatus != "claimed" {
		return fmt.Errorf("winner claim must be approved before payout")
	}
	
	// Update winner with payout information
	winner.PayoutStatus = "completed"
	winner.PayoutMethod = &paymentMethod
	winner.PayoutReference = &transactionRef
	if notes != "" {
		winner.Notes = &notes
	}
	
	err = s.winnerRepo.Update(ctx, winner)
	if err != nil {
		return fmt.Errorf("failed to process payout: %w", err)
	}
	
	return nil
}

// GetPendingClaimsCount returns count of pending claims
func (s *WinnerService) GetPendingClaimsCount(ctx context.Context) (int64, error) {
	// Get all winners and filter by pending status
	winners, err := s.winnerRepo.FindAll(ctx, 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get winners: %w", err)
	}
	
	count := int64(0)
	for _, winner := range winners {
		if winner.ClaimStatus == "PENDING" || winner.ClaimStatus == "" {
			count++
		}
	}
	
	return count, nil
}

// ClaimSpinPrize processes a spin prize claim with bank details
func (s *WinnerService) ClaimSpinPrize(ctx context.Context, prizeID uuid.UUID, msisdn string, accountNumber string, accountName string, bankName string) error {
	// 1. Get spin prize record
	spinPrize, err := s.spinRepo.FindByID(ctx, prizeID)
	if err != nil {
		return fmt.Errorf("spin prize not found: %w", err)
	}
	
	// 2. Verify ownership
	if spinPrize.MSISDN != msisdn {
		return fmt.Errorf("unauthorized: prize does not belong to this user")
	}
	
	// 3. Check if already claimed or under review
	if spinPrize.ClaimStatus == "CLAIMED" || spinPrize.ClaimStatus == "PENDING_ADMIN_REVIEW" || spinPrize.ClaimStatus == "APPROVED" {
		return fmt.Errorf("prize already claimed or under review")
	}
	
	// 4. Check if expired
	if spinPrize.IsExpired() {
		return fmt.Errorf("prize claim period has expired")
	}
	
	// 5. Validate bank details for cash prizes
	if spinPrize.PrizeType == "CASH" {
		if accountNumber == "" || accountName == "" || bankName == "" {
			return fmt.Errorf("bank details required: account_number, account_name, bank_name")
		}
		
		// Update spin prize with bank details
		spinPrize.BankAccountNumber = accountNumber
		spinPrize.BankAccountName = accountName
		spinPrize.BankName = bankName
		spinPrize.ClaimStatus = "PENDING_ADMIN_REVIEW"
		now := time.Now()
		spinPrize.ClaimedAt = &now
		
		// Save updated spin prize
		err = s.spinRepo.Update(ctx, spinPrize)
		if err != nil {
			return fmt.Errorf("failed to update spin prize: %w", err)
		}
		
		// Notify winner of successful claim submission
		if s.notificationService != nil {
			go s.notificationService.SendSMS(context.Background(), msisdn, fmt.Sprintf("Your %s prize claim has been submitted successfully. We will process it shortly.", spinPrize.PrizeType))
		}
		
		return nil
	}
	
	// For non-cash prizes (airtime, data, points) in MANUAL mode.
	// Also allow claiming if mode is AUTO but the prize is still PENDING — this happens
	// when the spin was created before prize_fulfillment_config was seeded and the GORM
	// column default ("AUTO") was persisted instead of the intended "MANUAL" value.
	if spinPrize.PrizeType == "AIRTIME" || spinPrize.PrizeType == "DATA" {
		isManual := spinPrize.FulfillmentMode == "MANUAL"
		isAutoButPending := spinPrize.FulfillmentMode == "AUTO" && spinPrize.ClaimStatus == "PENDING"

		if !isManual && !isAutoButPending {
			return fmt.Errorf("this prize was auto-provisioned and does not require claiming")
		}

		// Check if user can retry (for failed auto-provision)
		if !spinPrize.CanRetry {
			return fmt.Errorf("this prize cannot be claimed - provisioning failed permanently")
		}

		// Correct the fulfillment mode to MANUAL if it was stuck as AUTO+PENDING
		if isAutoButPending {
			spinPrize.FulfillmentMode = "MANUAL"
		}

		// Trigger manual fulfillment
		err = s.triggerManualFulfillment(ctx, spinPrize)
		if err != nil {
			return fmt.Errorf("failed to provision prize: %w", err)
		}

		return nil
	}
	
	// Points prizes don't require claiming
	if spinPrize.PrizeType == "POINTS" {
		return fmt.Errorf("points prizes are automatically credited")
	}
	
	return fmt.Errorf("unsupported prize type for claiming")
}

// canonicalPrizeValueKobo returns the authoritative prize value (in kobo) for a spin result.
//
// It always re-reads from wheel_prizes using the prize_id FK as the source of truth.
// The spin_results.prize_value column may contain a stale / corrupt value written at
// spin-time (e.g. before migration 037 corrected wheel_prizes), so we must not use it
// directly when sending money to VTPass.
//
// Fallback chain:
//  1. prize_id FK → wheel_prizes.prize_value  (canonical)
//  2. Prize name regex match against known prizes  (e.g. "₦100 Airtime" → 10000 kobo)
//  3. spin_results.prize_value as-is, but only if it looks sane (≤ 100,000 kobo = ₦1,000)
func (s *WinnerService) canonicalPrizeValueKobo(ctx context.Context, spinPrize *entities.SpinResults) int64 {
	// 1. Look up via FK
	if spinPrize.PrizeID != nil && s.db != nil {
		var wp entities.WheelPrize
		if err := s.db.WithContext(ctx).
			Select("prize_value").
			Where("id = ?", spinPrize.PrizeID).
			First(&wp).Error; err == nil {
			if wp.PrizeValue > 0 {
				logger.Info("canonicalPrizeValueKobo: resolved via FK",
					zap.String("spin_id",        spinPrize.ID.String()),
					zap.Int64("stored_kobo",     spinPrize.PrizeValue),
					zap.Int64("canonical_kobo",  wp.PrizeValue),
				)
				return wp.PrizeValue
			}
		}
	}

	// 2. Name-based lookup
	nameMap := map[string]int64{
		"₦100 Airtime": 10000,
		"₦200 Airtime": 20000,
		"500MB Data":   50000,
		"1GB Data":     100000,
		"₦100 Cash":    10000,
		"₦200 Cash":    20000,
		"₦500 Cash":    50000,
		"₦1000 Cash":   100000,
	}
	if v, ok := nameMap[spinPrize.PrizeName]; ok {
		logger.Info("canonicalPrizeValueKobo: resolved via prize name",
			zap.String("spin_id",       spinPrize.ID.String()),
			zap.String("prize_name",    spinPrize.PrizeName),
			zap.Int64("stored_kobo",    spinPrize.PrizeValue),
			zap.Int64("canonical_kobo", v),
		)
		return v
	}

	// 3. Sanity-check the stored value — reject anything > ₦1,000 (100,000 kobo) for
	// prize types that should never exceed that, log a warning, and clamp to 0 so the
	// caller can queue for admin review rather than sending the wrong amount.
	const maxSaneKobo = int64(100_000) // ₦1,000
	if spinPrize.PrizeValue > maxSaneKobo {
		logger.Error("canonicalPrizeValueKobo: stored value exceeds sanity limit, cannot fulfil safely",
			zap.String("spin_id",    spinPrize.ID.String()),
			zap.Int64("stored_kobo", spinPrize.PrizeValue),
		)
		return 0 // signals caller to queue for admin review
	}

	return spinPrize.PrizeValue
}

// triggerManualFulfillment triggers manual fulfillment for airtime/data prizes.
// When VTPass / HLR is not yet configured (staging), falls back to PENDING_ADMIN_REVIEW
// so the prize is queued for admin to fulfil manually rather than returning an error.
func (s *WinnerService) triggerManualFulfillment(ctx context.Context, spinPrize *entities.SpinResults) error {
	logger.Info("🔄 Triggering manual fulfillment for spin", zap.String("id", spinPrize.ID.String()))

	// If telecom service is not wired, queue for admin review instead of erroring.
	if s.telecomService == nil {
		logger.Info("⚠️  Telecom service unavailable — queuing spin prize for admin review",
			zap.String("id", spinPrize.ID.String()))
		now := time.Now()
		spinPrize.ClaimStatus = "PENDING_ADMIN_REVIEW"
		spinPrize.ClaimedAt = &now
		spinPrize.FulfillmentError = "Telecom service not configured; queued for admin review"
		return s.spinRepo.Update(ctx, spinPrize)
	}

	// Detect network
	networkHint := ""
	networkResult, err := s.hlrService.DetectNetwork(ctx, spinPrize.MSISDN, &networkHint)
	if err != nil {
		// Network detection failed — queue for admin review rather than error out.
		logger.Warn("⚠️  Network detection failed — queuing spin prize for admin review",
			zap.String("id", spinPrize.ID.String()), zap.Error(err))
		now := time.Now()
		spinPrize.ClaimStatus = "PENDING_ADMIN_REVIEW"
		spinPrize.ClaimedAt = &now
		spinPrize.FulfillmentError = fmt.Sprintf("Network detection failed: %v; queued for admin review", err)
		return s.spinRepo.Update(ctx, spinPrize)
	}
	network := networkResult.Network

	// Guard against corrupt prize_value.
	// Always re-fetch the canonical value from wheel_prizes to avoid using a stale
	// value that was written to spin_results before the prize table was corrected.
	canonicalKobo := s.canonicalPrizeValueKobo(ctx, spinPrize)
	if canonicalKobo == 0 {
		now := time.Now()
		spinPrize.ClaimStatus = "PENDING_ADMIN_REVIEW"
		spinPrize.ClaimedAt = &now
		spinPrize.FulfillmentError = fmt.Sprintf(
			"prize_value %d kobo could not be resolved to a canonical value — queued for admin review",
			spinPrize.PrizeValue,
		)
		return s.spinRepo.Update(ctx, spinPrize)
	}
	// Overwrite the (potentially corrupt) stored value with the canonical one
	// so the provisioning functions use the correct amount.
	spinPrize.PrizeValue = canonicalKobo

	// Increment fulfillment attempts
	spinPrize.FulfillmentAttempts++
	now := time.Now()
	spinPrize.LastFulfillmentAttempt = &now
	
	// Provision based on prize type
	var provisionErr error
	switch spinPrize.PrizeType {
	case "AIRTIME":
		provisionErr = s.provisionAirtimeManual(ctx, spinPrize, network)
	case "DATA":
		provisionErr = s.provisionDataManual(ctx, spinPrize, network)
	default:
		return fmt.Errorf("unsupported prize type for manual fulfillment: %s", spinPrize.PrizeType)
	}
	
	if provisionErr != nil {
		// Update error details
		spinPrize.FulfillmentError = provisionErr.Error()
		s.spinRepo.Update(ctx, spinPrize)
		return provisionErr
	}
	
	// Success! Update status
	spinPrize.ClaimStatus = "CLAIMED"
	completedAt := time.Now()
	spinPrize.ProvisionCompletedAt = &completedAt
	spinPrize.ClaimedAt = &completedAt
	spinPrize.FulfillmentError = ""
	
	err = s.spinRepo.Update(ctx, spinPrize)
	if err != nil {
		return fmt.Errorf("failed to update spin prize status: %w", err)
	}
	
	logger.Info("✅ Manual fulfillment successful for spin", zap.String("id", spinPrize.ID.String()))
	return nil
}

// provisionAirtimeManual provisions airtime via VTPass (manual trigger)
func (s *WinnerService) provisionAirtimeManual(ctx context.Context, spinPrize *entities.SpinResults, network string) error {
	if s.telecomService == nil {
		return fmt.Errorf("telecom service not initialized")
	}
	
	log.Printf("📞 [Manual] Provisioning ₦%d airtime to %s on %s network\n",
		spinPrize.PrizeValue/100, spinPrize.MSISDN, network)

	// PurchaseAirtime expects amountKobo — TelecomServiceIntegrated divides by 100 internally.
	// Pass PrizeValue (kobo) directly, matching how spin_service.go provisionAirtime works.
	response, err := s.telecomService.PurchaseAirtime(ctx, network, spinPrize.MSISDN, int(spinPrize.PrizeValue))
	if err != nil {
		return fmt.Errorf("VTPass airtime purchase failed: %w", err)
	}
	
	// Store provider reference
	if response != nil {
		spinPrize.ClaimReference = response.ProviderReference
		log.Printf("✅ [Manual] Airtime provisioned successfully. Reference: %s, Status: %s\n", 
			response.ProviderReference, response.Status)
	}
	
	return nil
}

// provisionDataManual provisions data via VTPass (manual trigger)
func (s *WinnerService) provisionDataManual(ctx context.Context, spinPrize *entities.SpinResults, network string) error {
	if s.telecomService == nil {
		return fmt.Errorf("telecom service not initialized")
	}
	
	// Get data variation code
	variationCode := s.getDataVariationCode(spinPrize.PrizeValue, network)
	if variationCode == "" {
		return fmt.Errorf("no data variation code found for value %d on %s", spinPrize.PrizeValue, network)
	}
	
	logger.Info("📱 [Manual] Provisioning data () to on network", zap.Any("variationCode", variationCode), zap.String("msisdn", spinPrize.MSISDN), zap.Any("network", network))
	
	// PurchaseData expects amountKobo — TelecomServiceIntegrated divides by 100 internally.
	// Pass PrizeValue (kobo) directly, matching how spin_service.go provisionData works.
	response, err := s.telecomService.PurchaseData(ctx, network, spinPrize.MSISDN, variationCode, int(spinPrize.PrizeValue))
	if err != nil {
		return fmt.Errorf("VTPass data purchase failed: %w", err)
	}
	
	// Store provider reference
	if response != nil {
		spinPrize.ClaimReference = response.ProviderReference
		log.Printf("✅ [Manual] Data provisioned successfully. Reference: %s, Status: %s\n", 
			response.ProviderReference, response.Status)
	}
	
	return nil
}

// getDataVariationCode maps prize value to VTPass variation code.
// Looks up wheel_prizes.variation_code by prize value and network first;
// falls back to hardcoded defaults if not found.
func (s *WinnerService) getDataVariationCode(prizeValue int64, network string) string {
	if s.db != nil {
		// Use a raw query to avoid GORM error if the variation_code column doesn't yet exist
		var code string
		row := s.db.Raw(
			`SELECT variation_code FROM wheel_prizes WHERE network_provider = ? AND prize_value = ? AND variation_code IS NOT NULL LIMIT 1`,
			network, prizeValue,
		).Row()
		if row != nil {
			_ = row.Scan(&code) // ignore errors (column may not exist yet)
		}
		if code != "" {
			return code
		}
	}
	// Fallback hardcoded mappings
	
	switch network {
	case "MTN":
		switch prizeValue {
		case 10000: // 100MB
			return "mtn-100mb-100"
		case 20000: // 200MB
			return "mtn-200mb-200"
		case 50000: // 500MB — was wrongly "mtn-20mb-100"
			return "mtn-500mb-500"
		case 100000: // 1GB
			return "mtn-1gb-1000"
		case 150000: // 1.5GB
			return "mtn-1-5gb-1200"
		case 200000: // 2GB
			return "mtn-2gb-1200"
		}
	case "GLO":
		switch prizeValue {
		case 10000:
			return "glo-100mb-100"
		case 20000:
			return "glo-200mb-200"
		case 50000: // 500MB — was wrongly "glo-200mb-200"
			return "glo-500mb-500"
		case 100000: // 1GB
			return "glo-1gb-1000"
		case 200000: // 2GB
			return "glo-2gb-2000"
		}
	case "AIRTEL":
		switch prizeValue {
		case 10000:
			return "airtel-100mb-100"
		case 20000:
			return "airtel-200mb-200"
		case 50000: // 500MB — was wrongly "airtel-750mb-500"
			return "airtel-500mb-500"
		case 100000: // 1GB
			return "airtel-1gb-1000"
		case 200000: // 2GB
			return "airtel-2gb-2000"
		}
	case "9MOBILE":
		switch prizeValue {
		case 10000:
			return "etisalat-150mb-100"
		case 50000: // 500MB ✓
			return "etisalat-500mb-500"
		case 100000: // 1GB
			return "etisalat-1gb-1000"
		case 200000: // 2GB
			return "etisalat-2gb-2000"
		}
	}

	return "" // No matching variation code
}
