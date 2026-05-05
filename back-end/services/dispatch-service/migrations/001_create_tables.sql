-- Dispatch Service Database Schema

CREATE TABLE IF NOT EXISTS assignments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    awb VARCHAR(20),
    courier_id VARCHAR(36) NOT NULL,
    courier_name VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'ASSIGNED',
    pickup_lat DOUBLE PRECISION DEFAULT 0,
    pickup_lng DOUBLE PRECISION DEFAULT 0,
    pickup_address TEXT,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    picked_up_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assignments_order_id ON assignments(order_id);
CREATE INDEX idx_assignments_courier_id ON assignments(courier_id);
CREATE INDEX idx_assignments_status ON assignments(status);
