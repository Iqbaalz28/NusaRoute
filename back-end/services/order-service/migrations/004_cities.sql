-- City separated from detailed address (city is highlighted on the label).
ALTER TABLE orders ADD COLUMN IF NOT EXISTS sender_city   VARCHAR(100) DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS receiver_city VARCHAR(100) DEFAULT '';
