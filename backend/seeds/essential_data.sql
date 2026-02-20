-- ============================================================================
-- RechargeMax Essential Data Seed File
-- This file contains essential reference data that should be loaded on every deployment
-- ============================================================================

-- Clear existing essential data (in correct order to avoid foreign key constraints)
TRUNCATE TABLE data_plans CASCADE;
TRUNCATE TABLE networks CASCADE;
TRUNCATE TABLE subscription_tier CASCADE;
TRUNCATE TABLE wheel_prizes CASCADE;
TRUNCATE TABLE commission_tiers CASCADE;

-- ============================================================================
-- NETWORKS (Nigerian Mobile Networks)
-- ============================================================================
INSERT INTO networks (id, name, code, logo_url, is_active, ussd_code, api_config, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'MTN', 'MTN', 'https://example.com/mtn-logo.png', true, '*555#', '{"provider": "vtpass", "service_id": "mtn"}', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Airtel', 'AIRTEL', 'https://example.com/airtel-logo.png', true, '*140#', '{"provider": "vtpass", "service_id": "airtel"}', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Glo', 'GLO', 'https://example.com/glo-logo.png', true, '*777#', '{"provider": "vtpass", "service_id": "glo"}', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', '9mobile', '9MOBILE', 'https://example.com/9mobile-logo.png', true, '*229#', '{"provider": "vtpass", "service_id": "etisalat"}', NOW(), NOW());

-- ============================================================================
-- DATA PLANS (Popular Nigerian Data Plans)
-- ============================================================================
INSERT INTO data_plans (id, network_id, name, data_amount, validity_days, price, description, is_active, created_at, updated_at) VALUES
-- MTN Data Plans
('650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'MTN 1GB Daily', '1GB', 1, 35000, '1GB valid for 1 day', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'MTN 2GB Weekly', '2GB', 7, 100000, '2GB valid for 7 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'MTN 6GB Monthly', '6GB', 30, 150000, '6GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', 'MTN 12GB Monthly', '12GB', 30, 300000, '12GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', 'MTN 25GB Monthly', '25GB', 30, 500000, '25GB valid for 30 days', true, NOW(), NOW()),

-- Airtel Data Plans
('650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440002', 'Airtel 1.5GB Daily', '1.5GB', 1, 35000, '1.5GB valid for 1 day', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440002', 'Airtel 3GB Weekly', '3GB', 7, 100000, '3GB valid for 7 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440002', 'Airtel 6GB Monthly', '6GB', 30, 150000, '6GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440002', 'Airtel 11GB Monthly', '11GB', 30, 200000, '11GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440002', 'Airtel 40GB Monthly', '40GB', 30, 500000, '40GB valid for 30 days', true, NOW(), NOW()),

-- Glo Data Plans
('650e8400-e29b-41d4-a716-446655440011', '550e8400-e29b-41d4-a716-446655440003', 'Glo 1.6GB Weekly', '1.6GB', 7, 50000, '1.6GB valid for 7 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440003', 'Glo 5.8GB Monthly', '5.8GB', 30, 150000, '5.8GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440013', '550e8400-e29b-41d4-a716-446655440003', 'Glo 10GB Monthly', '10GB', 30, 200000, '10GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440014', '550e8400-e29b-41d4-a716-446655440003', 'Glo 29.5GB Monthly', '29.5GB', 30, 500000, '29.5GB valid for 30 days', true, NOW(), NOW()),

-- 9mobile Data Plans
('650e8400-e29b-41d4-a716-446655440015', '550e8400-e29b-41d4-a716-446655440004', '9mobile 1.5GB Weekly', '1.5GB', 7, 100000, '1.5GB valid for 7 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440016', '550e8400-e29b-41d4-a716-446655440004', '9mobile 4.5GB Monthly', '4.5GB', 30, 200000, '4.5GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440017', '550e8400-e29b-41d4-a716-446655440004', '9mobile 11GB Monthly', '11GB', 30, 300000, '11GB valid for 30 days', true, NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440018', '550e8400-e29b-41d4-a716-446655440004', '9mobile 40GB Monthly', '40GB', 30, 500000, '40GB valid for 30 days', true, NOW(), NOW());

-- ============================================================================
-- SUBSCRIPTION TIERS (Loyalty Tiers)
-- ============================================================================
INSERT INTO subscription_tier (id, name, min_points, max_points, benefits, multiplier, created_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440001', 'Bronze', 0, 999, '{"draw_entries": 1, "spin_discount": 0, "cashback": 0}', 1.0, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440002', 'Silver', 1000, 4999, '{"draw_entries": 2, "spin_discount": 5, "cashback": 1}', 1.2, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440003', 'Gold', 5000, 19999, '{"draw_entries": 3, "spin_discount": 10, "cashback": 2}', 1.5, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440004', 'Platinum', 20000, 99999, '{"draw_entries": 5, "spin_discount": 15, "cashback": 3}', 2.0, NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440005', 'Diamond', 100000, NULL, '{"draw_entries": 10, "spin_discount": 20, "cashback": 5}', 3.0, NOW(), NOW());

-- ============================================================================
-- WHEEL PRIZES (Spin Wheel Rewards)
-- ============================================================================
INSERT INTO wheel_prizes (id, name, prize_type, value, probability, min_recharge_amount, is_active, created_at, updated_at) VALUES
('850e8400-e29b-41d4-a716-446655440001', '₦50 Airtime', 'AIRTIME', 5000, 25.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440002', '₦100 Airtime', 'AIRTIME', 10000, 20.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440003', '₦200 Airtime', 'AIRTIME', 20000, 15.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440004', '₦500 Airtime', 'AIRTIME', 50000, 10.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440005', '1GB Data', 'DATA', 100000, 12.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440006', '2GB Data', 'DATA', 200000, 8.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440007', '100 Points', 'POINTS', 100, 5.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440008', '500 Points', 'POINTS', 500, 3.0, 100000, true, NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440009', 'Better Luck', 'NONE', 0, 2.0, 100000, true, NOW(), NOW());

-- ============================================================================
-- COMMISSION TIERS (Affiliate Commission Structure)
-- ============================================================================
INSERT INTO commission_tiers (id, tier_name, min_referrals, max_referrals, commission_rate, bonus_threshold, bonus_amount, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440001', 'BRONZE', 0, 9, 2.0, 10000000, 50000, NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440002', 'SILVER', 10, 49, 3.0, 50000000, 200000, NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440003', 'GOLD', 50, 199, 4.0, 100000000, 500000, NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440004', 'PLATINUM', 200, NULL, 5.0, 500000000, 2000000, NOW(), NOW());

-- ============================================================================
-- End of Essential Data Seed File
-- ============================================================================
