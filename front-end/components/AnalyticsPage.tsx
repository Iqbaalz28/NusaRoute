"use client";

import React, { useEffect, useState } from "react";
import { apiGet } from "@/lib/api";

interface Overview { total_paket: number; total_transaksi: number; total_pendapatan: number; avg_durasi_jam: number; success_rate: number; }
interface MonthlyPoint { tahun: number; bulan: number; nama_bulan: string; paket: number; pendapatan: number; avg_durasi_jam: number; }
interface ProvinceRow { provinsi: string; paket: number; pendapatan: number; }
interface ServiceRow { nama_layanan: string; transaksi: number; avg_berat_kg: number; total_ongkir: number; }
interface StatusRow { kategori_status: string; paket: number; }
interface DriverRow { provinsi: string; avg_pendapatan_driver: number; avg_performa_driver: number; }

// ---- ML analytics ----
interface MLPoint { label: string; paket: number; pendapatan: number; is_forecast: boolean; is_anomaly: boolean; }
interface ForecastResult {
  series: MLPoint[]; paket_slope: number; paket_r2: number; rev_slope: number; rev_r2: number;
  trend_label: string; next_paket: number; next_pendapatan: number; growth_pct: number;
  anomaly_count: number; horizon_months: number;
}
interface SegProvince { provinsi: string; paket: number; pendapatan: number; success_rate: number; cluster: number; tier: string; }
interface ClusterSummary { cluster: number; tier: string; count: number; avg_paket: number; avg_pendapatan: number; avg_success_rate: number; }
interface SegmentResult { provinces: SegProvince[]; clusters: ClusterSummary[]; k: number; }
interface RiskRow { layanan: string; provinsi: string; total_paket: number; gagal_paket: number; raw_rate: number; risk_score: number; risk_level: string; }
interface RiskResult { rows: RiskRow[]; global_rate: number; prior_weight: number; high_count: number; }

const TIER_COLOR: Record<string, string> = {
  "Bintang (Prioritas)": "bg-green-500",
  "Berkembang": "bg-amber-500",
  "Perlu Perhatian": "bg-red-500",
};
const RISK_COLOR: Record<string, string> = {
  Tinggi: "bg-red-500 text-red-700 bg-red-50",
  Sedang: "bg-amber-500 text-amber-700 bg-amber-50",
  Rendah: "bg-green-500 text-green-700 bg-green-50",
};

const rp = (n: number) => `Rp ${Math.round(n).toLocaleString("id-ID")}`;
const rpShort = (n: number) => {
  if (n >= 1e9) return `Rp ${(n / 1e9).toFixed(1)} M`;
  if (n >= 1e6) return `Rp ${(n / 1e6).toFixed(1)} jt`;
  return rp(n);
};
const num = (n: number) => Math.round(n).toLocaleString("id-ID");

const STATUS_COLOR: Record<string, string> = {
  Sukses: "bg-green-500",
  Proses: "bg-amber-500",
  Gagal: "bg-red-500",
};

export const AnalyticsPage = () => {
  const [ov, setOv] = useState<Overview | null>(null);
  const [monthly, setMonthly] = useState<MonthlyPoint[]>([]);
  const [provinces, setProvinces] = useState<ProvinceRow[]>([]);
  const [services, setServices] = useState<ServiceRow[]>([]);
  const [statuses, setStatuses] = useState<StatusRow[]>([]);
  const [drivers, setDrivers] = useState<DriverRow[]>([]);
  const [forecast, setForecast] = useState<ForecastResult | null>(null);
  const [segments, setSegments] = useState<SegmentResult | null>(null);
  const [risk, setRisk] = useState<RiskResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const [o, m, p, s, st, d, fc, sg, rk] = await Promise.all([
          apiGet<Overview>("/api/v1/analytics/overview", { requiresAuth: true }),
          apiGet<MonthlyPoint[]>("/api/v1/analytics/monthly", { requiresAuth: true }),
          apiGet<ProvinceRow[]>("/api/v1/analytics/by-province", { requiresAuth: true }),
          apiGet<ServiceRow[]>("/api/v1/analytics/by-service", { requiresAuth: true }),
          apiGet<StatusRow[]>("/api/v1/analytics/by-status", { requiresAuth: true }),
          apiGet<DriverRow[]>("/api/v1/analytics/driver", { requiresAuth: true }),
          apiGet<ForecastResult>("/api/v1/analytics/ml/forecast", { requiresAuth: true }),
          apiGet<SegmentResult>("/api/v1/analytics/ml/segments", { requiresAuth: true }),
          apiGet<RiskResult>("/api/v1/analytics/ml/risk", { requiresAuth: true }),
        ]);
        setOv(o); setMonthly(m || []); setProvinces(p || []);
        setServices(s || []); setStatuses(st || []); setDrivers(d || []);
        setForecast(fc); setSegments(sg); setRisk(rk);
      } catch (err: any) {
        setError(err.message || "Gagal memuat analitik.");
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const card = "bg-surface border border-border rounded-card shadow-soft p-6";
  const maxRev = Math.max(1, ...monthly.map((m) => m.pendapatan));
  const maxProv = Math.max(1, ...provinces.map((p) => p.pendapatan));
  const totalStatus = Math.max(1, statuses.reduce((a, b) => a + b.paket, 0));
  const maxSvc = Math.max(1, ...services.map((s) => s.transaksi));

  if (loading) {
    return (
      <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          {[1, 2, 3, 4].map((i) => <div key={i} className="bg-surface border border-border rounded-card h-28 animate-pulse" />)}
        </div>
        <div className="bg-surface border border-border rounded-card h-72 animate-pulse" />
      </section>
    );
  }

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="mb-8">
        <h1 className="text-[2.5rem] font-bold leading-tight">Analitik — Data Warehouse</h1>
        <p className="text-muted">Agregat lintas-domain dari star schema (<code>fact_pengiriman</code> + dimensi). Periode Jan–Jun 2026.</p>
      </div>

      {error && <div className="text-red-500 text-sm font-bold bg-red-50 p-4 rounded-xl border border-red-100 mb-6">{error}</div>}

      {/* KPI cards */}
      {ov && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <div className={card}>
            <div className="text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1">Total Paket</div>
            <div className="text-3xl font-black">{num(ov.total_paket)}</div>
          </div>
          <div className={card}>
            <div className="text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1">Total Pendapatan</div>
            <div className="text-3xl font-black text-primary">{rpShort(ov.total_pendapatan)}</div>
          </div>
          <div className={card}>
            <div className="text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1">Rata-rata Durasi</div>
            <div className="text-3xl font-black">{ov.avg_durasi_jam.toFixed(1)} <span className="text-lg text-muted">jam</span></div>
          </div>
          <div className={card}>
            <div className="text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1">Tingkat Keberhasilan</div>
            <div className="text-3xl font-black text-green-600">{ov.success_rate.toFixed(1)}%</div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Monthly revenue trend */}
        <div className={card}>
          <h2 className="font-bold mb-1">Tren Pendapatan Bulanan</h2>
          <p className="text-muted text-[0.85rem] mb-5">Total biaya per bulan (SUM)</p>
          <div className="flex items-end gap-3 h-52">
            {monthly.map((m) => (
              <div key={`${m.tahun}-${m.bulan}`} className="flex-1 flex flex-col items-center justify-end gap-2 h-full">
                <div className="text-[0.7rem] font-bold text-muted">{rpShort(m.pendapatan)}</div>
                <div className="w-full bg-primary/80 rounded-t-lg transition-all hover:bg-primary" style={{ height: `${(m.pendapatan / maxRev) * 100}%` }} title={`${m.nama_bulan}: ${rp(m.pendapatan)} • ${num(m.paket)} paket`} />
                <div className="text-[0.7rem] font-semibold text-muted">{m.nama_bulan.slice(0, 3)}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Status distribution */}
        <div className={card}>
          <h2 className="font-bold mb-1">Distribusi Status Pengiriman</h2>
          <p className="text-muted text-[0.85rem] mb-5">Jumlah paket per kategori (COUNT)</p>
          <div className="flex flex-col gap-4 mt-8">
            {statuses.map((s) => (
              <div key={s.kategori_status}>
                <div className="flex justify-between text-[0.85rem] mb-1">
                  <span className="font-semibold">{s.kategori_status}</span>
                  <span className="text-muted">{num(s.paket)} ({((s.paket / totalStatus) * 100).toFixed(1)}%)</span>
                </div>
                <div className="h-3 bg-background rounded-full overflow-hidden">
                  <div className={`h-full rounded-full ${STATUS_COLOR[s.kategori_status] || "bg-gray-400"}`} style={{ width: `${(s.paket / totalStatus) * 100}%` }} />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Revenue by province */}
        <div className={card}>
          <h2 className="font-bold mb-1">Pendapatan per Provinsi Tujuan</h2>
          <p className="text-muted text-[0.85rem] mb-5">SUM total biaya</p>
          <div className="flex flex-col gap-3">
            {provinces.map((p) => (
              <div key={p.provinsi} className="flex items-center gap-3">
                <div className="w-32 text-[0.8rem] font-semibold truncate shrink-0">{p.provinsi}</div>
                <div className="flex-1 h-5 bg-background rounded-full overflow-hidden">
                  <div className="h-full bg-primary/70 rounded-full" style={{ width: `${(p.pendapatan / maxProv) * 100}%` }} />
                </div>
                <div className="w-20 text-right text-[0.78rem] text-muted shrink-0">{rpShort(p.pendapatan)}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Service breakdown */}
        <div className={card}>
          <h2 className="font-bold mb-1">Performa per Tipe Layanan</h2>
          <p className="text-muted text-[0.85rem] mb-5">Transaksi (COUNT), rata-rata berat (AVG), ongkir (SUM)</p>
          <div className="overflow-x-auto">
            <table className="w-full text-[0.85rem]">
              <thead>
                <tr className="text-left text-[0.7rem] uppercase tracking-wider text-muted border-b border-border">
                  <th className="py-2">Layanan</th><th className="py-2 text-right">Transaksi</th><th className="py-2 text-right">Berat</th><th className="py-2 text-right">Ongkir</th>
                </tr>
              </thead>
              <tbody>
                {services.map((s) => (
                  <tr key={s.nama_layanan} className="border-b border-border last:border-none">
                    <td className="py-2 font-semibold">{s.nama_layanan}</td>
                    <td className="py-2 text-right">{num(s.transaksi)}</td>
                    <td className="py-2 text-right text-muted">{s.avg_berat_kg.toFixed(1)} kg</td>
                    <td className="py-2 text-right text-muted">{rpShort(s.total_ongkir)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Driver metrics */}
      <div className={card}>
        <h2 className="font-bold mb-1">Metrik Driver per Provinsi</h2>
        <p className="text-muted text-[0.85rem] mb-5">Rata-rata pendapatan harian &amp; performa (skala 1–5)</p>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {drivers.map((d) => (
            <div key={d.provinsi} className="bg-background rounded-2xl p-4 flex items-center justify-between">
              <div>
                <div className="font-semibold text-[0.9rem]">{d.provinsi}</div>
                <div className="text-muted text-[0.8rem]">{rp(d.avg_pendapatan_driver)}/hari</div>
              </div>
              <div className="text-right">
                <div className="text-amber-500 font-bold">★ {d.avg_performa_driver.toFixed(2)}</div>
                <div className="text-muted text-[0.7rem]">performa</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ===================== ML ANALYTICS ===================== */}
      <div className="mt-12 mb-6">
        <h2 className="text-[1.6rem] font-bold leading-tight flex items-center gap-2">🤖 ML Analytics</h2>
        <p className="text-muted text-[0.9rem]">Lapisan prediktif di atas star schema — peramalan, segmentasi & skor risiko dihitung langsung dari data warehouse (model pure-Go, tanpa layanan eksternal).</p>
      </div>

      {/* Forecast */}
      {forecast && (
        <div className={`${card} mb-6`}>
          <div className="flex flex-wrap items-start justify-between gap-3 mb-1">
            <div>
              <h2 className="font-bold">Peramalan Permintaan &amp; Pendapatan</h2>
              <p className="text-muted text-[0.85rem]">Regresi linear (OLS) atas tren bulanan → proyeksi {forecast.horizon_months} bulan ke depan</p>
            </div>
            <div className="flex gap-2">
              <span className={`text-[0.75rem] font-bold px-3 py-1 rounded-full ${forecast.trend_label === "Naik" ? "bg-green-50 text-green-700" : forecast.trend_label === "Turun" ? "bg-red-50 text-red-700" : "bg-gray-100 text-gray-600"}`}>
                Tren: {forecast.trend_label}
              </span>
            </div>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 my-4">
            <div className="bg-background rounded-2xl p-3">
              <div className="text-[0.7rem] uppercase tracking-wider text-muted">Prediksi Paket (bln+1)</div>
              <div className="text-xl font-black">{num(forecast.next_paket)}</div>
            </div>
            <div className="bg-background rounded-2xl p-3">
              <div className="text-[0.7rem] uppercase tracking-wider text-muted">Prediksi Pendapatan</div>
              <div className="text-xl font-black text-primary">{rpShort(forecast.next_pendapatan)}</div>
            </div>
            <div className="bg-background rounded-2xl p-3">
              <div className="text-[0.7rem] uppercase tracking-wider text-muted">Pertumbuhan MoM</div>
              <div className={`text-xl font-black ${forecast.growth_pct >= 0 ? "text-green-600" : "text-red-600"}`}>{forecast.growth_pct >= 0 ? "+" : ""}{forecast.growth_pct.toFixed(1)}%</div>
            </div>
            <div className="bg-background rounded-2xl p-3">
              <div className="text-[0.7rem] uppercase tracking-wider text-muted">Akurasi Model (R²)</div>
              <div className="text-xl font-black">{(forecast.paket_r2 * 100).toFixed(0)}%</div>
            </div>
          </div>
          {(() => {
            const maxP = Math.max(1, ...forecast.series.map((s) => s.paket));
            return (
              <div className="flex items-end gap-2 h-48">
                {forecast.series.map((s, i) => (
                  <div key={i} className="flex-1 flex flex-col items-center justify-end gap-1 h-full">
                    <div className="text-[0.62rem] font-bold text-muted">{num(s.paket)}</div>
                    <div
                      className={`w-full rounded-t-lg transition-all ${s.is_forecast ? "bg-primary/30 border-2 border-dashed border-primary/60" : s.is_anomaly ? "bg-red-500" : "bg-primary/80"}`}
                      style={{ height: `${(s.paket / maxP) * 100}%` }}
                      title={`${s.label}: ${num(s.paket)} paket${s.is_forecast ? " (prediksi)" : ""}${s.is_anomaly ? " — ANOMALI" : ""}`}
                    />
                    <div className="text-[0.65rem] font-semibold text-muted">{s.label}</div>
                  </div>
                ))}
              </div>
            );
          })()}
          <div className="flex flex-wrap gap-4 mt-4 text-[0.72rem] text-muted">
            <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-primary/80 inline-block" /> Aktual</span>
            <span className="flex items-center gap-1"><span className="w-3 h-3 rounded border-2 border-dashed border-primary/60 bg-primary/30 inline-block" /> Prediksi</span>
            <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-red-500 inline-block" /> Anomali ({forecast.anomaly_count})</span>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Segmentation (K-Means) */}
        {segments && (
          <div className={card}>
            <h2 className="font-bold mb-1">Segmentasi Wilayah (K-Means)</h2>
            <p className="text-muted text-[0.85rem] mb-4">Klaster {segments.k} tier dari [volume, pendapatan, success-rate] provinsi tujuan</p>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 mb-4">
              {segments.clusters.map((c) => (
                <div key={c.cluster} className="bg-background rounded-2xl p-3">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`w-2.5 h-2.5 rounded-full ${TIER_COLOR[c.tier] || "bg-gray-400"}`} />
                    <span className="font-bold text-[0.8rem]">{c.tier}</span>
                  </div>
                  <div className="text-[0.72rem] text-muted">{c.count} provinsi</div>
                  <div className="text-[0.72rem] text-muted">~{rpShort(c.avg_pendapatan)} • {c.avg_success_rate.toFixed(0)}% sukses</div>
                </div>
              ))}
            </div>
            <div className="flex flex-col gap-2 max-h-64 overflow-y-auto">
              {segments.provinces.map((p) => (
                <div key={p.provinsi} className="flex items-center gap-2 text-[0.82rem]">
                  <span className={`w-2 h-2 rounded-full shrink-0 ${TIER_COLOR[p.tier] || "bg-gray-400"}`} />
                  <span className="font-semibold truncate flex-1">{p.provinsi}</span>
                  <span className="text-muted text-[0.74rem] shrink-0">{rpShort(p.pendapatan)} • {p.success_rate.toFixed(0)}%</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Risk scoring */}
        {risk && (
          <div className={card}>
            <h2 className="font-bold mb-1">Skor Risiko Kegagalan per Rute</h2>
            <p className="text-muted text-[0.85rem] mb-1">Probabilitas gagal (layanan × provinsi), dihaluskan empirical-Bayes</p>
            <p className="text-muted text-[0.72rem] mb-4">Rata-rata jaringan: {risk.global_rate.toFixed(1)}% • {risk.high_count} rute risiko tinggi</p>
            <div className="flex flex-col gap-2 max-h-80 overflow-y-auto">
              {risk.rows.slice(0, 12).map((r, i) => {
                const cls = (RISK_COLOR[r.risk_level] || "bg-gray-400 text-gray-600 bg-gray-50").split(" ");
                return (
                  <div key={i} className="flex items-center gap-2 text-[0.8rem]">
                    <div className="w-40 truncate shrink-0">
                      <span className="font-semibold">{r.layanan}</span>
                      <span className="text-muted"> → {r.provinsi}</span>
                    </div>
                    <div className="flex-1 h-4 bg-background rounded-full overflow-hidden">
                      <div className={`h-full rounded-full ${cls[0]}`} style={{ width: `${Math.min(100, r.risk_score * 3)}%` }} />
                    </div>
                    <span className="w-12 text-right font-bold shrink-0">{r.risk_score.toFixed(1)}%</span>
                    <span className={`text-[0.65rem] font-bold px-2 py-0.5 rounded-full shrink-0 ${cls[1]} ${cls[2]}`}>{r.risk_level}</span>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </section>
  );
};
