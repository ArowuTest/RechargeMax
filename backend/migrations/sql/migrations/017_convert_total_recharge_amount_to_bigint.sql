-- Migration 017: Convert total_recharge_amount from numeric(12,2) to bigint (kobo)
--
-- IDEMPOTENT: wraps the ALTER inside a DO block that checks the current column type.
-- On the first run (column is numeric) it multiplies existing naira values × 100 → kobo.
-- On re-runs (column already bigint) the DO block is a no-op, preventing the
-- "bigint out of range" error that occurred when an already-kobo value was multiplied again.

DO $$
BEGIN
    IF (
        SELECT data_type
        FROM   information_schema.columns
        WHERE  table_schema = 'public'
          AND  table_name   = 'users'
          AND  column_name  = 'total_recharge_amount'
    ) <> 'bigint' THEN
        ALTER TABLE users
            ALTER COLUMN total_recharge_amount TYPE BIGINT
            USING (total_recharge_amount * 100)::BIGINT;

        ALTER TABLE users
            ALTER COLUMN total_recharge_amount SET DEFAULT 0;

        COMMENT ON COLUMN users.total_recharge_amount IS 'Total recharge amount in kobo (₦1 = 100 kobo)';
    END IF;
END $$;
