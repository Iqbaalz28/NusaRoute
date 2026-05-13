"use client";

import React from "react";

export const HubsPage = () => {
  const hubs = [
    { name: "Hub Jakarta Pusat", location: "Kemayoran, Jakarta Pusat", type: "Pusat Sortir Utama", coverage: "Jabodetabek", status: "Operasional" },
    { name: "Hub Bandung", location: "Soekarno-Hatta, Bandung", type: "Pusat Sortir Regional", coverage: "Jawa Barat", status: "Operasional" },
    { name: "Hub Surabaya", location: "Rungkut, Surabaya", type: "Pusat Sortir Regional", coverage: "Jawa Timur", status: "Operasional" },
    { name: "Hub Medan", location: "Tanjung Morawa, Deli Serdang", type: "Pusat Sortir Regional", coverage: "Sumatera Utara", status: "Operasional" },
    { name: "Hub Makassar", location: "Biringkanaya, Makassar", type: "Pusat Sortir Regional", coverage: "Sulawesi Selatan", status: "Operasional" },
    { name: "Hub Semarang", location: "Gayamsari, Semarang", type: "Pusat Sortir Regional", coverage: "Jawa Tengah", status: "Operasional" },
    { name: "Hub Palembang", location: "Sukabangun, Palembang", type: "Pusat Sortir Regional", coverage: "Sumatera Selatan", status: "Operasional" },
    { name: "Hub Denpasar", location: "Denpasar Selatan, Bali", type: "Pusat Sortir Regional", coverage: "Bali & Nusa Tenggara", status: "Operasional" },
  ];

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="mb-16">
        <h1 className="text-[3rem] font-bold mb-4">Jaringan Hub Kami</h1>
        <p className="text-muted text-lg max-w-[600px]">
          NusaRoute memiliki pusat penyortiran strategis untuk memastikan paket Anda sampai lebih cepat.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {hubs.map((hub, i) => (
          <div key={i} className="bg-surface border border-border rounded-card p-8 shadow-soft hover:shadow-hover transition-all group">
            <div className="flex items-center gap-2 mb-6">
              <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
              <span className="text-[0.7rem] font-bold uppercase tracking-wider text-green-600">{hub.status}</span>
            </div>
            <h3 className="text-xl font-bold mb-2 group-hover:text-primary transition-colors">{hub.name}</h3>
            <p className="text-muted text-[0.9rem] mb-6 flex items-center gap-2">
              <span>📍</span> {hub.location}
            </p>
            <div className="pt-6 border-t border-border grid grid-cols-2 gap-4">
              <div>
                <div className="text-[0.7rem] text-muted font-bold uppercase mb-1">Tipe</div>
                <div className="text-[0.85rem] font-medium">{hub.type}</div>
              </div>
              <div>
                <div className="text-[0.7rem] text-muted font-bold uppercase mb-1">Cakupan</div>
                <div className="text-[0.85rem] font-medium">{hub.coverage}</div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
};
