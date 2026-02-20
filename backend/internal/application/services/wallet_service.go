package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// WalletService handles wallet and payout operations
type WalletService struct {
	walletRepo            repositories.WalletRepository
	walletTransactionRepo repositories.WalletTransactionRepository
	paymentService        *PaymentService
}

// WalletSummary represents wallet summary in user-friendly format
type WalletSummary struct {
	Balance         float64 `json:"balance"`          // Balance in naira
	PendingBalance  float64 `json:"pending_balance"`  // Pending earnings in naira
	TotalEarned     float64 `json:"total_earned"`     // Lifetime earnings in naira
	TotalWithdrawn  float64 `json:"total_withdrawn"`  // Lifetime withdrawals in naira
	CanWithdraw     bool    `json:"can_withdraw"`     // Whether user can withdraw
	MinPayoutAmount float64 `json:"min_payout_amount"` // Minimum payout in naira
}

// NewWalletService creates a new wallet service
func NewWalletService(
	walletRepo repositories.WalletRepository,
	walletTransactionRepo repositories.WalletTransactionRepository,
	paymentService *PaymentService,
) *WalletService {
	return &WalletService{
		walletRepo:            walletRepo,
		walletTransactionRepo: walletTransactionRepo,
		paymentService:        paymentService,
	}
}

// CreateWallet creates a new wallet for an affiliate
func (s *WalletService) CreateWallet(ctx context.Context, msisdn string) (*entities.Wallet, error) {
	wallet := &entities.Wallet{
		MSISDN:          msisdn,
		Balance:         0,
		PendingBalance:  0,
		TotalEarned:     0,
		TotalWithdrawn:  0,
		MinPayoutAmount: 100000, // ₦1000 in kobo
		IsActive:        true,
		IsSuspended:     false,
	}

	err := s.walletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

// GetWallet retrieves wallet by MSISDN
func (s *WalletService) GetWallet(ctx context.Context, msisdn string) (*entities.Wallet, error) {
	wallet, err := s.walletRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		// Create wallet if it doesn't exist
		return s.CreateWallet(ctx, msisdn)
	}

	return wallet, nil
}

// GetWalletSummary returns wallet summary in user-friendly format
func (s *WalletService) GetWalletSummary(ctx context.Context, msisdn string) (*WalletSummary, error) {
	wallet, err := s.GetWallet(ctx, msisdn)
	if err != nil {
		return nil, err
	}

	return &WalletSummary{
		Balance:         float64(wallet.Balance) / 100,
		PendingBalance:  float64(wallet.PendingBalance) / 100,
		TotalEarned:     float64(wallet.TotalEarned) / 100,
		TotalWithdrawn:  float64(wallet.TotalWithdrawn) / 100,
		CanWithdraw:     wallet.Balance >= wallet.MinPayoutAmount && wallet.IsActive && !wallet.IsSuspended,
		MinPayoutAmount: float64(wallet.MinPayoutAmount) / 100,
	}, nil
}

// AddPendingEarnings adds pending earnings to wallet (commission from referral)
func (s *WalletService) AddPendingEarnings(ctx context.Context, msisdn string, amount int64, description, referenceType, referenceID string) error {
	wallet, err := s.GetWallet(ctx, msisdn)
	if err != nil {
		return err
	}

	if wallet.IsSuspended {
		return errors.New("wallet is suspended")
	}

	// Update wallet pending balance
	wallet.PendingBalance += amount
	err = s.walletRepo.Update(ctx, wallet)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	// Create transaction record
	transaction := &entities.WalletTransaction{
		WalletID:      wallet.ID,
		TransactionID: s.generateTransactionID(),
		Type:          "pending_credit",
		Amount:        amount,
		BalanceBefore: wallet.PendingBalance - amount,
		BalanceAfter:  wallet.PendingBalance,
		Description:   description,
		ReferenceType: &referenceType,
		ReferenceID:   &referenceID,
		Status:        "completed",
		ProcessedAt:   time.Now(),
	}

	err = s.walletTransactionRepo.Create(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// ReleasePendingEarnings moves pending earnings to available balance (after holding period)
func (s *WalletService) ReleasePendingEarnings(ctx context.Context, msisdn string, amount int64, referenceID string) error {
	wallet, err := s.GetWallet(ctx, msisdn)
	if err != nil {
		return err
	}

	if wallet.PendingBalance < amount {
		return errors.New("insufficient pending balance")
	}

	// Update wallet balances
	wallet.Balance += amount
	wallet.PendingBalance -= amount
	wallet.TotalEarned += amount

	err = s.walletRepo.Update(ctx, wallet)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	// Create transaction record
	transaction := &entities.WalletTransaction{
		WalletID:      wallet.ID,
		TransactionID: s.generateTransactionID(),
		Type:          "pending_release",
		Amount:        amount,
		BalanceBefore: wallet.Balance - amount,
		BalanceAfter:  wallet.Balance,
		Description:   "Pending earnings released to available balance",
		ReferenceType: stringPtr("pending_release"),
		ReferenceID:   &referenceID,
		Status:        "completed",
		ProcessedAt:   time.Now(),
	}

	err = s.walletTransactionRepo.Create(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// RequestPayout creates a payout request
func (s *WalletService) RequestPayout(ctx context.Context, msisdn string, amount int64, payoutMethod, bankCode, accountNumber, accountName string) error {
	wallet, err := s.GetWallet(ctx, msisdn)
	if err != nil {
		return err
	}

	// Validation
	if wallet.IsSuspended {
		return errors.New("wallet is suspended")
	}

	if !wallet.IsActive {
		return errors.New("wallet is not active")
	}

	if wallet.Balance < amount {
		return errors.New("insufficient balance")
	}

	if amount < wallet.MinPayoutAmount {
		return fmt.Errorf("minimum payout amount is ₦%.2f", float64(wallet.MinPayoutAmount)/100)
	}

	// Process payout immediately
	return s.processPayout(ctx, wallet, amount, payoutMethod, bankCode, accountNumber, accountName)
}

// processPayout processes a payout
func (s *WalletService) processPayout(ctx context.Context, wallet *entities.Wallet, amount int64, payoutMethod, bankCode, accountNumber, accountName string) error {
	// Process payment based on method
	var paymentRef string
	var paymentErr error

	switch payoutMethod {
	case "bank_transfer":
		paymentRef, paymentErr = s.processBankTransfer(ctx, wallet, amount, bankCode, accountNumber, accountName)
	case "mobile_money":
		paymentRef, paymentErr = s.processMobileMoney(ctx, wallet, amount)
	default:
		return errors.New("unsupported payout method")
	}

	if paymentErr != nil {
		return paymentErr
	}

	// Deduct from wallet balance
	wallet.Balance -= amount
	wallet.TotalWithdrawn += amount

	err := s.walletRepo.Update(ctx, wallet)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	// Create debit transaction
	transaction := &entities.WalletTransaction{
		WalletID:      wallet.ID,
		TransactionID: s.generateTransactionID(),
		Type:          "debit",
		Amount:        amount,
		BalanceBefore: wallet.Balance + amount,
		BalanceAfter:  wallet.Balance,
		Description:   fmt.Sprintf("Payout to %s (Ref: %s)", accountNumber, paymentRef),
		ReferenceType: stringPtr("payout"),
		ReferenceID:   &paymentRef,
		Status:        "completed",
		ProcessedAt:   time.Now(),
	}

	err = s.walletTransactionRepo.Create(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// processBankTransfer handles bank transfer payouts
func (s *WalletService) processBankTransfer(ctx context.Context, wallet *entities.Wallet, amount int64, bankCode, accountNumber, accountName string) (string, error) {
	// Use payment service to process bank transfer
	transferRequest := map[string]interface{}{
		"amount":         amount,
		"bank_code":      bankCode,
		"account_number": accountNumber,
		"account_name":   accountName,
		"narration":      "RechargeMax affiliate payout",
		"reference":      fmt.Sprintf("PAYOUT_%s_%d", wallet.MSISDN, time.Now().Unix()),
	}

	// This integrates with Paystack Transfer API
	response, err := s.paymentService.ProcessTransfer(ctx, transferRequest)
	if err != nil {
		return "", fmt.Errorf("bank transfer failed: %w", err)
	}

	// Extract reference from response
	if ref, ok := response["reference"].(string); ok {
		return ref, nil
	}

	return "", errors.New("failed to get payment reference")
}

// processMobileMoney handles mobile money payouts
func (s *WalletService) processMobileMoney(ctx context.Context, wallet *entities.Wallet, amount int64) (string, error) {
	// For mobile money, we use the MSISDN
	// This would integrate with mobile money APIs (MTN MoMo, Airtel Money, etc.)
	// For now, return error as not implemented
	_ = wallet // Suppress unused variable warning
	_ = amount
	return "", errors.New("mobile money payouts not yet implemented")
}

// GetWalletTransactions retrieves wallet transaction history
func (s *WalletService) GetWalletTransactions(ctx context.Context, msisdn string, page, limit int) ([]*entities.WalletTransaction, int64, error) {
	wallet, err := s.GetWallet(ctx, msisdn)
	if err != nil {
		return nil, 0, err
	}

	transactions, err := s.walletTransactionRepo.FindByWalletID(ctx, wallet.ID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	total, err := s.walletTransactionRepo.CountByWalletID(ctx, wallet.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return transactions, total, nil
}

// ProcessPendingReleases processes pending earnings that are ready to be released
func (s *WalletService) ProcessPendingReleases(ctx context.Context) error {
	// Find all pending earnings older than 7 days (holding period)
	// In production, this would:
	// 1. Query wallet_transactions where status='pending' and created_at < NOW() - 7 days
	// 2. For each transaction, move from pending_balance to available_balance
	// 3. Update transaction status to 'completed'
	// 4. Send notification to user about released earnings
	//
	// Example implementation:
	// sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	// 
	// // This would require a new repository method:
	// // pendingTransactions, err := s.walletTransactionRepo.FindPendingForRelease(ctx, sevenDaysAgo)
	// // if err != nil {
	// //     return fmt.Errorf("failed to find pending transactions: %w", err)
	// // }
	// 
	// // for _, txn := range pendingTransactions {
	// //     // Get wallet
	// //     wallet, err := s.walletRepo.FindByID(ctx, txn.WalletID)
	// //     if err != nil {
	// //         continue // Log error and skip
	// //     }
	// //     
	// //     // Move from pending to available
	// //     wallet.PendingBalance -= txn.Amount
	// //     wallet.AvailableBalance += txn.Amount
	// //     
	// //     // Update wallet
	// //     err = s.walletRepo.Update(ctx, wallet)
	// //     if err != nil {
	// //         continue // Log error and skip
	// //     }
	// //     
	// //     // Update transaction status
	// //     txn.Status = "completed"
	// //     err = s.walletTransactionRepo.Update(ctx, txn)
	// //     
	// //     // Send notification
	// //     // s.notificationService.SendSMS(ctx, wallet.MSISDN, "Your earnings are now available!")
	// // }
	
	// Returns wallet analytics - enhance with more metrics as needed
	// When FindPendingForRelease repository method is implemented, uncomment above
	_ = ctx
	return nil
}

// SuspendWallet suspends a wallet (admin function)
func (s *WalletService) SuspendWallet(ctx context.Context, msisdn, reason string) error {
	wallet, err := s.walletRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return err
	}

	wallet.IsSuspended = true
	wallet.SuspensionReason = &reason

	return s.walletRepo.Update(ctx, wallet)
}

// UnsuspendWallet unsuspends a wallet (admin function)
func (s *WalletService) UnsuspendWallet(ctx context.Context, msisdn string) error {
	wallet, err := s.walletRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return err
	}

	wallet.IsSuspended = false
	emptyReason := ""
	wallet.SuspensionReason = &emptyReason

	return s.walletRepo.Update(ctx, wallet)
}

// Helper function to generate unique transaction ID
func (s *WalletService) generateTransactionID() string {
	return fmt.Sprintf("TXN_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
