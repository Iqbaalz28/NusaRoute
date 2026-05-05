-- Pricing Service Database Schema

CREATE TABLE IF NOT EXISTS service_types (
    code VARCHAR(10) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_per_km DECIMAL(10,2) NOT NULL DEFAULT 50,
    price_per_kg DECIMAL(10,2) NOT NULL DEFAULT 3000,
    base_fee DECIMAL(10,2) NOT NULL DEFAULT 8000,
    est_days VARCHAR(20) NOT NULL DEFAULT '2-4 hari',
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS tariff_zones (
    id VARCHAR(36) PRIMARY KEY,
    zone_name VARCHAR(100) NOT NULL,
    min_dist_km DECIMAL(10,2) NOT NULL,
    max_dist_km DECIMAL(10,2) NOT NULL,
    multiplier DECIMAL(5,2) NOT NULL DEFAULT 1.0
);

-- Seed default service types
INSERT INTO service_types (code, name, description, price_per_km, price_per_kg, base_fee, est_days) VALUES
('REG', 'Reguler', 'Pengiriman reguler 2-4 hari kerja', 30, 2500, 8000, '2-4 hari'),
('YES', 'Yakin Esok Sampai', 'Pengiriman express 1 hari', 80, 5000, 15000, '1 hari'),
('CARGO', 'Kargo', 'Pengiriman barang besar/berat', 15, 1500, 5000, '5-7 hari'),
('SAME', 'Same Day', 'Pengiriman di hari yang sama', 150, 8000, 25000, '< 12 jam')
ON CONFLICT (code) DO NOTHING;

-- Seed tariff zones
INSERT INTO tariff_zones (id, zone_name, min_dist_km, max_dist_km, multiplier) VALUES
('z1', 'Dalam Kota', 0, 30, 1.0),
('z2', 'Antar Kota (Dekat)', 30, 150, 1.2),
('z3', 'Antar Provinsi', 150, 800, 1.5),
('z4', 'Lintas Pulau', 800, 5000, 2.0),
('z5', 'Indonesia Timur', 5000, 99999, 2.5)
ON CONFLICT (id) DO NOTHING;
