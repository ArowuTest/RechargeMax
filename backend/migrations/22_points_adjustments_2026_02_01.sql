-- Migration: Points Adjustments Table
-- Created: 2026-02-01
-- Purpose: Track manual points adjustments by admins

CREATE TABLE IF NOT EXISTS points_adjustments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    points INT NOT NULL,
    reason VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_points_adjustments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_points_adjustments_admin FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_points_adjustments_user_id ON points_adjustments(user_id);
CREATE INDEX IF NOT EXISTS idx_points_adjustments_created_by ON points_adjustments(created_by);
CREATE INDEX IF NOT EXISTS idx_points_adjustments_created_at ON points_adjustments(created_at DESC);

-- Add comment
COMMENT ON TABLE points_adjustments IS 'Tracks manual points adjustments made by administrators';
COMMENT ON COLUMN points_adjustments.points IS 'Points amount (positive for add, negative for deduct)';
COMMENT ON COLUMN points_adjustments.reason IS 'Reason for adjustment (e.g., manual_adjustment, compensation, correction)';
COMMENT ON COLUMN points_adjustments.created_by IS 'Admin user ID who made the adjustment';

