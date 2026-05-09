import React from "react";

export const HubsPage = () => {
  const hubs = [
    { name: "Hub Jakarta Pusat", code: "JKT-01", city: "Jakarta, DKI Jakarta", type: "SORTATION", lat: -6.1751, lng: 106.865 },
    { name: "Hub Bandung", code: "BDG-01", city: "Bandung, Jawa Barat", type: "SORTATION", lat: -6.9175, lng: 107.6191 },
    { name: "Hub Surabaya", code: "SBY-01", city: "Surabaya, Jawa Timur", type: "SORTATION", lat: -7.2575, lng: 112.7521 },
    { name: "Hub Semarang", code: "SMG-01", city: "Semarang, Jawa Tengah", type: "TRANSIT", lat: -6.9666, lng: 110.4196 },
    { name: "Hub Medan", code: "MDN-01", city: "Medan, Sumatera Utara", type: "SORTATION", lat: 3.5952, lng: 98.6722 },
    { name: "Hub Makassar", code: "MKS-01", city: "Makassar, Sulawesi Selatan", type: "SORTATION", lat: -5.1477, lng: 119.4327 },
    { name: "Hub Denpasar", code: "DPS-01", city: "Denpasar, Bali", type: "TRANSIT", lat: -8.6705, lng: 115.2126 },
    { name: "Hub Yogyakarta", code: "YK-01", city: "Yogyakarta, DI Yogyakarta", type: "DISTRIBUTION", lat: -7.7956, lng: 110.3695 },
  ];

  return (
    <section className="page page--active" style={{ paddingTop: '8rem' }}>
      <div className="page__header" style={{ maxWidth: '1200px', margin: '0 auto', textAlign: 'center' }}>
        <h1 className="page__title">Jaringan Hub Lumina</h1>
        <p className="page__subtitle">Hub operasional strategis untuk memastikan paket Anda sampai tepat waktu.</p>
      </div>
      
      <div className="card-grid" style={{ marginTop: '4rem', display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '2rem', maxWidth: '1200px', margin: '4rem auto' }}>
        {hubs.map((h, i) => (
          <div className="service-card" key={i} style={{ textAlign: 'center' }}>
            <div className="service-card__icon" style={{ margin: '0 auto 2rem' }}>📍</div>
            <h3 className="service-card__name">{h.name}</h3>
            <p className="service-card__desc">{h.city}</p>
            <span className="status-pill status-pill--transit" style={{ marginBottom: '1rem', display: 'inline-block' }}>{h.type}</span>
            <div style={{ fontSize: '0.8rem', color: 'var(--color-muted)', fontFamily: 'monospace' }}>
              {h.lat.toFixed(4)}, {h.lng.toFixed(4)}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
};
