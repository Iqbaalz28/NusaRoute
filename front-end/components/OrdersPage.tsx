"use client";

import React, { useState, useEffect } from "react";
import { QRCodeSVG } from "qrcode.react";
import { apiGet } from "@/lib/api";
import { Order } from "@/lib/types";

interface PaginatedResponse<T> {
  data: T[];
  meta?: {
    page: number;
    per_page: number;
    total: number;
  };
}

export const OrdersPage = () => {
  const [filter, setFilter] = useState("SEMUA");
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<{ awb: string; id: string; statusLabel: string } | null>(null);

  useEffect(() => {
    const fetchOrders = async () => {
      try {
        const res = await apiGet<PaginatedResponse<Order>>('/api/v1/orders', { requiresAuth: true });
        // The API returns PaginatedResponse, but our apiGet returns the payload directly if envelope is used
        // Let's handle both in case apiGet unpacks the 'data' field.
        const orderData = Array.isArray(res) ? res : (res as any)?.data || [];
        setOrders(orderData);
      } catch (err) {
        console.error("Failed to load orders:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchOrders();
  }, []);

  // Use a fallback for destination/service if missing, based on standard format
  const mappedOrders = orders.map(o => {
    return {
      id: o.id.substring(0, 12).toUpperCase(),
      awb: o.awb,
      date: new Date(o.created_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }),
      destination: o.receiver_address?.split(',')[0] || "Unknown", // Simplified destination
      service: o.service_type === 'EXPRESS' ? '⚡ Ekspres' : o.service_type === 'CARGO' ? '📦 Kargo' : '🚚 Standar',
      price: `Rp ${o.total_cost.toLocaleString('id-ID')}`,
      status: o.status,
      statusLabel: o.status === 'CREATED' ? 'Diproses' : 
                   o.status === 'PICKED_UP' ? 'Dijemput' :
                   o.status === 'IN_TRANSIT' ? 'Transit' :
                   o.status === 'DELIVERED' ? 'Terkirim' : o.status
    };
  });

  const filteredOrders = filter === "SEMUA" ? mappedOrders : mappedOrders.filter(o => {
    if (filter === "DIKIRIM") return ['PICKED_UP', 'IN_TRANSIT', 'COURIER_ASSIGNED'].includes(o.status);
    if (filter === "TERKIRIM") return o.status === 'DELIVERED';
    if (filter === "TRANSIT") return o.status === 'IN_TRANSIT';
    return o.status === filter;
  });

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-8 mb-12">
        <div>
          <h1 className="text-[3rem] font-bold mb-4">Riwayat Pesanan</h1>
          <p className="text-muted text-lg">Kelola dan pantau semua pengiriman Anda di satu tempat.</p>
        </div>
        <div className="flex gap-2 p-1.5 bg-border/30 rounded-2xl flex-wrap">
          {["SEMUA", "DIKIRIM", "TERKIRIM", "TRANSIT"].map((f) => (
            <button
              key={f}
              className={`px-5 py-2 rounded-xl text-[0.85rem] font-bold transition-all ${filter === f ? "bg-surface text-text shadow-sm" : "text-muted hover:text-text"}`}
              onClick={() => setFilter(f)}
            >
              {f === "SEMUA" ? "Semua" : f.charAt(0) + f.slice(1).toLowerCase()}
            </button>
          ))}
        </div>
      </div>

      <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden min-h-[400px]">
        {loading ? (
          <div className="flex items-center justify-center h-64 text-muted animate-pulse">Memuat data pesanan...</div>
        ) : mappedOrders.length === 0 ? (
          <div className="flex items-center justify-center h-64 text-muted">Belum ada riwayat pesanan.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-background/50 border-b border-border text-left">
                  <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">ID Pesanan</th>
                  <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Tanggal</th>
                  <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Tujuan</th>
                  <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Layanan</th>
                  <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Biaya</th>
                  <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Status</th>
                  <th className="p-6 text-[0.7rem] font-bold text-muted uppercase tracking-wider">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {filteredOrders.map((o) => (
                  <tr key={o.id} className="border-b border-border last:border-none hover:bg-background/30 transition-colors group">
                    <td className="p-6 font-bold text-[0.95rem]">{o.id}</td>
                    <td className="p-6 text-muted text-[0.9rem]">{o.date}</td>
                    <td className="p-6 text-text text-[0.9rem]">{o.destination}</td>
                    <td className="p-6 text-text text-[0.9rem] font-medium">{o.service}</td>
                    <td className="p-6 text-text text-[0.9rem] font-bold">{o.price}</td>
                    <td className="p-6">
                      <span className={`inline-flex px-3 py-1 rounded-full text-[0.75rem] font-bold uppercase ${
                        o.status === 'DELIVERED' ? 'bg-green-100 text-green-700' :
                        ['PICKED_UP', 'COURIER_ASSIGNED'].includes(o.status) ? 'bg-blue-100 text-blue-700' :
                        o.status === 'IN_TRANSIT' ? 'bg-orange-100 text-orange-700' :
                        'bg-gray-100 text-gray-700'
                      }`}>
                        {o.statusLabel}
                      </span>
                    </td>
                    <td className="p-6">
                      <button
                        className="text-primary font-bold text-[0.85rem] hover:underline"
                        onClick={() => setSelected({ awb: o.awb, id: o.id, statusLabel: o.statusLabel })}
                      >
                        Label QR
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {selected && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-text/40 backdrop-blur-md" onClick={() => setSelected(null)}></div>
          <div className="bg-surface w-full max-w-[400px] rounded-[32px] p-10 relative z-10 shadow-hover animate-fade-up text-center">
            <button onClick={() => setSelected(null)} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>
            <h2 className="text-xl font-bold mb-1">Label Pengiriman</h2>
            <p className="text-muted text-[0.85rem] mb-6">Scan QR ini di Konsol Scan Hub.</p>
            <div className="bg-white p-4 rounded-2xl inline-block border border-border mb-5">
              <QRCodeSVG value={selected.awb} size={180} level="M" />
            </div>
            <div className="text-[0.8rem] text-muted font-medium">Nomor Resi (AWB)</div>
            <div className="text-xl font-black tracking-wide text-primary">{selected.awb}</div>
            <div className="text-[0.85rem] text-muted mt-2">Status: <strong>{selected.statusLabel}</strong></div>
          </div>
        </div>
      )}
    </section>
  );
};
