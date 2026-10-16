-- =============================================================================
-- Migration 005 — Stripe customer ID
-- Adds the column needed to open a Stripe customer portal session; checkout
-- sessions correlate to a workspace via subscription metadata (org_id), set
-- at session-creation time in the Next.js app, not via a stored customer ID.
-- =============================================================================

ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS stripe_customer_id TEXT;
