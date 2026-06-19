-- Two-leg hub-and-spoke: each assignment is a leg (FIRST_MILE/LAST_MILE/DIRECT)
-- with its own pickup and dropoff points, plus the hub involved in the leg.
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS leg VARCHAR(20) DEFAULT 'DIRECT';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS dropoff_lat DOUBLE PRECISION DEFAULT 0;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS dropoff_lng DOUBLE PRECISION DEFAULT 0;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS dropoff_address TEXT DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS hub_id VARCHAR(36) DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS hub_name VARCHAR(255) DEFAULT '';
