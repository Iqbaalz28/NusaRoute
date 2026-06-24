-- Order Service Database Schema

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    awb VARCHAR(20) UNIQUE NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING_PAYMENT',
    service_type VARCHAR(10) NOT NULL DEFAULT 'REG',
    delivery_mode VARCHAR(10) NOT NULL DEFAULT 'VIA_HUB',
    sender_name VARCHAR(255) NOT NULL,
    sender_phone VARCHAR(20),
    sender_address TEXT NOT NULL,
    sender_lat DOUBLE PRECISION DEFAULT 0,
    sender_lng DOUBLE PRECISION DEFAULT 0,
    receiver_name VARCHAR(255) NOT NULL,
    receiver_phone VARCHAR(20),
    receiver_address TEXT NOT NULL,
    receiver_lat DOUBLE PRECISION DEFAULT 0,
    receiver_lng DOUBLE PRECISION DEFAULT 0,
    item_description TEXT NOT NULL DEFAULT '',
    weight_kg DECIMAL(10,2) NOT NULL DEFAULT 0,
    length_cm DECIMAL(10,2) DEFAULT 0,
    width_cm DECIMAL(10,2) DEFAULT 0,
    height_cm DECIMAL(10,2) DEFAULT 0,
    is_insured BOOLEAN DEFAULT false,
    insured_value DECIMAL(15,2) DEFAULT 0,
    shipping_cost DECIMAL(15,2) NOT NULL DEFAULT 0,
    insurance_cost DECIMAL(15,2) DEFAULT 0,
    total_cost DECIMAL(15,2) NOT NULL DEFAULT 0,
    delivery_attempts INT DEFAULT 0,
    courier_id VARCHAR(36),
    paid_at TIMESTAMP,
    picked_up_at TIMESTAMP,
    delivered_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_awb ON orders(awb);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

CREATE TABLE IF NOT EXISTS order_status_history (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id),
    status VARCHAR(30) NOT NULL,
    note TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_history_order_id ON order_status_history(order_id);
