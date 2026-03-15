-- Migration: Fix daily_subscriptions schema
-- Date: 2026-02-23
-- Description: Add subscription_code column and update status constraint to allow 'pending'

-- Add subscription_code column if it doesn't exist
ALTER TABLE daily_subscriptions 
ADD COLUMN IF NOT EXISTS subscription_code VARCHAR(20) UNIQUE;

-- Drop old status constraint
ALTER TABLE daily_subscriptions 
DROP CONSTRAINT IF EXISTS daily_subscriptions_status_check;

-- Add new status constraint that includes 'pending'
ALTER TABLE daily_subscriptions 
ADD CONSTRAINT daily_subscriptions_status_check 
CHECK (status IN ('active', 'pending', 'cancelled', 'expired', 'paused'));

-- Grant permissions to application user
GRANT ALL PRIVILEGES ON TABLE daily_subscriptions TO rechargemax;
