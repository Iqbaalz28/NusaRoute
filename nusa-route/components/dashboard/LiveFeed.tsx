"use client";

import React, { useState, useEffect } from "react";

interface Event {
  icon: string;
  text: string;
  time: string;
}

export const LiveFeed = () => {
  const [events, setEvents] = useState<Event[]>([
    { icon: "✅", text: "Paket NR8847291034 telah diterima oleh Siti Aminah di Jakarta Selatan", time: "2 detik lalu" },
    { icon: "🛵", text: "Kurir Rudi Hartono ditugaskan untuk menjemput paket NR7762830192", time: "15 detik lalu" },
    { icon: "📦", text: "Paket NR6621940283 tiba di Hub Surabaya (SBY-01) — scan inbound", time: "32 detik lalu" },
    { icon: "💳", text: "Pembayaran Rp 45.500 dikonfirmasi untuk order NR5529174620", time: "1 menit lalu" },
    { icon: "🔄", text: "Paket NR4418293710 disortir di Hub Bandung (BDG-01)", time: "2 menit lalu" },
    { icon: "⚠️", text: "Pengiriman NR3307182649 gagal — percobaan ke-2 dari 3", time: "3 menit lalu" },
    { icon: "🚚", text: "Paket NR2296071538 berangkat dari Hub Semarang ke Hub Jakarta", time: "5 menit lalu" },
  ]);

  useEffect(() => {
    const newEvents = [
      { icon: "✅", text: "Paket NR9934827162 telah diterima oleh Ahmad di Medan", time: "baru saja" },
      { icon: "📦", text: "Paket NR1123948271 tiba di Hub Makassar (MKS-01)", time: "baru saja" },
      { icon: "🛵", text: "Kurir Diana ditugaskan pickup NR8812736451 di Bandung", time: "baru saja" },
    ];
    let idx = 0;

    const interval = setInterval(() => {
      setEvents((prev) => {
        const updated = [newEvents[idx % newEvents.length], ...prev];
        return updated.slice(0, 7);
      });
      idx++;
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="card__body" id="eventFeed">
      {events.map((e, i) => (
        <div className="event-item" key={i}>
          <span className="event-item__icon">{e.icon}</span>
          <div>
            <div className="event-item__text">{e.text}</div>
            <div className="event-item__time">{e.time}</div>
          </div>
        </div>
      ))}
    </div>
  );
};
