import React from "react";

export const HubsPage = () => {
  const hubs = [
    { name: "Hub Jakarta Pusat", code: "JKT-01", city: "Jakarta, DKI Jakarta", type: "SORTATION", lat: -6.1751, lng: 106.865 },
    { name: "Hub Bandung", code: "BDG-01", city: "Bandung, Jawa Barat", type: "SORTATION", lat: -6.9175, lng: 107.6191 },
    { name: "Hub Surabaya", code: "SBY-01", city: "Surabaya, Jawa Timur", type: "SORTATION", lat: -7.2575, lng: 112.7521 },
    { name: "Hub Medan", code: "MDN-01", city: "Medan, Sumatera Utara", type: "SORTATION", lat: 3.5952, lng: 98.6722 },
  ];

  return (
    <section className="page page--active" style={{ paddingTop: '8rem' }}>
      <div className="page__header" style={{ maxWidth: '1200px', margin: '0 auto', textAlign: 'center' }}>
        <h1 className="page__title">Jaringan Hub Lumina</h1>
        <p className="page__subtitle">Hub operasional strategis untuk memastikan paket Anda sampai tepat waktu.</p>
      </div>
      
      <div className="card-grid" style={{ marginTop: '4rem' }}>
        {hubs.map((h, i) => (
          <div className="service-card" key={i} style={{ textAlign: 'center' }}>
            <div className="service-card__icon" style={{ margin: '0 auto 2rem' }}>📍</div>
            <h3 className="service-card__name">{h.name}</h3>
            <p className="service-card__desc">{h.city}</p>
            <span className="status-pill status-pill--transit" style={{ marginBottom: '1rem' }}>{h.type}</span>
            <div style={{ fontSize: '0.8rem', color: 'var(--color-muted)', fontFamily: 'monospace' }}>
              {h.lat.toFixed(4)}, {h.lng.toFixed(4)}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
};
