"use client";

import React, { useEffect, useState, useRef } from "react";

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
      setDisplayCount(Math.floor(target * eased));
      if (progress < 1) requestAnimationFrame(update);
    };
    requestAnimationFrame(update);
  };

  return (
    <div className="bg-surface border border-border rounded-[20px] p-6 relative overflow-hidden transition-all duration-300 hover:-translate-y-1 hover:border-primary hover:shadow-hover shadow-soft" ref={elementRef}>
      <div className="text-3xl mb-3">{icon}</div>
      <div className="text-[2rem] font-extrabold tracking-tight text-text leading-none">
        {displayCount.toLocaleString("id-ID")}
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
  return (
    <div className="max-w-[1200px] mx-auto px-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
      <StatItem icon="📦" count={15847} label="Paket Terkirim Hari Ini" trend="+12.5%" trendType="up" />
      <StatItem icon="🛵" count={2341} label="Kurir Aktif" trend="+5.2%" trendType="up" />
      <StatItem icon="🏢" count={128} label="Hub Operasional" trend="Stabil" trendType="neutral" />
      <StatItem icon="⚡" count={98} label="SLA Tercapai" trend="+0.8%" trendType="up" suffix="%" />
    </div>
  );
};
