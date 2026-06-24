-- =====================================================
-- NusaRoute Data Warehouse — Initialization Script
-- Database: nusaroute_datawarehouse (PostgreSQL 16)
-- =====================================================

-- =====================================================
-- 1. DIMENSION TABLES
-- =====================================================

-- Dimensi Tanggal
CREATE TABLE IF NOT EXISTS dim_date (
    date_key        INT PRIMARY KEY,                -- format: YYYYMMDD (e.g., 20260612)
    full_date       DATE NOT NULL,
    year            SMALLINT NOT NULL,
    quarter         SMALLINT NOT NULL,
    month           SMALLINT NOT NULL,
    week_of_year    SMALLINT NOT NULL,
    day_of_month    SMALLINT NOT NULL,
    day_of_week     SMALLINT NOT NULL,              -- 0=Sunday, 6=Saturday
    month_name      VARCHAR(20) NOT NULL,
    day_name        VARCHAR(20) NOT NULL,
    is_weekend      BOOLEAN NOT NULL DEFAULT false,
    is_holiday      BOOLEAN NOT NULL DEFAULT false
);

-- Dimensi Waktu
CREATE TABLE IF NOT EXISTS dim_time (
    time_key    INT PRIMARY KEY,                    -- format: HHmm (e.g., 1430)
    full_time   VARCHAR(5) NOT NULL,                -- "14:30"
    hour        SMALLINT NOT NULL,
    minute      SMALLINT NOT NULL,
    period      VARCHAR(2) NOT NULL,                -- AM / PM
    shift       VARCHAR(20) NOT NULL                -- PAGI (06-14), SIANG (14-22), MALAM (22-06)
);

-- Dimensi Pengguna (SCD Type 2)
CREATE TABLE IF NOT EXISTS dim_user (
    user_key    SERIAL PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,               -- natural key dari User Service
    email       VARCHAR(255),
    full_name   VARCHAR(255),
    phone       VARCHAR(20),
    role        VARCHAR(20),                        -- USER, COURIER, ADMIN
    is_active   BOOLEAN DEFAULT true,
    -- SCD Type 2 columns
    valid_from  TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_to    TIMESTAMP DEFAULT '9999-12-31 23:59:59',
    is_current  BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_dim_user_user_id ON dim_user(user_id);
CREATE INDEX IF NOT EXISTS idx_dim_user_current ON dim_user(user_id, is_current);

-- Dimensi Kurir (SCD Type 2)
CREATE TABLE IF NOT EXISTS dim_courier (
    courier_key      SERIAL PRIMARY KEY,
    courier_id       VARCHAR(36) NOT NULL,          -- natural key dari Courier Service
    full_name        VARCHAR(255),
    phone            VARCHAR(20),
    vehicle_type     VARCHAR(20),                   -- MOTORCYCLE, CAR, VAN
    vehicle_plate    VARCHAR(20),
    max_capacity_kg  DECIMAL(10,2),
    rating           DECIMAL(3,2) DEFAULT 5.0,
    total_deliveries INT DEFAULT 0,
    coverage_area    VARCHAR(100),
    is_active        BOOLEAN DEFAULT true,
    -- SCD Type 2 columns
    valid_from       TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_to         TIMESTAMP DEFAULT '9999-12-31 23:59:59',
    is_current       BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_dim_courier_courier_id ON dim_courier(courier_id);
CREATE INDEX IF NOT EXISTS idx_dim_courier_current ON dim_courier(courier_id, is_current);

-- Dimensi Hub
CREATE TABLE IF NOT EXISTS dim_hub (
    hub_key     SERIAL PRIMARY KEY,
    hub_id      VARCHAR(36) NOT NULL,               -- natural key dari Hub Service
    hub_name    VARCHAR(255),
    hub_code    VARCHAR(20),
    city        VARCHAR(100),
    province    VARCHAR(100),
    hub_type    VARCHAR(30),                        -- SORTATION, TRANSIT, DISTRIBUTION
    lat         DOUBLE PRECISION,
    lng         DOUBLE PRECISION,
    is_active   BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_dim_hub_hub_id ON dim_hub(hub_id);

-- Dimensi Jenis Layanan
CREATE TABLE IF NOT EXISTS dim_service_type (
    service_type_key  SERIAL PRIMARY KEY,
    code              VARCHAR(10) NOT NULL,          -- REG, YES, CARGO, SAME
    name              VARCHAR(100),
    description       TEXT,
    price_per_km      DECIMAL(10,2),
    price_per_kg      DECIMAL(10,2),
    base_fee          DECIMAL(10,2),
    est_days          VARCHAR(20),
    is_active         BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_dim_service_type_code ON dim_service_type(code);

-- Dimensi Lokasi
CREATE TABLE IF NOT EXISTS dim_location (
    location_key  SERIAL PRIMARY KEY,
    province      VARCHAR(100),
    city          VARCHAR(100),
    district      VARCHAR(100),
    postal_code   VARCHAR(10),
    full_address  TEXT,
    lat           DOUBLE PRECISION,
    lng           DOUBLE PRECISION,
    zone_name     VARCHAR(100)                      -- Dalam Kota, Antar Kota, Antar Provinsi, dll
);

CREATE INDEX IF NOT EXISTS idx_dim_location_city ON dim_location(province, city);
CREATE INDEX IF NOT EXISTS idx_dim_location_coords ON dim_location(lat, lng);

-- Dimensi Metode Pembayaran
CREATE TABLE IF NOT EXISTS dim_payment_method (
    payment_method_key  SERIAL PRIMARY KEY,
    method_code         VARCHAR(20) NOT NULL,        -- VA, E_WALLET, CARD, COD
    method_name         VARCHAR(100),
    method_category     VARCHAR(50)                  -- DIGITAL, CASH, CARD
);

-- Unknown Member (untuk FK yang belum diketahui)
-- Digunakan saat ETL tidak menemukan dimensi yang cocok
INSERT INTO dim_user (user_key, user_id, email, full_name, role, is_current)
    VALUES (-1, 'UNKNOWN', 'unknown', 'Unknown User', 'UNKNOWN', true)
    ON CONFLICT DO NOTHING;

INSERT INTO dim_courier (courier_key, courier_id, full_name, vehicle_type, is_current)
    VALUES (-1, 'UNKNOWN', 'Unassigned', 'UNKNOWN', true)
    ON CONFLICT DO NOTHING;

INSERT INTO dim_hub (hub_key, hub_id, hub_name, hub_code, city, province, hub_type)
    VALUES (-1, 'UNKNOWN', 'Unknown Hub', 'UNK', 'Unknown', 'Unknown', 'UNKNOWN')
    ON CONFLICT DO NOTHING;

INSERT INTO dim_service_type (service_type_key, code, name)
    VALUES (-1, 'UNK', 'Unknown Service')
    ON CONFLICT DO NOTHING;

INSERT INTO dim_location (location_key, province, city, zone_name)
    VALUES (-1, 'Unknown', 'Unknown', 'Unknown')
    ON CONFLICT DO NOTHING;

INSERT INTO dim_payment_method (payment_method_key, method_code, method_name, method_category)
    VALUES (-1, 'UNKNOWN', 'Unknown', 'UNKNOWN')
    ON CONFLICT DO NOTHING;

-- =====================================================
-- 2. FACT TABLES
-- =====================================================

-- Fakta Pengiriman (grain: 1 baris = 1 order)
CREATE TABLE IF NOT EXISTS fact_shipment (
    shipment_key            SERIAL PRIMARY KEY,
    order_id                VARCHAR(36) NOT NULL,
    awb                     VARCHAR(20) NOT NULL,
    sender_user_key         INT NOT NULL DEFAULT -1 REFERENCES dim_user(user_key),
    courier_key             INT NOT NULL DEFAULT -1 REFERENCES dim_courier(courier_key),
    service_type_key        INT NOT NULL DEFAULT -1 REFERENCES dim_service_type(service_type_key),
    origin_location_key     INT NOT NULL DEFAULT -1 REFERENCES dim_location(location_key),
    dest_location_key       INT NOT NULL DEFAULT -1 REFERENCES dim_location(location_key),
    order_date_key          INT REFERENCES dim_date(date_key),
    order_time_key          INT REFERENCES dim_time(time_key),
    paid_date_key           INT REFERENCES dim_date(date_key),
    delivered_date_key      INT REFERENCES dim_date(date_key),
    order_status            VARCHAR(30) NOT NULL,
    -- MEASURES
    shipping_cost           DECIMAL(15,2) DEFAULT 0,
    insurance_cost          DECIMAL(15,2) DEFAULT 0,
    total_cost              DECIMAL(15,2) DEFAULT 0,
    weight_kg               DECIMAL(10,2) DEFAULT 0,
    volumetric_kg           DECIMAL(10,2) DEFAULT 0,
    distance_km             DECIMAL(10,2) DEFAULT 0,
    delivery_attempts       INT DEFAULT 0,
    is_insured              BOOLEAN DEFAULT false,
    insured_value           DECIMAL(15,2) DEFAULT 0,
    lead_time_hours         INT,                    -- created_at → delivered_at (jam)
    pickup_to_deliver_hours INT                     -- picked_up_at → delivered_at (jam)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fact_shipment_order_id ON fact_shipment(order_id);
CREATE INDEX IF NOT EXISTS idx_fact_shipment_awb ON fact_shipment(awb);
CREATE INDEX IF NOT EXISTS idx_fact_shipment_order_date ON fact_shipment(order_date_key);
CREATE INDEX IF NOT EXISTS idx_fact_shipment_status ON fact_shipment(order_status);

-- Fakta Pembayaran (grain: 1 baris = 1 transaksi)
CREATE TABLE IF NOT EXISTS fact_payment (
    payment_key         SERIAL PRIMARY KEY,
    transaction_id      VARCHAR(36) NOT NULL,
    order_id            VARCHAR(36),
    payment_date_key    INT REFERENCES dim_date(date_key),
    payment_method_key  INT NOT NULL DEFAULT -1 REFERENCES dim_payment_method(payment_method_key),
    payment_status      VARCHAR(20) NOT NULL,
    -- MEASURES
    amount              DECIMAL(15,2) DEFAULT 0,
    is_refunded         BOOLEAN DEFAULT false,
    refund_amount       DECIMAL(15,2) DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fact_payment_txn_id ON fact_payment(transaction_id);
CREATE INDEX IF NOT EXISTS idx_fact_payment_order_id ON fact_payment(order_id);
CREATE INDEX IF NOT EXISTS idx_fact_payment_date ON fact_payment(payment_date_key);

-- Fakta Scan Hub (grain: 1 baris = 1 event scan)
CREATE TABLE IF NOT EXISTS fact_hub_scan (
    scan_key            SERIAL PRIMARY KEY,
    scan_log_id         VARCHAR(36) NOT NULL,
    awb                 VARCHAR(20),
    order_id            VARCHAR(36),
    hub_key             INT NOT NULL DEFAULT -1 REFERENCES dim_hub(hub_key),
    scan_date_key       INT REFERENCES dim_date(date_key),
    scan_time_key       INT REFERENCES dim_time(time_key),
    scan_type           VARCHAR(20) NOT NULL,       -- ARRIVED, SORTED, DEPARTED
    operator_id         VARCHAR(36),
    -- MEASURES
    dwell_time_minutes  INT                         -- ARRIVED → DEPARTED di hub yang sama
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fact_hub_scan_log_id ON fact_hub_scan(scan_log_id);
CREATE INDEX IF NOT EXISTS idx_fact_hub_scan_awb ON fact_hub_scan(awb);
CREATE INDEX IF NOT EXISTS idx_fact_hub_scan_date ON fact_hub_scan(scan_date_key);

-- Fakta Resolusi (grain: 1 baris = 1 tiket keluhan)
CREATE TABLE IF NOT EXISTS fact_resolution (
    resolution_key      SERIAL PRIMARY KEY,
    ticket_id           VARCHAR(36) NOT NULL,
    order_id            VARCHAR(36),
    awb                 VARCHAR(20),
    user_key            INT NOT NULL DEFAULT -1 REFERENCES dim_user(user_key),
    agent_key           INT NOT NULL DEFAULT -1 REFERENCES dim_courier(courier_key),
    created_date_key    INT REFERENCES dim_date(date_key),
    resolved_date_key   INT REFERENCES dim_date(date_key),
    ticket_type         VARCHAR(30),                -- LOST, DAMAGED, DELIVERY_FAILED, COMPLAINT
    priority            VARCHAR(20),                -- LOW, MEDIUM, HIGH, CRITICAL
    ticket_status       VARCHAR(20),                -- OPEN, IN_PROGRESS, RESOLVED, CLOSED
    resolution_type     VARCHAR(30),                -- REFUND, RESEND, RETURN, CLOSED
    -- MEASURES
    claim_amount        DECIMAL(15,2) DEFAULT 0,
    claim_status        VARCHAR(20),                -- PENDING, APPROVED, REJECTED, PAID
    resolution_time_hours INT                       -- lama penyelesaian tiket (jam)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fact_resolution_ticket_id ON fact_resolution(ticket_id);
CREATE INDEX IF NOT EXISTS idx_fact_resolution_order_id ON fact_resolution(order_id);
CREATE INDEX IF NOT EXISTS idx_fact_resolution_date ON fact_resolution(created_date_key);

-- =====================================================
-- 3. SEED: dim_date (2024-01-01 s/d 2030-12-31)
-- =====================================================

INSERT INTO dim_date (date_key, full_date, year, quarter, month, week_of_year,
                      day_of_month, day_of_week, month_name, day_name, is_weekend)
SELECT
    TO_CHAR(d, 'YYYYMMDD')::INT              AS date_key,
    d                                         AS full_date,
    EXTRACT(YEAR FROM d)::SMALLINT            AS year,
    EXTRACT(QUARTER FROM d)::SMALLINT         AS quarter,
    EXTRACT(MONTH FROM d)::SMALLINT           AS month,
    EXTRACT(WEEK FROM d)::SMALLINT            AS week_of_year,
    EXTRACT(DAY FROM d)::SMALLINT             AS day_of_month,
    EXTRACT(DOW FROM d)::SMALLINT             AS day_of_week,
    TO_CHAR(d, 'Month')                       AS month_name,
    TO_CHAR(d, 'Day')                         AS day_name,
    CASE WHEN EXTRACT(DOW FROM d) IN (0, 6) THEN true ELSE false END AS is_weekend
FROM generate_series('2024-01-01'::DATE, '2030-12-31'::DATE, '1 day'::INTERVAL) AS d
ON CONFLICT (date_key) DO NOTHING;

-- =====================================================
-- 4. SEED: dim_time (00:00 s/d 23:59)
-- =====================================================

INSERT INTO dim_time (time_key, full_time, hour, minute, period, shift)
SELECT
    h * 100 + m                               AS time_key,
    LPAD(h::TEXT, 2, '0') || ':' || LPAD(m::TEXT, 2, '0') AS full_time,
    h                                         AS hour,
    m                                         AS minute,
    CASE WHEN h < 12 THEN 'AM' ELSE 'PM' END AS period,
    CASE
        WHEN h >= 6  AND h < 14 THEN 'PAGI'
        WHEN h >= 14 AND h < 22 THEN 'SIANG'
        ELSE 'MALAM'
    END                                       AS shift
FROM generate_series(0, 23) AS h,
     generate_series(0, 59) AS m
ON CONFLICT (time_key) DO NOTHING;

-- =====================================================
-- 5. SEED: dim_payment_method
-- =====================================================

INSERT INTO dim_payment_method (method_code, method_name, method_category) VALUES
    ('VA',       'Virtual Account',     'DIGITAL'),
    ('E_WALLET', 'E-Wallet',            'DIGITAL'),
    ('CARD',     'Kartu Kredit/Debit',  'CARD'),
    ('COD',      'Cash On Delivery',    'CASH')
ON CONFLICT DO NOTHING;

-- =====================================================
-- 6. SEED: dim_service_type (dari Pricing Service)
-- =====================================================

INSERT INTO dim_service_type (code, name, description, price_per_km, price_per_kg, base_fee, est_days, is_active) VALUES
    ('REG',  'Reguler',            'Pengiriman reguler 2-4 hari kerja',  30,  2500, 8000,  '2-4 hari',  true),
    ('YES',  'Yakin Esok Sampai',  'Pengiriman express 1 hari',          80,  5000, 15000, '1 hari',    true),
    ('CARGO','Kargo',              'Pengiriman barang besar/berat',      15,  1500, 5000,  '5-7 hari',  true),
    ('SAME', 'Same Day',           'Pengiriman di hari yang sama',       150, 8000, 25000, '< 12 jam',  true)
ON CONFLICT DO NOTHING;

-- =====================================================
-- 7. SEED: dim_hub (dari Hub Service)
-- =====================================================

INSERT INTO dim_hub (hub_id, hub_name, hub_code, city, province, lat, lng, hub_type, is_active) VALUES
    ('hub-jkt-1',  'Hub Jakarta Pusat', 'JKT-01', 'Jakarta',     'DKI Jakarta',       -6.1751,  106.8650, 'SORTATION',    true),
    ('hub-bdg-1',  'Hub Bandung',       'BDG-01', 'Bandung',     'Jawa Barat',        -6.9175,  107.6191, 'SORTATION',    true),
    ('hub-sby-1',  'Hub Surabaya',      'SBY-01', 'Surabaya',    'Jawa Timur',        -7.2575,  112.7521, 'SORTATION',    true),
    ('hub-smg-1',  'Hub Semarang',      'SMG-01', 'Semarang',    'Jawa Tengah',       -6.9666,  110.4196, 'TRANSIT',      true),
    ('hub-mdn-1',  'Hub Medan',         'MDN-01', 'Medan',       'Sumatera Utara',     3.5952,   98.6722, 'SORTATION',    true),
    ('hub-mks-1',  'Hub Makassar',      'MKS-01', 'Makassar',    'Sulawesi Selatan',  -5.1477,  119.4327, 'SORTATION',    true),
    ('hub-bali-1', 'Hub Denpasar',      'DPS-01', 'Denpasar',    'Bali',              -8.6705,  115.2126, 'TRANSIT',      true),
    ('hub-yk-1',   'Hub Yogyakarta',    'YK-01',  'Yogyakarta',  'DI Yogyakarta',     -7.7956,  110.3695, 'DISTRIBUTION', true)
ON CONFLICT DO NOTHING;

-- =====================================================
-- 8. ETL LOG TABLE (untuk tracking proses ETL)
-- =====================================================

CREATE TABLE IF NOT EXISTS etl_log (
    id              SERIAL PRIMARY KEY,
    table_name      VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL,           -- RUNNING, SUCCESS, FAILED
    rows_extracted  INT DEFAULT 0,
    rows_loaded     INT DEFAULT 0,
    rows_rejected   INT DEFAULT 0,
    started_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMP,
    error_message   TEXT
);

CREATE TABLE IF NOT EXISTS etl_error_log (
    id              SERIAL PRIMARY KEY,
    table_name      VARCHAR(100) NOT NULL,
    source_id       VARCHAR(100),                   -- ID record yang bermasalah
    error_type      VARCHAR(50),                    -- MISSING_DIMENSION, DATA_ANOMALY, PARSE_ERROR
    error_detail    TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- =====================================================
-- DONE
-- =====================================================
-- Total: 8 dimension tables + 4 fact tables + 2 ETL tables
-- Seeded: dim_date (2557 rows), dim_time (1440 rows),
--         dim_payment_method (4 rows), dim_service_type (4 rows),
--         dim_hub (8 rows), unknown members (6 rows)
