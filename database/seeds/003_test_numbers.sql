-- ========================================
-- TEST NUMBERS FOR NETWORK VALIDATION
-- ========================================
-- Purpose: Pre-validated test numbers for each network
-- Use Case: Testing network validation without real HLR API
-- Format: 234 prefix (e.g., 2348031234567)
-- Cache TTL: 365 days (permanent for testing)
-- ========================================

-- Clean existing test numbers (if any)
-- Using user_selection as source (valid constraint value)
DELETE FROM network_cache WHERE hlr_provider = 'test_seed';

-- ========================================
-- MTN TEST NUMBERS (5 numbers)
-- ========================================
INSERT INTO network_cache (
  msisdn, 
  network, 
  lookup_source, 
  confidence_level, 
  hlr_provider,
  last_verified_at,
  cache_expires_at,
  verification_count,
  is_valid
) VALUES
  ('2348031234567', 'MTN', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2347031234568', 'MTN', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2349031234569', 'MTN', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2348061234570', 'MTN', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2348131234571', 'MTN', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true);

-- ========================================
-- AIRTEL TEST NUMBERS (5 numbers)
-- ========================================
INSERT INTO network_cache (
  msisdn, 
  network, 
  lookup_source, 
  confidence_level, 
  hlr_provider,
  last_verified_at,
  cache_expires_at,
  verification_count,
  is_valid
) VALUES
  ('2348021234567', 'AIRTEL', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2347081234568', 'AIRTEL', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2349021234569', 'AIRTEL', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2348081234570', 'AIRTEL', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2347011234571', 'AIRTEL', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true);

-- ========================================
-- GLO TEST NUMBERS (5 numbers)
-- ========================================
INSERT INTO network_cache (
  msisdn, 
  network, 
  lookup_source, 
  confidence_level, 
  hlr_provider,
  last_verified_at,
  cache_expires_at,
  verification_count,
  is_valid
) VALUES
  ('2348051234567', 'GLO', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2347051234568', 'GLO', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2349051234569', 'GLO', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2348071234570', 'GLO', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2348151234571', 'GLO', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true);

-- ========================================
-- 9MOBILE TEST NUMBERS (5 numbers)
-- ========================================
INSERT INTO network_cache (
  msisdn, 
  network, 
  lookup_source, 
  confidence_level, 
  hlr_provider,
  last_verified_at,
  cache_expires_at,
  verification_count,
  is_valid
) VALUES
  ('2348091234567', '9MOBILE', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2348181234568', '9MOBILE', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2349091234569', '9MOBILE', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2348171234570', '9MOBILE', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true),
  ('2349081234571', '9MOBILE', 'user_selection', 'high', 'test_seed', NOW(), NOW() + INTERVAL '365 days', 1, true);

-- ========================================
-- VERIFICATION QUERY
-- ========================================
\echo ''
\echo '========================================='
\echo 'TEST NUMBERS SEEDED SUCCESSFULLY'
\echo '========================================='
\echo ''

-- Count by network
SELECT 
  network,
  COUNT(*) as count,
  'user_selection' as source
FROM network_cache 
WHERE hlr_provider = 'test_seed'
GROUP BY network
ORDER BY network;

\echo ''
\echo 'Total test numbers:'
SELECT COUNT(*) as total_test_numbers 
FROM network_cache 
WHERE hlr_provider = 'test_seed';

\echo ''
\echo '========================================='
\echo 'SAMPLE TEST NUMBERS (for UI testing):'
\echo '========================================='
\echo ''
\echo 'MTN:     08031234567 (or 2348031234567)'
\echo 'AIRTEL:  08021234567 (or 2348021234567)'
\echo 'GLO:     08051234567 (or 2348051234567)'
\echo '9MOBILE: 08091234567 (or 2348091234567)'
\echo ''
\echo '========================================='
\echo 'TEST NUMBERS READY FOR VALIDATION!'
\echo '========================================='
