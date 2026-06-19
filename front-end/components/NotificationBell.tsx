"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import { apiGet, apiPut, API_BASE_URL } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import { NotificationItem } from "@/lib/types";

const timeAgo = (iso: string): string => {
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "baru saja";
  if (m < 60) return `${m} mnt lalu`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h} jam lalu`;
  return `${Math.floor(h / 24)} hari lalu`;
};

export const NotificationBell = () => {
  const { isAuthenticated } = useAuth();
  const [items, setItems] = useState<NotificationItem[]>([]);
  const [open, setOpen] = useState(false);
  const esRef = useRef<EventSource | null>(null);
  const wrapRef = useRef<HTMLDivElement | null>(null);

  const unread = items.filter((n) => !n.is_read).length;

  const load = useCallback(async () => {
    try {
      setItems((await apiGet<NotificationItem[]>("/api/v1/notifications", { requiresAuth: true })) || []);
    } catch { /* non-fatal */ }
  }, []);

  // Initial load + live stream.
  useEffect(() => {
    if (!isAuthenticated) { setItems([]); return; }
    load();

    const token = typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;
    if (token) {
      const es = new EventSource(`${API_BASE_URL}/api/v1/notifications/stream?token=${encodeURIComponent(token)}`);
      es.onmessage = (e) => {
        try {
          const n: NotificationItem = JSON.parse(e.data);
          if (!n?.id) return;
          setItems((prev) => (prev.some((x) => x.id === n.id) ? prev : [n, ...prev]));
        } catch { /* ignore keep-alive frames */ }
      };
      es.onerror = () => { /* EventSource auto-reconnects */ };
      esRef.current = es;
    }
    return () => { esRef.current?.close(); esRef.current = null; };
  }, [isAuthenticated, load]);

  // Close on outside click.
  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, [open]);

  const toggle = async () => {
    const next = !open;
    setOpen(next);
    // Opening with unread items → mark them read (server + locally).
    if (next && unread > 0) {
      setItems((prev) => prev.map((n) => ({ ...n, is_read: true })));
      try { await apiPut("/api/v1/notifications/read-all", {}, { requiresAuth: true }); } catch { /* non-fatal */ }
    }
  };

  if (!isAuthenticated) return null;

  return (
    <div className="relative" ref={wrapRef}>
      <button
        className="relative w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center hover:bg-primary/20 transition-all"
        onClick={toggle}
        aria-label="Notifikasi"
      >
        <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
          <path d="M13.73 21a2 2 0 0 1-3.46 0" />
        </svg>
        {unread > 0 && (
          <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] px-1 rounded-full bg-red-600 text-white text-[0.65rem] font-bold flex items-center justify-center">
            {unread > 9 ? "9+" : unread}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-[340px] max-h-[460px] overflow-y-auto bg-surface border border-border rounded-2xl shadow-hover z-[200] animate-fade-up">
          <div className="px-5 py-3 border-b border-border flex items-center justify-between">
            <span className="font-bold">Notifikasi</span>
            <button className="text-[0.8rem] text-primary font-bold hover:underline" onClick={load}>↻ Muat ulang</button>
          </div>
          {items.length === 0 ? (
            <div className="px-5 py-10 text-center text-muted text-sm">Belum ada notifikasi.</div>
          ) : (
            <ul>
              {items.map((n) => (
                <li key={n.id} className={`px-5 py-3 border-b border-border last:border-none ${!n.is_read ? "bg-primary/5" : ""}`}>
                  <div className="flex items-start gap-2">
                    {!n.is_read && <span className="mt-1.5 w-2 h-2 rounded-full bg-primary shrink-0" />}
                    <div className="min-w-0">
                      <div className="font-bold text-[0.9rem]">{n.title}</div>
                      <div className="text-muted text-[0.82rem] leading-snug">{n.message}</div>
                      <div className="text-muted/70 text-[0.7rem] mt-1">
                        {timeAgo(n.created_at)}{n.awb ? ` • ${n.awb}` : ""}
                      </div>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
};
