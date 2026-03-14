-- Migration 042: Fix FK type mismatch between draws and draw_types/prize_templates.
-- draws.draw_type_id is BIGINT but draw_types.id is INTEGER (serial).
-- draws.prize_template_id is BIGINT but prize_templates.id is INTEGER (serial).
-- Align draws FK columns to INTEGER to allow proper FK constraints.

-- Drop existing FK constraints if any (ignore errors if they don't exist)
ALTER TABLE draws DROP CONSTRAINT IF EXISTS draws_draw_type_id_fkey;
ALTER TABLE draws DROP CONSTRAINT IF EXISTS draws_prize_template_id_fkey;

-- Cast the FK columns to INTEGER
ALTER TABLE draws ALTER COLUMN draw_type_id     TYPE INTEGER USING draw_type_id::INTEGER;
ALTER TABLE draws ALTER COLUMN prize_template_id TYPE INTEGER USING prize_template_id::INTEGER;

-- Re-add FK constraints
ALTER TABLE draws
    ADD CONSTRAINT draws_draw_type_id_fkey
    FOREIGN KEY (draw_type_id) REFERENCES draw_types(id) ON DELETE SET NULL;

ALTER TABLE draws
    ADD CONSTRAINT draws_prize_template_id_fkey
    FOREIGN KEY (prize_template_id) REFERENCES prize_templates(id) ON DELETE SET NULL;
