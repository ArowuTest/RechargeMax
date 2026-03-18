-- Migration 036: Fix otp_rate_limits id column to use auto-increment sequence
-- The table was created with `id bigint NOT NULL` but the sequence was never created,
-- causing INSERT failures with "null value in column id violates not-null constraint".
-- This migration creates the sequence and wires it as the column default.

-- Create the sequence if it doesn't already exist
CREATE SEQUENCE IF NOT EXISTS public.otp_rate_limits_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- Set the column default to use the sequence
ALTER TABLE public.otp_rate_limits
    ALTER COLUMN id SET DEFAULT nextval('public.otp_rate_limits_id_seq'::regclass);

-- Set sequence ownership so it is dropped with the table
ALTER SEQUENCE public.otp_rate_limits_id_seq OWNED BY public.otp_rate_limits.id;
