"use client";

import React, { useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { apiPost } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import { PriceCalculationRequest, PriceCalculationResponse, Order } from "@/lib/types";

type Coords = { lat: number; lon: number };

// Pricing returns short codes (REG/YES/SAME/CARGO); the order service stores a
// human-friendly service_type consistent with the rest of the data.
const mapServiceType = (code: string): string => {
  const m: Record<string, string> = { REG: "REGULAR", YES: "EXPRESS", SAME: "SAMEDAY", CARGO: "CARGO" };
  return m[code] || code;
};

// Supported cities with real coordinates. Using a fixed dropdown list keeps the
// tariff calculator from receiving invalid free-text input and removes the
// dependency on an external geocoder.
type City = { name: string; lat: number; lng: number };
const CITIES: City[] = [
  { name: "Jakarta", lat: -6.2088, lng: 106.8456 },
  { name: "Bogor", lat: -6.5950, lng: 106.8166 },
  { name: "Bandung", lat: -6.9175, lng: 107.6191 },
  { name: "Semarang", lat: -6.9667, lng: 110.4167 },
  { name: "Yogyakarta", lat: -7.7956, lng: 110.3695 },
  { name: "Surabaya", lat: -7.2575, lng: 112.7521 },
  { name: "Malang", lat: -7.9819, lng: 112.6265 },
  { name: "Denpasar (Bali)", lat: -8.6705, lng: 115.2126 },
  { name: "Medan", lat: 3.5952, lng: 98.6722 },
  { name: "Padang", lat: -0.9471, lng: 100.4172 },
  { name: "Pekanbaru", lat: 0.5071, lng: 101.4478 },
  { name: "Palembang", lat: -2.9761, lng: 104.7754 },
  { name: "Batam", lat: 1.0456, lng: 104.0305 },
  { name: "Pontianak", lat: -0.0263, lng: 109.3425 },
  { name: "Banjarmasin", lat: -3.3194, lng: 114.5908 },
  { name: "Balikpapan", lat: -1.2379, lng: 116.8529 },
  { name: "Makassar", lat: -5.1477, lng: 119.4327 },
  { name: "Manado", lat: 1.4748, lng: 124.8421 },
];

export const ServicesPage = () => {
  const { isAuthenticated } = useAuth();

  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [weight, setWeight] = useState(1);
  const [dimension, setDimension] = useState("");
  const [result, setResult] = useState<PriceCalculationResponse[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Geocoded coordinates + parsed dimensions kept from the last calculation,
  // so the order can be created with the exact values that were priced.
  const [originCoords, setOriginCoords] = useState<Coords | null>(null);
  const [destCoords, setDestCoords] = useState<Coords | null>(null);
  const [dims, setDims] = useState({ l: 10, w: 10, h: 10 });

  // Checkout / order placement state
  const [selectedService, setSelectedService] = useState<PriceCalculationResponse | null>(null);
  const [form, setForm] = useState({
    sender_name: "",
    sender_phone: "",
    sender_address: "",
    receiver_name: "",
    receiver_phone: "",
    receiver_address: "",
    item_description: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [orderError, setOrderError] = useState<string | null>(null);
  const [orderResult, setOrderResult] = useState<Order | null>(null);

  const calculate = async () => {
    const originCity = CITIES.find(c => c.name === origin);
    const destCity = CITIES.find(c => c.name === destination);
    if (!originCity || !destCity || weight <= 0) {
      setError("Silakan pilih kota asal, tujuan, dan isi berat dengan benar.");
      return;
    }
    if (origin === destination) {
      setError("Kota asal dan tujuan tidak boleh sama.");
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const oCoords: Coords = { lat: originCity.lat, lon: originCity.lng };
      const dCoords: Coords = { lat: destCity.lat, lon: destCity.lng };

      let l = 10, w = 10, h = 10;
      if (dimension) {
        const ds = dimension.toLowerCase().split('x').map(d => parseFloat(d.trim()));
        if (ds.length === 3 && !ds.some(isNaN)) {
          l = ds[0]; w = ds[1]; h = ds[2];
        }
      }

      const reqBody: PriceCalculationRequest = {
        origin_lat: oCoords.lat,
        origin_lng: oCoords.lon,
        dest_lat: dCoords.lat,
        dest_lng: dCoords.lon,
        weight_kg: weight,
        length_cm: l,
        width_cm: w,
        height_cm: h,
        service_type: "", // Empty for 'compare all'
        is_insured: false,
        insured_value: 0
      };

      // 3. Call backend pricing API
      const data = await apiPost<PriceCalculationResponse[]>('/api/v1/pricing/compare', reqBody);
      setOriginCoords(oCoords);
      setDestCoords(dCoords);
      setDims({ l, w, h });
      setResult(data);
    } catch (err: any) {
      setError(err.message || "Terjadi kesalahan saat menghitung estimasi.");
    } finally {
      setLoading(false);
    }
  };

  const openCheckout = (service: PriceCalculationResponse) => {
    setSelectedService(service);
    setOrderResult(null);
    setOrderError(isAuthenticated ? null : "Silakan login terlebih dahulu untuk membuat pesanan.");
    setForm(f => ({
      ...f,
      sender_address: f.sender_address || origin,
      receiver_address: f.receiver_address || destination,
    }));
  };

  const closeCheckout = () => {
    setSelectedService(null);
    setOrderError(null);
    setSubmitting(false);
  };

  const placeOrder = async () => {
    if (!isAuthenticated) {
      setOrderError("Silakan login terlebih dahulu untuk membuat pesanan.");
      return;
    }
    if (!selectedService || !originCoords || !destCoords) return;
    if (!form.sender_name.trim() || !form.sender_phone.trim() || !form.receiver_name.trim() || !form.receiver_phone.trim()) {
      setOrderError("Lengkapi nama & nomor telepon pengirim dan penerima.");
      return;
    }

    setSubmitting(true);
    setOrderError(null);

    try {
      const body = {
        sender_name: form.sender_name,
        sender_phone: form.sender_phone,
        sender_address: form.sender_address || origin,
        sender_lat: originCoords.lat,
        sender_lng: originCoords.lon,
        receiver_name: form.receiver_name,
        receiver_phone: form.receiver_phone,
        receiver_address: form.receiver_address || destination,
        receiver_lat: destCoords.lat,
        receiver_lng: destCoords.lon,
        item_description: form.item_description || "Barang",
        weight_kg: weight,
        length_cm: dims.l,
        width_cm: dims.w,
        height_cm: dims.h,
        service_type: mapServiceType(selectedService.service_type),
        is_insured: false,
        insured_value: 0,
        shipping_cost: selectedService.total_cost,
        insurance_cost: 0,
        total_cost: selectedService.total_cost,
      };

      const order = await apiPost<Order>('/api/v1/orders', body, { requiresAuth: true });
      setOrderResult(order);
    } catch (err: any) {
      setOrderError(err.message || "Gagal membuat pesanan. Coba lagi.");
    } finally {
      setSubmitting(false);
    }
  };

  const getIcon = (code: string) => {
    switch (code) {
      case 'REG': return "🚚";
      case 'YES': return "⚡";
      case 'CARGO': return "📦";
      case 'SAME': return "🚀";
      default: return "📦";
    }
  };

  const inputClass = "w-full h-12 bg-background border border-border rounded-2xl px-5 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all";

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_1.5fr] gap-12 items-start">
        <div className="bg-surface border border-border rounded-card p-10 shadow-soft">
          <h1 className="text-[2.5rem] font-bold mb-4 leading-tight">Kalkulator Tarif</h1>
          <p className="text-muted mb-10">Dapatkan estimasi biaya pengiriman instan ke seluruh Indonesia.</p>

          <div className="flex flex-col gap-6">
            <div>
              <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Asal</label>
              <select
                className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none cursor-pointer focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all"
                value={origin}
                onChange={(e) => setOrigin(e.target.value)}
              >
                <option value="">Pilih kota asal</option>
                {CITIES.map((c) => (
                  <option key={c.name} value={c.name}>{c.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Tujuan</label>
              <select
                className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none cursor-pointer focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all"
                value={destination}
                onChange={(e) => setDestination(e.target.value)}
              >
                <option value="">Pilih kota tujuan</option>
                {CITIES.map((c) => (
                  <option key={c.name} value={c.name}>{c.name}</option>
                ))}
              </select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Berat (kg)</label>
                <input
                  type="number"
                  min="0.1"
                  step="0.1"
                  className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all"
                  value={weight}
                  onChange={(e) => setWeight(Number(e.target.value))}
                />
              </div>
              <div>
                <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Dimensi (opsional)</label>
                <input
                  type="text"
                  placeholder="P x L x T (cm)"
                  className="w-full h-14 bg-background border border-border rounded-2xl px-4 text-center text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all"
                  value={dimension}
                  onChange={(e) => setDimension(e.target.value)}
                />
              </div>
            </div>
            {error && (
              <div className="text-red-500 text-sm font-bold bg-red-50 p-3 rounded-xl border border-red-100">
                {error}
              </div>
            )}
            <button
              className="bg-primary text-white font-heading font-semibold h-16 rounded-pill hover:bg-primary-hover hover:-translate-y-0.5 transition-all text-lg mt-4 shadow-lg shadow-primary/20 flex items-center justify-center disabled:opacity-50"
              onClick={calculate}
              disabled={loading}
            >
              {loading ? "Menghitung..." : "Hitung Sekarang"}
            </button>
          </div>
        </div>

        <div className="flex flex-col gap-6">
          {!result && !loading ? (
            <div className="h-full min-h-[400px] border-2 border-dashed border-border rounded-card flex flex-col items-center justify-center text-center p-12">
              <div className="text-5xl mb-6 opacity-30">📊</div>
              <h3 className="text-xl font-bold mb-2">Belum ada estimasi</h3>
              <p className="text-muted max-w-[300px]">Masukkan kota asal dan tujuan untuk melihat perbandingan layanan dan harga secara akurat dari koordinat.</p>
            </div>
          ) : loading ? (
            <div className="h-full min-h-[400px] border-2 border-dashed border-border rounded-card flex flex-col items-center justify-center text-center p-12">
              <div className="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin mb-4"></div>
              <p className="text-muted font-bold">Mencari koordinat dan menghitung tarif...</p>
            </div>
          ) : result ? (
            <>
              <h2 className="text-2xl font-bold mb-4">Hasil Estimasi: {origin} → {destination}</h2>
              <p className="text-muted mb-4">Jarak: <strong>{result[0]?.distance_km.toLocaleString('id-ID')} km</strong> | Berat Tagihan: <strong>{result[0]?.chargeable_kg} kg</strong></p>
              {result.map((r, i) => (
                <div key={i} className="bg-surface border border-border rounded-card p-8 shadow-soft flex justify-between items-center group hover:border-primary transition-all">
                  <div className="flex gap-6 items-center">
                    <div className="w-14 h-14 bg-background rounded-2xl flex items-center justify-center text-2xl">{getIcon(r.service_type)}</div>
                    <div>
                      <h3 className="text-xl font-bold mb-1">{r.service_name}</h3>
                      <div className="flex items-center gap-2 mt-2">
                        <span className="text-[0.7rem] font-bold bg-border/50 px-2 py-1 rounded text-muted uppercase tracking-tighter">Estimasi {r.estimated_days}</span>
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-2xl font-black text-primary">Rp {r.total_cost.toLocaleString("id-ID")}</div>
                    <button
                      className="mt-3 text-[0.85rem] font-bold text-text hover:text-primary transition-colors"
                      onClick={() => openCheckout(r)}
                    >
                      Pilih Layanan →
                    </button>
                  </div>
                </div>
              ))}
              <div className="mt-6 p-6 bg-primary/5 rounded-2xl border border-primary/10">
                <p className="text-[0.85rem] text-muted italic">
                  *Harga di atas adalah estimasi dan dapat berubah sewaktu-waktu tergantung pada kondisi operasional dan biaya tambahan lainnya.
                </p>
              </div>
            </>
          ) : null}
        </div>
      </div>

      {/* Checkout / Order placement modal */}
      {selectedService && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-text/40 backdrop-blur-md" onClick={closeCheckout}></div>

          <div className="bg-surface w-full max-w-[560px] max-h-[90vh] overflow-y-auto rounded-[32px] p-10 relative z-10 shadow-hover animate-fade-up">
            {orderResult ? (
              <div className="text-center">
                <div className="text-5xl mb-4">✅</div>
                <h2 className="text-2xl font-bold mb-2">Pesanan Berhasil Dibuat!</h2>
                <p className="text-muted mb-6">Simpan nomor resi untuk melacak paketmu.</p>
                <div className="bg-background rounded-2xl p-6 mb-6 flex flex-col items-center">
                  <div className="bg-white p-3 rounded-xl mb-4 border border-border">
                    <QRCodeSVG value={orderResult.awb} size={148} level="M" />
                  </div>
                  <div className="text-[0.8rem] text-muted font-medium mb-1">Nomor Resi (AWB)</div>
                  <div className="text-2xl font-black tracking-wide text-primary">{orderResult.awb}</div>
                  <div className="text-[0.85rem] text-muted mt-3">Status: <strong>{orderResult.status}</strong></div>
                  <div className="text-[0.75rem] text-muted mt-2 text-center">Scan QR ini di Konsol Scan Hub untuk merekam pergerakan paket antar-hub.</div>
                </div>
                <button
                  className="bg-primary text-white font-heading font-semibold w-full h-14 rounded-pill hover:bg-primary-hover transition-all"
                  onClick={closeCheckout}
                >
                  Selesai
                </button>
              </div>
            ) : (
              <>
                <button onClick={closeCheckout} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>
                <h2 className="text-2xl font-bold mb-1">Detail Pengiriman</h2>
                <div className="flex items-center justify-between bg-background rounded-2xl px-5 py-3 my-5">
                  <span className="font-bold">{selectedService.service_name}</span>
                  <span className="font-black text-primary">Rp {selectedService.total_cost.toLocaleString("id-ID")}</span>
                </div>

                <div className="flex flex-col gap-4">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Nama Pengirim</label>
                      <input className={inputClass} value={form.sender_name} onChange={e => setForm({ ...form, sender_name: e.target.value })} placeholder="Nama pengirim" />
                    </div>
                    <div>
                      <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Telp. Pengirim</label>
                      <input className={inputClass} value={form.sender_phone} onChange={e => setForm({ ...form, sender_phone: e.target.value })} placeholder="08xxx" />
                    </div>
                  </div>
                  <div>
                    <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Alamat Pengirim</label>
                    <input className={inputClass} value={form.sender_address} onChange={e => setForm({ ...form, sender_address: e.target.value })} placeholder="Alamat lengkap asal" />
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Nama Penerima</label>
                      <input className={inputClass} value={form.receiver_name} onChange={e => setForm({ ...form, receiver_name: e.target.value })} placeholder="Nama penerima" />
                    </div>
                    <div>
                      <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Telp. Penerima</label>
                      <input className={inputClass} value={form.receiver_phone} onChange={e => setForm({ ...form, receiver_phone: e.target.value })} placeholder="08xxx" />
                    </div>
                  </div>
                  <div>
                    <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Alamat Penerima</label>
                    <input className={inputClass} value={form.receiver_address} onChange={e => setForm({ ...form, receiver_address: e.target.value })} placeholder="Alamat lengkap tujuan" />
                  </div>
                  <div>
                    <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Deskripsi Barang</label>
                    <input className={inputClass} value={form.item_description} onChange={e => setForm({ ...form, item_description: e.target.value })} placeholder="Misal: Dokumen, pakaian" />
                  </div>
                </div>

                {orderError && (
                  <div className="text-red-500 text-sm font-bold bg-red-50 p-3 rounded-xl border border-red-100 mt-5">
                    {orderError}
                  </div>
                )}

                <button
                  className="bg-primary text-white font-heading font-semibold w-full h-14 rounded-pill hover:bg-primary-hover transition-all mt-6 disabled:opacity-50"
                  onClick={placeOrder}
                  disabled={submitting || !isAuthenticated}
                >
                  {submitting ? "Memproses..." : `Buat Pesanan • Rp ${selectedService.total_cost.toLocaleString("id-ID")}`}
                </button>
              </>
            )}
          </div>
        </div>
      )}
    </section>
  );
};
