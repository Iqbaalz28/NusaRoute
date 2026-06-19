// Package repository runs aggregate (OLAP) queries against the NusaRoute data
// warehouse star schema (fact_pengiriman + conformed dimensions).
package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Overview struct {
	TotalPaket      int64   `json:"total_paket" db:"total_paket"`
	TotalTransaksi  int64   `json:"total_transaksi" db:"total_transaksi"`
	TotalPendapatan float64 `json:"total_pendapatan" db:"total_pendapatan"`
	AvgDurasiJam    float64 `json:"avg_durasi_jam" db:"avg_durasi_jam"`
	SuccessRate     float64 `json:"success_rate" db:"success_rate"` // % paket berstatus Sukses
}

type MonthlyPoint struct {
	Tahun       int     `json:"tahun" db:"tahun"`
	Bulan       int     `json:"bulan" db:"bulan"`
	NamaBulan   string  `json:"nama_bulan" db:"nama_bulan"`
	Paket       int64   `json:"paket" db:"paket"`
	Pendapatan  float64 `json:"pendapatan" db:"pendapatan"`
	AvgDurasi   float64 `json:"avg_durasi_jam" db:"avg_durasi_jam"`
}

type ProvinceRow struct {
	Provinsi   string  `json:"provinsi" db:"provinsi"`
	Paket      int64   `json:"paket" db:"paket"`
	Pendapatan float64 `json:"pendapatan" db:"pendapatan"`
}

type ServiceRow struct {
	NamaLayanan string  `json:"nama_layanan" db:"nama_layanan"`
	Transaksi   int64   `json:"transaksi" db:"transaksi"`
	AvgBerat    float64 `json:"avg_berat_kg" db:"avg_berat_kg"`
	TotalOngkir float64 `json:"total_ongkir" db:"total_ongkir"`
}

type StatusRow struct {
	KategoriStatus string `json:"kategori_status" db:"kategori_status"`
	Paket          int64  `json:"paket" db:"paket"`
}

type DriverRow struct {
	Provinsi          string  `json:"provinsi" db:"provinsi"`
	AvgPendapatan     float64 `json:"avg_pendapatan_driver" db:"avg_pendapatan_driver"`
	AvgPerforma       float64 `json:"avg_performa_driver" db:"avg_performa_driver"`
}

type AnalyticsRepository struct{ db *sqlx.DB }

func NewAnalyticsRepository(db *sqlx.DB) *AnalyticsRepository { return &AnalyticsRepository{db: db} }

// Overview returns headline KPIs across the whole warehouse.
func (r *AnalyticsRepository) Overview(ctx context.Context) (Overview, error) {
	var o Overview
	err := r.db.GetContext(ctx, &o, `
		SELECT
			COALESCE(SUM(f.jumlah_paket), 0)                          AS total_paket,
			COALESCE(SUM(f.jumlah_transaksi), 0)                      AS total_transaksi,
			COALESCE(SUM(f.total_biaya), 0)                           AS total_pendapatan,
			COALESCE(AVG(f.durasi_pengiriman_jam), 0)                 AS avg_durasi_jam,
			COALESCE(100.0 * SUM(CASE WHEN ds.kategori_status = 'Sukses' THEN f.jumlah_paket ELSE 0 END)
				/ NULLIF(SUM(f.jumlah_paket), 0), 0)                 AS success_rate
		FROM fact_pengiriman f
		JOIN dim_status_pengiriman ds ON f.status_key = ds.status_key`)
	return o, err
}

// Monthly trend of packages, revenue and average duration (README Query 1 shape).
func (r *AnalyticsRepository) Monthly(ctx context.Context) ([]MonthlyPoint, error) {
	var rows []MonthlyPoint
	err := r.db.SelectContext(ctx, &rows, `
		SELECT dw.tahun, dw.bulan, dw.nama_bulan,
			SUM(f.jumlah_paket)          AS paket,
			SUM(f.total_biaya)           AS pendapatan,
			AVG(f.durasi_pengiriman_jam) AS avg_durasi_jam
		FROM fact_pengiriman f
		JOIN dim_waktu dw ON f.tanggal_key = dw.tanggal_key
		GROUP BY dw.tahun, dw.bulan, dw.nama_bulan
		ORDER BY dw.tahun, dw.bulan`)
	return rows, err
}

// ByProvince aggregates packages + revenue per destination province.
func (r *AnalyticsRepository) ByProvince(ctx context.Context) ([]ProvinceRow, error) {
	var rows []ProvinceRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT dl.provinsi,
			SUM(f.jumlah_paket) AS paket,
			SUM(f.total_biaya)  AS pendapatan
		FROM fact_pengiriman f
		JOIN dim_lokasi dl ON f.tujuan_lokasi_key = dl.lokasi_key
		GROUP BY dl.provinsi
		ORDER BY pendapatan DESC`)
	return rows, err
}

// ByService aggregates per service type (README Query 2 shape).
func (r *AnalyticsRepository) ByService(ctx context.Context) ([]ServiceRow, error) {
	var rows []ServiceRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT dly.nama_layanan,
			SUM(f.jumlah_transaksi) AS transaksi,
			AVG(f.berat_kg / NULLIF(f.jumlah_paket,0)) AS avg_berat_kg,
			SUM(f.biaya_pengiriman) AS total_ongkir
		FROM fact_pengiriman f
		JOIN dim_layanan dly ON f.layanan_key = dly.layanan_key
		GROUP BY dly.nama_layanan
		ORDER BY transaksi DESC`)
	return rows, err
}

// ByStatus aggregates package count per status category (README Query 3 shape).
func (r *AnalyticsRepository) ByStatus(ctx context.Context) ([]StatusRow, error) {
	var rows []StatusRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT ds.kategori_status, SUM(f.jumlah_paket) AS paket
		FROM fact_pengiriman f
		JOIN dim_status_pengiriman ds ON f.status_key = ds.status_key
		GROUP BY ds.kategori_status
		ORDER BY paket DESC`)
	return rows, err
}

// Driver metrics per destination province (README Query 4 shape).
func (r *AnalyticsRepository) Driver(ctx context.Context) ([]DriverRow, error) {
	var rows []DriverRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT dl.provinsi,
			AVG(f.avg_pendapatan_driver_per_hari) AS avg_pendapatan_driver,
			AVG(f.avg_performa_driver)            AS avg_performa_driver
		FROM fact_pengiriman f
		JOIN dim_lokasi dl ON f.tujuan_lokasi_key = dl.lokasi_key
		GROUP BY dl.provinsi
		ORDER BY avg_pendapatan_driver DESC`)
	return rows, err
}
