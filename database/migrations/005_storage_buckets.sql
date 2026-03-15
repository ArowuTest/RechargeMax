-- Storage buckets: Supabase-specific, not applicable to Go/PostgreSQL deployment.
-- File uploads are handled directly by the Go backend using local disk or S3.
-- This migration is intentionally a no-op.
SELECT 1;
