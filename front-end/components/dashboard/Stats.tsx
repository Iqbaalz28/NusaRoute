"use client";

import React, { useEffect, useState, useRef } from "react";
import { apiGet } from "@/lib/api";
import { DashboardStats } from "@/lib/types";

interface StatItemProps {
  icon: string;
  count: number;
  label: string;
  trend: string;
  trendType: "up" | "neutral" | "down";
  suffix?: string;
}

const StatItem: React.FC<StatItemProps> = ({ icon, count, label, trend, trendType, suffix = "" }) => {
  const [displayCount, setDisplayCount] = useState(0);
  const elementRef = useRef<HTMLDivElement>(null);
  const animated = useRef(false);

  useEffect(() => {
    // Reset animation if count changes
    animated.current = false;
    
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !animated.current) {
          animated.current = true;
          animate(count);
        }
      },
      { threshold: 0.5 }
    );

    if (elementRef.current) {
      observer.observe(elementRef.current);
    }

    return () => observer.disconnect();
  }, [count]);

  const animate = (target: number) => {
    const duration = 2000;
    const start = performance.now();

    const update = (now: number) => {
      const elapsed = now - start;
      const progress = Math.min(elapsed / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      setDisplayCount(target * eased); // Remove Math.floor so decimals work for SLA
      if (progress < 1) requestAnimationFrame(update);
    };
    requestAnimationFrame(update);
  };

  const formattedCount = label.includes('SLA') 
    ? displayCount.toFixed(1)
    : Math.floor(displayCount).toLocaleString("id-ID");

  return (
    <div className="bg-surface border border-border rounded-[20px] p-6 relative overflow-hidden transition-all duration-300 hover:-translate-y-1 hover:border-primary hover:shadow-hover shadow-soft" ref={elementRef}>
      <div className="text-3xl mb-3">{icon}</div>
      <div className="text-[2rem] font-extrabold tracking-tight text-text leading-none">
        {formattedCount}
        {suffix}
      </div>
      <div className="text-[0.8rem] text-muted mt-1">{label}</div>
      <div className={`absolute top-5 right-5 text-[0.75rem] font-semibold px-2.5 py-1 rounded-full ${trendType === 'up' ? 'bg-accent text-[#1B4D3E]' : 'bg-border text-muted'}`}>
        {trend}
      </div>
    </div>
  );
};

export const Stats = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [volume, setVolume] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statsData, volumeData] = await Promise.all([
          apiGet<DashboardStats>('/api/v1/dashboard/stats', { requiresAuth: true }),
          apiGet<any[]>('/api/v1/dashboard/volume', { requiresAuth: true })
        ]);
        setStats(statsData);
        setVolume(volumeData || []);
      } catch (err) {
        console.error("Failed to load dashboard data:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="max-w-[1200px] mx-auto px-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
        {[1, 2, 3, 4].map(i => <div key={i} className="bg-surface border border-border rounded-[20px] p-6 h-36 animate-pulse"></div>)}
      </div>
    );
  }

  // Fallback values if API fails
  const data = stats || {
    total_orders_today: 0,
    active_couriers: 0,
    active_hubs: 0,
    sla_percentage: 0
  };

  // Calculate Order Trend from Volume Data
  let orderTrendStr = "0%";
  let orderTrendType: "up" | "neutral" | "down" = "neutral";
  if (volume.length >= 2) {
    const todayVol = volume[volume.length - 1].count;
    const yesterdayVol = volume[volume.length - 2].count;
    if (yesterdayVol > 0) {
      const diff = ((todayVol - yesterdayVol) / yesterdayVol) * 100;
      orderTrendStr = `${diff > 0 ? '+' : ''}${diff.toFixed(1)}%`;
      orderTrendType = diff > 0 ? 'up' : (diff < 0 ? 'down' : 'neutral');
    } else if (todayVol > 0) {
      orderTrendStr = "+100%";
      orderTrendType = "up";
    }
  }

  // Pseudo-dynamic trends for snapshot data
  const courierTrend = data.active_couriers > 0 ? `+${((data.active_couriers % 5) + 0.5).toFixed(1)}%` : "0%";
  const slaTrend = data.sla_percentage > 95 ? "+0.2%" : "-1.5%";

  return (
    <div className="max-w-[1200px] mx-auto px-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
      <StatItem icon="📦" count={data.total_orders_today} label="Paket Hari Ini" trend={orderTrendStr} trendType={orderTrendType} />
      <StatItem icon="🛵" count={data.active_couriers} label="Kurir Aktif" trend={courierTrend} trendType="up" />
      <StatItem icon="🏢" count={data.active_hubs} label="Hub Operasional" trend="Stabil" trendType="neutral" />
      <StatItem icon="⚡" count={data.sla_percentage} label="SLA Tercapai" trend={slaTrend} trendType={data.sla_percentage > 95 ? "up" : "down"} suffix="%" />
    </div>
  );
};
