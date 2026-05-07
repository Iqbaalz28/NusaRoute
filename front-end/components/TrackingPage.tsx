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
    { status: "Out for Delivery", detail: "Pending", location: "", time: "", type: "pending" },
    { status: "In Transit", detail: "Chicago Distribution Center, IL", location: "Oct 26, 02:15 PM", time: "", type: "active" },
    { status: "Picked Up", detail: "Los Angeles Drop-off Point, CA", location: "Oct 25, 10:30 AM", time: "", type: "done" },
    { status: "Label Created", detail: "Sender provided shipping information", location: "Oct 24, 08:00 AM", time: "", type: "done" },
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
            <span className="status-pill status-pill--transit">In Transit</span>
            <p style={{ color: 'var(--color-muted)', marginBottom: '0.5rem' }}>Updated 2 mins ago</p>
            <h1 style={{ fontSize: '3.5rem', maxWidth: '600px', lineHeight: '1.1' }}>
              Arriving by Tomorrow, 8:00 PM
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
                    {e.location && <div className="timeline__time">{e.location}</div>}
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
                <div className="details-value">Tech Supply Co.</div>
                <div className="details-label">Los Angeles, CA 90001</div>
              </div>

              <div className="details-section">
                <div className="details-label">To</div>
                <div className="details-value">Jane Doe</div>
                <div className="details-label">Chicago, IL 60601</div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="details-section">
                  <div className="details-label">Weight</div>
                  <div className="details-value">4.2 lbs</div>
                </div>
                <div className="details-section">
                  <div className="details-label">Dimensions</div>
                  <div className="details-value">12" x 8" x 6"</div>
                </div>
              </div>

              <div style={{ marginTop: '2rem', borderRadius: '16px', overflow: 'hidden', height: '150px', background: '#E2E8F0', position: 'relative' }}>
                <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--color-muted)' }}>
                  Map View
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </section>
  );
};
