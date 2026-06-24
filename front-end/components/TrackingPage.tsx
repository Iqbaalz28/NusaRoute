"use client";

import React, { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";

interface TrackingPageProps {
  initialAwb?: string;
}

interface TrackingEvent {
  awb: string;
  status: string;
  location: string;
  detail?: string;
  timestamp: string;
}

export const TrackingPage: React.FC<TrackingPageProps> = ({ initialAwb }) => {
  const [awb, setAwb] = useState(initialAwb || "");
  const [showResult, setShowResult] = useState(false);
  const [events, setEvents] = useState<TrackingEvent[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (initialAwb) {
      setAwb(initialAwb);
      handleTrack(initialAwb);
    }
  }, [initialAwb]);

  const handleTrack = async (trackAwb: string = awb) => {
    if (!trackAwb.trim()) return;
    
    setLoading(true);
    setError("");
    
    try {
      // The API returns a timeline object: { awb, status, events: [...] }.
      const res = await apiGet<{ events?: TrackingEvent[] }>(`/api/v1/tracking/${trackAwb}`);
      const evs = res?.events ?? [];
      if (evs.length > 0) {
        // Sort descending by time (most recent first)
        const sorted = [...evs].sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
        setEvents(sorted);
        setShowResult(true);
      } else {
        setError("Resi tidak ditemukan.");
      }
    } catch (err) {
      console.error(err);
      setError("Gagal melacak resi atau resi tidak ditemukan.");
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string, index: number) => {
    if (index === 0) return 'active';
    return 'done';
  };

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleString('id-ID', { day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' }) + ' WIB';
  };

  const currentStatus = events.length > 0 ? events[0].status : "";

  return (
    <section className="animate-fade-up min-h-screen">
      {!showResult ? (
        <div className="max-w-[680px] mx-auto px-8 pt-32 pb-16 text-center">
          <div className="text-center mb-16">
            <h1 className="text-[2.5rem] font-bold mb-4">Lacak Paket</h1>
            <p className="text-muted">Masukkan nomor resi untuk melihat status terkini.</p>
          </div>
          <div className="bg-surface rounded-full p-2 shadow-soft max-w-[540px] w-full mx-auto flex gap-2 border border-black/5 relative">
            <input 
              type="text" 
              className="flex-1 border-none outline-none px-6 text-base font-body bg-transparent placeholder:text-muted" 
              placeholder="Masukkan Nomor Resi (misal: NSR...)"
              value={awb}
              onChange={(e) => setAwb(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleTrack()}
              disabled={loading}
            />
            <button 
              className={`bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover transition-all ${loading ? 'opacity-70' : ''}`} 
              onClick={() => handleTrack()}
              disabled={loading}
            >
              {loading ? 'Melacak...' : 'Lacak →'}
            </button>
            {error && <div className="absolute -bottom-8 left-0 right-0 text-red-500 text-sm font-semibold">{error}</div>}
          </div>
        </div>
      ) : (
        <>
          <div className="pt-32 max-w-[1000px] mx-auto mb-16 px-8 relative">
            <button onClick={() => setShowResult(false)} className="absolute top-16 left-8 text-muted hover:text-text text-sm font-semibold mb-4">← Kembali Cari Lain</button>
            <span className="inline-block px-5 py-2 bg-accent text-[#1B4D3E] rounded-full font-semibold text-[0.9rem] mb-6">
              {currentStatus}
            </span>
            <p className="text-muted text-[0.9rem] mb-2">Diperbarui {events.length > 0 ? getRelativeTime(events[0].timestamp) : 'baru saja'}</p>
            <h1 className="text-[3.5rem] font-bold max-w-[600px] leading-[1.1]">
              Detail Pelacakan
            </h1>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-[1.5fr_1fr] gap-12 max-w-[1000px] mx-auto px-8 mb-24">
            <div className="relative pl-12 before:content-[''] before:absolute before:left-[7px] before:top-0 before:bottom-0 before:w-[2px] before:bg-border">
              {events.map((e, i) => {
                const type = getStatusColor(e.status, i);
                return (
                  <div className={`relative pb-12 group`} key={i}>
                    <div className={`absolute -left-12 top-1 w-4 h-4 rounded-full border-[3px] border-white z-10 ${type === 'active' ? 'bg-primary ring-4 ring-primary/10' : 'bg-text'}`}></div>
                    <div className="bg-surface rounded-[20px] p-6 shadow-soft inline-block min-w-[300px]">
                      <div className="font-bold text-lg mb-1">{e.status}</div>
                      <div className="text-muted text-[0.9rem]">{e.detail}</div>
                      <div className="text-muted text-[0.8rem] mt-2">📍 {e.location} · {formatDate(e.timestamp)}</div>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="bg-surface rounded-card p-10 shadow-soft h-fit">
              <h2 className="text-xl font-bold mb-8">Informasi Resi</h2>
              
              <div className="mb-8">
                <div className="text-[0.85rem] text-muted font-medium mb-2">Nomor Resi</div>
                <div className="text-base font-semibold">#{awb}</div>
              </div>

              <div className="mb-8">
                <div className="text-[0.85rem] text-muted font-medium mb-2">Layanan</div>
                <div className="text-base font-semibold">Reguler / Sesuai Layanan API</div>
              </div>

              <div className="mt-8 rounded-2xl overflow-hidden h-[150px] bg-border relative">
                <div className="absolute inset-0 flex items-center justify-center text-muted">
                  Tampilan Peta Interaktif
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </section>
  );
};

const getRelativeTime = (dateStr: string) => {
  const diff = Math.floor((new Date().getTime() - new Date(dateStr).getTime()) / 1000);
  if (diff < 60) return `${diff} detik lalu`;
  if (diff < 3600) return `${Math.floor(diff / 60)} menit lalu`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} jam lalu`;
  return `${Math.floor(diff / 86400)} hari lalu`;
};
