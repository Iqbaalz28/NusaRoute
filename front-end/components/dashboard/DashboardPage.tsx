import React from "react";
import { Hero } from "./Hero";
import { Stats } from "./Stats";
import { VolumeChart } from "./VolumeChart";
import { LiveFeed } from "./LiveFeed";
import { ServiceComparison } from "./ServiceComparison";

interface DashboardPageProps {
  onStartShipping: () => void;
  onTrackPackage: (awb: string) => void;
  onNavigate: (page: string) => void;
}

export const DashboardPage: React.FC<DashboardPageProps> = ({ onStartShipping, onTrackPackage, onNavigate }) => {
  return (
    <section className="animate-fade-up pb-24">
      <Hero onTrack={onTrackPackage} onNavigate={onNavigate} />
      
      <Stats />

      <div className="max-w-[1200px] mx-auto px-6 grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-5 mb-8">
        <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden">
          <div className="flex items-center justify-between p-5 border-b border-border">
            <h2 className="font-heading font-bold text-base text-text">📊 Volume Pengiriman (7 Hari)</h2>
          </div>
          <div className="p-6">
            <VolumeChart />
          </div>
        </div>
        
        <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden">
          <div className="flex items-center justify-between p-5 border-b border-border">
            <h2 className="font-heading font-bold text-base text-text">🔴 Aktivitas Terbaru</h2>
            <span className="inline-flex items-center gap-1.5 px-3 py-1 bg-red-600/10 text-red-600 rounded-full text-[0.7rem] font-bold uppercase before:content-[''] before:w-1.5 before:h-1.5 before:rounded-full before:bg-red-600 before:animate-pulse">
              LANGSUNG
            </span>
          </div>
          <LiveFeed />
        </div>
      </div>

      <div className="max-w-[1280px] mx-auto px-6 my-8">
        <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="font-heading font-bold text-base text-text">🚚 Perbandingan Layanan</h2>
          </div>
          <ServiceComparison onNavigate={onNavigate} />
        </div>
      </div>

      <div className="max-w-[1200px] mx-auto px-8 mb-24">
        <div className="bg-[#2D3142] rounded-card p-16 flex flex-col md:flex-row justify-between items-center text-white gap-8">
          <div>
            <h2 className="font-heading font-bold text-[2.5rem] leading-tight mb-4 text-white">Butuh estimasi cepat?</h2>
            <p className="text-white/60 max-w-[400px]">
              Hitung biaya pengiriman secara instan berdasarkan tujuan dan detail paket.
            </p>
          </div>
          <button className="bg-primary text-white font-heading font-semibold px-10 py-4 rounded-pill hover:bg-primary-hover hover:-translate-y-0.5 transition-all text-lg flex items-center" onClick={onStartShipping}>
            <span className="mr-2">📋</span> Hitung Tarif
          </button>
        </div>
      </div>
    </section>
  );
};
