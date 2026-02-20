# VTPass Data Plans Integration

**Date:** February 4, 2026  
**Status:** ✅ Completed  
**Environment:** Production-Ready

---

## Overview

This document describes the integration of real VTPass data plans into the RechargeMax platform, replacing all hardcoded data with database-driven content sourced directly from the VTPass API.

---

## Changes Made

### 1. Database Seeding with Real VTPass Data

**File:** `database/08_vtpass_data_plans_seed.sql`

- Fetched real data plans from VTPass Sandbox API
- Generated SQL seed file with **49 data plans** across 2 networks:
  - **MTN Nigeria**: 28 plans
  - **Airtel Nigeria**: 21 plans
- All plans include:
  - VTPass variation code (plan_code)
  - Plan name and description
  - Data amount (e.g., "1GB", "500MB")
  - Price in Naira
  - Validity period in days

**Sample Plans:**

| Network | Plan Name | Data | Price | Validity |
|---------|-----------|------|-------|----------|
| MTN | 100 Naira - 100MB - 1 Day | 100MB | ₦100 | 1 day |
| MTN | 1500 Naira - 3GB - 30 days | 3GB | ₦1,500 | 30 days |
| Airtel | 1,000 Naira - 1.5GB - 30 Days | 1.5GB | ₦1,000 | 30 days |

### 2. Backend Code Changes

**File:** `backend/internal/application/services/network_config_service.go`

**Before:**
```go
// Had hardcoded fallback with switch statement for each network
switch network {
case "MTN":
    packages = []DataPackage{
        {ID: "MTN_500MB", Name: "500MB Daily", ...},
        // ... more hardcoded plans
    }
// ... more hardcoded networks
}
```

**After:**
```go
// Fetch data plans from database ONLY - no hardcoded fallback
if s.dataPlanRepo == nil {
    return nil, fmt.Errorf("data plan repository not initialized")
}

plans, err := s.dataPlanRepo.FindByNetworkCode(ctx, network)
if err != nil {
    return nil, fmt.Errorf("failed to load data plans from database: %w", err)
}
```

**Key Improvements:**
- ✅ Removed ALL hardcoded data plans (60+ lines of hardcoded data eliminated)
- ✅ Database-only queries with proper error handling
- ✅ No fallback to hardcoded values
- ✅ Production-ready, scalable architecture

### 3. Frontend Updates

**File:** `frontend/src/components/recharge/PremiumRechargeForm.tsx`

- Fixed API base URL to use environment variable
- Data plan dropdown now fetches from `/api/v1/networks/{networkId}/bundles`
- Displays real VTPass plans with accurate pricing and data amounts

---

## VTPass API Integration

### API Credentials (Sandbox)

```env
VTPASS_API_KEY=c5bd97e357820f85ace13c7926e9c925
VTPASS_PUBLIC_KEY=PK_761bd8cb3f9783c8f94258234e4618fbecadca22b9e
VTPASS_SECRET_KEY=SK_8221e655104cd4459259c3d4d840103565d4376b027
VTPASS_MODE=sandbox
VTPASS_BASE_URL=https://sandbox.vtpass.com/api
```

### Service IDs

| Network | Service ID |
|---------|------------|
| MTN | `mtn-data` |
| Airtel | `airtel-data` |
| Glo | `glo-data` |
| 9mobile | `etisalat-data` |

### API Endpoints Used

1. **Get Service Variations (Data Plans)**
   ```
   GET https://sandbox.vtpass.com/api/service-variations?serviceID={service_id}
   Headers:
     - api-key: {VTPASS_API_KEY}
     - public-key: {VTPASS_PUBLIC_KEY}
   ```

---

## Database Schema

### Table: `data_plans_2026_01_30_14_00`

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key |
| network_id | uuid | Foreign key to network_configs |
| plan_code | text | VTPass variation code (unique) |
| plan_name | text | Display name |
| data_amount | text | Data size (e.g., "1GB") |
| price | numeric(10,2) | Price in Naira |
| validity_days | integer | Validity period |
| description | text | Full description |
| is_active | boolean | Active status |
| created_at | timestamp | Creation timestamp |
| updated_at | timestamp | Last update timestamp |

### Constraints

- **Primary Key:** `id`
- **Unique Constraint:** `(network_id, plan_code)`
- **Foreign Key:** `network_id` references `network_configs_2026_01_30_14_00(id)`
- **Check Constraints:**
  - `price > 0`
  - `validity_days > 0`

---

## Seed Data Generation Process

### Script: `generate_vtpass_seed.py`

**Process:**
1. Fetch data plans from VTPass API for each network
2. Parse plan names to extract:
   - Data size (GB/MB)
   - Validity period (days)
   - Clean display name
3. Filter out non-data plans (voice bundles, etc.)
4. Generate SQL INSERT statements with proper schema mapping
5. Handle conflicts with `ON CONFLICT DO UPDATE`

**Usage:**
```bash
python3 generate_vtpass_seed.py
```

**Output:**
```
✅ Generated SQL for MTN: 28 plans
✅ Generated SQL for Airtel: 21 plans
✅ SQL seed file generated: /home/ubuntu/vtpass_data_plans_seed.sql
```

---

## Testing Results

### Database Verification

```sql
SELECT n.network_name, COUNT(d.id) as plan_count 
FROM data_plans_2026_01_30_14_00 d 
JOIN network_configs_2026_01_30_14_00 n ON d.network_id = n.id 
GROUP BY n.network_name 
ORDER BY n.network_name;
```

**Result:**
```
  network_name  | plan_count 
----------------+------------
 Airtel Nigeria |         21
 MTN Nigeria    |         28
```

### Frontend Testing

- ✅ Data plan dropdown loads successfully
- ✅ Displays 28 MTN plans when MTN network selected
- ✅ Displays 21 Airtel plans when Airtel network selected
- ✅ Plan details show correct pricing and data amounts
- ✅ Payment button updates with selected plan price

---

## Production Deployment

### Prerequisites

1. **VTPass Production Credentials**
   - Update `.env` with production API keys
   - Change `VTPASS_MODE=production`
   - Update `VTPASS_BASE_URL=https://vtpass.com/api`

2. **Database Migration**
   ```bash
   psql -h $DB_HOST -U $DB_USER $DB_NAME -f database/08_vtpass_data_plans_seed.sql
   ```

3. **Backend Rebuild**
   ```bash
   cd backend
   go build -o rechargemax ./cmd/server
   ```

### Data Plan Refresh Strategy

**Option 1: Manual Refresh**
```bash
python3 generate_vtpass_seed.py
psql -h $DB_HOST -U $DB_USER $DB_NAME -f vtpass_data_plans_seed.sql
```

**Option 2: Automated Refresh (Recommended)**
- Create a cron job to fetch and update plans daily/weekly
- Use `ON CONFLICT DO UPDATE` to handle changes
- Log all updates for audit trail

**Option 3: Real-time API Integration**
- Fetch plans directly from VTPass API on each request
- Cache results for 24 hours
- Fallback to database if API is unavailable

---

## Future Enhancements

### 1. Add Glo and 9mobile Plans
```bash
# Fetch Glo plans
curl -s -X GET "https://sandbox.vtpass.com/api/service-variations?serviceID=glo-data" \
  -H "api-key: $VTPASS_API_KEY" \
  -H "public-key: $VTPASS_PUBLIC_KEY" > glo_data_plans.json

# Fetch 9mobile plans
curl -s -X GET "https://sandbox.vtpass.com/api/service-variations?serviceID=etisalat-data" \
  -H "api-key: $VTPASS_API_KEY" \
  -H "public-key: $VTPASS_PUBLIC_KEY" > 9mobile_data_plans.json
```

### 2. Plan Categorization
- Add `category` field (Daily, Weekly, Monthly, SME, etc.)
- Filter plans by category in frontend

### 3. Plan Popularity Tracking
- Track which plans are purchased most
- Sort plans by popularity in dropdown

### 4. Price Change Alerts
- Monitor VTPass API for price changes
- Notify admins when plans are updated

---

## Compliance & Best Practices

### ✅ Achieved

1. **No Hardcoded Data**: All data plans sourced from database
2. **Production-Ready**: Scalable architecture with proper error handling
3. **Real VTPass Integration**: Uses actual VTPass API data
4. **Database-Driven**: Easy to update plans without code changes
5. **Proper Schema**: Foreign keys, constraints, and indexes in place

### ⚠️ Recommendations

1. **Add Glo and 9mobile**: Complete all 4 networks
2. **Implement Caching**: Reduce database queries
3. **Add Monitoring**: Track API failures and data staleness
4. **Version Control**: Track plan changes over time

---

## Support & Maintenance

### Troubleshooting

**Issue:** Data plans not loading in frontend

**Solution:**
1. Check database connection
2. Verify seed file was applied: `SELECT COUNT(*) FROM data_plans_2026_01_30_14_00;`
3. Check backend logs for errors
4. Verify API endpoint is accessible

**Issue:** Plans are outdated

**Solution:**
1. Re-run seed generation script
2. Apply updated seed file to database
3. Restart backend service

### Contact

For VTPass API issues, contact VTPass support:
- Website: https://vtpass.com
- Email: support@vtpass.com
- Phone: +234 xxx xxx xxxx

---

**Last Updated:** February 4, 2026  
**Maintained By:** RechargeMax Development Team
