package repository

import (
	"context"
	"math"
	"sort"

	"github.com/nusaroute/services/analytics-service/internal/ml"
)

// ===========================================================================
// ML layer — turns warehouse aggregates into predictions, segments and risk
// scores. Each method pulls raw aggregates from the star schema and runs a
// pure-Go model from the internal/ml package over them.
// ===========================================================================

var monthsID = []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// ---------------------------------------------------------------------------
// 1) Demand & revenue forecasting (linear regression) + anomaly flags
// ---------------------------------------------------------------------------

type MLPoint struct {
	Label      string  `json:"label"`
	Paket      float64 `json:"paket"`
	Pendapatan float64 `json:"pendapatan"`
	IsForecast bool    `json:"is_forecast"`
	IsAnomaly  bool    `json:"is_anomaly"` // history points >2σ off the trend line
}

type ForecastResult struct {
	Series        []MLPoint `json:"series"` // history followed by forecast, in order
	PaketSlope    float64   `json:"paket_slope"`
	PaketR2       float64   `json:"paket_r2"`
	RevSlope      float64   `json:"rev_slope"`
	RevR2         float64   `json:"rev_r2"`
	TrendLabel    string    `json:"trend_label"`
	NextPaket     float64   `json:"next_paket"`
	NextPendapatan float64  `json:"next_pendapatan"`
	GrowthPct     float64   `json:"growth_pct"` // forecast next-month vs. last actual month
	AnomalyCount  int       `json:"anomaly_count"`
	HorizonMonths int       `json:"horizon_months"`
}

// Forecast fits a least-squares trend over the monthly package & revenue series
// and projects the next `HorizonMonths` months. History points lying more than
// 2 residual-std-devs off the fitted line are flagged as anomalies.
func (r *AnalyticsRepository) Forecast(ctx context.Context) (ForecastResult, error) {
	m, err := r.Monthly(ctx)
	if err != nil {
		return ForecastResult{}, err
	}
	res := ForecastResult{HorizonMonths: 3}
	if len(m) == 0 {
		return res, nil
	}

	pk := make([]float64, len(m))
	rv := make([]float64, len(m))
	for i, p := range m {
		pk[i] = float64(p.Paket)
		rv[i] = p.Pendapatan
	}
	lp := ml.FitLine(pk)
	lr := ml.FitLine(rv)
	res.PaketSlope, res.PaketR2 = lp.Slope, lp.R2
	res.RevSlope, res.RevR2 = lr.Slope, lr.R2

	for i, p := range m {
		anom := lp.StdResid > 0 && math.Abs(pk[i]-lp.At(float64(i))) > 2*lp.StdResid
		if anom {
			res.AnomalyCount++
		}
		label := p.NamaBulan
		if len(label) > 3 {
			label = label[:3]
		}
		res.Series = append(res.Series, MLPoint{Label: label, Paket: pk[i], Pendapatan: rv[i], IsAnomaly: anom})
	}

	last := m[len(m)-1]
	for h := 1; h <= res.HorizonMonths; h++ {
		x := float64(len(m) - 1 + h)
		res.Series = append(res.Series, MLPoint{
			Label:      monthsID[(last.Bulan-1+h)%12],
			Paket:      math.Max(0, lp.At(x)),
			Pendapatan: math.Max(0, lr.At(x)),
			IsForecast: true,
		})
	}
	res.NextPaket = math.Max(0, lp.At(float64(len(m))))
	res.NextPendapatan = math.Max(0, lr.At(float64(len(m))))
	if last.Paket > 0 {
		res.GrowthPct = 100 * (res.NextPaket - float64(last.Paket)) / float64(last.Paket)
	}
	switch {
	case lp.Slope > 0.5:
		res.TrendLabel = "Naik"
	case lp.Slope < -0.5:
		res.TrendLabel = "Turun"
	default:
		res.TrendLabel = "Stabil"
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// 2) Province segmentation (K-Means, k=3)
// ---------------------------------------------------------------------------

type SegProvince struct {
	Provinsi    string  `json:"provinsi" db:"provinsi"`
	Paket       int64   `json:"paket" db:"paket"`
	Pendapatan  float64 `json:"pendapatan" db:"pendapatan"`
	SuccessRate float64 `json:"success_rate" db:"success_rate"`
	Cluster     int     `json:"cluster"`
	Tier        string  `json:"tier"`
}

type ClusterSummary struct {
	Cluster       int     `json:"cluster"`
	Tier          string  `json:"tier"`
	Count         int     `json:"count"`
	AvgPaket      float64 `json:"avg_paket"`
	AvgPendapatan float64 `json:"avg_pendapatan"`
	AvgSuccess    float64 `json:"avg_success_rate"`
}

type SegmentResult struct {
	Provinces []SegProvince    `json:"provinces"`
	Clusters  []ClusterSummary `json:"clusters"`
	K         int              `json:"k"`
}

// Segments clusters destination provinces on [volume, revenue, success-rate]
// into 3 tiers and labels each tier by its average revenue ranking.
func (r *AnalyticsRepository) Segments(ctx context.Context) (SegmentResult, error) {
	var prov []SegProvince
	err := r.db.SelectContext(ctx, &prov, `
		SELECT dl.provinsi,
			SUM(f.jumlah_paket) AS paket,
			SUM(f.total_biaya)  AS pendapatan,
			COALESCE(100.0 * SUM(CASE WHEN ds.kategori_status = 'Sukses' THEN f.jumlah_paket ELSE 0 END)
				/ NULLIF(SUM(f.jumlah_paket), 0), 0) AS success_rate
		FROM fact_pengiriman f
		JOIN dim_lokasi dl ON f.tujuan_lokasi_key = dl.lokasi_key
		JOIN dim_status_pengiriman ds ON f.status_key = ds.status_key
		GROUP BY dl.provinsi`)
	if err != nil {
		return SegmentResult{}, err
	}
	res := SegmentResult{Provinces: prov, K: 3}
	if len(prov) == 0 {
		return res, nil
	}

	rows := make([][]float64, len(prov))
	for i, p := range prov {
		rows[i] = []float64{float64(p.Paket), p.Pendapatan, p.SuccessRate}
	}
	k := 3
	if len(prov) < k {
		k = len(prov)
	}
	assign := ml.KMeans(rows, k, 50)

	// Aggregate each cluster.
	sums := make([]ClusterSummary, k)
	for c := range sums {
		sums[c].Cluster = c
	}
	for i, p := range prov {
		c := assign[i]
		sums[c].Count++
		sums[c].AvgPaket += float64(p.Paket)
		sums[c].AvgPendapatan += p.Pendapatan
		sums[c].AvgSuccess += p.SuccessRate
		res.Provinces[i].Cluster = c
	}
	for c := range sums {
		if sums[c].Count > 0 {
			n := float64(sums[c].Count)
			sums[c].AvgPaket /= n
			sums[c].AvgPendapatan /= n
			sums[c].AvgSuccess /= n
		}
	}

	// Label clusters by average revenue: richest → "Bintang", etc.
	rank := make([]int, k)
	for c := range rank {
		rank[c] = c
	}
	sort.Slice(rank, func(a, b int) bool { return sums[rank[a]].AvgPendapatan > sums[rank[b]].AvgPendapatan })
	tiers := []string{"Bintang (Prioritas)", "Berkembang", "Perlu Perhatian"}
	tierOf := make(map[int]string, k)
	for pos, c := range rank {
		t := "Lainnya"
		if pos < len(tiers) {
			t = tiers[pos]
		}
		sums[c].Tier = t
		tierOf[c] = t
	}
	for i := range res.Provinces {
		res.Provinces[i].Tier = tierOf[res.Provinces[i].Cluster]
	}
	sort.Slice(res.Provinces, func(a, b int) bool { return res.Provinces[a].Pendapatan > res.Provinces[b].Pendapatan })
	// Emit clusters in tier order (best first).
	for _, c := range rank {
		res.Clusters = append(res.Clusters, sums[c])
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// 3) Delivery-failure risk scoring (empirical-Bayes smoothed probability)
// ---------------------------------------------------------------------------

type riskRaw struct {
	Layanan  string `db:"layanan"`
	Provinsi string `db:"provinsi"`
	Total    int64  `db:"total_paket"`
	Gagal    int64  `db:"gagal_paket"`
}

type RiskRow struct {
	Layanan   string  `json:"layanan"`
	Provinsi  string  `json:"provinsi"`
	Total     int64   `json:"total_paket"`
	Gagal     int64   `json:"gagal_paket"`
	RawRate   float64 `json:"raw_rate"`   // naive gagal/total %
	RiskScore float64 `json:"risk_score"` // smoothed probability %
	RiskLevel string  `json:"risk_level"` // Tinggi / Sedang / Rendah
}

type RiskResult struct {
	Rows        []RiskRow `json:"rows"`
	GlobalRate  float64   `json:"global_rate"`  // network-wide failure %
	PriorWeight float64   `json:"prior_weight"` // m: strength of the prior
	HighCount   int       `json:"high_count"`
}

// Risk computes a per-route (service × destination province) failure
// probability. Small samples are pulled toward the network-wide failure rate
// using empirical-Bayes shrinkage: score = (gagal + m*global) / (total + m).
// This prevents a route with 1/2 failures from outranking a route with 50/400.
func (r *AnalyticsRepository) Risk(ctx context.Context) (RiskResult, error) {
	var raw []riskRaw
	err := r.db.SelectContext(ctx, &raw, `
		SELECT dly.nama_layanan AS layanan, dl.provinsi,
			SUM(f.jumlah_paket) AS total_paket,
			SUM(CASE WHEN ds.kategori_status = 'Gagal' THEN f.jumlah_paket ELSE 0 END) AS gagal_paket
		FROM fact_pengiriman f
		JOIN dim_layanan dly ON f.layanan_key = dly.layanan_key
		JOIN dim_lokasi dl ON f.tujuan_lokasi_key = dl.lokasi_key
		JOIN dim_status_pengiriman ds ON f.status_key = ds.status_key
		GROUP BY dly.nama_layanan, dl.provinsi
		HAVING SUM(f.jumlah_paket) > 0`)
	if err != nil {
		return RiskResult{}, err
	}
	res := RiskResult{PriorWeight: 80}
	if len(raw) == 0 {
		return res, nil
	}

	var totAll, totGagal int64
	for _, x := range raw {
		totAll += x.Total
		totGagal += x.Gagal
	}
	global := 0.0
	if totAll > 0 {
		global = float64(totGagal) / float64(totAll)
	}
	res.GlobalRate = global * 100
	m := res.PriorWeight

	for _, x := range raw {
		score := (float64(x.Gagal) + m*global) / (float64(x.Total) + m)
		row := RiskRow{
			Layanan:   x.Layanan,
			Provinsi:  x.Provinsi,
			Total:     x.Total,
			Gagal:     x.Gagal,
			RawRate:   100 * float64(x.Gagal) / float64(x.Total),
			RiskScore: score * 100,
		}
		switch {
		case score >= global*1.4:
			row.RiskLevel = "Tinggi"
			res.HighCount++
		case score >= global*0.8:
			row.RiskLevel = "Sedang"
		default:
			row.RiskLevel = "Rendah"
		}
		res.Rows = append(res.Rows, row)
	}
	sort.Slice(res.Rows, func(a, b int) bool { return res.Rows[a].RiskScore > res.Rows[b].RiskScore })
	return res, nil
}
