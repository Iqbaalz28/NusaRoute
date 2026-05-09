"use client";

import React, { useState } from "react";

export const ServicesPage = () => {
  const [origin, setOrigin] = useState("");
  const [dest, setDest] = useState("");
  const [weight, setWeight] = useState(0);
  const [length, setLength] = useState(0);
  const [width, setWidth] = useState(0);
  const [height, setHeight] = useState(0);
  const [showResults, setShowResults] = useState(false);

  const handleCalculate = () => {
    if (origin && dest && weight > 0) {
      setShowResults(true);
    }
  };

  return (
    <section className="page page--active">
      <div className="calc-container">
        <div className="calc-header">
          <h1>Get an Estimate</h1>
          <p>Enter your shipment details to find the best rate.</p>
        </div>

        <div className="location-box">
          <div className="location-input-group">
            <input 
              type="text" 
              className="location-input" 
              placeholder="Origin ZIP or City" 
              value={origin}
              onChange={(e) => setOrigin(e.target.value)}
            />
            <input 
              type="text" 
              className="location-input" 
              placeholder="Destination ZIP or City" 
              value={dest}
              onChange={(e) => setDest(e.target.value)}
            />
          </div>
        </div>

        <h3 style={{ marginBottom: '1.5rem', fontSize: '1.25rem' }}>Package Details</h3>
        <div className="package-grid">
          <div className="package-input-box">
            <label>Length (cm)</label>
            <input type="number" value={length} onChange={(e) => setLength(Number(e.target.value))} />
          </div>
          <div className="package-input-box">
            <label>Width (cm)</label>
            <input type="number" value={width} onChange={(e) => setWidth(Number(e.target.value))} />
          </div>
          <div className="package-input-box">
            <label>Height (cm)</label>
            <input type="number" value={height} onChange={(e) => setHeight(Number(e.target.value))} />
          </div>
        </div>
        
        <div className="package-input-box" style={{ marginBottom: '3rem', border: '1px solid var(--color-primary)' }}>
          <label style={{ color: 'var(--color-primary)' }}>Weight (kg)</label>
          <input type="number" value={weight} onChange={(e) => setWeight(Number(e.target.value))} />
        </div>

        <button className="btn btn--primary btn--lg btn--full" onClick={handleCalculate}>
          Calculate Rate <span style={{ marginLeft: '8px' }}>→</span>
        </button>

        {showResults && (
          <div className="results-banner">
            <div className="recommend-label">
              <span>✔</span> Recommended Option
            </div>
            
            <div className="result-row">
              <div>
                <div className="result-name">Standard</div>
                <div style={{ opacity: 0.6, fontSize: '0.9rem', marginTop: '0.25rem' }}>
                  📅 Estimated delivery: 3-5 days
                </div>
              </div>
              <div className="result-price">Rp 12.000</div>
            </div>

            <div className="result-row" style={{ padding: '1rem 0', opacity: 0.8 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                <span style={{ fontSize: '1.5rem' }}>⚡</span>
                <div>
                  <div style={{ fontWeight: 600 }}>Express</div>
                  <div style={{ fontSize: '0.8rem', opacity: 0.6 }}>Next Day Delivery</div>
                </div>
              </div>
              <div style={{ fontWeight: 700 }}>Rp 25.000</div>
            </div>

            <div className="result-row" style={{ padding: '1rem 0', opacity: 0.8 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                <span style={{ fontSize: '1.5rem' }}>📦</span>
                <div>
                  <div style={{ fontWeight: 600 }}>Cargo</div>
                  <div style={{ fontSize: '0.8rem', opacity: 0.6 }}>5-7 days for large items</div>
                </div>
              </div>
              <div style={{ fontWeight: 700 }}>Rp 45.000</div>
            </div>
          </div>
        )}
      </div>
    </section>
  );
};
