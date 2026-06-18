"use client";

import React, { useEffect, useState, useCallback } from "react";
import { apiGet, apiPost } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import { Assignment, Order } from "@/lib/types";

const STATUS_LABEL: Record<string, string> = {
  ASSIGNED: "Menunggu Dijemput",
  PICKED_UP: "Dalam Perjalanan",
  COMPLETED: "Selesai",
  NO_SHOW: "Tidak Hadir",
  REASSIGNED: "Dialihkan",
};

const STATUS_STYLE: Record<string, string> = {
  ASSIGNED: "bg-amber-100 text-amber-700",
  PICKED_UP: "bg-blue-100 text-blue-700",
  COMPLETED: "bg-green-100 text-green-700",
  NO_SHOW: "bg-red-100 text-red-700",
  REASSIGNED: "bg-gray-100 text-gray-600",
};

export const CourierPage = () => {
  const { isAuthenticated } = useAuth();
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actingId, setActingId] = useState<string | null>(null);

  const fetchAssignments = useCallback(async () => {
    if (!isAuthenticated) {
      setError("Silakan login terlebih dahulu untuk mengakses halaman Kurir.");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const data = await apiGet<Assignment[]>("/api/v1/dispatch/assignments", { requiresAuth: true });
      setAssignments(data || []);
    } catch (err: any) {
      setError(err.message || "Gagal memuat daftar tugas.");
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  useEffect(() => {
    fetchAssignments();
  }, [fetchAssignments]);

  const handlePickup = async (a: Assignment) => {
    setActingId(a.id);
    setError(null);
    try {
      await apiPost("/api/v1/dispatch/pickup", { order_id: a.order_id }, { requiresAuth: true });
      await fetchAssignments();
    } catch (err: any) {
      setError(err.message || "Gagal menjemput paket.");
    } finally {
      setActingId(null);
    }
  };

  const handleDeliver = async (a: Assignment) => {
    setActingId(a.id);
    setError(null);
    try {
      // Look up the receiver name so the delivery record is meaningful.
      let receiverName = "";
      try {
        const order = await apiGet<Order>(`/api/v1/orders/${a.order_id}`, { requiresAuth: true });
        receiverName = order?.receiver_name || "";
      } catch {
        // non-fatal: deliver without the receiver name
      }
      await apiPost("/api/v1/dispatch/deliver", { order_id: a.order_id, receiver_name: receiverName }, { requiresAuth: true });
      await fetchAssignments();
    } catch (err: any) {
      setError(err.message || "Gagal menyelesaikan pengiriman.");
    } finally {
      setActingId(null);
    }
  };

  const active = assignments.filter((a) => a.status === "ASSIGNED" || a.status === "PICKED_UP");
  const history = assignments.filter((a) => a.status !== "ASSIGNED" && a.status !== "PICKED_UP");

  const renderCard = (a: Assignment) => {
    const busy = actingId === a.id;
    return (
      <div key={a.id} className="bg-surface border border-border rounded-card p-6 shadow-soft flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <span className="font-bold text-lg tracking-wide">{a.awb}</span>
            <span className={`text-[0.7rem] font-bold px-2.5 py-1 rounded-full ${STATUS_STYLE[a.status] || "bg-gray-100 text-gray-600"}`}>
              {STATUS_LABEL[a.status] || a.status}
            </span>
          </div>
          <div className="text-muted text-[0.9rem]">👤 Kurir: {a.courier_name || a.courier_id}</div>
          {a.pickup_address && <div className="text-muted text-[0.85rem]">📍 Jemput: {a.pickup_address}</div>}
        </div>
        <div className="shrink-0">
          {a.status === "ASSIGNED" && (
            <button
              className="bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover transition-all disabled:opacity-50"
              onClick={() => handlePickup(a)}
              disabled={busy}
            >
              {busy ? "Memproses..." : "🛵 Jemput Paket"}
            </button>
          )}
          {a.status === "PICKED_UP" && (
            <button
              className="bg-green-600 text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-green-700 transition-all disabled:opacity-50"
              onClick={() => handleDeliver(a)}
              disabled={busy}
            >
              {busy ? "Memproses..." : "📬 Antar (Selesai)"}
            </button>
          )}
        </div>
      </div>
    );
  };

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[900px] mx-auto min-h-screen">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-[2.5rem] font-bold leading-tight">Konsol Kurir</h1>
          <p className="text-muted">Jemput paket yang ditugaskan, lalu tandai sudah diantar.</p>
        </div>
        <button
          className="text-[0.9rem] font-bold text-primary hover:text-primary-hover transition-colors"
          onClick={fetchAssignments}
          disabled={loading}
        >
          ↻ Muat ulang
        </button>
      </div>

      {error && (
        <div className="text-red-500 text-sm font-bold bg-red-50 p-4 rounded-xl border border-red-100 mb-6">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex flex-col gap-4">
          {[1, 2, 3].map((i) => <div key={i} className="bg-surface border border-border rounded-card h-24 animate-pulse" />)}
        </div>
      ) : (
        <>
          <h2 className="text-lg font-bold mb-4">Tugas Aktif ({active.length})</h2>
          {active.length === 0 ? (
            <div className="border-2 border-dashed border-border rounded-card p-10 text-center text-muted mb-10">
              Tidak ada tugas aktif. Buat pesanan & bayar agar kurir ditugaskan.
            </div>
          ) : (
            <div className="flex flex-col gap-4 mb-10">{active.map(renderCard)}</div>
          )}

          {history.length > 0 && (
            <>
              <h2 className="text-lg font-bold mb-4">Riwayat</h2>
              <div className="flex flex-col gap-4 opacity-80">{history.slice(0, 20).map(renderCard)}</div>
            </>
          )}
        </>
      )}
    </section>
  );
};
