"use client";

import React, { useState } from "react";

export const OrdersPage = () => {
  const [filter, setFilter] = useState("all");

  const orders = [
    { awb: "NR8847291034", sender: "Budi S.", receiver: "Siti A.", service: "YES", status: "DELIVERED", cost: 47500, date: "05 Mei 2026" },
    { awb: "NR7762830192", sender: "Rina W.", receiver: "Ahmad K.", service: "REG", status: "IN_TRANSIT", cost: 18000, date: "04 Mei 2026" },
    { awb: "NR6621940283", sender: "Deni P.", receiver: "Maya L.", service: "SAME", status: "IN_TRANSIT", cost: 85000, date: "04 Mei 2026" },
    { awb: "NR5529174620", sender: "Farhan R.", receiver: "Lisa N.", service: "REG", status: "PENDING_PAYMENT", cost: 22500, date: "04 Mei 2026" },
    { awb: "NR4418293710", sender: "Galih M.", receiver: "Putri H.", service: "CARGO", status: "IN_TRANSIT", cost: 125000, date: "03 Mei 2026" },
    { awb: "NR3307182649", sender: "Hadi S.", receiver: "Wulan D.", service: "YES", status: "DELIVERY_FAILED", cost: 35000, date: "03 Mei 2026" },
    { awb: "NR2296071538", sender: "Ivan T.", receiver: "Joko P.", service: "REG", status: "DELIVERED", cost: 15500, date: "02 Mei 2026" },
    { awb: "NR1185960427", sender: "Kiki A.", receiver: "Lani B.", service: "SAME", status: "CANCELLED", cost: 92000, date: "01 Mei 2026" },
  ];

  const filteredOrders = filter === "all" ? orders : orders.filter((o) => o.status === filter);

  const getStatusBadge = (status: string) => {
    const map: Record<string, [string, string]> = {
      PENDING_PAYMENT: ["badge--pending", "Menunggu Bayar"],
      IN_TRANSIT: ["badge--transit", "Dalam Perjalanan"],
      DELIVERED: ["badge--delivered", "Terkirim"],
      CANCELLED: ["badge--cancelled", "Dibatalkan"],
      DELIVERY_FAILED: ["badge--cancelled", "Gagal Kirim"],
    };
    const [cls, label] = map[status] || ["", status];
    return <span className={`badge badge--status ${cls}`}>{label}</span>;
  };

  const filters = [
    { id: "all", label: "Semua" },
    { id: "PENDING_PAYMENT", label: "Menunggu Bayar" },
    { id: "IN_TRANSIT", label: "Dalam Perjalanan" },
    { id: "DELIVERED", label: "Terkirim" },
    { id: "CANCELLED", label: "Dibatalkan" },
  ];

  return (
    <section className="page page--active">
      <div className="page__header">
        <h1 className="page__title">Daftar Pesanan</h1>
        <p className="page__subtitle">Kelola semua pesanan pengiriman Anda.</p>
      </div>
      <div className="orders-filter">
        {filters.map((f) => (
          <button
            key={f.id}
            className={`filter-btn ${filter === f.id ? "filter-btn--active" : ""}`}
            onClick={() => setFilter(f.id)}
          >
            {f.label}
          </button>
        ))}
      </div>
      <div className="orders-table-wrapper">
        <table className="table">
          <thead>
            <tr>
              <th>AWB</th>
              <th>Pengirim</th>
              <th>Penerima</th>
              <th>Layanan</th>
              <th>Status</th>
              <th>Total</th>
              <th>Tanggal</th>
            </tr>
          </thead>
          <tbody>
            {filteredOrders.map((o, i) => (
              <tr key={i}>
                <td>
                  <strong>{o.awb}</strong>
                </td>
                <td>{o.sender}</td>
                <td>{o.receiver}</td>
                <td>{o.service}</td>
                <td>{getStatusBadge(o.status)}</td>
                <td>Rp {o.cost.toLocaleString("id-ID")}</td>
                <td>{o.date}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
};
