-- ============================================================================
-- RECHARGEMAX MASTER PRODUCTION SEED DATA (Schema-Aligned)
-- ============================================================================
-- This file contains ALL essential data needed for a production-ready platform
-- Aligned with actual GORM-generated database schema
-- ============================================================================

\echo '========================================================================'
\echo 'RECHARGEMAX MASTER PRODUCTION SEED - LOADING...'
\echo '========================================================================'

-- ============================================================================
-- PART 1: ESSENTIAL REFERENCE DATA
-- ============================================================================

\echo ''
\echo '--- Part 1: Loading Essential Reference Data ---'

-- Clean existing reference data
\echo 'Cleaning existing reference data...'
TRUNCATE TABLE data_plans CASCADE;
TRUNCATE TABLE network_configs CASCADE;
TRUNCATE TABLE subscription_tiers CASCADE;
TRUNCATE TABLE wheel_prizes CASCADE;
TRUNCATE TABLE admin_users CASCADE;

-- ============================================================================
-- 1.1 NETWORK CONFIGURATIONS (4 Nigerian Networks)
-- ============================================================================

\echo 'Loading network configurations...'

INSERT INTO network_configs (
  id,
  network_name,
  network_code,
  is_active,
  airtime_enabled,
  data_enabled,
  commission_rate,
  minimum_amount,
  maximum_amount,
  logo_url,
  brand_color,
  sort_order,
  created_at,
  updated_at
) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'MTN Nigeria', 'MTN', true, true, true, 2.50, 5000, 5000000, 'https://example.com/mtn-logo.png', '#FFCC00', 1, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Airtel Nigeria', 'AIRTEL', true, true, true, 2.50, 5000, 5000000, 'https://example.com/airtel-logo.png', '#FF0000', 2, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Glo Mobile', 'GLO', true, true, true, 3.00, 5000, 5000000, 'https://example.com/glo-logo.png', '#00AA00', 3, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', '9mobile', '9MOBILE', true, true, true, 3.50, 5000, 5000000, 'https://example.com/9mobile-logo.png', '#006600', 4, NOW(), NOW());

\echo '✓ 4 network configurations loaded'

-- ============================================================================
-- 1.2 DATA PLANS (66 plans across all networks)
-- ============================================================================

\echo 'Loading data plans...'

-- MTN Data Plans (18 plans)
INSERT INTO data_plans (
  id,
  network_provider,
  plan_name,
  data_amount,
  price,
  validity_days,
  plan_code,
  is_active,
  sort_order,
  description,
  created_at,
  updated_at
) VALUES
('650e8400-e29b-41d4-a716-446655440001', 'MTN', 'MTN 500MB Daily', '500MB', 350.00, 1, 'MTN-500MB-1D', true, 1, '500MB valid for 1 day', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440002', 'MTN', 'MTN 1GB Daily', '1GB', 500.00, 1, 'MTN-1GB-1D', true, 2, '1GB valid for 1 day', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440003', 'MTN', 'MTN 2GB Weekly', '2GB', 1000.00, 7, 'MTN-2GB-7D', true, 3, '2GB valid for 7 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440004', 'MTN', 'MTN 3GB Weekly', '3GB', 1500.00, 7, 'MTN-3GB-7D', true, 4, '3GB valid for 7 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440005', 'MTN', 'MTN 6GB Monthly', '6GB', 1500.00, 30, 'MTN-6GB-30D', true, 5, '6GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440006', 'MTN', 'MTN 10GB Monthly', '10GB', 2500.00, 30, 'MTN-10GB-30D', true, 6, '10GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440007', 'MTN', 'MTN 12GB Monthly', '12GB', 3000.00, 30, 'MTN-12GB-30D', true, 7, '12GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440008', 'MTN', 'MTN 20GB Monthly', '20GB', 5000.00, 30, 'MTN-20GB-30D', true, 8, '20GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440009', 'MTN', 'MTN 25GB Monthly', '25GB', 6000.00, 30, 'MTN-25GB-30D', true, 9, '25GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440010', 'MTN', 'MTN 40GB Monthly', '40GB', 10000.00, 30, 'MTN-40GB-30D', true, 10, '40GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440011', 'MTN', 'MTN 75GB Monthly', '75GB', 15000.00, 30, 'MTN-75GB-30D', true, 11, '75GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440012', 'MTN', 'MTN 100GB Monthly', '100GB', 20000.00, 30, 'MTN-100GB-30D', true, 12, '100GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440013', 'MTN', 'MTN 120GB Monthly', '120GB', 30000.00, 30, 'MTN-120GB-30D', true, 13, '120GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440014', 'MTN', 'MTN 150GB Monthly', '150GB', 35000.00, 30, 'MTN-150GB-30D', true, 14, '150GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440015', 'MTN', 'MTN 200GB Monthly', '200GB', 50000.00, 30, 'MTN-200GB-30D', true, 15, '200GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440016', 'MTN', 'MTN 300GB Monthly', '300GB', 75000.00, 30, 'MTN-300GB-30D', true, 16, '300GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440017', 'MTN', 'MTN 400GB Monthly', '400GB', 100000.00, 30, 'MTN-400GB-30D', true, 17, '400GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440018', 'MTN', 'MTN 500GB Monthly', '500GB', 150000.00, 30, 'MTN-500GB-30D', true, 18, '500GB valid for 30 days', NOW(), NOW());

-- Airtel Data Plans (18 plans)
INSERT INTO data_plans (
  id,
  network_provider,
  plan_name,
  data_amount,
  price,
  validity_days,
  plan_code,
  is_active,
  sort_order,
  description,
  created_at,
  updated_at
) VALUES
('650e8400-e29b-41d4-a716-446655440019', 'AIRTEL', 'Airtel 750MB Weekly', '750MB', 500.00, 7, 'AIRTEL-750MB-7D', true, 1, '750MB valid for 7 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440020', 'AIRTEL', 'Airtel 1.5GB Daily', '1.5GB', 350.00, 1, 'AIRTEL-1.5GB-1D', true, 2, '1.5GB valid for 1 day', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440021', 'AIRTEL', 'Airtel 2GB Weekly', '2GB', 1000.00, 7, 'AIRTEL-2GB-7D', true, 3, '2GB valid for 7 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440022', 'AIRTEL', 'Airtel 3GB Monthly', '3GB', 1500.00, 30, 'AIRTEL-3GB-30D', true, 4, '3GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440023', 'AIRTEL', 'Airtel 6GB Monthly', '6GB', 1500.00, 30, 'AIRTEL-6GB-30D', true, 5, '6GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440024', 'AIRTEL', 'Airtel 10GB Monthly', '10GB', 2500.00, 30, 'AIRTEL-10GB-30D', true, 6, '10GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440025', 'AIRTEL', 'Airtel 11GB Monthly', '11GB', 2000.00, 30, 'AIRTEL-11GB-30D', true, 7, '11GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440026', 'AIRTEL', 'Airtel 20GB Monthly', '20GB', 5000.00, 30, 'AIRTEL-20GB-30D', true, 8, '20GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440027', 'AIRTEL', 'Airtel 40GB Monthly', '40GB', 5000.00, 30, 'AIRTEL-40GB-30D', true, 9, '40GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440028', 'AIRTEL', 'Airtel 75GB Monthly', '75GB', 10000.00, 30, 'AIRTEL-75GB-30D', true, 10, '75GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440029', 'AIRTEL', 'Airtel 100GB Monthly', '100GB', 15000.00, 30, 'AIRTEL-100GB-30D', true, 11, '100GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440030', 'AIRTEL', 'Airtel 120GB Monthly', '120GB', 20000.00, 30, 'AIRTEL-120GB-30D', true, 12, '120GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440031', 'AIRTEL', 'Airtel 150GB Monthly', '150GB', 30000.00, 30, 'AIRTEL-150GB-30D', true, 13, '150GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440032', 'AIRTEL', 'Airtel 200GB Monthly', '200GB', 35000.00, 30, 'AIRTEL-200GB-30D', true, 14, '200GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440033', 'AIRTEL', 'Airtel 300GB Monthly', '300GB', 50000.00, 30, 'AIRTEL-300GB-30D', true, 15, '300GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440034', 'AIRTEL', 'Airtel 400GB Monthly', '400GB', 75000.00, 30, 'AIRTEL-400GB-30D', true, 16, '400GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440035', 'AIRTEL', 'Airtel 500GB Monthly', '500GB', 100000.00, 30, 'AIRTEL-500GB-30D', true, 17, '500GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440036', 'AIRTEL', 'Airtel 1TB Monthly', '1TB', 150000.00, 30, 'AIRTEL-1TB-30D', true, 18, '1TB valid for 30 days', NOW(), NOW());

-- Glo Data Plans (15 plans)
INSERT INTO data_plans (
  id,
  network_provider,
  plan_name,
  data_amount,
  price,
  validity_days,
  plan_code,
  is_active,
  sort_order,
  description,
  created_at,
  updated_at
) VALUES
('650e8400-e29b-41d4-a716-446655440037', 'GLO', 'Glo 500MB Daily', '500MB', 100.00, 1, 'GLO-500MB-1D', true, 1, '500MB valid for 1 day', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440038', 'GLO', 'Glo 1.6GB Weekly', '1.6GB', 500.00, 7, 'GLO-1.6GB-7D', true, 2, '1.6GB valid for 7 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440039', 'GLO', 'Glo 2.9GB Monthly', '2.9GB', 1000.00, 30, 'GLO-2.9GB-30D', true, 3, '2.9GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440040', 'GLO', 'Glo 5.8GB Monthly', '5.8GB', 2000.00, 30, 'GLO-5.8GB-30D', true, 4, '5.8GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440041', 'GLO', 'Glo 7.7GB Monthly', '7.7GB', 2500.00, 30, 'GLO-7.7GB-30D', true, 5, '7.7GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440042', 'GLO', 'Glo 10GB Monthly', '10GB', 3000.00, 30, 'GLO-10GB-30D', true, 6, '10GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440043', 'GLO', 'Glo 13.25GB Monthly', '13.25GB', 4000.00, 30, 'GLO-13.25GB-30D', true, 7, '13.25GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440044', 'GLO', 'Glo 18GB Monthly', '18GB', 5000.00, 30, 'GLO-18GB-30D', true, 8, '18GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440045', 'GLO', 'Glo 29.5GB Monthly', '29.5GB', 8000.00, 30, 'GLO-29.5GB-30D', true, 9, '29.5GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440046', 'GLO', 'Glo 50GB Monthly', '50GB', 10000.00, 30, 'GLO-50GB-30D', true, 10, '50GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440047', 'GLO', 'Glo 93GB Monthly', '93GB', 15000.00, 30, 'GLO-93GB-30D', true, 11, '93GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440048', 'GLO', 'Glo 119GB Monthly', '119GB', 18000.00, 30, 'GLO-119GB-30D', true, 12, '119GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440049', 'GLO', 'Glo 138GB Monthly', '138GB', 20000.00, 30, 'GLO-138GB-30D', true, 13, '138GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440050', 'GLO', 'Glo 200GB Monthly', '200GB', 30000.00, 30, 'GLO-200GB-30D', true, 14, '200GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440051', 'GLO', 'Glo 425GB Monthly', '425GB', 50000.00, 30, 'GLO-425GB-30D', true, 15, '425GB valid for 30 days', NOW(), NOW());

-- 9mobile Data Plans (15 plans)
INSERT INTO data_plans (
  id,
  network_provider,
  plan_name,
  data_amount,
  price,
  validity_days,
  plan_code,
  is_active,
  sort_order,
  description,
  created_at,
  updated_at
) VALUES
('650e8400-e29b-41d4-a716-446655440052', '9MOBILE', '9mobile 500MB Weekly', '500MB', 500.00, 7, '9MOBILE-500MB-7D', true, 1, '500MB valid for 7 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440053', '9MOBILE', '9mobile 1.5GB Monthly', '1.5GB', 1000.00, 30, '9MOBILE-1.5GB-30D', true, 2, '1.5GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440054', '9MOBILE', '9mobile 2GB Monthly', '2GB', 1200.00, 30, '9MOBILE-2GB-30D', true, 3, '2GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440055', '9MOBILE', '9mobile 3GB Monthly', '3GB', 1500.00, 30, '9MOBILE-3GB-30D', true, 4, '3GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440056', '9MOBILE', '9mobile 4.5GB Monthly', '4.5GB', 2000.00, 30, '9MOBILE-4.5GB-30D', true, 5, '4.5GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440057', '9MOBILE', '9mobile 11GB Monthly', '11GB', 4000.00, 30, '9MOBILE-11GB-30D', true, 6, '11GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440058', '9MOBILE', '9mobile 15GB Monthly', '15GB', 5000.00, 30, '9MOBILE-15GB-30D', true, 7, '15GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440059', '9MOBILE', '9mobile 27.5GB Monthly', '27.5GB', 8000.00, 30, '9MOBILE-27.5GB-30D', true, 8, '27.5GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440060', '9MOBILE', '9mobile 40GB Monthly', '40GB', 10000.00, 30, '9MOBILE-40GB-30D', true, 9, '40GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440061', '9MOBILE', '9mobile 75GB Monthly', '75GB', 15000.00, 30, '9MOBILE-75GB-30D', true, 10, '75GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440062', '9MOBILE', '9mobile 100GB Monthly', '100GB', 18000.00, 30, '9MOBILE-100GB-30D', true, 11, '100GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440063', '9MOBILE', '9mobile 120GB Monthly', '120GB', 20000.00, 30, '9MOBILE-120GB-30D', true, 12, '120GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440064', '9MOBILE', '9mobile 150GB Monthly', '150GB', 25000.00, 30, '9MOBILE-150GB-30D', true, 13, '150GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440065', '9MOBILE', '9mobile 200GB Monthly', '200GB', 30000.00, 30, '9MOBILE-200GB-30D', true, 14, '200GB valid for 30 days', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440066', '9MOBILE', '9mobile 500GB Monthly', '500GB', 50000.00, 30, '9MOBILE-500GB-30D', true, 15, '500GB valid for 30 days', NOW(), NOW());

\echo '✓ 66 data plans loaded (MTN: 18, Airtel: 18, Glo: 15, 9mobile: 15)'

-- ============================================================================
-- 1.3 SUBSCRIPTION TIERS (4 draw entry tiers)
-- ============================================================================

\echo 'Loading subscription tiers...'

INSERT INTO subscription_tiers (
  id,
  name,
  description,
  entries,
  is_active,
  sort_order,
  created_at,
  updated_at
) VALUES
('750e8400-e29b-41d4-a716-446655440001', 'BRONZE', 'Entry level tier - 1 draw entry per recharge', 1, true, 1, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440002', 'SILVER', 'Intermediate tier - 2 draw entries per recharge', 2, true, 2, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440003', 'GOLD', 'Advanced tier - 3 draw entries per recharge', 3, true, 3, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440004', 'PLATINUM', 'Premium tier - 5 draw entries per recharge', 5, true, 4, NOW(), NOW());

\echo '✓ 4 subscription tiers loaded'

-- ============================================================================
-- 1.4 WHEEL PRIZES (15 prizes)
-- ============================================================================

\echo 'Loading wheel prizes...'

INSERT INTO wheel_prizes (
  id,
  prize_code,
  prize_name,
  prize_type,
  prize_value,
  probability,
  is_active,
  sort_order,
  created_at,
  updated_at
) VALUES
('850e8400-e29b-41d4-a716-446655440001', 'NONE', 'Better Luck Next Time', 'NONE', 0, 40.50, true, 1, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440002', 'PTS10', '10 Points', 'POINTS', 10, 25.00, true, 2, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440003', 'PTS25', '25 Points', 'POINTS', 25, 15.00, true, 3, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440004', 'PTS50', '50 Points', 'POINTS', 50, 8.00, true, 4, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440005', 'PTS100', '100 Points', 'POINTS', 100, 5.00, true, 5, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440006', 'AIR50', '₦50 Airtime', 'AIRTIME', 50, 3.00, true, 6, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440007', 'AIR100', '₦100 Airtime', 'AIRTIME', 100, 1.50, true, 7, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440008', 'AIR200', '₦200 Airtime', 'AIRTIME', 200, 0.75, true, 8, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440009', 'AIR500', '₦500 Airtime', 'AIRTIME', 500, 0.50, true, 9, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440010', 'DATA500', '500MB Data', 'DATA', 500, 0.30, true, 10, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440011', 'DATA1GB', '1GB Data', 'DATA', 1000, 0.20, true, 11, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440012', 'DATA2GB', '2GB Data', 'DATA', 2000, 0.10, true, 12, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440013', 'CASH1K', '₦1,000 Cash', 'CASH', 1000, 0.10, true, 13, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440014', 'CASH5K', '₦5,000 Cash', 'CASH', 5000, 0.03, true, 14, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440015', 'IPHONE15', 'iPhone 15 Pro', 'PHYSICAL', 500000, 0.02, true, 15, NOW(), NOW());

\echo '✓ 15 wheel prizes loaded (probabilities sum to 100%)'

-- ============================================================================
-- 1.5 ADMIN USER (Super Administrator)
-- ============================================================================

\echo 'Loading admin user...'

INSERT INTO admin_users (
  id,
  email,
  password_hash,
  full_name,
  role,
  permissions,
  is_active,
  created_at,
  updated_at
) VALUES
('950e8400-e29b-41d4-a716-446655440001', 
 'admin@rechargemax.ng', 
 '$2a$10$GSv3/EaeIzohXsGy6jIMfuoOCMkBLZJF/OiqtG7kVdVoD/dKXypoe',
 'Super Administrator',
 'SUPER_ADMIN',
 '["view_analytics","manage_users","manage_transactions","manage_networks","manage_prizes","manage_affiliates","manage_settings","manage_admins","view_monitoring","manage_draws"]'::jsonb,
 true,
 NOW(),
 NOW());

\echo '✓ Admin user loaded (admin@rechargemax.ng / Admin@123456)'

\echo ''
\echo '--- Part 1 Complete: Essential Reference Data Loaded ---'
\echo ''

\echo '========================================================================'
\echo 'MASTER PRODUCTION SEED COMPLETE!'
\echo '========================================================================'
\echo 'Summary:'
\echo '  ✓ 4 Network Configurations'
\echo '  ✓ 66 Data Plans (MTN: 18, Airtel: 18, Glo: 15, 9mobile: 15)'
\echo '  ✓ 4 Subscription Tiers (Draw Entry Levels)'
\echo '  ✓ 15 Wheel Prizes'
\echo '  ✓ 1 Admin User (admin@rechargemax.ng)'
\echo '========================================================================'
\echo 'Platform is ready for testing!'
\echo '========================================================================'
