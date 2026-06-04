"use client";

import React, { useState, useEffect } from "react";
import { apiGet } from "@/lib/api";

interface TrackingEvent {
  awb: string;
  status: string;
  location: string;
  detail?: string;
  timestamp: string;
}

interface Event {
  icon: string;
  text: string;
  time: string;
}

const getIcon = (status: string) => {
  switch (status) {
    case 'DELIVERED': return '✅';
    case 'PICKED_UP': return '📦';
    case 'IN_TRANSIT': return '🚚';
    case 'COURIER_ASSIGNED': return '🛵';
    case 'DELIVERY_FAILED': return '⚠️';
    default: return '🔄';
  }
};

const getRelativeTime = (dateStr: string) => {
  const diff = Math.floor((new Date().getTime() - new Date(dateStr).getTime()) / 1000);
  if (diff < 60) return `${diff} detik lalu`;
  if (diff < 3600) return `${Math.floor(diff / 60)} menit lalu`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} jam lalu`;
  return `${Math.floor(diff / 86400)} hari lalu`;
};

export const LiveFeed = () => {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchEvents = async () => {
    try {
      const data = await apiGet<TrackingEvent[]>('/api/v1/tracking/recent');
      if (data && Array.isArray(data)) {
        const mapped = data.map(d => ({
          icon: getIcon(d.status),
          text: d.detail || `Paket ${d.awb} ${d.status} di ${d.location}`,
          time: getRelativeTime(d.timestamp)
        }));
        setEvents(mapped);
      }
    } catch (err) {
      console.error("Failed to load live feed:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEvents();
    const interval = setInterval(fetchEvents, 10000); // refresh every 10s
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="p-6 h-64 flex flex-col gap-4 animate-pulse">
        {[1, 2, 3, 4].map(i => (
          <div key={i} className="flex gap-3 items-center">
            <div className="w-8 h-8 bg-border rounded-full flex-shrink-0"></div>
            <div className="flex-1 space-y-2">
              <div className="h-4 bg-border rounded w-3/4"></div>
              <div className="h-3 bg-border rounded w-1/4"></div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="p-6" id="eventFeed">
      {events.length === 0 ? (
        <div className="text-center text-muted py-8 text-sm border-2 border-dashed border-border rounded-xl">Belum ada aktivitas tracking terbaru</div>
      ) : (
        events.map((e, i) => (
          <div className="flex gap-3 items-start py-3 border-b border-border last:border-none" key={i}>
            <span className="text-xl flex-shrink-0">{e.icon}</span>
            <div>
              <div className="text-[0.8rem] leading-snug text-text">{e.text}</div>
              <div className="text-[0.7rem] text-muted mt-0.5">{e.time}</div>
            </div>
          </div>
        ))
      )}
    </div>
  );
};
