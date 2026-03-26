-- Migration 058: Add REJECTED to draw_winners.claim_status CHECK constraint
-- The RejectClaim service function sets claim_status = 'REJECTED' but the
-- original constraint only allows: PENDING, CLAIMED, EXPIRED.

ALTER TABLE public.draw_winners
  DROP CONSTRAINT IF EXISTS draw_winners_claim_status_check;

ALTER TABLE public.draw_winners
  ADD CONSTRAINT draw_winners_claim_status_check
  CHECK (claim_status = ANY (ARRAY[
    'PENDING'::text,
    'CLAIMED'::text,
    'EXPIRED'::text,
    'REJECTED'::text,
    'APPROVED'::text,
    'PENDING_ADMIN_REVIEW'::text
  ]));
