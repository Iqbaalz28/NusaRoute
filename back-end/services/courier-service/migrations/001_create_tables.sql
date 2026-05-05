-- Courier Service Database Schema

CREATE TABLE IF NOT EXISTS couriers (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    vehicle_type VARCHAR(20) NOT NULL DEFAULT 'MOTORCYCLE',
    vehicle_plate VARCHAR(20),
    max_capacity_kg DECIMAL(10,2) DEFAULT 30,
    current_lat DOUBLE PRECISION DEFAULT 0,
    current_lng DOUBLE PRECISION DEFAULT 0,
    is_online BOOLEAN DEFAULT false,
    is_available BOOLEAN DEFAULT true,
    rating DECIMAL(3,2) DEFAULT 5.0,
    total_deliveries INT DEFAULT 0,
    coverage_area VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_couriers_user_id ON couriers(user_id);
CREATE INDEX idx_couriers_online ON couriers(is_online, is_available);
CREATE INDEX idx_couriers_location ON couriers(current_lat, current_lng);
