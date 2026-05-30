-- BarterSwap — schema initial (placeholder, will be completed in Jalon 1).
-- This file is applied at every startup; keep statements idempotent.

CREATE TABLE IF NOT EXISTS _migrations_marker (
    id          SERIAL PRIMARY KEY,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
