import React from "react";

interface HeroProps {
  onStartShipping: () => void;
  onTrackPackage: () => void;
}

export const Hero: React.FC<HeroProps> = ({ onStartShipping, onTrackPackage }) => {
  return (
    <div className="hero">
      <div className="hero__bg">
        <div className="hero__orb hero__orb--1"></div>
        <div className="hero__orb hero__orb--2"></div>
        <div className="hero__orb hero__orb--3"></div>
      </div>
      <div className="hero__content">
        <div className="hero__badge">🚀 Platform Pengiriman Terdepan</div>
        <h1 className="hero__title">
          Kirim Paket ke Seluruh <span className="gradient-text">Nusantara</span>
        </h1>
        <p className="hero__subtitle">
          Jangkau 514 kota/kabupaten di Indonesia dengan layanan pengiriman terpercaya, cepat, dan aman. Dilengkapi tracking real-time berbasis GPS.
        </p>
        <div className="hero__cta">
          <button className="btn btn--primary btn--lg" onClick={onStartShipping}>
            Mulai Kirim Paket
          </button>
          <button className="btn btn--glass btn--lg" onClick={onTrackPackage}>
            Lacak Paket
          </button>
        </div>
      </div>
    </div>
  );
};
