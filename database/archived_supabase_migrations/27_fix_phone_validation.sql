-- Migration: Fix Phone Number Validation (Local + International Formats)
-- Description: Accepts both local (0803...) and international (234803...) formats
-- Created: 2026-02-01
-- Updated: 2026-02-01 (Added international format support)
-- Champion Developer Review: Phone Validation Fix

-- ============================================================================
-- IMPORTANT: NUMBER PORTABILITY IN NIGERIA
-- ============================================================================
-- Nigerian users can port their numbers between networks (MTN, Airtel, Glo, 9mobile)
-- Example: 0803 could be MTN today, Glo tomorrow
-- Therefore: We ONLY validate format, NOT network prefix
-- Network detection: Done via VTU provider API (VTPass, etc.)
-- ============================================================================

-- ============================================================================
-- SUPPORTED FORMATS
-- ============================================================================
-- Local Format:         0803XXXXXXX  (11 digits, starts with 0)
-- International Format: 234803XXXXXXX (13 digits, starts with 234)
-- 
-- Both formats are accepted in input
-- Backend normalizes to local format (0803...) for storage
-- ============================================================================

-- ============================================================================
-- VALIDATION STRATEGY
-- ============================================================================
-- ✅ DO: Accept both local (0803...) and international (234803...) formats
-- ✅ DO: Normalize international to local format before storage
-- ✅ DO: Validate mobile prefix (07, 08, 09)
-- ❌ DON'T: Assume network based on prefix
-- ❌ DON'T: Hardcode network mappings
-- ============================================================================

-- ============================================================================
-- USERS TABLE - Fix MSISDN validation
-- ============================================================================

-- Drop old constraint
ALTER TABLE users 
DROP CONSTRAINT IF EXISTS users_msisdn_check;

-- Add new constraint: Accepts both local and international formats
-- Local:         ^0[7-9][0-9]{9}$        (11 digits: 0803XXXXXXX)
-- International: ^234[7-9][0-9]{9}$      (13 digits: 234803XXXXXXX)
ALTER TABLE users 
ADD CONSTRAINT users_msisdn_check 
CHECK (msisdn ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$');

-- ============================================================================
-- TRANSACTIONS TABLE - Fix MSISDN validation
-- ============================================================================

-- Drop old constraint
ALTER TABLE transactions 
DROP CONSTRAINT IF EXISTS transactions_msisdn_check;

-- Add new constraint: Accepts both formats
ALTER TABLE transactions 
ADD CONSTRAINT transactions_msisdn_check 
CHECK (msisdn ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$');

-- ============================================================================
-- VTU_TRANSACTIONS TABLE - Fix phone_number validation
-- ============================================================================

-- Drop old constraint
ALTER TABLE vtu_transactions 
DROP CONSTRAINT IF EXISTS vtu_transactions_phone_number_check;

-- Add new constraint: Accepts both formats
ALTER TABLE vtu_transactions 
ADD CONSTRAINT vtu_transactions_phone_number_check 
CHECK (phone_number ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$');

-- ============================================================================
-- OTP_CODES TABLE - Fix MSISDN validation
-- ============================================================================

-- Drop old constraint
ALTER TABLE otp_codes 
DROP CONSTRAINT IF EXISTS otp_codes_msisdn_check;

-- Add new constraint: Accepts both formats
ALTER TABLE otp_codes 
ADD CONSTRAINT otp_codes_msisdn_check 
CHECK (msisdn ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$');

-- ============================================================================
-- WHEEL_SPINS TABLE - Fix MSISDN validation
-- ============================================================================

-- Drop old constraint
ALTER TABLE wheel_spins 
DROP CONSTRAINT IF EXISTS wheel_spins_msisdn_check;

-- Add new constraint: Accepts both formats
ALTER TABLE wheel_spins 
ADD CONSTRAINT wheel_spins_msisdn_check 
CHECK (msisdn ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$');

-- ============================================================================
-- DRAW_WINNERS TABLE - Fix MSISDN validation
-- ============================================================================

-- Drop old constraint
ALTER TABLE draw_winners 
DROP CONSTRAINT IF EXISTS draw_winners_msisdn_check;

-- Add new constraint: Accepts both formats
ALTER TABLE draw_winners 
ADD CONSTRAINT draw_winners_msisdn_check 
CHECK (msisdn ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$');

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Test the new regex pattern
DO $$
BEGIN
    -- Valid LOCAL format numbers (11 digits)
    ASSERT '07012345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid local 070 number rejected';
    ASSERT '07112345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid local 071 number rejected';
    ASSERT '08012345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid local 080 number rejected';
    ASSERT '08112345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid local 081 number rejected';
    ASSERT '09012345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid local 090 number rejected';
    ASSERT '09112345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid local 091 number rejected';
    
    -- Valid INTERNATIONAL format numbers (13 digits)
    ASSERT '2347012345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid intl 234-70 number rejected';
    ASSERT '2347112345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid intl 234-71 number rejected';
    ASSERT '2348012345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid intl 234-80 number rejected';
    ASSERT '2348112345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid intl 234-81 number rejected';
    ASSERT '2349012345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid intl 234-90 number rejected';
    ASSERT '2349112345678' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Valid intl 234-91 number rejected';
    
    -- Invalid numbers (should fail)
    ASSERT NOT '0601234567890' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Invalid 060 landline accepted';
    ASSERT NOT '070123456' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Short local number accepted';
    ASSERT NOT '07012345678910' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Long local number accepted';
    ASSERT NOT '70123456789' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Number without leading 0 accepted';
    ASSERT NOT '234601234567890' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Invalid 234-60 landline accepted';
    ASSERT NOT '23470123456' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Short intl number accepted';
    ASSERT NOT '1234567890' ~ '^(0[7-9][0-9]{9}|234[7-9][0-9]{9})$', 'Random number accepted';
    
    RAISE NOTICE 'Phone validation tests passed!';
END $$;

-- ============================================================================
-- NOTES ON PHONE NUMBER NORMALIZATION
-- ============================================================================

/*
NORMALIZATION STRATEGY:

When user enters a phone number, the backend should normalize it to local format
before storing in the database.

Example:
  User Input:    234803XXXXXXX  (international)
  Normalized:    0803XXXXXXX    (local)
  Stored:        0803XXXXXXX    (local)

This ensures:
- Consistent storage format
- Easier querying and indexing
- Simpler API integration (most Nigerian APIs expect local format)

BACKEND NORMALIZATION FUNCTION (Go):

func NormalizePhoneNumber(phone string) string {
    // Remove all non-digit characters
    phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
    
    // Convert international format to local format
    if strings.HasPrefix(phone, "234") && len(phone) == 13 {
        phone = "0" + phone[3:]  // 234803XXXXXXX -> 0803XXXXXXX
    }
    
    return phone
}

FRONTEND NORMALIZATION (TypeScript):

function normalizePhoneNumber(phone: string): string {
    // Remove all non-digit characters
    phone = phone.replace(/\D/g, '');
    
    // Convert international format to local format
    if (phone.startsWith('234') && phone.length === 13) {
        phone = '0' + phone.substring(3);  // 234803XXXXXXX -> 0803XXXXXXX
    }
    
    return phone;
}

USAGE:
1. User enters: "234 803 123 4567" or "+234-803-123-4567"
2. Frontend removes formatting: "2348031234567"
3. Frontend normalizes: "08031234567"
4. Backend validates: ✅ Matches regex
5. Backend stores: "08031234567"
*/

-- ============================================================================
-- NOTES ON NETWORK DETECTION
-- ============================================================================

/*
HOW NETWORK DETECTION WORKS:

1. User enters phone number (e.g., 08012345678 or 2348012345678)
2. Backend normalizes to local format (08012345678)
3. Backend validates format
4. Backend calls VTU provider API (VTPass, etc.) with normalized number
5. VTU provider returns actual network (MTN, Airtel, Glo, 9mobile)
6. Backend stores network in transaction record

WHY WE DON'T HARDCODE PREFIXES:

- Number Portability: Users can port numbers between networks
- Prefix Changes: Networks can acquire new prefixes
- Accuracy: Only the network provider knows the true network
- Maintainability: No need to update code when prefixes change

EXAMPLE API CALL (VTPass):

POST https://api.vtpass.com/api/merchant-verify
{
  "serviceID": "mtn",
  "billersCode": "08012345678"
}

Response:
{
  "code": "000",
  "content": {
    "Customer_Name": "John Doe",
    "Current_Bouquet": "MTN",
    "Status": "ACTIVE"
  }
}

The API tells us the ACTUAL network, regardless of prefix.
*/

-- ============================================================================
-- NIGERIAN PHONE NUMBER FORMATS
-- ============================================================================

/*
LOCAL FORMAT (11 digits):
- Format: 0[7-9][0-9]{9}
- Example: 08012345678
- Used by: Most Nigerian services and APIs

INTERNATIONAL FORMAT (13 digits):
- Format: 234[7-9][0-9]{9}
- Example: 2348012345678
- Used by: International calls, some apps

CONVERSION:
- Local → International: Replace leading 0 with 234
  08012345678 → 2348012345678
  
- International → Local: Replace leading 234 with 0
  2348012345678 → 08012345678

STORAGE RECOMMENDATION:
- Always store in LOCAL format (0803...)
- Normalize international format to local before storage
- Display in user's preferred format

VALIDATION:
- Accept both formats in input
- Validate format (not network)
- Normalize before storage
*/

-- ============================================================================
-- EXAMPLES OF VALID NUMBERS
-- ============================================================================

/*
LOCAL FORMAT (11 digits):
070XXXXXXXX ✅
071XXXXXXXX ✅
072XXXXXXXX ✅
...
079XXXXXXXX ✅

080XXXXXXXX ✅
081XXXXXXXX ✅
082XXXXXXXX ✅
...
089XXXXXXXX ✅

090XXXXXXXX ✅
091XXXXXXXX ✅
092XXXXXXXX ✅
...
099XXXXXXXX ✅

INTERNATIONAL FORMAT (13 digits):
2347XXXXXXXXX ✅
2348XXXXXXXXX ✅
2349XXXXXXXXX ✅

INVALID FORMATS:
060XXXXXXXX ❌ (landline)
01XXXXXXXXX ❌ (landline)
070XXXXXXX  ❌ (too short)
070XXXXXXXXX ❌ (too long)
70XXXXXXXX  ❌ (missing leading 0)
234601234567890 ❌ (landline)
23470123456 ❌ (too short)
*/

-- ============================================================================
-- ROLLBACK
-- ============================================================================

/*
-- To rollback to old validation:

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_msisdn_check;
ALTER TABLE users ADD CONSTRAINT users_msisdn_check CHECK (msisdn ~ '^(070|080|081|090|091)[0-9]{8}$');

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_msisdn_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_msisdn_check CHECK (msisdn ~ '^(070|080|081|090|091)[0-9]{8}$');

ALTER TABLE vtu_transactions DROP CONSTRAINT IF EXISTS vtu_transactions_phone_number_check;
ALTER TABLE vtu_transactions ADD CONSTRAINT vtu_transactions_phone_number_check CHECK (phone_number ~ '^(070|080|081|090|091)[0-9]{8}$');

ALTER TABLE otp_codes DROP CONSTRAINT IF EXISTS otp_codes_msisdn_check;
ALTER TABLE otp_codes ADD CONSTRAINT otp_codes_msisdn_check CHECK (msisdn ~ '^(070|080|081|090|091)[0-9]{8}$');

ALTER TABLE wheel_spins DROP CONSTRAINT IF EXISTS wheel_spins_msisdn_check;
ALTER TABLE wheel_spins ADD CONSTRAINT wheel_spins_msisdn_check CHECK (msisdn ~ '^(070|080|081|090|091)[0-9]{8}$');

ALTER TABLE draw_winners DROP CONSTRAINT IF EXISTS draw_winners_msisdn_check;
ALTER TABLE draw_winners ADD CONSTRAINT draw_winners_msisdn_check CHECK (msisdn ~ '^(070|080|081|090|091)[0-9]{8}$');
*/
