-- Migration: Remove redundant points calculation trigger
-- Date: 2026-02-20
-- Reason: Business logic moved to application layer for better maintainability
--
-- STRATEGIC DECISION:
-- The trigger_process_transaction trigger was attempting to calculate points
-- and update transactions AFTER the application layer already did so.
-- This caused:
-- 1. Duplicate logic (harder to maintain)
-- 2. The trigger's UPDATE didn't persist (PostgreSQL RETURNING limitation)
-- 3. Business logic split between Go and PL/pgSQL (harder to test)
--
-- SOLUTION:
-- Remove the trigger and keep all business logic in Go code (RechargeService)
-- This provides:
-- - Single source of truth
-- - Better testability
-- - Easier debugging
-- - More flexibility for promotions/bonuses
--
-- We keep the helper functions for potential admin/reporting use

-- Drop the trigger
DROP TRIGGER IF EXISTS trigger_process_transaction ON public.transactions;

-- Drop the trigger function
DROP FUNCTION IF EXISTS public.process_successful_transaction(uuid);

-- Keep calculate_points_earned() and calculate_draw_entries() functions
-- These can still be useful for admin tools, reporting, or manual queries

-- Verify: The update_transactions_updated_at trigger remains active
-- This trigger updates the updated_at timestamp and should NOT be removed

-- Migration complete
-- All points calculation now handled by RechargeService.ProcessSuccessfulPayment()
