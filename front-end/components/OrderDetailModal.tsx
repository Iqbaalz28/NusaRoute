"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import { apiGet } from "@/lib/api";
import { Order } from "@/lib/types";
import { PaymentPanel } from "./PaymentPanel";
import { ResolutionPanel } from "./ResolutionPanel";

interface TrackEvent { status: string; location?: string; detail?: string; timestamp: string; }
interface Payment { status: string; method: string; amount: number; paid_at?: string; }

const ORDER_STATUS_LABEL: Record<string, string> = {
  PENDING_PAYMENT: "Menunggu Pembayaran",
  READY_FOR_PICKUP: "Siap Dijemput",
  PICKED_UP: "Dijemput",
  IN_TRANSIT: "Dalam Perjalanan",
  OUT_FOR_DELIVERY: "Sedang Diantar",
  DELIVERED: "Terkirim",
  DELIVERY_FAILED: "Gagal Kirim",
  CANCELLED: "Dibatalkan",
};
const statusStyle = (s: string) =>
  s === "DELIVERED" ? "bg-green-100 text-green-700" :
  s === "CANCELLED" || s.includes("FAILED") ? "bg-red-100 text-red-700" :
  s === "IN_TRANSIT" || s === "OUT_FOR_DELIVERY" ? "bg-orange-100 text-orange-700" :
  s === "PENDING_PAYMENT" ? "bg-gray-200 text-gray-600" :
  "bg-blue-100 text-blue-700";

const rp = (n?: number) => `Rp ${Math.round(n || 0).toLocaleString("id-ID")}`;
const fmt = (iso: string) => { try { return new Date(iso).toLocaleString("id-ID", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" }); } catch { return iso; } };

// OrderDetailModal shows the full event-driven state of one order: payment status,
// current shipment status, and the tracking timeline (fed by package.* events). It
// polls every 4s while open so status/timeline updates appear live. If the order is
// still PENDING_PAYMENT it embeds the PaymentPanel so the customer can pay here.
export const OrderDetailModal = ({ orderId, awb, onClose }: { orderId: string; awb: string; onClose: () => void }) => {
  const [order, setOrder] = useState<Order | null>(null);
  const [payment, setPayment] = useState<Payment | null>(null);
  const [events, setEvents] = useState<TrackEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const timer = useRef<ReturnType<typeof setInterval> | null>(null);

  const refresh = useCallback(async () => {
    const [o, p, t] = await Promise.allSettled([
      apiGet<Order>(`/api/v1/orders/${orderId}`, { requiresAuth: true }),
      apiGet<Payment>(`/api/v1/payments/${orderId}`, { requiresAuth: true }),
      apiGet<any>(`/api/v1/tracking/${awb}`),
    ]);
    if (o.status === "fulfilled") setOrder(o.value);
    setPayment(p.status === "fulfilled" ? p.value : null); // 404 = belum ada pembayaran
    if (t.status === "fulfilled" && t.value) {
      const tl = t.value;
      const evs: TrackEvent[] = Array.isArray(tl) ? tl : tl.events || [];
      setEvents(evs.slice().sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()));
    }
    setLoading(false);
  }, [orderId, awb]);

  useEffect(() => {
    refresh();
    timer.current = setInterval(refresh, 4000); // live: reflect event-driven updates
    return () => { if (timer.current) clearInterval(timer.current); };
  }, [refresh]);

  return (
    <div className="fixed inset-0 z-[250] flex items-center justify-center p-6">
      <div className="absolute inset-0 bg-text/50 backdrop-blur-md" onClick={onClose}></div>
      <div className="bg-surface w-full max-w-[560px] max-h-[90vh] overflow-y-auto rounded-[32px] p-8 relative z-10 shadow-hover animate-fade-up">
        <button onClick={onClose} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>

        {loading || !order ? (
          <div className="h-48 flex items-center justify-center text-muted animate-pulse">Memuat detail…</div>
        ) : (
          <>
            <div className="flex items-center gap-3 mb-1 flex-wrap">
              <h2 className="text-2xl font-bold tracking-wide">{order.awb}</h2>
              <span className={`text-[0.75rem] font-bold px-3 py-1 rounded-full ${statusStyle(order.status)}`}>
                {ORDER_STATUS_LABEL[order.status] || order.status}
              </span>
              <span className="inline-flex items-center gap-1 text-[0.7rem] text-muted ml-auto">
                <span className="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse" /> live
              </span>
            </div>
            {order.item_description && <p className="text-muted mb-4">{order.item_description}</p>}

            {/* Pembayaran */}
            <div className="mb-5">
              <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-2">Pembayaran</div>
              {order.status === "PENDING_PAYMENT" ? (
                <PaymentPanel orderId={orderId} amount={order.total_cost} onPaid={refresh} />
              ) : (
                <div className="bg-background rounded-2xl p-4 flex items-center justify-between text-[0.9rem]">
                  <span className="text-muted">{payment ? payment.method : "—"}</span>
                  <span className="font-bold">{rp(order.total_cost)}</span>
                  <span className="text-[0.75rem] font-bold px-3 py-1 rounded-full bg-green-100 text-green-700">
                    {payment?.status === "PAID" ? "Lunas" : "Dibayar"}
                  </span>
                </div>
              )}
            </div>

            {/* Rute */}
            <div className="grid grid-cols-2 gap-3 mb-5 text-[0.85rem]">
              <div className="bg-background rounded-2xl p-4">
                <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-1">📤 Pengirim</div>
                <div className="font-semibold">{order.sender_name}</div>
                {order.sender_city && <div className="text-primary font-bold">📍 {order.sender_city}</div>}
              </div>
              <div className="bg-background rounded-2xl p-4">
                <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-1">📥 Penerima</div>
                <div className="font-semibold">{order.receiver_name}</div>
                {order.receiver_city && <div className="text-primary font-bold">📍 {order.receiver_city}</div>}
              </div>
            </div>

            {/* Timeline */}
            <div>
              <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-3">Riwayat Perjalanan</div>
              {events.length === 0 ? (
                <div className="text-muted text-sm">Belum ada aktivitas pengiriman.</div>
              ) : (
                <ul className="flex flex-col gap-0">
                  {events.map((e, i) => (
                    <li key={i} className="flex gap-3">
                      <div className="flex flex-col items-center">
                        <span className={`w-3 h-3 rounded-full ${i === 0 ? "bg-primary" : "bg-border"}`} />
                        {i < events.length - 1 && <span className="w-0.5 flex-1 bg-border" />}
                      </div>
                      <div className="pb-4">
                        <div className="font-semibold text-[0.85rem]">{ORDER_STATUS_LABEL[e.status] || e.status}</div>
                        <div className="text-muted text-[0.8rem]">{e.detail || e.location || ""}</div>
                        <div className="text-muted/70 text-[0.7rem]">{fmt(e.timestamp)}{e.location ? ` • ${e.location}` : ""}</div>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            {/* Bantuan & klaim asuransi */}
            <div className="mt-6 pt-5 border-t border-border">
              <ResolutionPanel orderId={orderId} awb={order.awb} isInsured={!!order.is_insured} insuredValue={order.insured_value || 0} />
            </div>
          </>
        )}
      </div>
    </div>
  );
};
