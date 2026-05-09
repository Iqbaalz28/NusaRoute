"use client";

import React, { useState, useEffect } from "react";

interface TrackingPageProps {
  initialAwb?: string;
}

export const TrackingPage: React.FC<TrackingPageProps> = ({ initialAwb }) => {
  const [awb, setAwb] = useState(initialAwb || "");
  const [showResult, setShowResult] = useState(!!initialAwb);

  useEffect(() => {
    if (initialAwb) {
      setAwb(initialAwb);
      setShowResult(true);
    }
  }, [initialAwb]);

  const events = [
    { status: "Package Received", detail: "Package has been received by recipient", location: "South Jakarta", time: "05 May 2026, 14:32 WIB", type: "done" },
    { status: "Out for Delivery", detail: "Courier is on the way to the recipient's address", location: "South Jakarta", time: "05 May 2026, 13:45 WIB", type: "active" },
    { status: "Arrived at Hub Jakarta", detail: "Package sorted at Central Jakarta Hub (JKT-01)", location: "Central Jakarta Hub", time: "05 May 2026, 08:12 WIB", type: "done" },
    { status: "Departed from Hub Bandung", detail: "Package departed for Jakarta Hub", location: "Bandung Hub", time: "04 May 2026, 22:30 WIB", type: "done" },
    { status: "Arrived at Hub Bandung", detail: "Package arrived and scanned at Bandung Hub (BDG-01)", location: "Bandung Hub", time: "04 May 2026, 18:15 WIB", type: "done" },
    { status: "Package Picked Up", detail: "Package picked up by courier from sender's address", location: "Bandung", time: "04 May 2026, 15:20 WIB", type: "done" },
  ];

  const handleTrack = () => {
    if (awb.trim()) {
      setShowResult(true);
    }
  };

  return (
    <section className="page page--active">
      {!showResult ? (
        <div className="calc-container" style={{ textAlign: 'center' }}>
          <div className="calc-header">
            <h1>Lacak Paket</h1>
            <p>Masukkan nomor resi untuk melihat status terkini.</p>
          </div>
          <div className="hero__track-container" style={{ margin: '0 auto' }}>
            <input 
              type="text" 
              className="hero__track-input" 
              placeholder="Enter Tracking Number (e.g. LUM...)"
              value={awb}
              onChange={(e) => setAwb(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleTrack()}
            />
            <button className="btn btn--primary" onClick={handleTrack}>Track →</button>
          </div>
        </div>
      ) : (
        <>
          <div className="track-header">
            <span className="status-pill status-pill--delivered">Delivered</span>
            <p style={{ color: 'var(--color-muted)', marginBottom: '0.5rem' }}>Updated 2 mins ago</p>
            <h1 style={{ fontSize: '3.5rem', maxWidth: '600px', lineHeight: '1.1' }}>
              Package Delivered Successfully
            </h1>
          </div>

          <div className="track-grid">
            <div className="timeline">
              {events.map((e, i) => (
                <div 
                  className={`timeline__item ${e.type === 'active' ? 'timeline__item--active' : ''} ${e.type === 'done' ? 'timeline__item--done' : ''}`} 
                  key={i}
                >
                  <div className="timeline__dot"></div>
                  <div className="timeline__content">
                    <div className="timeline__status">{e.status}</div>
                    <div className="timeline__location">{e.detail}</div>
                    <div className="timeline__time">📍 {e.location} · {e.time}</div>
                  </div>
                </div>
              ))}
            </div>

            <div className="shipment-details">
              <h2 style={{ marginBottom: '2rem' }}>Shipment Details</h2>
              
              <div className="details-section">
                <div className="details-label">Tracking Number</div>
                <div className="details-value">#{awb}</div>
              </div>

              <div className="details-section">
                <div className="details-label">Service Tier</div>
                <div className="details-value">⚡ Express Delivery</div>
              </div>

              <div className="details-section">
                <div className="details-label">From</div>
                <div className="details-value">Budi S.</div>
                <div className="details-label">Bandung, Jawa Barat</div>
              </div>

              <div className="details-section">
                <div className="details-label">To</div>
                <div className="details-value">Siti Aminah</div>
                <div className="details-label">Jakarta Selatan, DKI Jakarta</div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="details-section">
                  <div className="details-label">Weight</div>
                  <div className="details-value">2.0 kg</div>
                </div>
                <div className="details-section">
                  <div className="details-label">Dimensions</div>
                  <div className="details-value">30cm x 20cm x 15cm</div>
                </div>
              </div>

              <div style={{ marginTop: '2rem', borderRadius: '16px', overflow: 'hidden', height: '150px', background: '#E2E8F0', position: 'relative' }}>
                <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--color-muted)' }}>
                  Map View (Lumina Tracking)
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </section>
  );
};
