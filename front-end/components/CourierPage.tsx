"use client";

import React, { useEffect, useState, useCallback, useRef } from "react";
import { apiGet, apiPost, apiPut, API_BASE_URL } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import { Assignment, Order } from "@/lib/types";
import { haversineKm } from "@/lib/geo";
import { CITIES } from "@/lib/cities";
import { QrScanner } from "./QrScanner";

const STATUS_LABEL: Record<string, string> = {
  OPEN: "Tersedia",
  ASSIGNED: "Diambil — menuju penjemputan",
  PICKED_UP: "Dalam Perjalanan",
  COMPLETED: "Selesai",
  NO_SHOW: "Tidak Hadir",
  REASSIGNED: "Dialihkan",
};

const STATUS_STYLE: Record<string, string> = {
  OPEN: "bg-purple-100 text-purple-700",
  ASSIGNED: "bg-amber-100 text-amber-700",
  PICKED_UP: "bg-blue-100 text-blue-700",
  COMPLETED: "bg-green-100 text-green-700",
  NO_SHOW: "bg-red-100 text-red-700",
  REASSIGNED: "bg-gray-100 text-gray-600",
};

type ScanTarget = { assignment: Assignment; mode: "pickup" | "deliver" };

export const CourierPage = () => {
  const { isAuthenticated, user } = useAuth();
  const [jobs, setJobs] = useState<Assignment[]>([]);
  const [tasks, setTasks] = useState<Assignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);
  const [actingId, setActingId] = useState<string | null>(null);
  const [scan, setScan] = useState<ScanTarget | null>(null);
  const [coords, setCoords] = useState<{ lat: number; lng: number } | null>(null);
  const [simCity, setSimCity] = useState(""); // "" = lokasi GPS asli; nama kota = simulasi
  const [radiusKm, setRadiusKm] = useState<number | null>(75); // null = tampilkan semua area
  const [detail, setDetail] = useState<Order | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const esRef = useRef<EventSource | null>(null);
  const courierIdRef = useRef<string>("");

  const courierName = user?.full_name || user?.email || "Kurir";

  // Persist the courier's location so dispatch/courier-service knows where they are.
  const persistLocation = async (c: { lat: number; lng: number }) => {
    if (!courierIdRef.current) return;
    try { await apiPut(`/api/v1/couriers/location?courier_id=${courierIdRef.current}`, c, { requiresAuth: true }); } catch { /* non-fatal */ }
  };

  // Location simulator: pretend the courier is in a chosen city (testing the
  // nearest-first job board without a real GPS move), or fall back to real GPS.
  const onLocationChange = async (cityName: string) => {
    setSimCity(cityName);
    if (!cityName) {
      // Back to real GPS.
      if (typeof navigator !== "undefined" && navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
          (pos) => { const c = { lat: pos.coords.latitude, lng: pos.coords.longitude }; setCoords(c); persistLocation(c); },
          () => setError("Tidak bisa mengambil lokasi GPS — pilih kota untuk simulasi."),
          { enableHighAccuracy: true, timeout: 8000 }
        );
      }
      return;
    }
    const city = CITIES.find((c) => c.name === cityName);
    if (city) {
      const c = { lat: city.lat, lng: city.lng };
      setCoords(c);
      await persistLocation(c);
      setInfo(`📍 Simulasi lokasi: ${city.name}. Job board diurutkan dari sini.`);
    }
  };

  // Fetch and show the full order/AWB detail for a job or task.
  const openDetail = async (orderId: string) => {
    setDetailOpen(true);
    setDetail(null);
    setDetailLoading(true);
    try {
      setDetail(await apiGet<Order>(`/api/v1/orders/${orderId}`, { requiresAuth: true }));
    } catch (err: any) {
      setError(err.message || "Gagal memuat detail order.");
      setDetailOpen(false);
    } finally {
      setDetailLoading(false);
    }
  };
  const closeDetail = () => { setDetailOpen(false); setDetail(null); };

  const fetchAll = useCallback(async () => {
    try {
      const [openJobs, myTasks] = await Promise.all([
        apiGet<Assignment[]>("/api/v1/dispatch/jobs", { requiresAuth: true }),
        apiGet<Assignment[]>("/api/v1/dispatch/my", { requiresAuth: true }),
      ]);
      setJobs(openJobs || []);
      setTasks(myTasks || []);
      setError(null);
    } catch (err: any) {
      setError(err.message || "Gagal memuat data.");
    } finally {
      setLoading(false);
    }
  }, []);

  // Bootstrap: link courier account, load data, open the real-time job stream.
  useEffect(() => {
    if (!isAuthenticated) {
      setError("Silakan login sebagai kurir untuk mengakses konsol ini.");
      setLoading(false);
      return;
    }

    let cancelled = false;
    (async () => {
      try {
        const c = await apiPost<any>("/api/v1/couriers/ensure", { full_name: courierName, phone: user?.phone, email: user?.email }, { requiresAuth: true });
        courierIdRef.current = c?.id || "";
      } catch {
        // non-fatal: courier row may already exist
      }
      // Browser geolocation → sort jobs by distance + persist courier location.
      // Skipped once the user has picked a simulated city.
      if (!simCity && typeof navigator !== "undefined" && navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
          (pos) => {
            const c = { lat: pos.coords.latitude, lng: pos.coords.longitude };
            setCoords(c);
            persistLocation(c);
          },
          () => { /* izin lokasi ditolak — job tetap tampil, hanya tak terurut jarak */ },
          { enableHighAccuracy: true, timeout: 8000 }
        );
      }
      if (!cancelled) await fetchAll();
    })();

    // Real-time job board via Server-Sent Events (token in query — EventSource
    // can't set Authorization headers).
    const token = typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;
    if (token) {
      const es = new EventSource(`${API_BASE_URL}/api/v1/dispatch/jobs/stream?token=${encodeURIComponent(token)}`);
      es.onmessage = (e) => {
        try {
          const job: Assignment = JSON.parse(e.data);
          if (!job?.order_id) return;
          setJobs((prev) => (prev.some((j) => j.order_id === job.order_id) ? prev : [job, ...prev]));
          setInfo(`📢 Job baru tersedia: ${job.awb}`);
        } catch {
          /* ignore keep-alive comments / malformed frames */
        }
      };
      es.onerror = () => { /* EventSource auto-reconnects */ };
      esRef.current = es;
    }

    return () => {
      cancelled = true;
      esRef.current?.close();
      esRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated]);

  const claim = async (job: Assignment) => {
    setActingId(job.id);
    setError(null);
    setInfo(null);
    try {
      await apiPost("/api/v1/dispatch/claim", { order_id: job.order_id, courier_name: courierName }, { requiresAuth: true });
      setJobs((prev) => prev.filter((j) => j.order_id !== job.order_id));
      await fetchAll();
      setInfo(`✅ Job ${job.awb} berhasil diambil.`);
    } catch (err: any) {
      // Someone else won the race, or another error.
      setError(err.message || "Gagal mengambil job.");
      await fetchAll(); // refresh board to drop the taken job
    } finally {
      setActingId(null);
    }
  };

  const onScanned = async (value: string) => {
    if (!scan) return;
    const target = scan;
    setScan(null);
    const a = target.assignment;

    if (value.trim().toLowerCase() !== a.awb.trim().toLowerCase()) {
      setError(`QR (${value}) tidak cocok dengan paket ini (${a.awb}).`);
      return;
    }

    setActingId(a.id);
    setError(null);
    try {
      if (target.mode === "pickup") {
        await apiPost("/api/v1/dispatch/pickup", { order_id: a.order_id, awb: a.awb }, { requiresAuth: true });
        setInfo(`📦 Paket ${a.awb} dijemput.`);
      } else {
        let receiverName = "";
        try {
          const order = await apiGet<Order>(`/api/v1/orders/${a.order_id}`, { requiresAuth: true });
          receiverName = order?.receiver_name || "";
        } catch { /* non-fatal */ }
        await apiPost("/api/v1/dispatch/deliver", { order_id: a.order_id, awb: a.awb, receiver_name: receiverName }, { requiresAuth: true });
        setInfo(`🏠 Paket ${a.awb} terkirim.`);
      }
      await fetchAll();
    } catch (err: any) {
      setError(err.message || "Aksi gagal.");
    } finally {
      setActingId(null);
    }
  };

  const active = tasks.filter((t) => t.status === "ASSIGNED" || t.status === "PICKED_UP");
  const history = tasks.filter((t) => t.status === "COMPLETED" || t.status === "NO_SHOW" || t.status === "REASSIGNED");

  // Distance from courier to a job's pickup point (km), when location is known.
  const jobDistance = (j: Assignment): number | null => {
    if (!coords || j.pickup_lat == null || j.pickup_lng == null) return null;
    if (j.pickup_lat === 0 && j.pickup_lng === 0) return null;
    return haversineKm(coords.lat, coords.lng, j.pickup_lat, j.pickup_lng);
  };
  // Nearest jobs first when location is available.
  const sortedJobs = coords
    ? [...jobs].sort((a, b) => (jobDistance(a) ?? Infinity) - (jobDistance(b) ?? Infinity))
    : jobs;

  // Only surface jobs in the courier's area. Without a known location (GPS denied
  // and no simulation) or with the radius set to "Semua" we show everything;
  // jobs whose pickup has no coordinates are never hidden.
  const inArea = (j: Assignment): boolean => {
    if (radiusKm == null || !coords) return true;
    const d = jobDistance(j);
    return d == null ? true : d <= radiusKm;
  };
  const visibleJobs = sortedJobs.filter(inArea);
  const hiddenCount = jobs.length - visibleJobs.length;

  const badge = (status: string) => (
    <span className={`text-[0.7rem] font-bold px-2.5 py-1 rounded-full ${STATUS_STYLE[status] || "bg-gray-100 text-gray-600"}`}>
      {STATUS_LABEL[status] || status}
    </span>
  );

  const LEG_LABEL: Record<string, string> = {
    FIRST_MILE: "Jemput → Hub",
    LAST_MILE: "Hub → Antar",
    DIRECT: "Langsung",
  };
  const legChip = (leg?: string) =>
    leg ? <span className="text-[0.7rem] font-bold px-2.5 py-1 rounded-full bg-indigo-100 text-indigo-700">{LEG_LABEL[leg] || leg}</span> : null;

  // Route line: where the courier picks up → where they drop off (per leg).
  const routeLine = (a: Assignment) => (
    <div className="text-muted text-[0.85rem]">
      📍 {a.pickup_address || "-"} <span className="mx-1">→</span> 🎯 {a.dropoff_address || "-"}
    </div>
  );

  // Scan button label depends on leg + stage. Note: the first-mile leg has no
  // courier deliver step — the origin hub operator scans the package in (the
  // receiving party scans), which closes the leg automatically.
  const scanLabel = (a: Assignment, acting: boolean) => {
    if (acting) return "Memproses...";
    if (a.status === "ASSIGNED") return a.leg === "LAST_MILE" ? "📷 Scan Ambil dari Hub" : "📷 Scan Jemput";
    return "📷 Scan Antar"; // deliver step only exists for LAST_MILE / DIRECT
  };

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[900px] mx-auto min-h-screen">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-[2.5rem] font-bold leading-tight">Konsol Kurir</h1>
          <p className="text-muted">Ambil job yang tersedia, jemput &amp; antar paket dengan scan QR.</p>
        </div>
        <button className="text-[0.9rem] font-bold text-primary hover:text-primary-hover transition-colors" onClick={fetchAll} disabled={loading}>
          ↻ Muat ulang
        </button>
      </div>

      {/* Location simulator — pretend to be in a city to test nearest-first jobs. */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-3 mb-6 bg-surface border border-border rounded-2xl p-4">
        <div className="text-[0.85rem] font-bold flex items-center gap-2">
          📍 Lokasi kurir
          <span className="font-normal text-muted">
            {coords ? `(${coords.lat.toFixed(3)}, ${coords.lng.toFixed(3)})` : "belum diketahui"}
          </span>
        </div>
        <select
          className="h-10 bg-background border border-border rounded-xl px-3 text-[0.9rem] font-semibold outline-none focus:border-primary transition-all sm:ml-auto"
          value={simCity}
          onChange={(e) => onLocationChange(e.target.value)}
        >
          <option value="">🛰️ GPS asli (otomatis)</option>
          {CITIES.map((c) => (
            <option key={c.name} value={c.name}>🧪 Simulasi: {c.name}</option>
          ))}
        </select>
        <select
          className="h-10 bg-background border border-border rounded-xl px-3 text-[0.9rem] font-semibold outline-none focus:border-primary transition-all"
          value={radiusKm ?? "all"}
          onChange={(e) => setRadiusKm(e.target.value === "all" ? null : Number(e.target.value))}
          title="Hanya tampilkan job dalam radius ini dari lokasi kurir"
        >
          {[25, 50, 75, 150].map((r) => (
            <option key={r} value={r}>📍 Radius {r} km</option>
          ))}
          <option value="all">🌐 Semua area</option>
        </select>
      </div>

      {error && <div className="text-red-500 text-sm font-bold bg-red-50 p-4 rounded-xl border border-red-100 mb-4">{error}</div>}
      {info && <div className="text-green-700 text-sm font-bold bg-green-50 p-4 rounded-xl border border-green-100 mb-4">{info}</div>}

      {loading ? (
        <div className="flex flex-col gap-4">
          {[1, 2, 3].map((i) => <div key={i} className="bg-surface border border-border rounded-card h-24 animate-pulse" />)}
        </div>
      ) : (
        <>
          {/* Job board */}
          <div className="flex items-center gap-2 mb-4 flex-wrap">
            <h2 className="text-lg font-bold">Job Tersedia ({visibleJobs.length})</h2>
            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 bg-red-600/10 text-red-600 rounded-full text-[0.65rem] font-bold uppercase before:content-[''] before:w-1.5 before:h-1.5 before:rounded-full before:bg-red-600 before:animate-pulse">
              live
            </span>
            {hiddenCount > 0 && (
              <span className="text-[0.8rem] text-muted">{hiddenCount} job di luar area disembunyikan</span>
            )}
          </div>
          {visibleJobs.length === 0 ? (
            <div className="border-2 border-dashed border-border rounded-card p-8 text-center text-muted mb-10">
              {hiddenCount > 0
                ? `Tidak ada job dalam radius ${radiusKm} km. Perluas radius atau ubah lokasi.`
                : "Belum ada job. Job baru akan muncul otomatis di sini."}
            </div>
          ) : (
            <div className="flex flex-col gap-4 mb-10">
              {visibleJobs.map((j) => {
                const dist = jobDistance(j);
                return (
                <div key={j.id} className="bg-surface border border-border rounded-card p-6 shadow-soft flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div>
                    <div className="flex items-center gap-2 mb-1 flex-wrap">
                      <span className="font-bold text-lg tracking-wide">{j.awb}</span>
                      {legChip(j.leg)}
                      {dist != null && <span className="text-[0.7rem] font-bold px-2.5 py-1 rounded-full bg-teal-100 text-teal-700">📏 {dist.toFixed(1)} km</span>}
                    </div>
                    {routeLine(j)}
                    <button onClick={() => openDetail(j.order_id)} className="text-primary text-[0.8rem] font-bold hover:underline mt-1">ℹ️ Lihat detail</button>
                  </div>
                  <button
                    className="shrink-0 bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover transition-all disabled:opacity-50"
                    onClick={() => claim(j)}
                    disabled={actingId === j.id}
                  >
                    {actingId === j.id ? "Mengambil..." : "✋ Ambil Job"}
                  </button>
                </div>
                );
              })}
            </div>
          )}

          {/* My active tasks */}
          <h2 className="text-lg font-bold mb-4">Tugas Saya ({active.length})</h2>
          {active.length === 0 ? (
            <div className="border-2 border-dashed border-border rounded-card p-8 text-center text-muted mb-10">
              Belum ada tugas aktif. Ambil job di atas untuk memulai.
            </div>
          ) : (
            <div className="flex flex-col gap-4 mb-10">
              {active.map((a) => (
                <div key={a.id} className="bg-surface border border-border rounded-card p-6 shadow-soft flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div>
                    <div className="flex items-center gap-2 mb-1 flex-wrap">
                      <span className="font-bold text-lg tracking-wide">{a.awb}</span>
                      {legChip(a.leg)}
                      {badge(a.status)}
                    </div>
                    {routeLine(a)}
                    <button onClick={() => openDetail(a.order_id)} className="text-primary text-[0.8rem] font-bold hover:underline mt-1">ℹ️ Lihat detail</button>
                  </div>
                  <div className="shrink-0">
                    {a.status === "ASSIGNED" && (
                      <button
                        className="bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover transition-all disabled:opacity-50"
                        onClick={() => setScan({ assignment: a, mode: "pickup" })}
                        disabled={actingId === a.id}
                      >
                        {scanLabel(a, actingId === a.id)}
                      </button>
                    )}
                    {/* First-mile: the courier just drops at the hub — the hub
                        operator scans it in to close this leg (no courier scan). */}
                    {a.status === "PICKED_UP" && a.leg === "FIRST_MILE" && (
                      <div className="text-[0.85rem] text-muted font-semibold text-right max-w-[220px]">
                        🏢 Antar ke <span className="text-text">{a.dropoff_address || "hub asal"}</span>. Operator hub akan scan masuk saat paket tiba.
                      </div>
                    )}
                    {a.status === "PICKED_UP" && a.leg !== "FIRST_MILE" && (
                      <button
                        className="bg-green-600 text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-green-700 transition-all disabled:opacity-50"
                        onClick={() => setScan({ assignment: a, mode: "deliver" })}
                        disabled={actingId === a.id}
                      >
                        {scanLabel(a, actingId === a.id)}
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* History */}
          {history.length > 0 && (
            <>
              <h2 className="text-lg font-bold mb-4">Riwayat</h2>
              <div className="flex flex-col gap-4 opacity-80">
                {history.slice(0, 20).map((a) => (
                  <div key={a.id} className="bg-surface border border-border rounded-card p-6 shadow-soft flex items-center justify-between gap-4">
                    <div className="flex items-center gap-3">
                      <span className="font-bold text-lg tracking-wide">{a.awb}</span>
                      {badge(a.status)}
                    </div>
                    <button onClick={() => openDetail(a.order_id)} className="text-primary text-[0.8rem] font-bold hover:underline">ℹ️ Detail</button>
                  </div>
                ))}
              </div>
            </>
          )}
        </>
      )}

      {scan && (
        <QrScanner
          title={scan.mode === "pickup" ? "Scan QR Penjemputan" : "Scan QR Pengantaran"}
          hint={scan.mode === "pickup" ? `Pindai QR/AWB dari HP pengirim untuk paket ${scan.assignment.awb}.` : `Pindai QR pada label paket ${scan.assignment.awb} untuk konfirmasi sampai.`}
          onResult={onScanned}
          onClose={() => setScan(null)}
        />
      )}

      {detailOpen && (
        <div className="fixed inset-0 z-[250] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-text/50 backdrop-blur-md" onClick={closeDetail}></div>
          <div className="bg-surface w-full max-w-[520px] max-h-[90vh] overflow-y-auto rounded-[32px] p-8 relative z-10 shadow-hover animate-fade-up">
            <button onClick={closeDetail} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>
            {detailLoading || !detail ? (
              <div className="h-48 flex items-center justify-center text-muted animate-pulse">Memuat detail...</div>
            ) : (
              <>
                <div className="flex items-center gap-3 mb-1">
                  <h2 className="text-2xl font-bold tracking-wide">{detail.awb}</h2>
                  {badge(detail.status)}
                </div>
                {detail.item_description && <p className="text-muted mb-5">{detail.item_description}</p>}

                {(detail.origin_hub_name || detail.dest_hub_name) && (
                  <div className="bg-background rounded-2xl p-4 mb-4 text-[0.9rem]">
                    <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-1">Rute Hub</div>
                    <div className="font-semibold">
                      {(detail.origin_hub_name || "-")}{detail.origin_hub_code ? ` (${detail.origin_hub_code})` : ""} → {(detail.dest_hub_name || "-")}{detail.dest_hub_code ? ` (${detail.dest_hub_code})` : ""}
                    </div>
                  </div>
                )}

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
                  <div className="bg-background rounded-2xl p-4">
                    <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-1">📤 Pengirim</div>
                    <div className="font-semibold">{detail.sender_name}</div>
                    {detail.sender_city && <div className="text-primary font-bold text-[0.9rem]">📍 {detail.sender_city}</div>}
                    <div className="text-muted text-[0.85rem]">{detail.sender_phone}</div>
                    <div className="text-muted text-[0.8rem] mt-1">{detail.sender_address}</div>
                  </div>
                  <div className="bg-background rounded-2xl p-4">
                    <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-1">📥 Penerima</div>
                    <div className="font-semibold">{detail.receiver_name}</div>
                    {detail.receiver_city && <div className="text-primary font-bold text-[0.9rem]">📍 {detail.receiver_city}</div>}
                    <div className="text-muted text-[0.85rem]">{detail.receiver_phone}</div>
                    <div className="text-muted text-[0.8rem] mt-1">{detail.receiver_address}</div>
                  </div>
                </div>

                <div className="flex flex-wrap gap-x-6 gap-y-2 text-[0.85rem] border-t border-border pt-4">
                  <div><span className="text-muted">Layanan:</span> <strong>{detail.service_type}</strong></div>
                  <div><span className="text-muted">Berat:</span> <strong>{detail.weight_kg} kg</strong></div>
                  {(detail.length_cm || detail.width_cm || detail.height_cm) ? (
                    <div><span className="text-muted">Dimensi:</span> <strong>{detail.length_cm}×{detail.width_cm}×{detail.height_cm} cm</strong></div>
                  ) : null}
                  <div><span className="text-muted">Biaya:</span> <strong>Rp {detail.total_cost?.toLocaleString("id-ID")}</strong></div>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </section>
  );
};
