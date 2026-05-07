"use client";

import React, { useState } from "react";

interface HeroProps {
  onTrack: (awb: string) => void;
  onNavigate: (page: string) => void;
}

export const Hero: React.FC<HeroProps> = ({ onTrack, onNavigate }) => {
  const [activeTab, setActiveTab] = useState<"track" | "hub" | "rate">("track");
  const [awb, setAwb] = useState("");

  const handleTrack = () => {
    if (awb.trim()) {
      onTrack(awb);
    }
  };

  return (
    <div className="hero">
      {/* Background Ambient Glows */}
      <div 
        style={{ 
          position: 'absolute', top: '20%', left: '50%', transform: 'translateX(-50%)',
          width: '600px', height: '600px', 
          background: 'radial-gradient(circle, rgba(255,107,74,0.06) 0%, transparent 70%)', 
          filter: 'blur(60px)', zIndex: 0 
        }} 
      />

      {/* Content */}
      <div className="animate-fade-up" style={{ position: 'relative', zIndex: 1, maxWidth: '800px', width: '100%' }}>
        <h1 className="hero__title" style={{ fontSize: 'clamp(2.5rem, 6vw, 4rem)', maxWidth: 'none', marginBottom: '1rem' }}>
          Shipping, <span style={{ color: 'var(--color-primary)' }}>Simplified.</span>
        </h1>
        
        <p className="hero__subtitle" style={{ fontSize: '1.125rem', maxWidth: '560px', margin: '0 auto 2.5rem' }}>
          Lumina delivers your parcels across the archipelago with speed, 
          transparency, and care.
        </p>
        
        {/* Compact Action Hub */}
        <div 
          className="glass" 
          style={{ 
            borderRadius: '32px', padding: '1.5rem', margin: '0 auto', maxWidth: '640px',
            border: '1px solid rgba(255,255,255,0.6)',
            boxShadow: '0 10px 30px rgba(45, 49, 66, 0.06)'
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem', marginBottom: '1.5rem' }}>
            <button 
              className={`nav__link ${activeTab === 'track' ? 'nav__link--active' : ''}`}
              onClick={() => setActiveTab('track')}
              style={{ padding: '0.5rem 1.25rem', borderRadius: '99px', fontSize: '0.9rem' }}
            >
              🔍 Track
            </button>
            <button 
              className={`nav__link ${activeTab === 'hub' ? 'nav__link--active' : ''}`}
              onClick={() => setActiveTab('hub')}
              style={{ padding: '0.5rem 1.25rem', borderRadius: '99px', fontSize: '0.9rem' }}
            >
              📍 Hubs
            </button>
            <button 
              className={`nav__link ${activeTab === 'rate' ? 'nav__link--active' : ''}`}
              onClick={() => setActiveTab('rate')}
              style={{ padding: '0.5rem 1.25rem', borderRadius: '99px', fontSize: '0.9rem' }}
            >
              💰 Rates
            </button>
          </div>

          <div style={{ display: 'flex', gap: '0.75rem', maxWidth: '560px', margin: '0 auto' }}>
            {activeTab === 'track' && (
              <>
                <input 
                  type="text" 
                  className="hero__track-input" 
                  placeholder="Enter AWB Number..."
                  value={awb}
                  onChange={(e) => setAwb(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleTrack()}
                  style={{ background: '#fff', borderRadius: '99px', flex: 1, height: '52px', fontSize: '1rem', padding: '0 1.5rem' }}
                />
                <button className="btn btn--primary" onClick={handleTrack} style={{ height: '52px', padding: '0 2rem' }}>Track</button>
              </>
            )}
            {activeTab === 'hub' && (
              <>
                <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#fff', borderRadius: '99px', padding: '0 1.5rem', height: '52px', fontSize: '0.9rem', color: 'var(--color-muted)' }}>
                  8 sorting centers across Indonesia
                </div>
                <button className="btn btn--primary" onClick={() => onNavigate('hubs')} style={{ height: '52px' }}>View</button>
              </>
            )}
            {activeTab === 'rate' && (
              <>
                <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#fff', borderRadius: '99px', padding: '0 1.5rem', height: '52px', fontSize: '0.9rem', color: 'var(--color-muted)' }}>
                  Rates starting from Rp 8.000
                </div>
                <button className="btn btn--primary" onClick={() => onNavigate('services')} style={{ height: '52px' }}>Check</button>
              </>
            )}
          </div>
        </div>

        {/* Minimal Stats Row */}
        <div style={{ display: 'flex', justifyContent: 'center', gap: '3rem', marginTop: '3rem' }}>
          <div>
            <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>514</div>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-muted)', fontWeight: 600 }}>CITIES</div>
          </div>
          <div style={{ width: '1px', height: '30px', background: 'var(--color-border)' }}></div>
          <div>
            <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>98.5%</div>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-muted)', fontWeight: 600 }}>SUCCESS</div>
          </div>
          <div style={{ width: '1px', height: '30px', background: 'var(--color-border)' }}></div>
          <div>
            <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>2.4k</div>
            <div style={{ fontSize: '0.7rem', color: 'var(--color-muted)', fontWeight: 600 }}>COURIERS</div>
          </div>
        </div>
      </div>
    </div>
  );
};
