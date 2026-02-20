package utils

import (
	"errors"
	"regexp"
	"strings"
)

// Phone number formats
const (
	LocalFormatLength        = 11  // 08012345678
	InternationalFormatLength = 13 // 2348012345678
	NigeriaCountryCode       = "234"
)

// NormalizeMSISDNToInternational converts any Nigerian phone number to international format
// This is the PRIMARY normalization function used throughout the application
// Accepts:
//   - Local format: 08012345678 (11 digits)
//   - International format: 2348012345678 (13 digits)
//   - Formatted: +234 803 123 4567, 234-803-123-4567, etc.
// Returns: 2348012345678 (international format)
func NormalizeMSISDNToInternational(phone string) (string, error) {
	if phone == "" {
		return "", errors.New("phone number is required")
	}
	
	// Remove all non-digit characters (spaces, dashes, plus signs, etc.)
	digitsOnly := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	
	// Handle local format (08XXXXXXXXXX → 234XXXXXXXXXX)
	if strings.HasPrefix(digitsOnly, "0") && len(digitsOnly) == LocalFormatLength {
		digitsOnly = NigeriaCountryCode + digitsOnly[1:] // Replace 0 with 234
	}
	
	// Validate final format
	if len(digitsOnly) != InternationalFormatLength {
		return "", errors.New("invalid phone number length")
	}
	
	if !strings.HasPrefix(digitsOnly, NigeriaCountryCode) {
		return "", errors.New("phone number must be a Nigerian number (234)")
	}
	
	// Validate mobile number (4th digit must be 7, 8, or 9)
	if len(digitsOnly) >= 4 {
		fourthDigit := string(digitsOnly[3])
		if fourthDigit != "7" && fourthDigit != "8" && fourthDigit != "9" {
			return "", errors.New("invalid mobile number: must start with 07, 08, or 09")
		}
	}
	
	return digitsOnly, nil
}

// NormalizeMSISDN is an alias for NormalizeMSISDNToInternational
// Provided for backward compatibility and cleaner API
func NormalizeMSISDN(phone string) (string, error) {
	return NormalizeMSISDNToInternational(phone)
}

// NormalizePhoneNumber converts phone numbers to local Nigerian format (08012345678)
// DEPRECATED: Use NormalizeMSISDN() instead for consistent international format
// Kept for backward compatibility only
func NormalizePhoneNumber(phone string) (string, error) {
	// Convert to international first
	international, err := NormalizeMSISDNToInternational(phone)
	if err != nil {
		return "", err
	}
	// Convert to local
	return "0" + international[3:], nil
}

// ToInternationalFormat converts local format to international format
// 08012345678 → 2348012345678
func ToInternationalFormat(phone string) (string, error) {
	// First normalize to ensure valid local format
	normalized, err := NormalizePhoneNumber(phone)
	if err != nil {
		return "", err
	}
	
	// Replace leading 0 with 234
	return NigeriaCountryCode + normalized[1:], nil
}

// FormatPhoneNumber formats phone number for display
// Options: "local" (08012345678), "international" (+234 803 123 4567), "compact" (0803-123-4567)
func FormatPhoneNumber(phone string, format string) (string, error) {
	// Normalize first
	normalized, err := NormalizePhoneNumber(phone)
	if err != nil {
		return "", err
	}
	
	switch format {
	case "local":
		// 08012345678
		return normalized, nil
		
	case "international":
		// +234 803 123 4567
		intl, _ := ToInternationalFormat(normalized)
		return "+" + intl[:3] + " " + intl[3:6] + " " + intl[6:9] + " " + intl[9:], nil
		
	case "compact":
		// 0803-123-4567
		return normalized[:4] + "-" + normalized[4:7] + "-" + normalized[7:], nil
		
	default:
		return normalized, nil
	}
}

// ValidatePhoneNumber validates phone number format (accepts both local and international)
func ValidatePhoneNumber(phone string) error {
	_, err := NormalizePhoneNumber(phone)
	return err
}

// IsLocalFormat checks if phone number is in local format (08012345678)
func IsLocalFormat(phone string) bool {
	digitsOnly := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	return len(digitsOnly) == LocalFormatLength && strings.HasPrefix(digitsOnly, "0")
}

// IsInternationalFormat checks if phone number is in international format (2348012345678)
func IsInternationalFormat(phone string) bool {
	digitsOnly := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	return len(digitsOnly) == InternationalFormatLength && strings.HasPrefix(digitsOnly, NigeriaCountryCode)
}

// ExtractDigits removes all non-digit characters from phone number
func ExtractDigits(phone string) string {
	return regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
}

// MaskMSISDN masks phone number for privacy
// 2348031234567 → 234803***4567
func MaskMSISDN(phone string) string {
	normalized, err := NormalizeMSISDN(phone)
	if err != nil {
		return "***"
	}
	
	// 234803***4567
	return normalized[:6] + "***" + normalized[9:]
}

// MaskPhoneNumber masks phone number for privacy
// DEPRECATED: Use MaskMSISDN() instead
// 08012345678 → 0801***5678
// 2348012345678 → 234801***5678
func MaskPhoneNumber(phone string) string {
	digitsOnly := ExtractDigits(phone)
	
	if len(digitsOnly) == LocalFormatLength {
		// 08012345678 → 0801***5678
		return digitsOnly[:4] + "***" + digitsOnly[7:]
	} else if len(digitsOnly) == InternationalFormatLength {
		// 2348012345678 → 234801***5678
		return digitsOnly[:6] + "***" + digitsOnly[9:]
	}
	
	// Fallback: mask middle portion
	if len(digitsOnly) > 6 {
		return digitsOnly[:3] + "***" + digitsOnly[len(digitsOnly)-3:]
	}
	
	return "***"
}

// IsMobileNumber checks if phone number is a mobile number (not landline)
// Mobile numbers start with 07, 08, or 09
func IsMobileNumber(phone string) bool {
	normalized, err := NormalizePhoneNumber(phone)
	if err != nil {
		return false
	}
	
	secondDigit := string(normalized[1])
	return secondDigit == "7" || secondDigit == "8" || secondDigit == "9"
}

// GetPhoneNumberType returns the type of phone number
// Returns: "mobile", "landline", "invalid"
func GetPhoneNumberType(phone string) string {
	digitsOnly := ExtractDigits(phone)
	
	// Convert international to local if needed
	if strings.HasPrefix(digitsOnly, NigeriaCountryCode) && len(digitsOnly) == InternationalFormatLength {
		digitsOnly = "0" + digitsOnly[3:]
	}
	
	if len(digitsOnly) != LocalFormatLength {
		return "invalid"
	}
	
	if !strings.HasPrefix(digitsOnly, "0") {
		return "invalid"
	}
	
	secondDigit := string(digitsOnly[1])
	if secondDigit == "7" || secondDigit == "8" || secondDigit == "9" {
		return "mobile"
	}
	
	return "landline"
}

// CompareMSISDN compares two phone numbers after normalization
// Returns true if they represent the same phone number
func CompareMSISDN(phone1, phone2 string) bool {
	normalized1, err1 := NormalizeMSISDN(phone1)
	normalized2, err2 := NormalizeMSISDN(phone2)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return normalized1 == normalized2
}

// ComparePhoneNumbers checks if two phone numbers are the same
// DEPRECATED: Use CompareMSISDN() instead
func ComparePhoneNumbers(phone1, phone2 string) bool {
	normalized1, err1 := NormalizePhoneNumber(phone1)
	normalized2, err2 := NormalizePhoneNumber(phone2)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return normalized1 == normalized2
}
