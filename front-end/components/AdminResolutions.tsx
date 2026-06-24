"use client";

import React, { useCallback, useEffect, useState } from "react";
import { apiGet, apiPut } from "@/lib/api";

interface Ticket { id: string; order_id: string; awb: string; type: string; priority: string; status: string; description: string; resolution?: string; created_at: string; }
interface Claim { id: string; ticket_id: string; order_id: string; claim_type: string; amount: number; status: string; }

const TYPE_LABEL: Record<string, string> = { LOST: "Hilang", DAMAGED: "Rusak", DELIVERY_FAILED: "Gagal Kirim", COMPLAINT: "Komplain" };
const PRIO_STYLE: Record<string, string> = { CRITICAL: "bg-red-100 text-red-700", HIGH: "bg-orange-100 text-orange-700", MEDIUM: "bg-amber-100 text-amber-700", LOW: "bg-gray-100 text-gray-600" };
const claimStyle = (s: string) => s === "PAID" ? "bg-green-100 text-green-700" : s === "APPROVED" ? "bg-blue-100 text-blue-700" : s === "REJECTED" ? "bg-red-100 text-red-700" : "bg-amber-100 text-amber-700";
const rp = (n: number) => `Rp ${Math.round(n).toLocaleString("id-ID")}`;

export const AdminResolutions = () => {
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [claims, setClaims] = useState<Claim[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [amounts, setAmounts] = useState<Record<string, number>>({});

  const load = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const [t, c] = await Promise.all([
        apiGet<Ticket[]>("/api/v1/resolutions/admin/tickets", { requiresAuth: true }),
        apiGet<Claim[]>("/api/v1/resolutions/admin/claims", { requiresAuth: true }),
      ]);
      setTickets(t || []);
      setClaims(c || []);
      setAmounts(Object.fromEntries((c || []).map((x) => [x.id, x.amount])));
    } catch (err: any) {
      setError(err.message || "Gagal memuat data resolusi.");
    } finally { setLoading(false); }
  }, []);

  useEffect(() => { load(); }, [load]);

  const updateTicket = async (id: string, body: { status?: string; resolution?: string }) => {
    setBusyId(id); setError(null);
    try {
      await apiPut(`/api/v1/resolutions/admin/tickets/${id}`, body, { requiresAuth: true });
      await load();
    } catch (err: any) { setError(err.message || "Gagal memperbarui tiket."); }
    finally { setBusyId(null); }
  };

  const updateClaim = async (c: Claim, status: string) => {
    setBusyId(c.id); setError(null);
    try {
      await apiPut(`/api/v1/resolutions/admin/claims/${c.id}`, { status, amount: amounts[c.id] ?? c.amount }, { requiresAuth: true });
      await load();
    } catch (err: any) { setError(err.message || "Gagal memperbarui klaim."); }
    finally { setBusyId(null); }
  };

  const sel = "text-[0.8rem] font-semibold rounded-lg border border-border bg-background px-2 py-1 outline-none focus:border-primary";

  if (loading) return <div className="bg-surface border border-border rounded-card h-64 flex items-center justify-center text-muted animate-pulse">Memuat resolusi…</div>;

  return (
    <>
      {error && <div className="text-red-500 text-sm font-bold bg-red-50 p-3 rounded-xl border border-red-100 mb-4">{error}</div>}

      {/* Tickets */}
      <h2 className="text-lg font-bold mb-3">Tiket ({tickets.length})</h2>
      <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden mb-8">
        {tickets.length === 0 ? (
          <div className="h-32 flex items-center justify-center text-muted">Tidak ada tiket.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-[0.85rem]">
              <thead>
                <tr className="bg-background/50 border-b border-border text-left text-[0.7rem] uppercase tracking-wider text-muted">
                  <th className="p-3">Jenis</th><th className="p-3">AWB</th><th className="p-3">Prioritas</th><th className="p-3">Deskripsi</th><th className="p-3">Status</th><th className="p-3">Resolusi</th>
                </tr>
              </thead>
              <tbody>
                {tickets.map((t) => (
                  <tr key={t.id} className={`border-b border-border last:border-none ${busyId === t.id ? "opacity-50" : ""}`}>
                    <td className="p-3 font-bold">{TYPE_LABEL[t.type] || t.type}</td>
                    <td className="p-3 font-mono text-[0.8rem]">{t.awb}</td>
                    <td className="p-3"><span className={`text-[0.7rem] font-bold px-2 py-0.5 rounded-full ${PRIO_STYLE[t.priority] || "bg-gray-100"}`}>{t.priority}</span></td>
                    <td className="p-3 text-muted max-w-[220px] truncate" title={t.description}>{t.description}</td>
                    <td className="p-3">
                      <select className={sel} value={t.status} disabled={busyId === t.id} onChange={(e) => updateTicket(t.id, { status: e.target.value })}>
                        {["OPEN", "IN_PROGRESS", "RESOLVED", "CLOSED"].map((s) => <option key={s} value={s}>{s}</option>)}
                      </select>
                    </td>
                    <td className="p-3">
                      <select className={sel} value={t.resolution || ""} disabled={busyId === t.id} onChange={(e) => updateTicket(t.id, { status: "RESOLVED", resolution: e.target.value })}>
                        <option value="">—</option>
                        {["REFUND", "RESEND", "RETURN", "CLOSED"].map((s) => <option key={s} value={s}>{s}</option>)}
                      </select>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Claims */}
      <h2 className="text-lg font-bold mb-3">Klaim Asuransi ({claims.length})</h2>
      <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden">
        {claims.length === 0 ? (
          <div className="h-32 flex items-center justify-center text-muted">Tidak ada klaim.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-[0.85rem]">
              <thead>
                <tr className="bg-background/50 border-b border-border text-left text-[0.7rem] uppercase tracking-wider text-muted">
                  <th className="p-3">Tipe</th><th className="p-3">Order</th><th className="p-3">Jumlah (Rp)</th><th className="p-3">Status</th><th className="p-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {claims.map((c) => (
                  <tr key={c.id} className={`border-b border-border last:border-none ${busyId === c.id ? "opacity-50" : ""}`}>
                    <td className="p-3 font-bold">{c.claim_type}</td>
                    <td className="p-3 font-mono text-[0.78rem]">{c.order_id.substring(0, 12)}</td>
                    <td className="p-3">
                      <input type="number" className="w-28 rounded-lg border border-border bg-background px-2 py-1 text-[0.82rem] outline-none focus:border-primary"
                        value={amounts[c.id] ?? c.amount}
                        onChange={(e) => setAmounts((a) => ({ ...a, [c.id]: Number(e.target.value) }))} />
                    </td>
                    <td className="p-3"><span className={`text-[0.7rem] font-bold px-2 py-0.5 rounded-full ${claimStyle(c.status)}`}>{c.status}</span></td>
                    <td className="p-3">
                      <div className="flex gap-2">
                        <button className="text-blue-600 font-bold hover:underline disabled:opacity-40" disabled={busyId === c.id || c.status !== "PENDING"} onClick={() => updateClaim(c, "APPROVED")}>Setujui</button>
                        <button className="text-green-600 font-bold hover:underline disabled:opacity-40" disabled={busyId === c.id || c.status === "PAID" || c.status === "REJECTED"} onClick={() => updateClaim(c, "PAID")}>Bayar</button>
                        <button className="text-red-500 font-bold hover:underline disabled:opacity-40" disabled={busyId === c.id || c.status === "PAID"} onClick={() => updateClaim(c, "REJECTED")}>Tolak</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="mt-4 text-right">
        <button className="text-[0.85rem] font-bold text-primary hover:underline" onClick={load}>↻ Muat ulang</button>
      </div>
    </>
  );
};
