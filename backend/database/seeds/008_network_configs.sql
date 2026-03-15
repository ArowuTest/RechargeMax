-- 008_network_configs.sql
-- Default telecom network configurations for Nigeria
-- Admin can enable/disable any network from the admin panel.
-- AirtimeEnabled and DataEnabled default to TRUE for all networks.

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
  sort_order
)
VALUES
  (
    uuid_generate_v4(),
    'MTN Nigeria',
    'MTN',
    TRUE,
    TRUE,
    TRUE,
    0.0,
    50000,   -- ₦500 minimum (in kobo)
    500000000, -- ₦5,000,000 maximum (in kobo)
    '/images/networks/mtn.png',
    '#FFD700',
    1
  ),
  (
    uuid_generate_v4(),
    'Glo Mobile',
    'GLO',
    TRUE,
    TRUE,
    TRUE,
    0.0,
    50000,
    500000000,
    '/images/networks/glo.png',
    '#00A651',
    2
  ),
  (
    uuid_generate_v4(),
    'Airtel Nigeria',
    'AIRTEL',
    TRUE,
    TRUE,
    TRUE,
    0.0,
    50000,
    500000000,
    '/images/networks/airtel.png',
    '#E40000',
    3
  ),
  (
    uuid_generate_v4(),
    '9mobile',
    '9MOBILE',
    TRUE,
    TRUE,
    TRUE,
    0.0,
    50000,
    500000000,
    '/images/networks/9mobile.png',
    '#006E51',
    4
  )
ON CONFLICT (network_code) DO NOTHING;
