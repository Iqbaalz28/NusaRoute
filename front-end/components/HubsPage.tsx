"use client";

import React, { useEffect, useState } from "react";
import { apiGet } from "@/lib/api";
import { Hub } from "@/lib/types";

export const HubsPage = () => {
  const [hubs, setHubs] = useState<Hub[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHubs = async () => {
      try {
        const data = await apiGet<Hub[]>('/api/v1/hub/list');
        setHubs(data);
        setError(null);
      } catch (err: any) {
        setError(err.message || "Failed to load hubs");
      } finally {
        setLoading(false);
      }
    };
    fetchHubs();
  }, []);

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="mb-16">
        <h1 className="text-[3rem] font-bold mb-4">Jaringan Hub Kami</h1>
        <p className="text-muted text-lg max-w-[600px]">
          NusaRoute memiliki pusat penyortiran strategis untuk memastikan paket Anda sampai lebih cepat.
        </p>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="bg-surface border border-border rounded-card p-8 h-48 animate-pulse"></div>
          ))}
        </div>
      ) : error ? (
        <div className="bg-red-50 text-red-500 p-4 rounded-card text-center border border-red-100">
          <p>{error}</p>
          <button onClick={() => window.location.reload()} className="mt-2 text-sm font-bold underline">Coba Lagi</button>
        </div>
      ) : hubs.length === 0 ? (
        <div className="text-center text-muted py-12">Belum ada hub yang beroperasi.</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {hubs.map((hub) => (
            <div key={hub.id} className="bg-surface border border-border rounded-card p-8 shadow-soft hover:shadow-hover transition-all group">
              <div className="flex items-center gap-2 mb-6">
                <span className={`w-2 h-2 rounded-full ${hub.is_active ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}></span>
                <span className={`text-[0.7rem] font-bold uppercase tracking-wider ${hub.is_active ? 'text-green-600' : 'text-red-600'}`}>
                  {hub.is_active ? "Operasional" : "Non-aktif"}
                </span>
              </div>
              <h3 className="text-xl font-bold mb-2 group-hover:text-primary transition-colors">{hub.name}</h3>
              <p className="text-muted text-[0.9rem] mb-6 flex items-center gap-2">
                <span>📍</span> {hub.city}, {hub.province}
              </p>
              <div className="pt-6 border-t border-border grid grid-cols-2 gap-4">
                <div>
                  <div className="text-[0.7rem] text-muted font-bold uppercase mb-1">Tipe</div>
                  <div className="text-[0.85rem] font-medium">{hub.type}</div>
                </div>
                <div>
                  <div className="text-[0.7rem] text-muted font-bold uppercase mb-1">Kode</div>
                  <div className="text-[0.85rem] font-medium">{hub.code}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
};
