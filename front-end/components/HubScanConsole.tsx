"use client";

import React, { useEffect, useRef, useState } from "react";
import { apiGet, apiPost } from "@/lib/api";
import { Hub } from "@/lib/types";

type ScanType = "inbound" | "sort" | "outbound";

const SCAN_META: Record<ScanType, { label: string; emoji: string; desc: string }> = {
  inbound: { label: "Scan Masuk", emoji: "📥", desc: "Paket tiba di hub (ARRIVED)" },
  sort: { label: "Sortir", emoji: "🔀", desc: "Paket disortir (SORTED)" },
  outbound: { label: "Scan Keluar", emoji: "📤", desc: "Paket berangkat dari hub (DEPARTED)" },
};

interface LogEntry { time: string; text: string; ok: boolean; }

export const HubScanConsole = () => {
  const [hubs, setHubs] = useState<Hub[]>([]);
  const [hubId, setHubId] = useState("");
  const [awb, setAwb] = useState("");
  const [operator, setOperator] = useState("OP-01");
  const [busy, setBusy] = useState<ScanType | null>(null);
  const [log, setLog] = useState<LogEntry[]>([]);

  // Camera scanning state
  const [scanning, setScanning] = useState(false);
  const [camError, setCamError] = useState<string | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    apiGet<Hub[]>("/api/v1/hub/list")
      .then((data) => {
        setHubs(data || []);
        if (data && data.length > 0) setHubId(data[0].id);
      })
      .catch(() => {});
    return () => stopCamera();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const pushLog = (text: string, ok: boolean) => {
    const time = new Date().toLocaleTimeString("id-ID");
    setLog((l) => [{ time, text, ok }, ...l].slice(0, 20));
  };

  const stopCamera = () => {
    if (rafRef.current) cancelAnimationFrame(rafRef.current);
    rafRef.current = null;
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((t) => t.stop());
      streamRef.current = null;
    }
    setScanning(false);
  };

  const startCamera = async () => {
    setCamError(null);
    const BarcodeDetectorCtor = (window as any).BarcodeDetector;
    if (!BarcodeDetectorCtor) {
      setCamError("Pemindai QR kamera tidak didukung browser ini. Masukkan AWB secara manual.");
      return;
    }
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" } });
      streamRef.current = stream;
      setScanning(true);
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        await videoRef.current.play();
      }
      const detector = new BarcodeDetectorCtor({ formats: ["qr_code"] });
      const tick = async () => {
        if (!videoRef.current || !streamRef.current) return;
        try {
          const codes = await detector.detect(videoRef.current);
          if (codes && codes.length > 0) {
            const value = String(codes[0].rawValue || "").trim();
            if (value) {
              setAwb(value);
              pushLog(`QR terbaca: ${value}`, true);
              stopCamera();
              return;
            }
          }
        } catch {
          /* transient detect error — keep trying */
        }
        rafRef.current = requestAnimationFrame(tick);
      };
      rafRef.current = requestAnimationFrame(tick);
    } catch (e: any) {
      setCamError(e?.message || "Gagal mengakses kamera.");
      stopCamera();
    }
  };

  const doScan = async (type: ScanType) => {
    if (!hubId) { pushLog("Pilih hub dulu.", false); return; }
    if (!awb.trim()) { pushLog("Masukkan/scan AWB dulu.", false); return; }
    setBusy(type);
    try {
      await apiPost(`/api/v1/hub/scan/${type}`, { awb: awb.trim(), hub_id: hubId, operator_id: operator });
      const hubName = hubs.find((h) => h.id === hubId)?.name || hubId;
      pushLog(`${SCAN_META[type].label} OK — ${awb.trim()} @ ${hubName}`, true);
    } catch (err: any) {
      pushLog(`${SCAN_META[type].label} gagal: ${err.message || "error"}`, false);
    } finally {
      setBusy(null);
    }
  };

  const inputClass = "w-full h-12 bg-background border border-border rounded-2xl px-5 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all";

  return (
    <div className="bg-surface border border-border rounded-card p-8 shadow-soft mb-12">
      <h2 className="text-2xl font-bold mb-1">Konsol Scan Hub</h2>
      <p className="text-muted mb-6">Pindai QR pada label AWB (atau ketik manual) lalu rekam pergerakan paket di hub.</p>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Left: inputs */}
        <div className="flex flex-col gap-4">
          <div>
            <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Hub</label>
            <select className={inputClass} value={hubId} onChange={(e) => setHubId(e.target.value)}>
              {hubs.length === 0 && <option value="">Memuat hub...</option>}
              {hubs.map((h) => (
                <option key={h.id} value={h.id}>{h.name} ({h.code})</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Nomor Resi (AWB)</label>
            <div className="flex gap-2">
              <input className={inputClass} value={awb} onChange={(e) => setAwb(e.target.value)} placeholder="NR..." />
              <button
                className="shrink-0 px-4 rounded-2xl border border-border font-semibold hover:border-primary hover:text-primary transition-all"
                onClick={scanning ? stopCamera : startCamera}
              >
                {scanning ? "Tutup" : "📷 Scan QR"}
              </button>
            </div>
          </div>
          <div>
            <label className="block text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">ID Operator</label>
            <input className={inputClass} value={operator} onChange={(e) => setOperator(e.target.value)} placeholder="OP-01" />
          </div>

          {scanning && (
            <div className="rounded-2xl overflow-hidden border border-border bg-black aspect-video relative">
              <video ref={videoRef} className="w-full h-full object-cover" muted playsInline />
              <div className="absolute inset-0 border-[3px] border-primary/60 m-10 rounded-xl pointer-events-none"></div>
            </div>
          )}
          {camError && <div className="text-red-500 text-sm font-semibold">{camError}</div>}

          <div className="grid grid-cols-3 gap-3 mt-2">
            {(["inbound", "sort", "outbound"] as ScanType[]).map((t) => (
              <button
                key={t}
                className="flex flex-col items-center gap-1 bg-primary/10 text-primary font-heading font-bold py-4 rounded-2xl hover:bg-primary/20 transition-all disabled:opacity-50"
                onClick={() => doScan(t)}
                disabled={busy !== null}
                title={SCAN_META[t].desc}
              >
                <span className="text-2xl">{SCAN_META[t].emoji}</span>
                <span className="text-[0.85rem]">{busy === t ? "..." : SCAN_META[t].label}</span>
              </button>
            ))}
          </div>
        </div>

        {/* Right: activity log */}
        <div>
          <div className="text-[0.75rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Aktivitas Scan</div>
          <div className="bg-background rounded-2xl p-4 h-[260px] overflow-y-auto border border-border">
            {log.length === 0 ? (
              <div className="text-muted text-sm h-full flex items-center justify-center">Belum ada aktivitas.</div>
            ) : (
              <ul className="flex flex-col gap-2">
                {log.map((e, i) => (
                  <li key={i} className={`text-[0.85rem] flex gap-2 ${e.ok ? "text-text" : "text-red-500"}`}>
                    <span className="text-muted tabular-nums">{e.time}</span>
                    <span>{e.ok ? "✓" : "✗"} {e.text}</span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
