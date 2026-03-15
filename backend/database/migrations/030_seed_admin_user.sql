-- Seed default admin user (idempotent)
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
) VALUES (
  '950e8400-e29b-41d4-a716-446655440001',
  'admin@rechargemax.ng',
  '$2a$10$GSv3/EaeIzohXsGy6jIMfuoOCMkBLZJF/OiqtG7kVdVoD/dKXypoe',
  'Super Administrator',
  'SUPER_ADMIN',
  '["view_analytics","manage_users","manage_transactions","manage_networks","manage_prizes","manage_affiliates","manage_settings","manage_admins","view_monitoring","manage_draws"]'::jsonb,
  true,
  NOW(),
  NOW()
) ON CONFLICT (email) DO NOTHING;

-- Seed network configurations (idempotent)
INSERT INTO network_configs (id, network_name, network_code, is_active, airtime_enabled, data_enabled, commission_rate, minimum_amount, maximum_amount, logo_url, brand_color, sort_order, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'MTN Nigeria',   'MTN',     true, true, true, 2.50, 5000, 5000000, 'https://upload.wikimedia.org/wikipedia/commons/thumb/9/93/New-mtn-logo.jpg/240px-New-mtn-logo.jpg',    '#FFCC00', 1, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Airtel Nigeria', 'AIRTEL',  true, true, true, 2.50, 5000, 5000000, 'https://upload.wikimedia.org/wikipedia/commons/thumb/f/fb/Airtel_logo.svg/240px-Airtel_logo.svg.png', '#FF0000', 2, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Glo Mobile',    'GLO',     true, true, true, 3.00, 5000, 5000000, 'https://upload.wikimedia.org/wikipedia/en/thumb/3/3a/Globacom.png/240px-Globacom.png',                '#00AA00', 3, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', '9mobile',       '9MOBILE', true, true, true, 3.50, 5000, 5000000, 'https://upload.wikimedia.org/wikipedia/en/thumb/2/2a/9mobile_logo.svg/240px-9mobile_logo.svg.png',  '#006600', 4, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Seed subscription tiers
INSERT INTO subscription_tiers (id, name, description, daily_amount, weekly_amount, monthly_amount, draw_entries_per_recharge, spin_credits_per_day, is_active, sort_order, created_at, updated_at) VALUES
('a50e8400-e29b-41d4-a716-446655440001', 'Bronze',   'Basic daily subscription',    200.00,  1400.00,  6000.00,  1, 1, true, 1, NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440002', 'Silver',   'Enhanced daily subscription', 500.00,  3500.00, 15000.00,  2, 2, true, 2, NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440003', 'Gold',     'Premium daily subscription', 1000.00,  7000.00, 30000.00,  5, 3, true, 3, NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440004', 'Platinum', 'Elite daily subscription',   2000.00, 14000.00, 60000.00, 10, 5, true, 4, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
