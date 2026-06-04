"use client";

import React, { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";
import { Hub, ServiceInfo } from "@/lib/types";

interface HeroProps {
  onTrack: (awb: string) => void;
  onNavigate: (page: string) => void;
}

export const Hero: React.FC<HeroProps> = ({ onTrack, onNavigate }) => {
  const [activeTab, setActiveTab] = useState<"track" | "hub" | "rate">("track");
  const [awb, setAwb] = useState("");
  
  const [hubCount, setHubCount] = useState<number | null>(null);
  const [minPrice, setMinPrice] = useState<number | null>(null);
  const [stats, setStats] = useState<any>(null);

  useEffect(() => {
    // Fetch quick stats independently so failure of one doesn't block the other
    apiGet<Hub[]>('/api/v1/hub/list')
      .then(data => setHubCount(data.length))
      .catch(console.error);

    apiGet<ServiceInfo[]>('/api/v1/pricing/services')
      .then(data => {
        if (data.length > 0) {
          setMinPrice(Math.min(...data.map(s => s.base_fee)));
        }
      })
      .catch(console.error);

    apiGet<any>('/api/v1/dashboard/stats', { requiresAuth: true })
      .then(data => setStats(data))
      .catch(console.error);
  }, []);

  const handleTrack = () => {
    if (awb.trim()) {
      onTrack(awb);
    }
  };

  return (
    <div className="min-h-screen pt-20 px-8 pb-16 max-w-[1200px] mx-auto flex flex-col items-center justify-center relative">
      {/* Background Ambient Glows */}
      <div 
        className="absolute top-1/5 left-1/2 -translate-x-1/2 w-[600px] h-[600px] bg-[radial-gradient(circle,rgba(255,107,74,0.06)_0%,transparent_70%)] blur-[60px] z-0"
      />

      {/* Content */}
      <div className="animate-fade-up relative z-[1] max-w-[800px] w-full text-center">
        <h1 className="font-heading font-extrabold text-[clamp(2.5rem,6vw,4rem)] leading-none mb-4">
          Pengiriman, <span className="text-primary">Jadi Mudah.</span>
        </h1>
        
        <p className="font-body text-[1.125rem] text-muted max-w-[560px] mx-auto mb-10">
          NusaRoute mengirimkan paket Anda ke seluruh Nusantara dengan kecepatan, 
          transparansi, dan perhatian penuh.
        </p>
        
        {/* Compact Action Hub */}
        <div 
          className="glass rounded-[32px] p-6 mx-auto max-w-[640px] border border-white/60 shadow-soft"
        >
          <div className="flex justify-center gap-4 mb-6">
            <button 
              className={`font-body text-[0.9rem] font-medium transition-all duration-300 px-5 py-2 rounded-full ${activeTab === 'track' ? 'text-primary bg-primary/5' : 'text-text opacity-80 hover:opacity-100 hover:text-primary hover:bg-primary/5'}`}
              onClick={() => setActiveTab('track')}
            >
              🔍 Lacak
            </button>
            <button 
              className={`font-body text-[0.9rem] font-medium transition-all duration-300 px-5 py-2 rounded-full ${activeTab === 'hub' ? 'text-primary bg-primary/5' : 'text-text opacity-80 hover:opacity-100 hover:text-primary hover:bg-primary/5'}`}
              onClick={() => setActiveTab('hub')}
            >
              📍 Hub
            </button>
            <button 
              className={`font-body text-[0.9rem] font-medium transition-all duration-300 px-5 py-2 rounded-full ${activeTab === 'rate' ? 'text-primary bg-primary/5' : 'text-text opacity-80 hover:opacity-100 hover:text-primary hover:bg-primary/5'}`}
              onClick={() => setActiveTab('rate')}
            >
              💰 Tarif
            </button>
          </div>

          <div className="flex gap-3 max-w-[560px] mx-auto">
            {activeTab === 'track' && (
              <>
                <input 
                  type="text" 
                  className="flex-1 h-[52px] bg-white rounded-full px-6 text-base font-body outline-none border border-transparent focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all" 
                  placeholder="Masukkan Nomor Resi..."
                  value={awb}
                  onChange={(e) => setAwb(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleTrack()}
                />
                <button className="h-[52px] px-8 bg-primary text-white font-heading font-semibold rounded-pill hover:bg-primary-hover transition-all" onClick={handleTrack}>Lacak</button>
              </>
            )}
            {activeTab === 'hub' && (
              <>
                <div className="flex-1 flex items-center justify-center bg-white rounded-full px-6 h-[52px] text-[0.9rem] text-muted">
                   {hubCount !== null ? `${hubCount} pusat penyortiran di seluruh Indonesia` : 'Memuat data hub...'}
                </div>
                <button className="h-[52px] px-8 bg-primary text-white font-heading font-semibold rounded-pill hover:bg-primary-hover transition-all" onClick={() => onNavigate('hubs')}>Lihat</button>
              </>
            )}
            {activeTab === 'rate' && (
              <>
                <div className="flex-1 flex items-center justify-center bg-white rounded-full px-6 h-[52px] text-[0.9rem] text-muted">
                   {minPrice !== null ? `Tarif mulai dari Rp ${minPrice.toLocaleString('id-ID')}` : 'Memuat tarif...'}
                </div>
                <button className="h-[52px] px-8 bg-primary text-white font-heading font-semibold rounded-pill hover:bg-primary-hover transition-all" onClick={() => onNavigate('services')}>Cek</button>
              </>
            )}
          </div>
        </div>

        {/* Minimal Stats Row (Dynamic Real Data) */}
        <div className="flex justify-center gap-12 mt-12">
          <div>
            <div className="text-xl font-bold">{stats ? stats.total_cities : '...'}</div>
            <div className="text-[0.7rem] text-muted font-semibold uppercase tracking-wider">KOTA</div>
          </div>
          <div className="w-[1px] h-[30px] bg-border"></div>
          <div>
            <div className="text-xl font-bold">{stats ? stats.sla_percentage.toFixed(1) + '%' : '...'}</div>
            <div className="text-[0.7rem] text-muted font-semibold uppercase tracking-wider">SUKSES</div>
          </div>
          <div className="w-[1px] h-[30px] bg-border"></div>
          <div>
            <div className="text-xl font-bold">{stats ? (stats.active_couriers >= 1000 ? (stats.active_couriers/1000).toFixed(1) + 'k' : stats.active_couriers) : '...'}</div>
            <div className="text-[0.7rem] text-muted font-semibold uppercase tracking-wider">KURIR</div>
          </div>
        </div>
      </div>
    </div>
  );
};
