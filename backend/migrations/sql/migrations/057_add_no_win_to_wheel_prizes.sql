-- Migration 057: Add is_no_win and no_win_message columns to wheel_prizes
-- Allows admin to configure "Try Again" / "Better Luck Next Time" slots on the
-- spin wheel that do not award a prize. When the wheel lands on such a slot:
--   - backend returns no_win=true (no spin_result record is created)
--   - frontend shows a "you didn't win" message + CTA to spin again / recharge

ALTER TABLE wheel_prizes
  ADD COLUMN IF NOT EXISTS is_no_win      BOOLEAN      NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS no_win_message VARCHAR(200) NOT NULL DEFAULT '';

-- Also extend the spin_results claim_status check constraint to allow NO_WIN
ALTER TABLE spin_results DROP CONSTRAINT IF EXISTS spin_results_claim_status_check;
ALTER TABLE spin_results ADD CONSTRAINT spin_results_claim_status_check
  CHECK (claim_status IN ('PENDING','CLAIMED','EXPIRED','PENDING_ADMIN_REVIEW','APPROVED','REJECTED','NO_WIN'));

-- And allow NO_WIN as a prize_type in spin_results (no constraint there currently, but document it)
-- NO_WIN rows have prize_value=0 and are only used for eligibility tracking (no prize awarded)
