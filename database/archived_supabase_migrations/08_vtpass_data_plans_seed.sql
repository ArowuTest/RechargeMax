-- ============================================================================
-- VTPass Data Plans Seed File
-- Generated: 2026-02-04 07:51:49
-- Source: VTPass API (Sandbox)
-- ============================================================================

-- Clear existing data plans
DELETE FROM data_plans_2026_01_30_14_00;


-- MTN Data Plans from VTPass

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-10mb-100',
    'N100 100MB - 24 hrs',
    '100MB',
    100.0,
    30,
    'N100 100MB - 24 hrs - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-50mb-200',
    'N200 200MB - 2 days',
    '200MB',
    200.0,
    2,
    'N200 200MB - 2 days - Valid for 2 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-100mb-1000',
    'N1000 1.5GB - 30 days',
    '1.5GB',
    1000.0,
    30,
    'N1000 1.5GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-500mb-2000',
    'N2000 4.5GB - 30 days',
    '4.5GB',
    2000.0,
    30,
    'N2000 4.5GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-20hrs-1500',
    'N1500 6GB - 7 days',
    '6GB',
    1500.0,
    7,
    'N1500 6GB - 7 days - Valid for 7 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-3gb-2500',
    'N2500 6GB - 30 days',
    '6GB',
    2500.0,
    30,
    'N2500 6GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-data-3000',
    'N3000 8GB - 30 days',
    '8GB',
    3000.0,
    30,
    'N3000 8GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-1gb-3500',
    'N3500 10GB - 30 days',
    '10GB',
    3500.0,
    30,
    'N3500 10GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-100hr-5000',
    'N5000 15GB - 30 days',
    '15GB',
    5000.0,
    30,
    'N5000 15GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-3gb-6000',
    'N6000 20GB - 30 days',
    '20GB',
    6000.0,
    30,
    'N6000 20GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-40gb-10000',
    'N10000 40GB - 30 days',
    '40GB',
    10000.0,
    30,
    'N10000 40GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-75gb-15000',
    'N15000 75GB - 30 days',
    '75GB',
    15000.0,
    30,
    'N15000 75GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-110gb-20000',
    'N20000 110GB - 30 days',
    '110GB',
    20000.0,
    30,
    'N20000 110GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-3gb-1500',
    'N1500 3GB - 30 days',
    '3GB',
    1500.0,
    30,
    'N1500 3GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-25gb-sme-10000',
    'N10,000 25GB SME Mobile Data ( 1 Month)',
    '25GB',
    10000.0,
    30,
    'N10,000 25GB SME Mobile Data ( 1 Month) - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-165gb-sme-50000',
    'N50,000 165GB SME Mobile Data (2-Months)',
    '165GB',
    50000.0,
    60,
    'N50,000 165GB SME Mobile Data (2-Months) - Valid for 60 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-360gb-sme-100000',
    'N100,000 360GB SME Mobile Data (3 Months)',
    '360GB',
    100000.0,
    90,
    'N100,000 360GB SME Mobile Data (3 Months) - Valid for 90 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-4-5tb-450000',
    'N450,000 4.5TB Mobile Data (1 Year)',
    '4.5TB',
    450000.0,
    365,
    'N450,000 4.5TB Mobile Data (1 Year) - Valid for 365 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-1tb-110000',
    'N100,000 1TB Mobile Data (1 Year)',
    '1TB',
    100000.0,
    365,
    'N100,000 1TB Mobile Data (1 Year) - Valid for 365 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-2-5gb-600',
    'N600 2.5GB - 2 days',
    '2.5GB',
    600.0,
    2,
    'N600 2.5GB - 2 days - Valid for 2 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-120gb-22000',
    'N22000 120GB Monthly Plan + 80mins',
    '120GB',
    22000.0,
    30,
    'N22000 120GB Monthly Plan + 80mins - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-100gb-20000',
    '100GB 2-Month Plan',
    '100GB',
    20000.0,
    60,
    '100GB 2-Month Plan - Valid for 60 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-160gb-30000',
    'N30,000 160GB 2-Month Plan',
    '160GB',
    30000.0,
    60,
    'N30,000 160GB 2-Month Plan - Valid for 60 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-400gb-50000',
    'N50,000 400GB 3-Month Plan',
    '400GB',
    50000.0,
    90,
    'N50,000 400GB 3-Month Plan - Valid for 90 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-600gb-75000',
    'N75,000 600GB 3-Months Plan',
    '600GB',
    75000.0,
    90,
    'N75,000 600GB 3-Months Plan - Valid for 90 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-3gb-800',
    'N800 3GB - 2 days',
    '3GB',
    800.0,
    2,
    'N800 3GB - 2 days - Valid for 2 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-7gb-2000',
    'N2000 7GB - 7 days',
    '7GB',
    2000.0,
    7,
    'N2000 7GB - 7 days - Valid for 7 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    'mtn-xtradata-200',
    'N200 Xtradata',
    'N/A',
    200.0,
    30,
    'N200 Xtradata - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();


-- Airtel Data Plans from VTPass

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-100',
    '100 Naira - 75MB - 1Day',
    '75MB',
    99.0,
    1,
    '100 Naira - 75MB - 1Day - Valid for 1 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-200',
    '200 Naira - 200MB - 3Days',
    '200MB',
    199.03,
    3,
    '200 Naira - 200MB - 3Days - Valid for 3 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-300',
    '300 Naira - 350MB - 7 Days',
    '350MB',
    299.02,
    7,
    '300 Naira - 350MB - 7 Days - Valid for 7 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-500',
    '500 Naira - 750MB - 14 Days',
    '750MB',
    499.0,
    14,
    '500 Naira - 750MB - 14 Days - Valid for 14 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-1000',
    '1,000 Naira - 1.5GB - 30 Days',
    '1.5GB',
    999.0,
    30,
    '1,000 Naira - 1.5GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-1500',
    '1,500 Naira - 3GB - 30 Days',
    '3GB',
    1499.01,
    30,
    '1,500 Naira - 3GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-2000',
    '2,000 Naira - 4.5GB - 30 Days',
    '4.5GB',
    1999.0,
    30,
    '2,000 Naira - 4.5GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-3000',
    '3,000 Naira - 8GB - 30 Days',
    '8GB',
    2999.02,
    30,
    '3,000 Naira - 8GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-4000',
    '4,000 Naira - 11GB - 30 Days',
    '11GB',
    3999.01,
    30,
    '4,000 Naira - 11GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-5000',
    '5,000 Naira - 15GB - 30 Days',
    '15GB',
    4999.0,
    30,
    '5,000 Naira - 15GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-1500-2',
    'Binge Data - 1,500 Naira (7 Days) - 6GB',
    '6GB',
    1499.03,
    7,
    'Binge Data - 1,500 Naira (7 Days) - 6GB - Valid for 7 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-10000',
    '10,000 Naira - 40GB - 30 Days',
    '40GB',
    9999.0,
    30,
    '10,000 Naira - 40GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-15000',
    '15,000 Naira - 75GB - 30 Days',
    '75GB',
    14999.0,
    30,
    '15,000 Naira - 75GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-20000',
    '20,000 Naira - 110GB - 30 Days',
    '110GB',
    19999.02,
    30,
    '20,000 Naira - 110GB - 30 Days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-600',
    'Data - 600 Naira - 1GB - 14 days',
    '1GB',
    600.0,
    14,
    'Data - 600 Naira - 1GB - 14 days - Valid for 14 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-1000-7',
    'Data - 1000 Naira - 1.5GB - 7 days',
    '1.5GB',
    1000.0,
    7,
    'Data - 1000 Naira - 1.5GB - 7 days - Valid for 7 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-2000-7',
    'Data - 2000 Naira - 7GB - 7 days',
    '7GB',
    2000.0,
    7,
    'Data - 2000 Naira - 7GB - 7 days - Valid for 7 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-5000-7',
    'Data - 5000 Naira - 25GB - 7 days',
    '25GB',
    5000.0,
    7,
    'Data - 5000 Naira - 25GB - 7 days - Valid for 7 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-400-1',
    'Data - 400 Naira - 1.5GB - 1 day',
    '1.5GB',
    400.0,
    1,
    'Data - 400 Naira - 1.5GB - 1 day - Valid for 1 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-800-2',
    'Data - 800 Naira - 3.5GB - 2 days',
    '3.5GB',
    800.0,
    2,
    'Data - 800 Naira - 3.5GB - 2 days - Valid for 2 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO data_plans_2026_01_30_14_00 (
    id, network_id, plan_code, plan_name, data_amount, price, validity_days, 
    description, is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '22222222-2222-2222-2222-222222222222',
    'airt-6000-30',
    'Data - 6000 Naira - 23GB - 30 days',
    '23GB',
    6000.0,
    30,
    'Data - 6000 Naira - 23GB - 30 days - Valid for 30 days',
    true,
    NOW(),
    NOW()
) ON CONFLICT (network_id, plan_code) DO UPDATE SET
    plan_name = EXCLUDED.plan_name,
    data_amount = EXCLUDED.data_amount,
    price = EXCLUDED.price,
    validity_days = EXCLUDED.validity_days,
    description = EXCLUDED.description,
    updated_at = NOW();


-- End of VTPass Data Plans Seed
