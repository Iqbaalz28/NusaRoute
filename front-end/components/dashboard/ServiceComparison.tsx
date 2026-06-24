"use client";

import React, { useEffect, useState } from "react";
import { apiGet } from "@/lib/api";
import { ServiceInfo } from "@/lib/types";

interface ServiceComparisonProps {
  onNavigate?: (page: string) => void;
}

export const ServiceComparison: React.FC<ServiceComparisonProps> = ({ onNavigate }) => {
  const [services, setServices] = useState<ServiceInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchServices = async () => {
      try {
        const data = await apiGet<ServiceInfo[]>('/api/v1/pricing/services');
        // Filter out inactive services just in case
        setServices(data.filter(s => s.is_active));
      } catch (err) {
        console.error("Failed to load services:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchServices();
  }, []);

  const getIcon = (code: string) => {
    switch (code) {
      case 'REG': return "🚚";
      case 'YES': return "⚡";
      case 'CARGO': return "📦";
      case 'SAME': return "🚀";
      default: return "📦";
    }
  };

  const isPopular = (code: string) => code === 'YES';

  return (
    <>
      <h2 className="text-center font-heading font-bold text-[2.5rem] mb-2 leading-tight">Pilih kecepatan Anda.</h2>
      <p className="text-center text-muted mb-16">Tingkatan pengiriman fleksibel yang dirancang untuk memenuhi kebutuhan pengiriman spesifik Anda.</p>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-10 max-w-[1200px] mx-auto px-8 mb-24">
        {loading ? (
          [1, 2, 3].map((i) => (
            <div key={i} className="bg-surface rounded-[32px] p-12 h-64 animate-pulse"></div>
          ))
        ) : (
          services.map((s) => {
            const pop = isPopular(s.code);
            return (
              <div className="bg-surface rounded-[32px] p-12 shadow-[0_10px_40px_rgba(45,49,66,0.04)] transition-all duration-300 hover:-translate-y-2 hover:shadow-[0_20px_60px_rgba(45,49,66,0.08)] flex flex-col relative border border-black/5" key={s.code}>
                {pop && (
                  <span className="absolute -top-3 right-6 bg-primary text-white px-4 py-1.5 rounded-xl text-[0.75rem] font-extrabold uppercase tracking-widest shadow-[0_4px_12px_rgba(255,107,74,0.3)]">
                    Populer
                  </span>
                )}
                <div className={`w-16 h-16 rounded-full flex items-center justify-center text-2xl mb-10 ${pop ? 'bg-primary text-white' : 'bg-[#FFF0ED] text-primary'}`}>
                  {getIcon(s.code)}
                </div>
                <h3 className="text-[1.75rem] font-bold mb-5">{s.name}</h3>
                <p className="text-muted text-base mb-10 flex-1 leading-relaxed">{s.description}</p>
                <div className="text-sm font-bold text-gray-400 mb-4">Estimasi: {s.estimated_days}</div>
                <button
                  type="button"
                  onClick={() => onNavigate?.("services")}
                  className="font-heading font-semibold text-primary text-[1.1rem] flex items-center gap-2 group mt-auto cursor-pointer"
                >
                  Pelajari selengkapnya <span className="transition-transform group-hover:translate-x-1">→</span>
                </button>
              </div>
            );
          })
        )}
      </div>
    </>
  );
};
