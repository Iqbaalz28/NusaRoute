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
    <div className="stat-card" ref={elementRef}>
      <div className={`stat-card__icon`}>{icon}</div>
      <div className="stat-card__value">
        {displayCount.toLocaleString("id-ID")}
        {suffix}
      </div>
      <div className="stat-card__label">{label}</div>
      <div className={`stat-card__trend stat-card__trend--${trendType}`}>{trend}</div>
    </div>
  );
};

export const Stats = () => {
  return (
    <div className="stats-grid">
      <StatItem icon="📦" count={15847} label="Paket Terkirim Hari Ini" trend="+12.5%" trendType="up" />
      <StatItem icon="🛵" count={2341} label="Kurir Aktif" trend="+5.2%" trendType="up" />
      <StatItem icon="🏢" count={128} label="Hub Operasional" trend="Stabil" trendType="neutral" />
      <StatItem icon="⚡" count={98} label="SLA Tercapai" trend="+0.8%" trendType="up" suffix="%" />
    </div>
  );
};
