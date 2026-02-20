# RechargeMax Database Migrations & Updates

## Overview
This document tracks all database schema changes and migrations applied to support the complete business logic.

## Migration Files

### 1. `backend/migration_business_logic.sql`
**Purpose:** Add core business logic support to database schema

**Changes:**
- Added `prefixes` column to `network_configs` table for phone number validation
- Created `spin_tiers` table with 5 tiers (Bronze, Silver, Gold, Platinum, Diamond)
- Updated admin roles to include VIEWER role
- Created helper functions:
  - `validate_phone_network(phone_number, network_code)` - Validates phone numbers
  - `get_spin_tier(daily_amount)` - Returns appropriate spin tier based on recharge amount

**Network Prefixes:**
- **MTN:** 0803, 0703, 0903, 0806, 0813, 0810, 0814, 0816, 0906
- **Airtel:** 0802, 0708, 0902, 0808, 0701, 0812, 0901, 0907
- **Glo:** 0805, 0705, 0905, 0807, 0815, 0811
- **9mobile:** 0809, 0818, 0909, 0817, 0908

**Spin Tiers:**
| Tier     | Daily Amount Range | Spins/Day | Icon |
|----------|-------------------|-----------|------|
| Bronze   | ₦1,000 - ₦4,999   | 1         | 🥉   |
| Silver   | ₦5,000 - ₦9,999   | 2         | 🥈   |
| Gold     | ₦10,000 - ₦19,999 | 3         | 🥇   |
| Platinum | ₦20,000 - ₦49,999 | 5         | 💎   |
| Diamond  | ₦50,000+          | 10        | 💠   |

## Current Active Draws

1. **Daily Cash Draw** - ₦50,000 (ends daily at 11:59 PM)
2. **Monthly Super Prize** - ₦5,000,000 (ends in 29 days)

## Database Functions

### validate_phone_network(phone_number text, network_code text)
```sql
SELECT validate_phone_network('08031234567', 'MTN');
-- Returns: true
```

### get_spin_tier(daily_amount numeric)
```sql
SELECT * FROM get_spin_tier(15000.00);
-- Returns: Gold tier with 3 spins/day
```

## Backend Updates

### CORS Configuration
- Added port 8081 to allowed origins in `internal/middleware/middleware.go`

### Platform Handler
- Fixed `GetPlatformStatistics` to use uppercase 'ACTIVE' status

### API Endpoints
Added missing functions to `frontend/src/lib/api.ts`:
- `getAvailableSpins()`
- `consumeSpin()`
- `recordTransactionPrize()`
- `getTierProgress()`
- `getSpinTiers()`

## Frontend Updates

### Components
- Fixed `Header.tsx` - Changed react-router-dom import to use proxy
- Fixed `EnterpriseHomePage.tsx` - Updated hero background to blue gradient
- Removed all Supabase dependencies

### Environment Configuration
- Updated `.env` to use `VITE_API_BASE_URL=http://localhost:8080`

## How to Apply Migrations

### Fresh Database
```bash
# Apply business logic migration
PGPASSWORD=rechargemax123 psql -h localhost -U rechargemax -d rechargemax -f backend/migration_business_logic.sql
```

### Existing Database
Migrations are idempotent and safe to re-run.

## Verification Queries

### Check Network Configurations
```sql
SELECT network_name, network_code, prefixes, is_active
FROM network_configs
ORDER BY sort_order;
```

### Check Spin Tiers
```sql
SELECT tier_name, min_daily_amount, max_daily_amount, spins_per_day
FROM spin_tiers
WHERE is_active = true
ORDER BY sort_order;
```

### Check Active Draws
```sql
SELECT name, prize_pool, type, status,
    EXTRACT(EPOCH FROM (end_time - NOW())) / 3600 as hours_remaining
FROM draws
WHERE status = 'ACTIVE'
ORDER BY prize_pool;
```
