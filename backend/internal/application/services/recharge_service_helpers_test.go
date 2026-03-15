package services

// recharge_service_test.go — tests for unexported helpers and integration logic.
// Since normalizePhoneToInternational is unexported, these tests live in the
// same package (white-box testing).

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ─── normalizePhoneToInternational ────────────────────────────────────────────

func TestNormalizePhoneToInternational_LocalFormat(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"08012345678", "2348012345678"},   // standard local 08XX
		{"07011223344", "2347011223344"},   // 07XX prefix
		{"09099887766", "2349099887766"},   // 09XX prefix
		{"0701 123 4567", "2347011234567"}, // spaces stripped
		{"+2348012345678", "2348012345678"}, // already international with +
		{"2348012345678", "2348012345678"},  // already international without +
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizePhoneToInternational(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestNormalizePhoneToInternational_FallbackForUnknownFormats(t *testing.T) {
	// 10-digit number (not 11 local, not 13 intl) — returned as digits only
	result := normalizePhoneToInternational("0123456789")
	assert.Equal(t, "0123456789", result)
}

func TestNormalizePhoneToInternational_EmptyString(t *testing.T) {
	result := normalizePhoneToInternational("")
	assert.Equal(t, "", result)
}

func TestNormalizePhoneToInternational_StripsNonDigits(t *testing.T) {
	// Dashes, dots, spaces all stripped
	result := normalizePhoneToInternational("080-123-45678")
	assert.Equal(t, "2348012345678", result)
}
