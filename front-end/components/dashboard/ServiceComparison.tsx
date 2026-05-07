import React from "react";

export const ServiceComparison = () => {
  const services = [
    { icon: "📦", name: "Reguler (REG)", desc: "2-4 hari kerja", price: "Mulai Rp 8.000" },
    { icon: "⚡", name: "YES", desc: "Yakin Esok Sampai", price: "Mulai Rp 15.000" },
    { icon: "🚀", name: "Same Day", desc: "< 12 jam", price: "Mulai Rp 25.000" },
    { icon: "🏗️", name: "Kargo", desc: "5-7 hari, barang besar", price: "Mulai Rp 5.000" },
  ];

  return (
    <div className="service-comparison">
      {services.map((s, i) => (
        <div className="svc-card" key={i}>
          <div className="svc-card__icon">{s.icon}</div>
          <div className="svc-card__name">{s.name}</div>
          <div className="svc-card__desc">{s.desc}</div>
          <div className="svc-card__price">{s.price}</div>
        </div>
      ))}
    </div>
  );
};
