import React from "react";
import { Hero } from "./Hero";
import { Stats } from "./Stats";
import { VolumeChart } from "./VolumeChart";
import { LiveFeed } from "./LiveFeed";
import { ServiceComparison } from "./ServiceComparison";

interface DashboardPageProps {
  onStartShipping: () => void;
  onTrackPackage: () => void;
}

export const DashboardPage: React.FC<DashboardPageProps> = ({ onStartShipping, onTrackPackage }) => {
  return (
    <section className="page page--active">
      <Hero onStartShipping={onStartShipping} onTrackPackage={onTrackPackage} />
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
      <div className="card card--glass" style={{ maxWidth: "1280px", margin: "0 auto 2rem", padding: "0 1.5rem" }}>
        <div className="card__header">
          <h2 className="card__title">🚚 Perbandingan Layanan</h2>
        </div>
        <div className="card__body">
          <ServiceComparison />
        </div>
      </div>
    </section>
  );
};
