package utils

import (
	"fmt"
	"math"
)

// Currency conversion constants
const (
	KoboPerNaira = 100
)

// NairaToKobo converts Naira (as float64) to kobo (as int64)
// Example: 1000.50 Naira = 100050 kobo
func NairaToKobo(naira float64) int64 {
	return int64(math.Round(naira * KoboPerNaira))
}

// KoboToNaira converts kobo (as int64) to Naira (as float64)
// Example: 100050 kobo = 1000.50 Naira
func KoboToNaira(kobo int64) float64 {
	return float64(kobo) / KoboPerNaira
}

// FormatNaira formats kobo as a Naira string with currency symbol
// Example: 100050 kobo = "₦1,000.50"
func FormatNaira(kobo int64) string {
	naira := KoboToNaira(kobo)
	return fmt.Sprintf("₦%.2f", naira)
}

// FormatNairaWithCommas formats kobo as a Naira string with thousands separators
// Example: 10005000 kobo = "₦100,050.00"
func FormatNairaWithCommas(kobo int64) string {
	naira := KoboToNaira(kobo)
	
	// Split into integer and decimal parts
	intPart := int64(naira)
	decPart := int64(math.Round((naira - float64(intPart)) * 100))
	
	// Format integer part with commas
	intStr := fmt.Sprintf("%d", intPart)
	if len(intStr) > 3 {
		// Add commas every 3 digits from right
		result := ""
		for i, digit := range reverse(intStr) {
			if i > 0 && i%3 == 0 {
				result = "," + result
			}
			result = string(digit) + result
		}
		intStr = result
	}
	
	return fmt.Sprintf("₦%s.%02d", intStr, decPart)
}

// ParseNairaToKobo parses a Naira string to kobo
// Supports formats: "1000", "1000.50", "₦1000", "₦1,000.50"
func ParseNairaToKobo(nairaStr string) (int64, error) {
	// Remove currency symbol and commas
	cleaned := ""
	for _, ch := range nairaStr {
		if ch >= '0' && ch <= '9' || ch == '.' {
			cleaned += string(ch)
		}
	}
	
	var naira float64
	_, err := fmt.Sscanf(cleaned, "%f", &naira)
	if err != nil {
		return 0, fmt.Errorf("invalid naira amount: %s", nairaStr)
	}
	
	return NairaToKobo(naira), nil
}

// ValidateAmount validates that an amount in kobo is within acceptable range
func ValidateAmount(kobo int64, min, max int64) error {
	if kobo < min {
		return fmt.Errorf("amount %s is below minimum %s", FormatNaira(kobo), FormatNaira(min))
	}
	if kobo > max {
		return fmt.Errorf("amount %s exceeds maximum %s", FormatNaira(kobo), FormatNaira(max))
	}
	return nil
}

// Helper function to reverse a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsValidKoboAmount checks if a kobo amount is valid (positive)
func IsValidKoboAmount(kobo int64) bool {
	return kobo > 0
}

// CalculatePercentage calculates a percentage of an amount in kobo
// Example: CalculatePercentage(100000, 10) = 10000 (10% of ₦1000)
func CalculatePercentage(kobo int64, percentage float64) int64 {
	return int64(math.Round(float64(kobo) * percentage / 100.0))
}

// AddAmounts adds multiple kobo amounts safely
func AddAmounts(amounts ...int64) int64 {
	total := int64(0)
	for _, amount := range amounts {
		total += amount
	}
	return total
}

// SubtractAmount subtracts one amount from another, ensuring non-negative result
func SubtractAmount(from, subtract int64) (int64, error) {
	result := from - subtract
	if result < 0 {
		return 0, fmt.Errorf("insufficient funds: %s - %s = %s", 
			FormatNaira(from), FormatNaira(subtract), FormatNaira(result))
	}
	return result, nil
}
