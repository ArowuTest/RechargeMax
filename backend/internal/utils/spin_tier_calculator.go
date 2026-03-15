package utils

import (
	"fmt"

	"gorm.io/gorm"
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

// --- DB-backed tier calculator ---

// SpinTierDB represents a spin tier from the database
type SpinTierDB struct {
	ID               string  `gorm:"column:id;primaryKey"`
	TierName         string  `gorm:"column:tier_name"`
	TierDisplayName  string  `gorm:"column:tier_display_name"`
	MinDailyAmount   int64   `gorm:"column:min_daily_amount"`   // Amount in kobo
	MaxDailyAmount   int64   `gorm:"column:max_daily_amount"`   // Amount in kobo
	SpinsPerDay      int     `gorm:"column:spins_per_day"`
	TierColor        string  `gorm:"column:tier_color"`
	TierIcon         string  `gorm:"column:tier_icon"`
	TierBadge        string  `gorm:"column:tier_badge"`
	Description      string  `gorm:"column:description"`
	SortOrder        int     `gorm:"column:sort_order"`
	IsActive         bool    `gorm:"column:is_active"`
	CreatedBy        *string `gorm:"column:created_by"` // UUID of admin who created
	UpdatedBy        *string `gorm:"column:updated_by"` // UUID of admin who last updated
}

func (SpinTierDB) TableName() string {
	return "spin_tiers"
}

// SpinTierCalculatorDB handles spin tier calculations using database
type SpinTierCalculatorDB struct {
	db *gorm.DB
}

// NewSpinTierCalculatorDB creates a new database-driven spin tier calculator
func NewSpinTierCalculatorDB(db *gorm.DB) *SpinTierCalculatorDB {
	return &SpinTierCalculatorDB{db: db}
}

// GetSpinTierFromDB returns the spin tier for a given daily recharge amount (in kobo)
func (c *SpinTierCalculatorDB) GetSpinTierFromDB(dailyRechargeAmountKobo int64) (*SpinTierDB, error) {
	if dailyRechargeAmountKobo < 100000 { // ₦1,000 in kobo
		return nil, fmt.Errorf("minimum recharge amount for spins is ₦1,000")
	}

	var tier SpinTierDB
	err := c.db.Where("is_active = ? AND min_daily_amount <= ? AND max_daily_amount >= ?",
		true, dailyRechargeAmountKobo, dailyRechargeAmountKobo).
		Order("sort_order ASC").
		First(&tier).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Convert kobo to Naira for error message
			nairaAmount := float64(dailyRechargeAmountKobo) / 100.0
			return nil, fmt.Errorf("unable to determine spin tier for amount: ₦%.2f", nairaAmount)
		}
		return nil, fmt.Errorf("database error while fetching spin tier: %v", err)
	}

	return &tier, nil
}

// CalculateSpinsEarnedFromDB calculates the number of spins earned for a daily recharge amount (in kobo)
func (c *SpinTierCalculatorDB) CalculateSpinsEarnedFromDB(dailyRechargeAmountKobo int64) (int, string, error) {
	tier, err := c.GetSpinTierFromDB(dailyRechargeAmountKobo)
	if err != nil {
		return 0, "", err
	}

	return tier.SpinsPerDay, tier.TierDisplayName, nil
}

// GetAllTiersFromDB returns all active spin tiers ordered by sort_order
func (c *SpinTierCalculatorDB) GetAllTiersFromDB() ([]SpinTierDB, error) {
	var tiers []SpinTierDB
	err := c.db.Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&tiers).Error

	if err != nil {
		return nil, fmt.Errorf("database error while fetching all tiers: %v", err)
	}

	return tiers, nil
}

// GetTierByNameFromDB returns a tier by its name
func (c *SpinTierCalculatorDB) GetTierByNameFromDB(name string) (*SpinTierDB, error) {
	var tier SpinTierDB
	err := c.db.Where("tier_name = ? AND is_active = ?", name, true).
		First(&tier).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tier not found: %s", name)
		}
		return nil, fmt.Errorf("database error while fetching tier: %v", err)
	}

	return &tier, nil
}

// GetTierByIDFromDB returns a tier by its ID
func (c *SpinTierCalculatorDB) GetTierByIDFromDB(id string) (*SpinTierDB, error) {
	var tier SpinTierDB
	err := c.db.Where("id = ?", id).First(&tier).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tier not found with ID: %s", id)
		}
		return nil, fmt.Errorf("database error while fetching tier: %v", err)
	}

	return &tier, nil
}

// ValidateTierConfiguration validates that tier ranges don't overlap and cover all amounts
func (c *SpinTierCalculatorDB) ValidateTierConfiguration() error {
	tiers, err := c.GetAllTiersFromDB()
	if err != nil {
		return err
	}

	if len(tiers) == 0 {
		return fmt.Errorf("no active tiers found in database")
	}

	// Check for gaps and overlaps
	for i := 0; i < len(tiers)-1; i++ {
		currentTier := tiers[i]
		nextTier := tiers[i+1]

		// Check for overlap
		if currentTier.MaxDailyAmount >= nextTier.MinDailyAmount {
			return fmt.Errorf("tier overlap detected: %s (max: %d) overlaps with %s (min: %d)",
				currentTier.TierDisplayName, currentTier.MaxDailyAmount,
				nextTier.TierDisplayName, nextTier.MinDailyAmount)
		}

		// Check for gap (max of current should be min of next - 1)
		expectedNextMin := currentTier.MaxDailyAmount + 1
		if nextTier.MinDailyAmount != expectedNextMin {
			return fmt.Errorf("tier gap detected: %s (max: %d) has gap before %s (min: %d)",
				currentTier.TierDisplayName, currentTier.MaxDailyAmount,
				nextTier.TierDisplayName, nextTier.MinDailyAmount)
		}
	}

	return nil
}

// RefreshTierCache can be called after admin updates tiers
// This is a placeholder for future caching implementation
func (c *SpinTierCalculatorDB) RefreshTierCache() error {
	// For now, just validate the configuration
	return c.ValidateTierConfiguration()
}
