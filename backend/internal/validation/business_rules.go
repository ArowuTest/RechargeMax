package validation

import (
	"fmt"
	"time"
)

// Business rule constants
const (
	// Recharge rules
	MinRechargeAmount = 50.0
	MaxRechargeAmount = 50000.0
	PointsPerNaira    = 200.0 // ₦200 = 1 point
	
	// Subscription rules
	SubscriptionDailyAmount = 20.0 // ₦20 per day
	SubscriptionPointsPerDay = 1   // 1 point per day
	
	// Wheel spin rules
	MinSpinEligibilityAmount = 1000.0 // ₦1000 minimum to be eligible for spin
	
	// Commission rules
	DefaultCommissionRate = 1.0  // 1% default
	MinCommissionRate     = 0.0  // 0% minimum
	MaxCommissionRate     = 100.0 // 100% maximum
	
	// Draw rules
	MinDrawParticipants = 10 // Minimum participants for a draw
	
	// Fraud detection thresholds
	MaxTransactionsPerHour = 10
	MaxTransactionsPerDay  = 50
	MaxAmountPerDay        = 100000.0 // ₦100,000 per day
)

// ValidateRechargeEligibility validates if a recharge is eligible
func ValidateRechargeEligibility(amount float64, userStatus string) error {
	// Validate amount
	if err := ValidateRechargeAmount(amount); err != nil {
		return err
	}
	
	// Check user status
	if userStatus == "SUSPENDED" || userStatus == "BANNED" {
		return fmt.Errorf("account is suspended or banned")
	}
	
	return nil
}

// ValidateSpinEligibility validates if a user is eligible for wheel spin
func ValidateSpinEligibility(rechargeAmount float64, hasUsedSpin bool) error {
	// Check minimum amount
	if rechargeAmount < MinSpinEligibilityAmount {
		return fmt.Errorf("recharge amount must be at least ₦%.2f to be eligible for spin", MinSpinEligibilityAmount)
	}
	
	// Check if spin already used
	if hasUsedSpin {
		return fmt.Errorf("spin already used for this recharge")
	}
	
	return nil
}

// CalculatePoints calculates points earned from a recharge amount
func CalculatePoints(amount float64) int64 {
	return int64(amount / PointsPerNaira)
}

// ValidatePointsCalculation validates if points calculation is correct
func ValidatePointsCalculation(amount float64, points int64) error {
	expectedPoints := CalculatePoints(amount)
	if points != expectedPoints {
		return fmt.Errorf("incorrect points calculation: expected %d, got %d", expectedPoints, points)
	}
	return nil
}

// CalculateCommission calculates commission amount
func CalculateCommission(amount float64, rate float64) float64 {
	return amount * (rate / 100.0)
}

// ValidateCommissionCalculation validates if commission calculation is correct
func ValidateCommissionCalculation(amount, rate, commission float64) error {
	expectedCommission := CalculateCommission(amount, rate)
	// Allow small floating point differences
	diff := expectedCommission - commission
	if diff < -0.01 || diff > 0.01 {
		return fmt.Errorf("incorrect commission calculation: expected %.2f, got %.2f", expectedCommission, commission)
	}
	return nil
}

// ValidateSubscriptionEligibility validates if a user can subscribe
func ValidateSubscriptionEligibility(hasActiveSubscription bool, userStatus string) error {
	// Check if already subscribed
	if hasActiveSubscription {
		return fmt.Errorf("user already has an active subscription")
	}
	
	// Check user status
	if userStatus == "SUSPENDED" || userStatus == "BANNED" {
		return fmt.Errorf("account is suspended or banned")
	}
	
	return nil
}

// ValidateDrawEligibility validates if a draw can be executed
func ValidateDrawEligibility(totalParticipants int, drawDate time.Time) error {
	// Check minimum participants
	if totalParticipants < MinDrawParticipants {
		return fmt.Errorf("draw requires at least %d participants, got %d", MinDrawParticipants, totalParticipants)
	}
	
	// Check if draw date has passed
	if time.Now().Before(drawDate) {
		return fmt.Errorf("draw date has not yet arrived")
	}
	
	return nil
}

// ValidatePrizeValue validates prize value based on type
func ValidatePrizeValue(prizeType string, value float64) error {
	switch prizeType {
	case "AIRTIME":
		// Airtime: ₦50 - ₦10,000
		if value < 50 || value > 10000 {
			return fmt.Errorf("airtime prize must be between ₦50 and ₦10,000")
		}
	case "DATA":
		// Data: ₦100 - ₦20,000
		if value < 100 || value > 20000 {
			return fmt.Errorf("data prize must be between ₦100 and ₦20,000")
		}
	case "CASH":
		// Cash: ₦100 - ₦1,000,000
		if value < 100 || value > 1000000 {
			return fmt.Errorf("cash prize must be between ₦100 and ₦1,000,000")
		}
	case "PHYSICAL":
		// Physical prizes: any positive value (represents estimated value)
		if value <= 0 {
			return fmt.Errorf("physical prize value must be positive")
		}
	default:
		return fmt.Errorf("invalid prize type")
	}
	
	return nil
}

// ValidateFraudThresholds validates if transaction exceeds fraud detection thresholds
func ValidateFraudThresholds(transactionsInHour, transactionsInDay int, amountInDay float64) error {
	if transactionsInHour > MaxTransactionsPerHour {
		return fmt.Errorf("exceeded maximum transactions per hour (%d)", MaxTransactionsPerHour)
	}
	
	if transactionsInDay > MaxTransactionsPerDay {
		return fmt.Errorf("exceeded maximum transactions per day (%d)", MaxTransactionsPerDay)
	}
	
	if amountInDay > MaxAmountPerDay {
		return fmt.Errorf("exceeded maximum amount per day (₦%.2f)", MaxAmountPerDay)
	}
	
	return nil
}

// ValidateAffiliateEligibility validates if a user can become an affiliate
func ValidateAffiliateEligibility(totalRecharges int, totalAmount float64, accountAge time.Duration) error {
	// Minimum 5 recharges
	if totalRecharges < 5 {
		return fmt.Errorf("minimum 5 recharges required to become an affiliate")
	}
	
	// Minimum ₦5,000 total recharge amount
	if totalAmount < 5000 {
		return fmt.Errorf("minimum ₦5,000 total recharge amount required to become an affiliate")
	}
	
	// Account must be at least 7 days old
	if accountAge < 7*24*time.Hour {
		return fmt.Errorf("account must be at least 7 days old to become an affiliate")
	}
	
	return nil
}

// ValidateWithdrawalEligibility validates if a withdrawal can be processed
func ValidateWithdrawalEligibility(availableBalance, withdrawalAmount, minimumBalance float64) error {
	// Check if sufficient balance
	if withdrawalAmount > availableBalance {
		return fmt.Errorf("insufficient balance: available ₦%.2f, requested ₦%.2f", availableBalance, withdrawalAmount)
	}
	
	// Check minimum withdrawal amount
	if withdrawalAmount < 1000 {
		return fmt.Errorf("minimum withdrawal amount is ₦1,000")
	}
	
	// Check if withdrawal would leave account below minimum balance
	if availableBalance-withdrawalAmount < minimumBalance {
		return fmt.Errorf("withdrawal would leave balance below minimum (₦%.2f)", minimumBalance)
	}
	
	return nil
}

// ValidatePrizeClaimEligibility validates if a prize can be claimed
func ValidatePrizeClaimEligibility(claimStatus string, claimDeadline time.Time) error {
	// Check claim status
	if claimStatus != "UNCLAIMED" {
		return fmt.Errorf("prize has already been claimed or is being processed")
	}
	
	// Check if claim deadline has passed
	if time.Now().After(claimDeadline) {
		return fmt.Errorf("claim deadline has passed")
	}
	
	return nil
}

// ValidateReferralEligibility validates if a referral is valid
func ValidateReferralEligibility(referrerID, refereeID string, referrerStatus string) error {
	// Check if referrer and referee are different
	if referrerID == refereeID {
		return fmt.Errorf("cannot refer yourself")
	}
	
	// Check referrer status
	if referrerStatus != "ACTIVE" {
		return fmt.Errorf("referrer account is not active")
	}
	
	return nil
}

// ValidateNetworkOperationWindow validates if operation is within allowed time window
func ValidateNetworkOperationWindow(currentTime time.Time) error {
	// Nigerian networks typically have maintenance windows
	// Example: 2 AM - 4 AM maintenance window
	hour := currentTime.Hour()
	if hour >= 2 && hour < 4 {
		return fmt.Errorf("network operations are unavailable during maintenance window (2 AM - 4 AM)")
	}
	
	return nil
}

// ValidatePrizeProbabilitySum validates if total prize probabilities sum to 1.0
func ValidatePrizeProbabilitySum(probabilities []float64) error {
	var sum float64
	for _, prob := range probabilities {
		sum += prob
	}
	
	// Allow small floating point differences
	if sum < 0.99 || sum > 1.01 {
		return fmt.Errorf("total prize probabilities must sum to 1.0, got %.4f", sum)
	}
	
	return nil
}
