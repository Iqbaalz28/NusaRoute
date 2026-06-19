"use client";

import React, { useCallback, useEffect, useState } from "react";
import { apiGet, apiPost } from "@/lib/api";

interface Ticket { id: string; order_id: string; awb: string; type: string; priority: string; status: string; description: string; resolution?: string; created_at: string; }
interface Claim { id: string; ticket_id: string; order_id: string; claim_type: string; amount: number; status: string; }

const TYPE_LABEL: Record<string, string> = { LOST: "Paket Hilang", DAMAGED: "Paket Rusak", DELIVERY_FAILED: "Gagal Kirim", COMPLAINT: "Komplain" };
const TICKET_STATUS: Record<string, string> = { OPEN: "Terbuka", IN_PROGRESS: "Diproses", RESOLVED: "Selesai", CLOSED: "Ditutup" };
const CLAIM_STATUS: Record<string, string> = { PENDING: "Menunggu Review", APPROVED: "Disetujui", REJECTED: "Ditolak", PAID: "Dana Dicairkan" };
const claimStyle = (s: string) =>
  s === "PAID" ? "bg-green-100 text-green-700" :
  s === "APPROVED" ? "bg-blue-100 text-blue-700" :
  s === "REJECTED" ? "bg-red-100 text-red-700" : "bg-amber-100 text-amber-700";
const rp = (n: number) => `Rp ${Math.round(n).toLocaleString("id-ID")}`;

// ResolutionPanel — customer self-service for problems & insurance on one order:
// file a ticket (lost/damaged/failed/complaint), and for an INSURED order file an
// insurance claim against a lost/damaged ticket. Shows current ticket/claim status.
export const ResolutionPanel = ({ orderId, awb, isInsured, insuredValue }: { orderId: string; awb: string; isInsured: boolean; insuredValue: number }) => {
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [claims, setClaims] = useState<Claim[]>([]);
  const [reporting, setReporting] = useState(false);
  const [type, setType] = useState("LOST");
  const [desc, setDesc] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await apiGet<{ tickets: Ticket[]; claims: Claim[] }>(`/api/v1/resolutions/order/${orderId}`, { requiresAuth: true });
      setTickets(res?.tickets || []);
      setClaims(res?.claims || []);
    } catch { /* non-fatal */ }
  }, [orderId]);

  useEffect(() => { load(); }, [load]);

  const submitTicket = async () => {
    if (!desc.trim()) { setError("Jelaskan masalahnya dulu."); return; }
    setBusy(true); setError(null);
    try {
      await apiPost("/api/v1/resolutions/tickets", { order_id: orderId, awb, type, description: desc.trim() }, { requiresAuth: true });
      setDesc(""); setReporting(false);
      await load();
    } catch (err: any) {
      setError(err.message || "Gagal mengirim laporan.");
    } finally { setBusy(false); }
  };

  const fileClaim = async (ticket: Ticket) => {
    setBusy(true); setError(null);
    try {
      await apiPost("/api/v1/resolutions/claims", { ticket_id: ticket.id, order_id: orderId, claim_type: "INSURANCE", amount: insuredValue }, { requiresAuth: true });
      await load();
    } catch (err: any) {
      setError(err.message || "Gagal mengajukan klaim.");
    } finally { setBusy(false); }
  };

  const claimForTicket = (tid: string) => claims.find((c) => c.ticket_id === tid);

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <div className="text-[0.7rem] font-bold uppercase tracking-wider text-muted">Bantuan &amp; Klaim</div>
        {isInsured ? (
          <span className="text-[0.7rem] font-bold px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700">🛡 Diasuransikan {rp(insuredValue)}</span>
        ) : (
          <span className="text-[0.7rem] text-muted">Tanpa asuransi</span>
        )}
      </div>

      {error && <div className="text-red-500 text-[0.8rem] font-bold mb-2">{error}</div>}

      {/* Existing tickets */}
      {tickets.length > 0 && (
        <div className="flex flex-col gap-2 mb-3">
          {tickets.map((t) => {
            const claim = claimForTicket(t.id);
            const canClaim = isInsured && !claim && (t.type === "LOST" || t.type === "DAMAGED") && t.status !== "CLOSED";
            return (
              <div key={t.id} className="bg-background rounded-2xl p-3 text-[0.85rem]">
                <div className="flex items-center justify-between gap-2">
                  <span className="font-bold">{TYPE_LABEL[t.type] || t.type}</span>
                  <span className="text-[0.7rem] font-bold px-2 py-0.5 rounded-full bg-gray-200 text-gray-600">{TICKET_STATUS[t.status] || t.status}</span>
                </div>
                {t.description && <div className="text-muted text-[0.8rem] mt-1">{t.description}</div>}
                {t.resolution && <div className="text-[0.78rem] mt-1">Resolusi: <strong>{t.resolution}</strong></div>}
                {claim && (
                  <div className="flex items-center justify-between mt-2 pt-2 border-t border-border">
                    <span className="text-[0.8rem]">🛡 Klaim {rp(claim.amount)}</span>
                    <span className={`text-[0.7rem] font-bold px-2 py-0.5 rounded-full ${claimStyle(claim.status)}`}>{CLAIM_STATUS[claim.status] || claim.status}</span>
                  </div>
                )}
                {canClaim && (
                  <button className="mt-2 w-full bg-emerald-600 text-white font-semibold text-[0.85rem] py-2 rounded-pill hover:bg-emerald-700 disabled:opacity-50" onClick={() => fileClaim(t)} disabled={busy}>
                    🛡 Ajukan Klaim Asuransi • {rp(insuredValue)}
                  </button>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Report problem */}
      {reporting ? (
        <div className="bg-background rounded-2xl p-3">
          <select className="w-full h-10 bg-surface border border-border rounded-xl px-3 text-[0.9rem] mb-2 outline-none focus:border-primary" value={type} onChange={(e) => setType(e.target.value)}>
            <option value="LOST">Paket Hilang</option>
            <option value="DAMAGED">Paket Rusak</option>
            <option value="DELIVERY_FAILED">Gagal Kirim</option>
            <option value="COMPLAINT">Komplain</option>
          </select>
          <textarea className="w-full bg-surface border border-border rounded-xl px-3 py-2 text-[0.9rem] mb-2 outline-none focus:border-primary resize-none" rows={3} placeholder="Jelaskan masalahnya…" value={desc} onChange={(e) => setDesc(e.target.value)} />
          <div className="flex gap-2">
            <button className="flex-1 bg-primary text-white font-semibold text-[0.85rem] py-2 rounded-pill hover:bg-primary-hover disabled:opacity-50" onClick={submitTicket} disabled={busy}>{busy ? "Mengirim…" : "Kirim Laporan"}</button>
            <button className="px-4 text-muted font-semibold text-[0.85rem]" onClick={() => setReporting(false)}>Batal</button>
          </div>
        </div>
      ) : (
        <button className="w-full border border-border rounded-pill py-2.5 text-[0.85rem] font-bold text-primary hover:border-primary transition-all" onClick={() => setReporting(true)}>
          + Laporkan Masalah
        </button>
      )}
    </div>
  );
};
