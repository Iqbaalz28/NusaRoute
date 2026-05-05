"use client";

import React, { useState } from "react";

export const TrackingPage = () => {
  const [awb, setAwb] = useState("");
  const [showResult, setShowResult] = useState(false);

  const events = [
    { status: "Paket Diterima", detail: "Paket telah diterima oleh Siti Aminah", location: "Jakarta Selatan", time: "05 Mei 2026, 14:32 WIB" },
    { status: "Sedang Diantar", detail: "Kurir Rudi sedang menuju alamat penerima", location: "Jakarta Selatan", time: "05 Mei 2026, 13:45 WIB" },
    { status: "Paket Tiba di Hub Jakarta", detail: "Paket disortir di Hub Jakarta Pusat (JKT-01)", location: "Hub Jakarta Pusat", time: "05 Mei 2026, 08:12 WIB" },
    { status: "Berangkat dari Hub Bandung", detail: "Paket berangkat menuju Hub Jakarta", location: "Hub Bandung", time: "04 Mei 2026, 22:30 WIB" },
    { status: "Paket Tiba di Hub Bandung", detail: "Paket tiba dan discan di Hub Bandung (BDG-01)", location: "Hub Bandung", time: "04 Mei 2026, 18:15 WIB" },
    { status: "Paket Dijemput", detail: "Paket dijemput oleh kurir dari alamat pengirim", location: "Bandung", time: "04 Mei 2026, 15:20 WIB" },
    { status: "Kurir Ditugaskan", detail: "Kurir Ahmad ditugaskan untuk penjemputan", location: "-", time: "04 Mei 2026, 14:55 WIB" },
    { status: "Pembayaran Dikonfirmasi", detail: "Pembayaran Rp 47.500 via VA berhasil", location: "-", time: "04 Mei 2026, 14:30 WIB" },
  ];

  const handleTrack = () => {
    if (awb.trim()) {
      setShowResult(true);
    }
  };

  return (
    <section className="page page--active">
      <div className="page__header">
        <h1 className="page__title">Lacak Paket</h1>
        <p className="page__subtitle">Masukkan nomor resi (AWB) untuk melacak status pengiriman paket Anda secara real-time.</p>
      </div>

      <div className="tracking-search">
        <div className="tracking-search__input-group">
          <input
            type="text"
            className="input input--lg"
            value={awb}
            onChange={(e) => setAwb(e.target.value)}
            placeholder="Masukkan nomor resi, contoh: NR0000000001"
            onKeyDown={(e) => e.key === "Enter" && handleTrack()}
          />
          <button className="btn btn--primary btn--lg" onClick={handleTrack}>
            Lacak
          </button>
        </div>
      </div>

      {showResult && (
        <div className="tracking-result">
          <div className="card card--glass">
            <div className="card__header">
              <h2 className="card__title">📦 {awb}</h2>
              <span className="badge badge--status badge--delivered">Terkirim</span>
            </div>
            <div className="card__body">
              <div className="timeline">
                {events.map((e, i) => (
                  <div className="timeline__item" key={i} style={{ animationDelay: `${i * 0.1}s` }}>
                    <div className="timeline__dot"></div>
                    <div className="timeline__status">{e.status}</div>
                    <div className="timeline__detail">{e.detail}</div>
                    <div className="timeline__time">
                      📍 {e.location} · {e.time}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </section>
  );
};
