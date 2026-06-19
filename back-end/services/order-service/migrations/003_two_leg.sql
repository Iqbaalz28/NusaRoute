-- Two-leg hub-and-spoke flow: how the package reaches the origin hub, and the
-- destination hub id (to detect arrival and open the last-mile job).
ALTER TABLE orders ADD COLUMN IF NOT EXISTS pickup_mode VARCHAR(20) DEFAULT 'COURIER'; -- COURIER | SELF_DROPOFF
ALTER TABLE orders ADD COLUMN IF NOT EXISTS dest_hub_id VARCHAR(36) DEFAULT '';
