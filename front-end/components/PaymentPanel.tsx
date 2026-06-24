"use client";

import React, { useState } from "react";
import { apiGet, apiPost } from "@/lib/api";
import { Order } from "@/lib/types";

interface Transaction {
  id: string;
  order_id: string;
  amount: number;
  method: string;
  status: string;
  payment_url?: string;
  external_id?: string;
  idempotency_key: string;
}

const METHODS = [
  { code: "VA", label: "Virtual Account", icon: "🏦" },
  { code: "E_WALLET", label: "E-Wallet", icon: "📱" },
  { code: "CARD", label: "Kartu Kredit/Debit", icon: "💳" },
];

const rp = (n: number) => `Rp ${Math.round(n).toLocaleString("id-ID")}`;

// PaymentPanel drives the event-driven payment flow for one order:
// initiate (create pending transaction) → "Saya sudah bayar" (simulate the
// gateway webhook with status PAID) → poll the order until the payment.success
// event has advanced it past PENDING_PAYMENT. onPaid fires on confirmation.
export const PaymentPanel = ({ orderId, amount, onPaid }: { orderId: string; amount: number; onPaid?: () => void }) => {
  const [method, setMethod] = useState("VA");
  const [tx, setTx] = useState<Transaction | null>(null);
  const [phase, setPhase] = useState<"choose" | "awaiting" | "processing" | "paid">("choose");
  const [error, setError] = useState<string | null>(null);

  const initiate = async () => {
    setError(null);
    setPhase("processing");
    try {
      const t = await apiPost<Transaction>("/api/v1/payments/initiate", { order_id: orderId, amount, method }, { requiresAuth: true });
      setTx(t);
      setPhase("awaiting");
    } catch (err: any) {
      setError(err.message || "Gagal memulai pembayaran.");
      setPhase("choose");
    }
  };

  const confirm = async () => {
    if (!tx) return;
    setError(null);
    setPhase("processing");
    try {
      // Simulate the payment gateway calling our webhook with a successful charge.
      await apiPost("/api/v1/payments/webhook", {
        order_id: orderId, idempotency_key: tx.idempotency_key,
        external_id: tx.external_id, amount, status: "PAID",
      }, { requiresAuth: true });

      // The webhook emits payment.success → order-service moves the order off
      // PENDING_PAYMENT. Poll until that event has been processed.
      for (let i = 0; i < 12; i++) {
        await new Promise((r) => setTimeout(r, 1500));
        try {
          const o = await apiGet<Order>(`/api/v1/orders/${orderId}`, { requiresAuth: true });
          if (o.status && o.status !== "PENDING_PAYMENT") {
            setPhase("paid");
            onPaid?.();
            return;
          }
        } catch { /* keep polling */ }
      }
      // Paid but the downstream event is lagging — still treat as success.
      setPhase("paid");
      onPaid?.();
    } catch (err: any) {
      setError(err.message || "Gagal konfirmasi pembayaran.");
      setPhase("awaiting");
    }
  };

  if (phase === "paid") {
    return (
      <div className="bg-green-50 border border-green-200 rounded-2xl p-5 text-center">
        <div className="text-3xl mb-1">✅</div>
        <div className="font-bold text-green-700">Pembayaran berhasil</div>
        <div className="text-[0.85rem] text-muted mt-1">Pesanan diteruskan untuk diproses kurir.</div>
      </div>
    );
  }

  return (
    <div className="bg-background rounded-2xl p-5">
      <div className="flex items-center justify-between mb-4">
        <span className="font-bold">Pembayaran</span>
        <span className="font-black text-primary">{rp(amount)}</span>
      </div>

      {error && <div className="text-red-500 text-[0.85rem] font-bold mb-3">{error}</div>}

      {phase === "choose" && (
        <>
          <div className="flex flex-col gap-2 mb-4">
            {METHODS.map((m) => (
              <label key={m.code} className={`flex items-center gap-3 px-4 py-3 rounded-xl border cursor-pointer transition-all ${method === m.code ? "border-primary bg-primary/5" : "border-border"}`}>
                <input type="radio" name="paymethod" value={m.code} checked={method === m.code} onChange={() => setMethod(m.code)} />
                <span className="text-xl">{m.icon}</span>
                <span className="font-semibold text-[0.9rem]">{m.label}</span>
              </label>
            ))}
          </div>
          <button className="bg-primary text-white font-heading font-semibold w-full h-12 rounded-pill hover:bg-primary-hover transition-all" onClick={initiate}>
            Bayar Sekarang • {rp(amount)}
          </button>
        </>
      )}

      {phase === "processing" && (
        <div className="text-center py-4 text-muted animate-pulse font-semibold">Memproses…</div>
      )}

      {phase === "awaiting" && tx && (
        <>
          <div className="bg-surface border border-border rounded-xl p-4 mb-4 text-[0.85rem]">
            <div className="flex justify-between"><span className="text-muted">Metode</span><span className="font-bold">{METHODS.find((m) => m.code === tx.method)?.label || tx.method}</span></div>
            <div className="flex justify-between mt-1"><span className="text-muted">Kode Bayar (VA)</span><span className="font-mono font-bold">{(tx.external_id || "").replace(/-/g, "").slice(0, 12).toUpperCase()}</span></div>
            <div className="flex justify-between mt-1"><span className="text-muted">Jumlah</span><span className="font-bold text-primary">{rp(amount)}</span></div>
          </div>
          <p className="text-[0.78rem] text-muted mb-3 text-center">Mode simulasi — tekan tombol di bawah untuk meniru konfirmasi dari payment gateway.</p>
          <button className="bg-green-600 text-white font-heading font-semibold w-full h-12 rounded-pill hover:bg-green-700 transition-all" onClick={confirm}>
            ✅ Saya sudah bayar
          </button>
        </>
      )}
    </div>
  );
};
