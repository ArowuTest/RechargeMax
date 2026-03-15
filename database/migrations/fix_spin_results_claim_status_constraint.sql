-- Migration: Fix spin_results claim_status constraint
-- Date: 2026-02-23
-- Description: Drop old restrictive constraint that only allowed PENDING, CLAIMED, EXPIRED
--              The correct constraint (chk_spin_results_claim_status) already exists with full status list

ALTER TABLE spin_results DROP CONSTRAINT IF EXISTS spin_results_claim_status_check;

-- Verify the correct constraint exists
-- It should allow: PENDING, CLAIMED, EXPIRED, PENDING_ADMIN_REVIEW, APPROVED, REJECTED
