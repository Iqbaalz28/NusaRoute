"use client";

import React, { useState } from "react";

export const ServicesPage = () => {
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [weight, setWeight] = useState(1);
  const [result, setResult] = useState<any>(null);

  const calculate = () => {
    if (origin && destination) {
      setResult([
        { name: "NusaRoute Standar", price: 8000 * weight, time: "3-5 Hari", icon: "🚚", desc: "Paling hemat untuk paket santai." },
        { name: "NusaRoute Ekspres", price: 15000 * weight, time: "1-2 Hari", icon: "⚡", desc: "Cepat dan andal untuk kebutuhan mendesak." },
        { name: "NusaRoute Kargo", price: 5000 * weight + 20000, time: "5-7 Hari", icon: "📦", desc: "Khusus paket besar di atas 10kg." },
      ]);
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
                placeholder="Kota Asal"
                value={origin}
                onChange={(e) => setOrigin(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Tujuan</label>
              <input 
                type="text" 
                className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all" 
                placeholder="Kota Tujuan"
                value={destination}
                onChange={(e) => setDestination(e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Berat (kg)</label>
                <input 
                  type="number" 
                  className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all" 
                  value={weight}
                  onChange={(e) => setWeight(Number(e.target.value))}
                />
              </div>
              <div>
                <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Dimensi (opsional)</label>
                <div className="h-14 bg-background border border-border rounded-2xl px-4 flex items-center justify-center text-muted text-[0.8rem]">
                  P x L x T
                </div>
              </div>
            </div>
            <button className="bg-primary text-white font-heading font-semibold h-16 rounded-pill hover:bg-primary-hover hover:-translate-y-0.5 transition-all text-lg mt-4 shadow-lg shadow-primary/20" onClick={calculate}>
              Hitung Sekarang
            </button>
          </div>
        </div>

        <div className="flex flex-col gap-6">
          {!result ? (
            <div className="h-full min-h-[400px] border-2 border-dashed border-border rounded-card flex flex-col items-center justify-center text-center p-12">
              <div className="text-5xl mb-6 opacity-30">📊</div>
              <h3 className="text-xl font-bold mb-2">Belum ada estimasi</h3>
              <p className="text-muted max-w-[300px]">Masukkan detail pengiriman Anda untuk melihat perbandingan layanan dan harga.</p>
            </div>
          ) : (
            <>
              <h2 className="text-2xl font-bold mb-4">Hasil Estimasi: {origin} → {destination}</h2>
              {result.map((r: any, i: number) => (
                <div key={i} className="bg-surface border border-border rounded-card p-8 shadow-soft flex justify-between items-center group hover:border-primary transition-all">
                  <div className="flex gap-6 items-center">
                    <div className="w-14 h-14 bg-background rounded-2xl flex items-center justify-center text-2xl">{r.icon}</div>
                    <div>
                      <h3 className="text-xl font-bold mb-1">{r.name}</h3>
                      <p className="text-muted text-[0.85rem]">{r.desc}</p>
                      <div className="flex items-center gap-2 mt-2">
                        <span className="text-[0.7rem] font-bold bg-border/50 px-2 py-1 rounded text-muted uppercase tracking-tighter">Estimasi {r.time}</span>
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-2xl font-black text-primary">Rp {r.price.toLocaleString("id-ID")}</div>
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
          )}
        </div>
      </div>
    </section>
  );
};
