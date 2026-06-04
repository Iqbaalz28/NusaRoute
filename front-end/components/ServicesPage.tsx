"use client";

import React, { useState } from "react";
import { apiPost } from "@/lib/api";
import { PriceCalculationRequest, PriceCalculationResponse } from "@/lib/types";

export const ServicesPage = () => {
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [weight, setWeight] = useState(1);
  const [dimension, setDimension] = useState("");
  const [result, setResult] = useState<PriceCalculationResponse[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const geocode = async (query: string): Promise<{ lat: number, lon: number }> => {
    const res = await fetch(`https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(query)}&format=json&limit=1`);
    const data = await res.json();
    if (data && data.length > 0) {
      return { lat: parseFloat(data[0].lat), lon: parseFloat(data[0].lon) };
    }
    throw new Error(`Lokasi tidak ditemukan: ${query}`);
  };

  const calculate = async () => {
    if (!origin || !destination || weight <= 0) {
      setError("Silakan isi kota asal, tujuan, dan berat dengan benar.");
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      // 1. Geocode origin
      const originCoords = await geocode(origin);
      // 2. Geocode destination
      const destCoords = await geocode(destination);

      let l = 10, w = 10, h = 10;
      if (dimension) {
        const dims = dimension.toLowerCase().split('x').map(d => parseFloat(d.trim()));
        if (dims.length === 3 && !dims.some(isNaN)) {
          l = dims[0]; w = dims[1]; h = dims[2];
        }
      }

      const reqBody: PriceCalculationRequest = {
        origin_lat: originCoords.lat,
        origin_lng: originCoords.lon,
        dest_lat: destCoords.lat,
        dest_lng: destCoords.lon,
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
      setResult(data);
    } catch (err: any) {
      setError(err.message || "Terjadi kesalahan saat menghitung estimasi.");
    } finally {
      setLoading(false);
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

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_1.5fr] gap-12 items-start">
        <div className="bg-surface border border-border rounded-card p-10 shadow-soft">
          <h1 className="text-[2.5rem] font-bold mb-4 leading-tight">Kalkulator Tarif</h1>
          <p className="text-muted mb-10">Dapatkan estimasi biaya pengiriman instan ke seluruh Indonesia.</p>

          <div className="flex flex-col gap-6">
            <div>
              <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Asal</label>
              <input 
                type="text" 
                className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all" 
                placeholder="Misal: Jakarta Pusat"
                value={origin}
                onChange={(e) => setOrigin(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Tujuan</label>
              <input 
                type="text" 
                className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all" 
                placeholder="Misal: Bandung"
                value={destination}
                onChange={(e) => setDestination(e.target.value)}
              />
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
                    <button className="mt-3 text-[0.85rem] font-bold text-text hover:text-primary transition-colors">Pilih Layanan →</button>
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
    </section>
  );
};
