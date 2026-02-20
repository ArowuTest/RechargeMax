package utils

import (
	"fmt"
)

// SpinTier represents a tier in the spin system
// All amounts are in kobo (1 Naira = 100 kobo)
type SpinTier struct {
	Name      string
	MinAmount int64 // Amount in kobo
	MaxAmount int64 // Amount in kobo
	SpinsPerDay int
}

// All spin tiers (amounts in kobo)
var SpinTiers = []SpinTier{
	{
		Name:        "Bronze",
		MinAmount:   100000,  // ₦1,000
		MaxAmount:   499999,  // ₦4,999.99
		SpinsPerDay: 1,
	},
	{
		Name:        "Silver",
		MinAmount:   500000,  // ₦5,000
		MaxAmount:   999999,  // ₦9,999.99
		SpinsPerDay: 2,
	},
	{
		Name:        "Gold",
		MinAmount:   1000000, // ₦10,000
		MaxAmount:   1999999, // ₦19,999.99
		SpinsPerDay: 3,
	},
	{
		Name:        "Platinum",
		MinAmount:   2000000, // ₦20,000
		MaxAmount:   4999999, // ₦49,999.99
		SpinsPerDay: 5,
	},
	{
		Name:        "Diamond",
		MinAmount:   5000000,      // ₦50,000
		MaxAmount:   99999999999,  // Effectively unlimited
		SpinsPerDay: 10,
	},
}

// GetSpinTier returns the spin tier for a given daily recharge amount (in kobo)
func GetSpinTier(dailyRechargeAmountKobo int64) (*SpinTier, error) {
	if dailyRechargeAmountKobo < 100000 { // ₦1,000 in kobo
		return nil, fmt.Errorf("minimum recharge amount for spins is ₦1,000")
	}
	
	for _, tier := range SpinTiers {
		if dailyRechargeAmountKobo >= tier.MinAmount && dailyRechargeAmountKobo <= tier.MaxAmount {
			return &tier, nil
		}
	}
	
	// Convert kobo to Naira for error message
	nairaAmount := float64(dailyRechargeAmountKobo) / 100.0
	return nil, fmt.Errorf("unable to determine spin tier for amount: ₦%.2f", nairaAmount)
}

// CalculateSpinsEarned calculates the number of spins earned for a daily recharge amount (in kobo)
func CalculateSpinsEarned(dailyRechargeAmountKobo int64) (int, string, error) {
	tier, err := GetSpinTier(dailyRechargeAmountKobo)
	if err != nil {
		return 0, "", err
	}
	
	return tier.SpinsPerDay, tier.Name, nil
}

// GetAllTiers returns all available spin tiers
func GetAllTiers() []SpinTier {
	return SpinTiers
}

// GetTierByName returns a tier by its name
func GetTierByName(name string) (*SpinTier, error) {
	for _, tier := range SpinTiers {
		if tier.Name == name {
			return &tier, nil
		}
	}
	return nil, fmt.Errorf("tier not found: %s", name)
}
