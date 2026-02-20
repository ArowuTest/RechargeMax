# Archived Supabase Migration Files

**Date Archived:** February 12, 2026  
**Reason:** These migration files contain table names with timestamp suffixes (e.g., `users_2026_01_30_14_00`) that are incompatible with the current GORM-based backend.

## Why These Files Are Not Used

1. **Wrong Table Names:** All tables have timestamp suffixes like `_2026_01_30_14_00`
2. **GORM Incompatibility:** The Go backend expects clean table names (`users`, `transactions`, etc.)
3. **Supabase-Specific:** Many files contain RLS policies, storage buckets, and auth triggers specific to Supabase
4. **Superseded:** The functionality is now handled by GORM AutoMigrate + files 19-37

## Files Archived (19 total)

- 01_core_tables_schema.sql
- 02_rls_policies.sql
- 03_seeded_data.sql
- 04_functions_triggers.sql
- 05_storage_buckets.sql
- 06_notification_system.sql
- 07_notification_templates_seed_fixed.sql
- 08_otp_verifications.sql
- 08_vtpass_data_plans_seed.sql
- 09_admin_activity_logs.sql
- 10_affiliate_payouts.sql
- 11_affiliate_analytics_bank.sql
- 12_payment_logs.sql
- 13_service_pricing.sql
- 14_vtu_transactions.sql
- 15_wallet_transactions.sql
- 16_commission_ledger.sql
- 17_withdrawal_bank.sql
- 18_api_webhook_logs.sql

## Current Migration Strategy

**Base Tables:** Created by GORM AutoMigrate in `/backend/cmd/server/main.go`  
**Enhancements:** Applied via migration files 19-37 in `/database/`  
**Seed Data:** Loaded from `/database/seeds/production_seed_final.sql`

## Reference Only

These files are kept for historical reference and documentation purposes. **DO NOT** execute them on the production database.
