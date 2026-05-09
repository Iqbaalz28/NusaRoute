"use client";

import React, { useState } from "react";

export const OrdersPage = () => {
  const [filter, setFilter] = useState("all");

  const orders = [
    { awb: "LUM88472910", sender: "Budi S.", receiver: "Siti A.", service: "Express", status: "DELIVERED", cost: 47500, date: "05 Mei 2026" },
    { awb: "LUM77628301", sender: "Rina W.", receiver: "Ahmad K.", service: "Standard", status: "IN_TRANSIT", cost: 18000, date: "04 Mei 2026" },
    { awb: "LUM66219402", sender: "Deni P.", receiver: "Maya L.", service: "Express", status: "IN_TRANSIT", cost: 85000, date: "04 Mei 2026" },
    { awb: "LUM55291746", sender: "Farhan R.", receiver: "Lisa N.", service: "Standard", status: "PENDING", cost: 22500, date: "04 Mei 2026" },
    { awb: "LUM44182937", sender: "Galih M.", receiver: "Putri H.", service: "Cargo", status: "IN_TRANSIT", cost: 125000, date: "03 Mei 2026" },
    { awb: "LUM33071826", sender: "Hadi S.", receiver: "Wulan D.", service: "Express", status: "FAILED", cost: 35000, date: "03 Mei 2026" },
    { awb: "LUM22960715", sender: "Ivan T.", receiver: "Joko P.", service: "Standard", status: "DELIVERED", cost: 15500, date: "02 Mei 2026" },
    { awb: "LUM11859604", sender: "Kiki A.", receiver: "Lani B.", service: "Express", status: "CANCELLED", cost: 92000, date: "01 Mei 2026" },
  ];

  const filteredOrders = filter === "all" ? orders : orders.filter((o) => o.status === filter);

  const getStatusBadge = (status: string) => {
    const map: Record<string, [string, string]> = {
      PENDING: ["status-pill--transit", "Pending"],
      IN_TRANSIT: ["status-pill--transit", "In Transit"],
      DELIVERED: ["status-pill--delivered", "Delivered"],
      CANCELLED: ["status-pill--transit", "Cancelled"],
      FAILED: ["status-pill--transit", "Failed"],
    };
    const [cls, label] = map[status] || ["", status];
    return <span className={`status-pill ${cls}`} style={{ marginBottom: 0, padding: '0.25rem 0.75rem', fontSize: '0.75rem' }}>{label}</span>;
  };

  const filters = [
    { id: "all", label: "All" },
    { id: "IN_TRANSIT", label: "In Transit" },
    { id: "DELIVERED", label: "Delivered" },
    { id: "PENDING", label: "Pending" },
  ];

  return (
    <section className="page page--active" style={{ paddingTop: '8rem' }}>
      <div className="page__header" style={{ maxWidth: '1200px', margin: '0 auto' }}>
        <h1 className="page__title">Shipment History</h1>
        <p className="page__subtitle">View and manage your recent package shipments.</p>
      </div>
      
      <div className="orders-filter" style={{ maxWidth: '1200px', margin: '2rem auto' }}>
        {filters.map((f) => (
          <button
            key={f.id}
            className={`filter-btn ${filter === f.id ? "filter-btn--active" : ""}`}
            onClick={() => setFilter(f.id)}
            style={{ 
              background: filter === f.id ? 'var(--color-primary)' : 'var(--color-surface)',
              color: filter === f.id ? '#fff' : 'var(--color-text)',
              border: 'none', padding: '0.5rem 1.5rem', borderRadius: '99px', marginRight: '0.5rem', cursor: 'pointer',
              fontFamily: 'var(--font-heading)', fontWeight: 600, fontSize: '0.9rem', boxShadow: 'var(--shadow-soft)'
            }}
          >
            {f.label}
          </button>
        ))}
      </div>

      <div className="orders-table-wrapper" style={{ maxWidth: '1200px', margin: '0 auto', background: 'var(--color-surface)', borderRadius: '24px', padding: '1rem', boxShadow: 'var(--shadow-soft)', overflow: 'hidden' }}>
        <table className="table" style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ textAlign: 'left', borderBottom: '1px solid var(--color-border)' }}>
              <th style={{ padding: '1.5rem' }}>AWB</th>
              <th style={{ padding: '1.5rem' }}>Sender</th>
              <th style={{ padding: '1.5rem' }}>Receiver</th>
              <th style={{ padding: '1.5rem' }}>Service</th>
              <th style={{ padding: '1.5rem' }}>Status</th>
              <th style={{ padding: '1.5rem' }}>Cost</th>
              <th style={{ padding: '1.5rem' }}>Date</th>
            </tr>
          </thead>
          <tbody>
            {filteredOrders.map((o, i) => (
              <tr key={i} style={{ borderBottom: i === filteredOrders.length - 1 ? 'none' : '1px solid var(--color-border)' }}>
                <td style={{ padding: '1.5rem' }}><strong>{o.awb}</strong></td>
                <td style={{ padding: '1.5rem' }}>{o.sender}</td>
                <td style={{ padding: '1.5rem' }}>{o.receiver}</td>
                <td style={{ padding: '1.5rem' }}>{o.service}</td>
                <td style={{ padding: '1.5rem' }}>{getStatusBadge(o.status)}</td>
                <td style={{ padding: '1.5rem', fontWeight: 700, color: 'var(--color-primary)' }}>Rp {o.cost.toLocaleString("id-ID")}</td>
                <td style={{ padding: '1.5rem' }}>{o.date}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
};
