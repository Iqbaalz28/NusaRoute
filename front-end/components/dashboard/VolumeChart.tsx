"use client";

import React, { useEffect, useRef } from "react";

export const VolumeChart = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const parent = canvas.parentElement;
    if (!parent) return;

    const width = parent.offsetWidth - 48;
    const height = 280;

    canvas.width = width * dpr;
    canvas.height = height * dpr;
    canvas.style.width = `${width}px`;
    canvas.style.height = `${height}px`;
    ctx.scale(dpr, dpr);

    const W = width, H = height;
    const data = [12400, 14200, 11800, 15600, 16800, 14900, 15847];
    const labels = ["Sen", "Sel", "Rab", "Kam", "Jum", "Sab", "Min"];
    const maxVal = Math.max(...data) * 1.15;
    const pad = { top: 20, right: 20, bottom: 40, left: 60 };
    const chartW = W - pad.left - pad.right;
    const chartH = H - pad.top - pad.bottom;

    // Grid — light theme
    ctx.strokeStyle = "rgba(226,232,240,0.8)";
    ctx.lineWidth = 1;
    for (let i = 0; i <= 4; i++) {
      const y = pad.top + (chartH / 4) * i;
      ctx.beginPath();
      ctx.moveTo(pad.left, y);
      ctx.lineTo(W - pad.right, y);
      ctx.stroke();
      const val = Math.round(maxVal - (maxVal / 4) * i);
      ctx.fillStyle = "#94a3b8";
      ctx.font = "11px Inter, sans-serif";
      ctx.textAlign = "right";
      ctx.fillText(val.toLocaleString("id-ID"), pad.left - 8, y + 4);
    }

    // Area gradient — NusaRoute Primary
    const gradient = ctx.createLinearGradient(0, pad.top, 0, H - pad.bottom);
    gradient.addColorStop(0, "rgba(255,107,74,0.15)");
    gradient.addColorStop(1, "rgba(255,107,74,0.01)");

    const points = data.map((v, i) => ({
      x: pad.left + (chartW / (data.length - 1)) * i,
      y: pad.top + chartH - (v / maxVal) * chartH,
    }));

    // Fill area
    ctx.beginPath();
    ctx.moveTo(points[0].x, H - pad.bottom);
    points.forEach((p) => ctx.lineTo(p.x, p.y));
    ctx.lineTo(points[points.length - 1].x, H - pad.bottom);
    ctx.closePath();
    ctx.fillStyle = gradient;
    ctx.fill();

    // Line
    ctx.beginPath();
    ctx.moveTo(points[0].x, points[0].y);
    for (let i = 1; i < points.length; i++) {
      const cp1x = (points[i - 1].x + points[i].x) / 2;
      ctx.bezierCurveTo(cp1x, points[i - 1].y, cp1x, points[i].y, points[i].x, points[i].y);
    }
    ctx.strokeStyle = "#FF6B4A";
    ctx.lineWidth = 3;
    ctx.stroke();

    // Dots
    points.forEach((p, i) => {
      ctx.beginPath();
      ctx.arc(p.x, p.y, 5, 0, Math.PI * 2);
      ctx.fillStyle = i === points.length - 1 ? "#FF6B4A" : "#2D3142";
      ctx.fill();
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = 2;
      ctx.stroke();
    });

    // Labels
    labels.forEach((l, i) => {
      ctx.fillStyle = "#8C93A8";
      ctx.font = "bold 11px var(--font-body), sans-serif";
      ctx.textAlign = "center";
      ctx.fillText(l, points[i].x, H - pad.bottom + 25);
    });
  }, []);

  return <canvas ref={canvasRef} className="w-full h-[280px]"></canvas>;
};
