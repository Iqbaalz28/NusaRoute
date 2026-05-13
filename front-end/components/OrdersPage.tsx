"use client";

import React, { useState } from "react";

export const OrdersPage = () => {
  const [filter, setFilter] = useState("SEMUA");

  const orders = [
    { id: "NR8847291034", date: "12 Mei 2026", destination: "Jakarta Selatan", service: "⚡ Ekspres", price: "Rp 32.000", status: "DIKIRIM", statusLabel: "Sedang Dikirim" },
    { id: "NR7762830192", date: "11 Mei 2026", destination: "Surabaya", service: "🚚 Standar", price: "Rp 18.000", status: "TERKIRIM", statusLabel: "Terkirim" },
    { id: "NR6621940283", date: "10 Mei 2026", destination: "Medan", service: "📦 Kargo", price: "Rp 145.000", status: "TRANSIT", statusLabel: "Transit" },
    { id: "NR5529174620", date: "09 Mei 2026", destination: "Bandung", service: "⚡ Ekspres", price: "Rp 24.000", status: "TERKIRIM", statusLabel: "Terkirim" },
    { id: "NR4418293710", date: "08 Mei 2026", destination: "Makassar", service: "🚚 Standar", price: "Rp 42.000", status: "DIKIRIM", statusLabel: "Sedang Dikirim" },
    { id: "NR3307182649", date: "07 Mei 2026", destination: "Semarang", service: "📦 Kargo", price: "Rp 88.000", status: "PROSES", statusLabel: "Diproses" },
    { id: "NR2296071538", date: "06 Mei 2026", destination: "Denpasar", service: "⚡ Ekspres", price: "Rp 55.000", status: "TERKIRIM", statusLabel: "Terkirim" },
  ];

  const filteredOrders = filter === "SEMUA" ? orders : orders.filter(o => o.status === filter);

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-8 mb-12">
        <div>
          <h1 className="text-[3rem] font-bold mb-4">Riwayat Pesanan</h1>
          <p className="text-muted text-lg">Kelola dan pantau semua pengiriman Anda di satu tempat.</p>
        </div>
        <div className="flex gap-2 p-1.5 bg-border/30 rounded-2xl">
          {["SEMUA", "DIKIRIM", "TERKIRIM", "TRANSIT"].map((f) => (
            <button
              key={f}
              className={`px-5 py-2 rounded-xl text-[0.85rem] font-bold transition-all ${filter === f ? "bg-surface text-text shadow-sm" : "text-muted hover:text-text"}`}
              onClick={() => setFilter(f)}
            >
              {f === "SEMUA" ? "Semua" : f.charAt(0) + f.slice(1).toLowerCase()}
            </button>
          ))}
        </div>
      </div>

      <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full border-collapse">
            <thead>
              <tr className="bg-background/50 border-b border-border text-left">
                <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">ID Pesanan</th>
                <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Tanggal</th>
                <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Tujuan</th>
                <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Layanan</th>
                <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Biaya</th>
                <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Status</th>
                <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {filteredOrders.map((o) => (
                <tr key={o.id} className="border-b border-border last:border-none hover:bg-background/30 transition-colors group">
                  <td className="p-6 font-bold text-[0.95rem]">{o.id}</td>
                  <td className="p-6 text-muted text-[0.9rem]">{o.date}</td>
                  <td className="p-6 text-text text-[0.9rem]">{o.destination}</td>
                  <td className="p-6 text-text text-[0.9rem] font-medium">{o.service}</td>
                  <td className="p-6 text-text text-[0.9rem] font-bold">{o.price}</td>
                  <td className="p-6">
                    <span className={`inline-flex px-3 py-1 rounded-full text-[0.75rem] font-bold uppercase ${
                      o.status === 'TERKIRIM' ? 'bg-green-100 text-green-700' :
                      o.status === 'DIKIRIM' ? 'bg-blue-100 text-blue-700' :
                      o.status === 'TRANSIT' ? 'bg-orange-100 text-orange-700' :
                      'bg-gray-100 text-gray-700'
                    }`}>
                      {o.statusLabel}
                    </span>
                  </td>
                  <td className="p-6">
                    <button className="text-primary font-bold text-[0.85rem] hover:underline">Detail</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
};
