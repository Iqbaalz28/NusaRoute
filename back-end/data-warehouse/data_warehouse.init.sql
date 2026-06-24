-- ============================================================================
-- NusaRoute — Data Warehouse (Star Schema) initialization
-- Target DB: nusaroute_warehouse
-- Implements the star schema documented in README "⭐ Star Schema — Data Warehouse":
-- one fact (fact_pengiriman) + 8 conformed dimensions. Self-contained: DDL +
-- dimension seeds + a representative aggregated fact seed for the analytics
-- dashboard. Safe to re-run (drops & recreates).
--
-- Load:  psql -U nusaroute -d nusaroute_warehouse -f data_warehouse.init.sql
-- ============================================================================

DROP TABLE IF EXISTS fact_pengiriman CASCADE;
DROP TABLE IF EXISTS dim_waktu CASCADE;
DROP TABLE IF EXISTS dim_lokasi CASCADE;
DROP TABLE IF EXISTS dim_layanan CASCADE;
DROP TABLE IF EXISTS dim_hub CASCADE;
DROP TABLE IF EXISTS dim_kendaraan CASCADE;
DROP TABLE IF EXISTS dim_metode_pembayaran CASCADE;
DROP TABLE IF EXISTS dim_status_pengiriman CASCADE;

-- ----------------------------------------------------------------------------
-- Dimension tables
-- ----------------------------------------------------------------------------
CREATE TABLE dim_waktu (
    tanggal_key   INT PRIMARY KEY,         -- surrogate key YYYYMMDD
    tanggal       DATE NOT NULL,
    hari          VARCHAR(10) NOT NULL,
    minggu_ke     INT NOT NULL,
    bulan         INT NOT NULL,
    nama_bulan    VARCHAR(20) NOT NULL,
    kuartal       INT NOT NULL,
    tahun         INT NOT NULL,
    is_weekend    BOOLEAN NOT NULL DEFAULT false,
    is_hari_libur BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE dim_lokasi (
    lokasi_key     SERIAL PRIMARY KEY,
    provinsi       VARCHAR(100) NOT NULL,
    kabupaten_kota VARCHAR(100) NOT NULL,
    kecamatan      VARCHAR(100),
    kelurahan      VARCHAR(100),
    kode_pos       VARCHAR(10),
    pulau          VARCHAR(50)
);

CREATE TABLE dim_layanan (
    layanan_key        SERIAL PRIMARY KEY,
    kode_layanan       VARCHAR(10) NOT NULL,
    nama_layanan       VARCHAR(50) NOT NULL,
    kategori_kecepatan VARCHAR(20),
    estimasi_hari      VARCHAR(10)
);

CREATE TABLE dim_hub (
    hub_key   SERIAL PRIMARY KEY,
    kode_hub  VARCHAR(20) NOT NULL,
    nama_hub  VARCHAR(100) NOT NULL,
    tipe_hub  VARCHAR(20),
    kota      VARCHAR(100),
    provinsi  VARCHAR(100)
);

CREATE TABLE dim_kendaraan (
    kendaraan_key      SERIAL PRIMARY KEY,
    tipe_kendaraan     VARCHAR(20) NOT NULL,
    kategori_kapasitas VARCHAR(20)
);

CREATE TABLE dim_metode_pembayaran (
    metode_bayar_key SERIAL PRIMARY KEY,
    kode_metode      VARCHAR(20) NOT NULL,
    nama_metode      VARCHAR(50) NOT NULL,
    kategori         VARCHAR(20)
);

CREATE TABLE dim_status_pengiriman (
    status_key      SERIAL PRIMARY KEY,
    kode_status     VARCHAR(30) NOT NULL,
    nama_status     VARCHAR(50) NOT NULL,
    kategori_status VARCHAR(20),
    is_final        BOOLEAN NOT NULL DEFAULT false
);

-- ----------------------------------------------------------------------------
-- Fact table
-- ----------------------------------------------------------------------------
CREATE TABLE fact_pengiriman (
    pengiriman_key                 BIGSERIAL PRIMARY KEY,
    tanggal_key                    INT NOT NULL REFERENCES dim_waktu(tanggal_key),
    layanan_key                    INT NOT NULL REFERENCES dim_layanan(layanan_key),
    asal_lokasi_key                INT NOT NULL REFERENCES dim_lokasi(lokasi_key),
    tujuan_lokasi_key              INT NOT NULL REFERENCES dim_lokasi(lokasi_key),
    hub_key                        INT NOT NULL REFERENCES dim_hub(hub_key),
    tipe_kendaraan_key             INT NOT NULL REFERENCES dim_kendaraan(kendaraan_key),
    metode_bayar_key               INT NOT NULL REFERENCES dim_metode_pembayaran(metode_bayar_key),
    status_key                     INT NOT NULL REFERENCES dim_status_pengiriman(status_key),
    -- Measures
    jumlah_paket                   INT,
    jumlah_transaksi               INT,
    berat_kg                       DECIMAL(12,2),
    jarak_km                       DECIMAL(12,2),
    biaya_pengiriman               DECIMAL(15,2),
    biaya_asuransi                 DECIMAL(15,2),
    total_biaya                    DECIMAL(15,2),
    durasi_pengiriman_jam          DECIMAL(10,2),
    jumlah_percobaan               INT,
    avg_pendapatan_driver_per_hari DECIMAL(15,2),
    avg_performa_driver            DECIMAL(3,2)
);

CREATE INDEX idx_fact_tanggal  ON fact_pengiriman(tanggal_key);
CREATE INDEX idx_fact_tujuan   ON fact_pengiriman(tujuan_lokasi_key);
CREATE INDEX idx_fact_layanan  ON fact_pengiriman(layanan_key);
CREATE INDEX idx_fact_status   ON fact_pengiriman(status_key);

-- ----------------------------------------------------------------------------
-- Seed: dim_waktu (1 Jan 2026 – 30 Jun 2026)
-- ----------------------------------------------------------------------------
INSERT INTO dim_waktu (tanggal_key, tanggal, hari, minggu_ke, bulan, nama_bulan, kuartal, tahun, is_weekend, is_hari_libur)
SELECT
    to_char(d, 'YYYYMMDD')::int,
    d::date,
    CASE EXTRACT(ISODOW FROM d)
        WHEN 1 THEN 'Senin' WHEN 2 THEN 'Selasa' WHEN 3 THEN 'Rabu'
        WHEN 4 THEN 'Kamis' WHEN 5 THEN 'Jumat' WHEN 6 THEN 'Sabtu' ELSE 'Minggu' END,
    EXTRACT(WEEK FROM d)::int,
    EXTRACT(MONTH FROM d)::int,
    CASE EXTRACT(MONTH FROM d)
        WHEN 1 THEN 'Januari' WHEN 2 THEN 'Februari' WHEN 3 THEN 'Maret'
        WHEN 4 THEN 'April' WHEN 5 THEN 'Mei' WHEN 6 THEN 'Juni'
        WHEN 7 THEN 'Juli' WHEN 8 THEN 'Agustus' WHEN 9 THEN 'September'
        WHEN 10 THEN 'Oktober' WHEN 11 THEN 'November' ELSE 'Desember' END,
    EXTRACT(QUARTER FROM d)::int,
    EXTRACT(YEAR FROM d)::int,
    EXTRACT(ISODOW FROM d) IN (6, 7),
    false
FROM generate_series('2026-01-01'::date, '2026-06-30'::date, '1 day') AS d;

-- ----------------------------------------------------------------------------
-- Seed: dim_lokasi (geographic hierarchy)
-- ----------------------------------------------------------------------------
INSERT INTO dim_lokasi (provinsi, kabupaten_kota, kecamatan, kelurahan, kode_pos, pulau) VALUES
('DKI Jakarta',        'Jakarta Pusat',  'Tanah Abang',   'Bendungan Hilir', '10210', 'Jawa'),
('Jawa Barat',         'Kota Bandung',   'Coblong',       'Dago',            '40135', 'Jawa'),
('Jawa Barat',         'Kota Bogor',     'Bogor Tengah',  'Pabaton',         '16121', 'Jawa'),
('Jawa Tengah',        'Kota Semarang',  'Semarang Tengah','Pekunden',       '50134', 'Jawa'),
('DI Yogyakarta',      'Kota Yogyakarta','Gondokusuman',  'Klitren',         '55222', 'Jawa'),
('Jawa Timur',         'Kota Surabaya',  'Gubeng',        'Airlangga',       '60286', 'Jawa'),
('Jawa Timur',         'Kota Malang',    'Klojen',        'Kauman',          '65119', 'Jawa'),
('Bali',               'Kota Denpasar',  'Denpasar Barat','Pemecutan',       '80119', 'Bali'),
('Sumatera Utara',     'Kota Medan',     'Medan Kota',    'Mesjid',          '20214', 'Sumatera'),
('Sumatera Selatan',   'Kota Palembang', 'Ilir Barat I',  'Bukit Lama',      '30139', 'Sumatera'),
('Kalimantan Timur',   'Kota Balikpapan','Balikpapan Kota','Klandasan Ilir', '76113', 'Kalimantan'),
('Sulawesi Selatan',   'Kota Makassar',  'Makassar',      'Maradekaya',      '90142', 'Sulawesi');

-- ----------------------------------------------------------------------------
-- Seed: dim_layanan
-- ----------------------------------------------------------------------------
INSERT INTO dim_layanan (kode_layanan, nama_layanan, kategori_kecepatan, estimasi_hari) VALUES
('REG',   'Regular',            'Standar',  '2-3'),
('YES',   'Yakin Esok Sampai',  'Ekspres',  '1'),
('CARGO', 'Kargo',              'Kargo',    '5-7'),
('SAME',  'Same Day',           'Same Day', '0');

-- ----------------------------------------------------------------------------
-- Seed: dim_hub
-- ----------------------------------------------------------------------------
INSERT INTO dim_hub (kode_hub, nama_hub, tipe_hub, kota, provinsi) VALUES
('JKT-01', 'Hub Jakarta Pusat', 'SORTATION',    'Jakarta',    'DKI Jakarta'),
('BDG-01', 'Hub Bandung',       'SORTATION',    'Bandung',    'Jawa Barat'),
('SMG-01', 'Hub Semarang',      'TRANSIT',      'Semarang',   'Jawa Tengah'),
('YK-01',  'Hub Yogyakarta',    'DISTRIBUTION', 'Yogyakarta', 'DI Yogyakarta'),
('SBY-01', 'Hub Surabaya',      'SORTATION',    'Surabaya',   'Jawa Timur'),
('DPS-01', 'Hub Denpasar',      'TRANSIT',      'Denpasar',   'Bali'),
('MDN-01', 'Hub Medan',         'SORTATION',    'Medan',      'Sumatera Utara'),
('MKS-01', 'Hub Makassar',      'SORTATION',    'Makassar',   'Sulawesi Selatan');

-- ----------------------------------------------------------------------------
-- Seed: dim_kendaraan
-- ----------------------------------------------------------------------------
INSERT INTO dim_kendaraan (tipe_kendaraan, kategori_kapasitas) VALUES
('MOTORCYCLE', 'Ringan (<=5kg)'),
('CAR',        'Sedang (<=20kg)'),
('VAN',        'Berat (>20kg)');

-- ----------------------------------------------------------------------------
-- Seed: dim_metode_pembayaran
-- ----------------------------------------------------------------------------
INSERT INTO dim_metode_pembayaran (kode_metode, nama_metode, kategori) VALUES
('VA',       'Virtual Account',         'Digital'),
('E_WALLET', 'E-Wallet',                'Digital'),
('CARD',     'Kartu Kredit/Debit',      'Digital'),
('COD',      'Bayar di Tempat (COD)',   'Tunai');

-- ----------------------------------------------------------------------------
-- Seed: dim_status_pengiriman
-- ----------------------------------------------------------------------------
INSERT INTO dim_status_pengiriman (kode_status, nama_status, kategori_status, is_final) VALUES
('DELIVERED',        'Terkirim',           'Sukses', true),
('IN_TRANSIT',       'Dalam Perjalanan',   'Proses', false),
('OUT_FOR_DELIVERY', 'Sedang Diantar',     'Proses', false),
('CANCELLED',        'Dibatalkan',         'Gagal',  true),
('DELIVERY_FAILED',  'Gagal Kirim',        'Gagal',  true),
('RETURN_TO_SENDER', 'Dikembalikan',       'Gagal',  true);

-- ----------------------------------------------------------------------------
-- Seed: fact_pengiriman — representative aggregated rows.
-- One row per (15th-of-month) x destination-location x service, with the
-- remaining dimensions sampled and measures scaled by package count. Status is
-- skewed toward DELIVERED to mimic a healthy network.
-- ----------------------------------------------------------------------------
-- The inner subquery (fenced with OFFSET 0 so it is materialized, not flattened)
-- evaluates random() once PER ROW; the outer SELECT then reuses those stable
-- per-row values across the derived measures. Dimension keys are 1..N contiguous
-- because the dimension seeds above are fresh SERIAL inserts. Status is skewed
-- toward DELIVERED (key 1 = 'Sukses') to mimic a healthy network.
INSERT INTO fact_pengiriman
    (tanggal_key, layanan_key, asal_lokasi_key, tujuan_lokasi_key, hub_key,
     tipe_kendaraan_key, metode_bayar_key, status_key,
     jumlah_paket, jumlah_transaksi, berat_kg, jarak_km,
     biaya_pengiriman, biaya_asuransi, total_biaya,
     durasi_pengiriman_jam, jumlah_percobaan,
     avg_pendapatan_driver_per_hari, avg_performa_driver)
SELECT
    b.tanggal_key,
    b.layanan_key,
    b.asal_lokasi_key,
    b.tujuan_lokasi_key,
    b.hub_key,
    b.tipe_kendaraan_key,
    b.metode_bayar_key,
    (CASE
        WHEN b.rstat < 0.70 THEN 1   -- DELIVERED       (Sukses)
        WHEN b.rstat < 0.82 THEN 2   -- IN_TRANSIT      (Proses)
        WHEN b.rstat < 0.90 THEN 3   -- OUT_FOR_DELIVERY(Proses)
        WHEN b.rstat < 0.95 THEN 4   -- CANCELLED       (Gagal)
        WHEN b.rstat < 0.98 THEN 5   -- DELIVERY_FAILED (Gagal)
        ELSE 6                       -- RETURN_TO_SENDER(Gagal)
    END)                                 AS status_key,
    b.pkt,
    b.pkt,
    round((b.pkt * b.wkg)::numeric, 2),
    round((b.pkt * b.dist)::numeric, 2),
    round((b.pkt * b.ongkir)::numeric, 2),
    round((b.pkt * b.ongkir * 0.05)::numeric, 2),
    round((b.pkt * b.ongkir * 1.05)::numeric, 2),
    round(b.dur::numeric, 2),
    b.att,
    round(b.pendapatan::numeric, 2),
    round(b.perf::numeric, 2)
FROM (
    SELECT
        w.tanggal_key,
        s.layanan_key,
        tuj.lokasi_key                       AS tujuan_lokasi_key,
        (1 + floor(random() * 12))::int      AS asal_lokasi_key,
        (1 + floor(random() * 8))::int       AS hub_key,
        (1 + floor(random() * 3))::int       AS tipe_kendaraan_key,
        (1 + floor(random() * 4))::int       AS metode_bayar_key,
        random()                             AS rstat,
        (5 + (random() * 145))::int          AS pkt,
        (0.5 + random() * 24)                AS wkg,
        (15 + random() * 1400)               AS dist,
        (9000 + random() * 90000)            AS ongkir,
        (4 + random() * 70)                  AS dur,
        (1 + (random() * 2))::int            AS att,
        (150000 + random() * 350000)         AS pendapatan,
        (3.2 + random() * 1.8)               AS perf
    FROM dim_waktu w
    JOIN dim_layanan s ON true
    JOIN dim_lokasi tuj ON true
    WHERE EXTRACT(DAY FROM w.tanggal) = 15
    OFFSET 0
) AS b;

-- Quick sanity output when run interactively.
DO $$
DECLARE n bigint;
BEGIN
    SELECT count(*) INTO n FROM fact_pengiriman;
    RAISE NOTICE 'fact_pengiriman seeded with % rows', n;
END $$;
