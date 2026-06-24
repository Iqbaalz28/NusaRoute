-- Routing hubs: nearest sortation hub to sender (origin) and receiver
-- (destination), stamped at order creation and shown on the label/tracking.
ALTER TABLE orders ADD COLUMN IF NOT EXISTS origin_hub_code VARCHAR(20) DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS origin_hub_name VARCHAR(255) DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS dest_hub_code   VARCHAR(20) DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS dest_hub_name   VARCHAR(255) DEFAULT '';
