package utils

import (
	"testing"
)

func TestNormalizePhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Local format
		{"Local format", "08012345678", "08012345678", false},
		{"Local format with spaces", "0801 234 5678", "08012345678", false},
		{"Local format with dashes", "0801-234-5678", "08012345678", false},
		
		// International format
		{"International format", "2348012345678", "08012345678", false},
		{"International format with plus", "+2348012345678", "08012345678", false},
		{"International format with spaces", "+234 801 234 5678", "08012345678", false},
		{"International format with dashes", "+234-801-234-5678", "08012345678", false},
		
		// Different prefixes
		{"070 prefix", "07012345678", "07012345678", false},
		{"071 prefix", "07112345678", "07112345678", false},
		{"080 prefix", "08012345678", "08012345678", false},
		{"081 prefix", "08112345678", "08112345678", false},
		{"090 prefix", "09012345678", "09012345678", false},
		{"091 prefix", "09112345678", "09112345678", false},
		
		// Invalid formats
		{"Empty", "", "", true},
		{"Too short", "0801234567", "", true},
		{"Too long", "080123456789", "", true},
		{"Landline", "01012345678", "", true},
		{"Invalid prefix", "06012345678", "", true},
		{"No leading zero", "8012345678", "", true},
		{"Random number", "1234567890", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizePhoneNumber(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizePhoneNumber() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("NormalizePhoneNumber() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("NormalizePhoneNumber() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestToInternationalFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"Local to international", "08012345678", "2348012345678", false},
		{"Already international", "2348012345678", "2348012345678", false},
		{"With formatting", "+234 801 234 5678", "2348012345678", false},
		{"Invalid", "1234567890", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToInternationalFormat(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ToInternationalFormat() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ToInternationalFormat() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("ToInternationalFormat() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestFormatPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		format   string
		expected string
		wantErr  bool
	}{
		{"Local format", "08012345678", "local", "08012345678", false},
		{"International format", "08012345678", "international", "+234 801 234 5678", false},
		{"Compact format", "08012345678", "compact", "0801-234-5678", false},
		{"Default format", "08012345678", "", "08012345678", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatPhoneNumber(tt.input, tt.format)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("FormatPhoneNumber() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("FormatPhoneNumber() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("FormatPhoneNumber() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestIsLocalFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Local format", "08012345678", true},
		{"Local with spaces", "0801 234 5678", true},
		{"International format", "2348012345678", false},
		{"Too short", "0801234567", false},
		{"Invalid", "1234567890", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLocalFormat(tt.input)
			if result != tt.expected {
				t.Errorf("IsLocalFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsInternationalFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"International format", "2348012345678", true},
		{"International with plus", "+2348012345678", true},
		{"Local format", "08012345678", false},
		{"Too short", "234801234567", false},
		{"Invalid", "1234567890", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInternationalFormat(tt.input)
			if result != tt.expected {
				t.Errorf("IsInternationalFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMaskPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Local format", "08012345678", "0801***5678"},
		{"International format", "2348012345678", "234801***5678"},
		{"Short number", "12345", "***"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskPhoneNumber(tt.input)
			if result != tt.expected {
				t.Errorf("MaskPhoneNumber() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsMobileNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Mobile 070", "07012345678", true},
		{"Mobile 080", "08012345678", true},
		{"Mobile 090", "09012345678", true},
		{"Landline 010", "01012345678", false},
		{"Landline 060", "06012345678", false},
		{"Invalid", "1234567890", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMobileNumber(tt.input)
			if result != tt.expected {
				t.Errorf("IsMobileNumber() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetPhoneNumberType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Mobile", "08012345678", "mobile"},
		{"Mobile international", "2348012345678", "mobile"},
		{"Landline", "01012345678", "landline"},
		{"Invalid", "1234567890", "invalid"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPhoneNumberType(tt.input)
			if result != tt.expected {
				t.Errorf("GetPhoneNumberType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComparePhoneNumbers(t *testing.T) {
	tests := []struct {
		name     string
		phone1   string
		phone2   string
		expected bool
	}{
		{"Same local format", "08012345678", "08012345678", true},
		{"Local vs international", "08012345678", "2348012345678", true},
		{"With formatting", "0801 234 5678", "+234-801-234-5678", true},
		{"Different numbers", "08012345678", "08112345678", false},
		{"Invalid vs valid", "1234567890", "08012345678", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePhoneNumbers(tt.phone1, tt.phone2)
			if result != tt.expected {
				t.Errorf("ComparePhoneNumbers() = %v, want %v", result, tt.expected)
			}
		})
	}
}
