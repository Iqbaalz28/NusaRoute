-- Hub Service Database Schema

CREATE TABLE IF NOT EXISTS hubs (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL,
    city VARCHAR(100) NOT NULL,
    province VARCHAR(100) NOT NULL,
    lat DOUBLE PRECISION DEFAULT 0,
    lng DOUBLE PRECISION DEFAULT 0,
    type VARCHAR(30) NOT NULL DEFAULT 'SORTATION',
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS scan_logs (
    id VARCHAR(36) PRIMARY KEY,
    awb VARCHAR(20) NOT NULL,
    order_id VARCHAR(36),
    hub_id VARCHAR(36) NOT NULL REFERENCES hubs(id),
    scan_type VARCHAR(20) NOT NULL,
    operator_id VARCHAR(36),
    note TEXT,
    scanned_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scan_logs_awb ON scan_logs(awb);
CREATE INDEX idx_scan_logs_hub_id ON scan_logs(hub_id);
CREATE INDEX idx_scan_logs_scanned_at ON scan_logs(scanned_at);

-- Seed some hubs across Indonesia
INSERT INTO hubs (id, name, code, city, province, lat, lng, type) VALUES
('hub-jkt-1', 'Hub Jakarta Pusat', 'JKT-01', 'Jakarta', 'DKI Jakarta', -6.1751, 106.8650, 'SORTATION'),
('hub-bdg-1', 'Hub Bandung', 'BDG-01', 'Bandung', 'Jawa Barat', -6.9175, 107.6191, 'SORTATION'),
('hub-sby-1', 'Hub Surabaya', 'SBY-01', 'Surabaya', 'Jawa Timur', -7.2575, 112.7521, 'SORTATION'),
('hub-smg-1', 'Hub Semarang', 'SMG-01', 'Semarang', 'Jawa Tengah', -6.9666, 110.4196, 'TRANSIT'),
('hub-mdn-1', 'Hub Medan', 'MDN-01', 'Medan', 'Sumatera Utara', 3.5952, 98.6722, 'SORTATION'),
('hub-mks-1', 'Hub Makassar', 'MKS-01', 'Makassar', 'Sulawesi Selatan', -5.1477, 119.4327, 'SORTATION'),
('hub-bali-1', 'Hub Denpasar', 'DPS-01', 'Denpasar', 'Bali', -8.6705, 115.2126, 'TRANSIT'),
('hub-yk-1', 'Hub Yogyakarta', 'YK-01', 'Yogyakarta', 'DI Yogyakarta', -7.7956, 110.3695, 'DISTRIBUTION')
ON CONFLICT (code) DO NOTHING;
