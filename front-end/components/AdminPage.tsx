"use client";

import React, { useCallback, useEffect, useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { apiGet, apiPost, apiPut } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import { Order } from "@/lib/types";

const STATUSES = [
  "PENDING_PAYMENT", "READY_FOR_PICKUP", "PICKED_UP", "IN_TRANSIT",
  "OUT_FOR_DELIVERY", "DELIVERED", "DELIVERY_FAILED", "RETURN_TO_SENDER",
  "CANCELLED", "LOST_SUSPECTED",
];

const PER_PAGE = 15;

const emptyForm = {
  sender_name: "", sender_phone: "", sender_address: "Jakarta",
  receiver_name: "", receiver_phone: "", receiver_address: "Bandung",
  item_description: "", weight_kg: 1, service_type: "REGULAR", total_cost: 25000,
};

export const AdminPage = () => {
  const { isAuthenticated } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [updatingId, setUpdatingId] = useState<string | null>(null);

  const [qrOrder, setQrOrder] = useState<Order | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [creating, setCreating] = useState(false);
  const [createErr, setCreateErr] = useState<string | null>(null);
  const [createdAwb, setCreatedAwb] = useState<string | null>(null);

  const fetchOrders = useCallback(async () => {
    if (!isAuthenticated) { setError("Silakan login untuk mengakses Admin."); setLoading(false); return; }
    setLoading(true);
    setError(null);
    try {
      const q = `?page=${page}&per_page=${PER_PAGE}${search ? `&search=${encodeURIComponent(search)}` : ""}`;
      const data = await apiGet<Order[]>(`/api/v1/orders/all${q}`, { requiresAuth: true });
      setOrders(data || []);
    } catch (err: any) {
      setError(err.message || "Gagal memuat pesanan.");
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated, page, search]);

  useEffect(() => { fetchOrders(); }, [fetchOrders]);

  const changeStatus = async (o: Order, status: string) => {
    if (status === o.status) return;
    setUpdatingId(o.id);
    setError(null);
    try {
      await apiPut("/api/v1/orders/status", { order_id: o.id, status }, { requiresAuth: true });
      setOrders((prev) => prev.map((x) => (x.id === o.id ? { ...x, status } : x)));
    } catch (err: any) {
      setError(err.message || "Gagal mengubah status.");
    } finally {
      setUpdatingId(null);
    }
  };

  const cancelOrder = async (o: Order) => {
    await changeStatus(o, "CANCELLED");
  };

  const submitCreate = async () => {
    if (!form.sender_name.trim() || !form.receiver_name.trim()) {
      setCreateErr("Nama pengirim dan penerima wajib diisi.");
      return;
    }
    setCreating(true);
    setCreateErr(null);
    try {
      const body = {
        ...form,
        sender_lat: -6.2, sender_lng: 106.8,
        receiver_lat: -6.9175, receiver_lng: 107.6191,
        length_cm: 10, width_cm: 10, height_cm: 10,
        is_insured: false, insured_value: 0,
        shipping_cost: form.total_cost, insurance_cost: 0,
      };
      const order = await apiPost<Order>("/api/v1/orders", body, { requiresAuth: true });
      setCreatedAwb(order.awb);
      setForm(emptyForm);
      fetchOrders();
    } catch (err: any) {
      setCreateErr(err.message || "Gagal membuat pesanan.");
    } finally {
      setCreating(false);
    }
  };

  const badge = (status: string) =>
    status === "DELIVERED" ? "bg-green-100 text-green-700" :
    status === "CANCELLED" || status.includes("FAILED") || status.includes("LOST") ? "bg-red-100 text-red-700" :
    ["PICKED_UP", "COURIER_ASSIGNED", "READY_FOR_PICKUP"].includes(status) ? "bg-blue-100 text-blue-700" :
    status === "IN_TRANSIT" || status === "OUT_FOR_DELIVERY" ? "bg-orange-100 text-orange-700" :
    "bg-gray-100 text-gray-700";

  const inputClass = "w-full h-11 bg-background border border-border rounded-xl px-4 text-base outline-none focus:border-primary transition-all";

  return (
    <section className="animate-fade-up pt-32 px-8 pb-24 max-w-[1200px] mx-auto min-h-screen">
      <div className="flex flex-col md:flex-row justify-between md:items-center gap-6 mb-8">
        <div>
          <h1 className="text-[2.5rem] font-bold leading-tight">Admin — Kelola Pesanan</h1>
          <p className="text-muted">Lihat semua pesanan, buat baru (AWB otomatis), ubah status, atau batalkan.</p>
        </div>
        <button
          className="bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover transition-all shrink-0"
          onClick={() => { setCreateOpen(true); setCreatedAwb(null); setCreateErr(null); }}
        >
          + Buat Pesanan
        </button>
      </div>

      <div className="flex gap-2 mb-6">
        <input
          className={inputClass}
          placeholder="Cari AWB / nama pengirim / penerima..."
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={(e) => { if (e.key === "Enter") { setPage(1); setSearch(searchInput); } }}
        />
        <button className="px-5 rounded-xl border border-border font-semibold hover:border-primary hover:text-primary transition-all" onClick={() => { setPage(1); setSearch(searchInput); }}>Cari</button>
        <button className="px-5 rounded-xl border border-border font-semibold hover:border-primary hover:text-primary transition-all" onClick={fetchOrders}>↻</button>
      </div>

      {error && <div className="text-red-500 text-sm font-bold bg-red-50 p-3 rounded-xl border border-red-100 mb-4">{error}</div>}

      <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden">
        {loading ? (
          <div className="h-64 flex items-center justify-center text-muted animate-pulse">Memuat...</div>
        ) : orders.length === 0 ? (
          <div className="h-64 flex items-center justify-center text-muted">Tidak ada pesanan.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-[0.9rem]">
              <thead>
                <tr className="bg-background/50 border-b border-border text-left text-[0.7rem] uppercase tracking-wider text-muted">
                  <th className="p-4">AWB</th>
                  <th className="p-4">Pengirim → Penerima</th>
                  <th className="p-4">Biaya</th>
                  <th className="p-4">Status</th>
                  <th className="p-4">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((o) => (
                  <tr key={o.id} className="border-b border-border last:border-none hover:bg-background/30">
                    <td className="p-4 font-bold">{o.awb}</td>
                    <td className="p-4 text-muted">{o.sender_name} → {o.receiver_name}</td>
                    <td className="p-4 font-semibold">Rp {o.total_cost?.toLocaleString("id-ID")}</td>
                    <td className="p-4">
                      <select
                        className={`text-[0.75rem] font-bold rounded-full px-3 py-1.5 border-none cursor-pointer ${badge(o.status)} ${updatingId === o.id ? "opacity-50" : ""}`}
                        value={STATUSES.includes(o.status) ? o.status : ""}
                        disabled={updatingId === o.id}
                        onChange={(e) => changeStatus(o, e.target.value)}
                      >
                        {!STATUSES.includes(o.status) && <option value="">{o.status}</option>}
                        {STATUSES.map((s) => <option key={s} value={s}>{s}</option>)}
                      </select>
                    </td>
                    <td className="p-4">
                      <div className="flex gap-3">
                        <button className="text-primary font-bold hover:underline" onClick={() => setQrOrder(o)}>Label QR</button>
                        {o.status !== "CANCELLED" && o.status !== "DELIVERED" && (
                          <button className="text-red-500 font-bold hover:underline" onClick={() => cancelOrder(o)} disabled={updatingId === o.id}>Batal</button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-center gap-4 mt-6">
        <button className="px-4 py-2 rounded-xl border border-border font-semibold disabled:opacity-40" disabled={page <= 1 || loading} onClick={() => setPage((p) => p - 1)}>← Sebelumnya</button>
        <span className="text-muted font-semibold">Halaman {page}</span>
        <button className="px-4 py-2 rounded-xl border border-border font-semibold disabled:opacity-40" disabled={orders.length < PER_PAGE || loading} onClick={() => setPage((p) => p + 1)}>Berikutnya →</button>
      </div>

      {/* QR label modal */}
      {qrOrder && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-text/40 backdrop-blur-md" onClick={() => setQrOrder(null)}></div>
          <div className="bg-surface w-full max-w-[400px] rounded-[32px] p-10 relative z-10 shadow-hover animate-fade-up text-center">
            <button onClick={() => setQrOrder(null)} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>
            <h2 className="text-xl font-bold mb-1">Label Pengiriman</h2>
            <p className="text-muted text-[0.85rem] mb-6">Scan di Konsol Scan Hub.</p>
            <div className="bg-white p-4 rounded-2xl inline-block border border-border mb-5"><QRCodeSVG value={qrOrder.awb} size={180} level="M" /></div>
            <div className="text-[0.8rem] text-muted font-medium">Nomor Resi (AWB)</div>
            <div className="text-xl font-black tracking-wide text-primary">{qrOrder.awb}</div>
            <div className="text-[0.85rem] text-muted mt-2">{qrOrder.sender_name} → {qrOrder.receiver_name}</div>
          </div>
        </div>
      )}

      {/* Create order modal */}
      {createOpen && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-text/40 backdrop-blur-md" onClick={() => setCreateOpen(false)}></div>
          <div className="bg-surface w-full max-w-[560px] max-h-[90vh] overflow-y-auto rounded-[32px] p-10 relative z-10 shadow-hover animate-fade-up">
            <button onClick={() => setCreateOpen(false)} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>
            {createdAwb ? (
              <div className="text-center">
                <div className="text-5xl mb-4">✅</div>
                <h2 className="text-2xl font-bold mb-2">Pesanan Dibuat</h2>
                <div className="bg-white p-4 rounded-2xl inline-block border border-border my-4"><QRCodeSVG value={createdAwb} size={150} level="M" /></div>
                <div className="text-2xl font-black text-primary">{createdAwb}</div>
                <button className="mt-6 bg-primary text-white font-heading font-semibold w-full h-12 rounded-pill hover:bg-primary-hover" onClick={() => setCreateOpen(false)}>Tutup</button>
              </div>
            ) : (
              <>
                <h2 className="text-2xl font-bold mb-6">Buat Pesanan Baru</h2>
                <div className="grid grid-cols-2 gap-3">
                  <input className={inputClass} placeholder="Nama Pengirim" value={form.sender_name} onChange={(e) => setForm({ ...form, sender_name: e.target.value })} />
                  <input className={inputClass} placeholder="Telp. Pengirim" value={form.sender_phone} onChange={(e) => setForm({ ...form, sender_phone: e.target.value })} />
                  <input className={inputClass} placeholder="Alamat Pengirim" value={form.sender_address} onChange={(e) => setForm({ ...form, sender_address: e.target.value })} />
                  <input className={inputClass} placeholder="Deskripsi Barang" value={form.item_description} onChange={(e) => setForm({ ...form, item_description: e.target.value })} />
                  <input className={inputClass} placeholder="Nama Penerima" value={form.receiver_name} onChange={(e) => setForm({ ...form, receiver_name: e.target.value })} />
                  <input className={inputClass} placeholder="Telp. Penerima" value={form.receiver_phone} onChange={(e) => setForm({ ...form, receiver_phone: e.target.value })} />
                  <input className={inputClass} placeholder="Alamat Penerima" value={form.receiver_address} onChange={(e) => setForm({ ...form, receiver_address: e.target.value })} />
                  <select className={inputClass} value={form.service_type} onChange={(e) => setForm({ ...form, service_type: e.target.value })}>
                    <option value="REGULAR">Reguler</option>
                    <option value="EXPRESS">Ekspres</option>
                    <option value="CARGO">Kargo</option>
                  </select>
                  <input className={inputClass} type="number" placeholder="Berat (kg)" value={form.weight_kg} onChange={(e) => setForm({ ...form, weight_kg: Number(e.target.value) })} />
                  <input className={inputClass} type="number" placeholder="Total Biaya" value={form.total_cost} onChange={(e) => setForm({ ...form, total_cost: Number(e.target.value) })} />
                </div>
                {createErr && <div className="text-red-500 text-sm font-bold mt-4">{createErr}</div>}
                <button className="bg-primary text-white font-heading font-semibold w-full h-12 rounded-pill hover:bg-primary-hover mt-6 disabled:opacity-50" onClick={submitCreate} disabled={creating}>
                  {creating ? "Membuat..." : "Buat Pesanan (AWB otomatis)"}
                </button>
              </>
            )}
          </div>
        </div>
      )}
    </section>
  );
};
