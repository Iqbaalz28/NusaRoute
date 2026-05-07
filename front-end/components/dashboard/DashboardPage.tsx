import React from "react";
import { Hero } from "./Hero";
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
      
      <ServiceComparison />

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
