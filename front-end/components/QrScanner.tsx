"use client";

import React, { useEffect, useRef, useState } from "react";

interface QrScannerProps {
  title?: string;
  hint?: string;
  onResult: (value: string) => void;
  onClose: () => void;
}

// Reusable QR scanning modal. Uses the browser BarcodeDetector API (same
// approach as the Hub scan console). Falls back to manual entry when the API
// is unavailable (e.g. desktop browsers without the feature).
export const QrScanner: React.FC<QrScannerProps> = ({ title = "Pindai QR", hint, onResult, onClose }) => {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const rafRef = useRef<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [manual, setManual] = useState("");
  const [supported, setSupported] = useState(true);

  const stop = () => {
    if (rafRef.current) cancelAnimationFrame(rafRef.current);
    rafRef.current = null;
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((t) => t.stop());
      streamRef.current = null;
    }
  };

  useEffect(() => {
    const Ctor = (window as any).BarcodeDetector;
    if (!Ctor) {
      setSupported(false);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" } });
        if (cancelled) {
          stream.getTracks().forEach((t) => t.stop());
          return;
        }
        streamRef.current = stream;
        if (videoRef.current) {
          videoRef.current.srcObject = stream;
          await videoRef.current.play();
        }
        const detector = new Ctor({ formats: ["qr_code"] });
        const tick = async () => {
          if (!videoRef.current || !streamRef.current) return;
          try {
            const codes = await detector.detect(videoRef.current);
            if (codes && codes.length > 0) {
              const value = String(codes[0].rawValue || "").trim();
              if (value) {
                stop();
                onResult(value);
                return;
              }
            }
          } catch {
            /* transient detect error — keep scanning */
          }
          rafRef.current = requestAnimationFrame(tick);
        };
        rafRef.current = requestAnimationFrame(tick);
      } catch (e: any) {
        setError(e?.message || "Gagal mengakses kamera.");
      }
    })();

    return () => {
      cancelled = true;
      stop();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const close = () => {
    stop();
    onClose();
  };

  return (
    <div className="fixed inset-0 z-[300] flex items-center justify-center p-6">
      <div className="absolute inset-0 bg-text/50 backdrop-blur-md" onClick={close}></div>
      <div className="bg-surface w-full max-w-[440px] rounded-[32px] p-8 relative z-10 shadow-hover animate-fade-up">
        <button onClick={close} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>
        <h2 className="text-xl font-bold mb-1">{title}</h2>
        {hint && <p className="text-muted text-[0.85rem] mb-4">{hint}</p>}

        {supported ? (
          <div className="rounded-2xl overflow-hidden border border-border bg-black aspect-video relative mb-4">
            <video ref={videoRef} className="w-full h-full object-cover" muted playsInline />
            <div className="absolute inset-0 border-[3px] border-primary/60 m-10 rounded-xl pointer-events-none"></div>
          </div>
        ) : (
          <div className="text-amber-600 text-sm font-semibold bg-amber-50 p-3 rounded-xl border border-amber-100 mb-4">
            Pemindai kamera tidak didukung browser ini. Masukkan AWB secara manual.
          </div>
        )}

        {error && <div className="text-red-500 text-sm font-semibold mb-4">{error}</div>}

        {/* Manual fallback — always available */}
        <div className="flex gap-2">
          <input
            className="w-full h-12 bg-background border border-border rounded-2xl px-5 text-base outline-none focus:border-primary transition-all"
            placeholder="Ketik AWB (NR...)"
            value={manual}
            onChange={(e) => setManual(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter" && manual.trim()) { stop(); onResult(manual.trim()); } }}
          />
          <button
            className="shrink-0 px-5 rounded-2xl bg-primary text-white font-semibold hover:bg-primary-hover transition-all disabled:opacity-50"
            disabled={!manual.trim()}
            onClick={() => { stop(); onResult(manual.trim()); }}
          >
            OK
          </button>
        </div>
      </div>
    </div>
  );
};
