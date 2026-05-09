import React from "react";
import { Hero } from "./Hero";
import { Stats } from "./Stats";
import { VolumeChart } from "./VolumeChart";
import { LiveFeed } from "./LiveFeed";
import { ServiceComparison } from "./ServiceComparison";

interface DashboardPageProps {
  onStartShipping: () => void;
  onTrackPackage: (awb: string) => void;
  onNavigate: (page: string) => void;
}

export const DashboardPage: React.FC<DashboardPageProps> = ({ onStartShipping, onTrackPackage, onNavigate }) => {
  return (
    <section className="page page--active" style={{ paddingBottom: '6rem' }}>
      <Hero onTrack={onTrackPackage} onNavigate={onNavigate} />
      
      <Stats />

      <div className="dashboard-grid">
        <div className="card card--glass">
          <div className="card__header">
            <h2 className="card__title">📊 Volume Pengiriman (7 Hari)</h2>
          </div>
          <div className="card__body">
            <VolumeChart />
          </div>
        </div>
        
        <div className="card card--glass">
          <div className="card__header">
            <h2 className="card__title">🔴 Live Event Feed</h2>
            <span className="badge badge--live">LIVE</span>
          </div>
          <LiveFeed />
        </div>
      </div>

      <div className="card card--glass" style={{ maxWidth: "1280px", margin: "2rem auto", padding: "0 1.5rem" }}>
        <div className="card__header">
          <h2 className="card__title">🚚 Perbandingan Layanan</h2>
        </div>
        <div className="card__body">
          <ServiceComparison />
        </div>
      </div>

      <div className="estimate-cta">
        <div className="estimate-card">
          <div>
            <h2 className="estimate-card__title">Need a quick estimate?</h2>
            <p className="estimate-card__text">
              Calculate shipping costs instantly based on destination and package details.
            </p>
          </div>
          <button className="btn btn--primary btn--lg" onClick={onStartShipping}>
            <span style={{ marginRight: '8px' }}>📋</span> Calculate Rate
          </button>
        </div>
      </div>
    </section>
  );
};
