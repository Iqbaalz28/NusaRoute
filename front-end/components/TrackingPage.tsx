"use client";

import React, { useState, useEffect } from "react";

interface TrackingPageProps {
  initialAwb?: string;
}

export const TrackingPage: React.FC<TrackingPageProps> = ({ initialAwb }) => {
  const [awb, setAwb] = useState(initialAwb || "");
  const [showResult, setShowResult] = useState(!!initialAwb);

  useEffect(() => {
    if (initialAwb) {
      setAwb(initialAwb);
      setShowResult(true);
    }
  }, [initialAwb]);

  const events = [
    { status: "Paket Diterima", detail: "Paket telah diterima oleh penerima", location: "Jakarta Selatan", time: "05 Mei 2026, 14:32 WIB", type: "done" },
    { status: "Sedang Dikirim", detail: "Kurir sedang dalam perjalanan ke alamat penerima", location: "Jakarta Selatan", time: "05 Mei 2026, 13:45 WIB", type: "active" },
    { status: "Tiba di Hub Jakarta", detail: "Paket disortir di Hub Jakarta Pusat (JKT-01)", location: "Hub Jakarta Pusat", time: "05 Mei 2026, 08:12 WIB", type: "done" },
    { status: "Berangkat dari Hub Bandung", detail: "Paket berangkat menuju Hub Jakarta", location: "Hub Bandung", time: "04 Mei 2026, 22:30 WIB", type: "done" },
    { status: "Tiba di Hub Bandung", detail: "Paket tiba dan dipindai di Hub Bandung (BDG-01)", location: "Hub Bandung", time: "04 Mei 2026, 18:15 WIB", type: "done" },
    { status: "Paket Dijemput", detail: "Paket dijemput oleh kurir dari alamat pengirim", location: "Bandung", time: "04 Mei 2026, 15:20 WIB", type: "done" },
  ];

  const handleTrack = () => {
    if (awb.trim()) {
      setShowResult(true);
    }
  };

  return (
    <section className="animate-fade-up min-h-screen">
      {!showResult ? (
        <div className="max-w-[680px] mx-auto px-8 pt-32 pb-16 text-center">
          <div className="text-center mb-16">
            <h1 className="text-[2.5rem] font-bold mb-4">Lacak Paket</h1>
            <p className="text-muted">Masukkan nomor resi untuk melihat status terkini.</p>
          </div>
          <div className="bg-surface rounded-full p-2 shadow-soft max-w-[540px] w-full mx-auto flex gap-2 border border-black/5">
            <input 
              type="text" 
              className="flex-1 border-none outline-none px-6 text-base font-body bg-transparent placeholder:text-muted" 
              placeholder="Masukkan Nomor Resi (misal: NR...)"
              value={awb}
              onChange={(e) => setAwb(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleTrack()}
            />
            <button className="bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover transition-all" onClick={handleTrack}>Lacak →</button>
          </div>
        </div>
      ) : (
        <>
          <div className="pt-32 max-w-[1000px] mx-auto mb-16 px-8">
            <span className="inline-block px-5 py-2 bg-accent text-[#1B4D3E] rounded-full font-semibold text-[0.9rem] mb-6">Terkirim</span>
            <p className="text-muted text-[0.9rem] mb-2">Diperbarui 2 menit yang lalu</p>
            <h1 className="text-[3.5rem] font-bold max-w-[600px] leading-[1.1]">
              Paket Berhasil Terkirim
            </h1>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-12 max-w-[1000px] mx-auto px-8 mb-24">
            <div className="relative pl-12 before:content-[''] before:absolute before:left-[7px] before:top-0 before:bottom-0 before:w-[2px] before:bg-border">
              {events.map((e, i) => (
                <div 
                  className={`relative pb-12 group`} 
                  key={i}
                >
                  <div className={`absolute -left-12 top-1 w-4 h-4 rounded-full border-[3px] border-white z-10 ${e.type === 'active' ? 'bg-primary ring-4 ring-primary/10' : e.type === 'done' ? 'bg-text' : 'bg-border'}`}></div>
                  <div className="bg-surface rounded-[20px] p-6 shadow-soft inline-block min-w-[300px]">
                    <div className="font-bold text-lg mb-1">{e.status}</div>
                    <div className="text-muted text-[0.9rem]">{e.detail}</div>
                    <div className="text-muted text-[0.8rem] mt-2">📍 {e.location} · {e.time}</div>
                  </div>
                </div>
              ))}
            </div>

            <div className="bg-surface rounded-card p-10 shadow-soft h-fit">
              <h2 className="text-xl font-bold mb-8">Detail Pengiriman</h2>
              
              <div className="mb-8">
                <div className="text-[0.85rem] text-muted font-medium mb-2">Nomor Resi</div>
                <div className="text-base font-semibold">#{awb}</div>
              </div>

              <div className="mb-8">
                <div className="text-[0.85rem] text-muted font-medium mb-2">Layanan</div>
                <div className="text-base font-semibold">⚡ Pengiriman Ekspres</div>
              </div>

              <div className="mb-8">
                <div className="text-[0.85rem] text-muted font-medium mb-2">Dari</div>
                <div className="text-base font-semibold">Budi S.</div>
                <div className="text-[0.85rem] text-muted">Bandung, Jawa Barat</div>
              </div>

              <div className="mb-8">
                <div className="text-[0.85rem] text-muted font-medium mb-2">Tujuan</div>
                <div className="text-base font-semibold">Siti Aminah</div>
                <div className="text-[0.85rem] text-muted">Jakarta Selatan, DKI Jakarta</div>
              </div>

              <div className="grid grid-cols-2 gap-4 mb-8">
                <div>
                  <div className="text-[0.85rem] text-muted font-medium mb-2">Berat</div>
                  <div className="text-base font-semibold">2.0 kg</div>
                </div>
                <div>
                  <div className="text-[0.85rem] text-muted font-medium mb-2">Dimensi</div>
                  <div className="text-base font-semibold">30cm x 20cm x 15cm</div>
                </div>
              </div>

              <div className="mt-8 rounded-2xl overflow-hidden h-[150px] bg-border relative">
                <div className="absolute inset-0 flex items-center justify-center text-muted">
                  Tampilan Peta (Pelacakan NusaRoute)
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </section>
  );
};
