"use client";

import React, { useState } from "react";

export const ServicesPage = () => {
  const [origin, setOrigin] = useState("-6.917,107.619");
  const [dest, setDest] = useState("-6.175,106.865");
  const [weight, setWeight] = useState(2);
  const [length, setLength] = useState(30);
  const [width, setWidth] = useState(20);
  const [height, setHeight] = useState(15);
  const [results, setResults] = useState<any[]>([]);

  const calculatePrice = () => {
    const originCoords = origin.split(",").map(Number);
    const destCoords = dest.split(",").map(Number);
    
    const distKm = haversine(originCoords[0], originCoords[1], destCoords[0], destCoords[1]);
    const volWeight = (length * width * height) / 6000;
    const chargeKg = Math.max(weight, volWeight);

    let multiplier = 1.0;
    if (distKm > 5000) multiplier = 2.5;
    else if (distKm > 800) multiplier = 2.0;
    else if (distKm > 150) multiplier = 1.5;
    else if (distKm > 30) multiplier = 1.2;

    const services = [
      { code: "REG", name: "Reguler", priceKm: 30, priceKg: 2500, base: 8000, est: "2-4 hari", recommended: true },
      { code: "YES", name: "Yakin Esok Sampai", priceKm: 80, priceKg: 5000, base: 15000, est: "1 hari", recommended: false },
      { code: "SAME", name: "Same Day", priceKm: 150, priceKg: 8000, base: 25000, est: "< 12 jam", recommended: false },
      { code: "CARGO", name: "Kargo", priceKm: 15, priceKg: 1500, base: 5000, est: "5-7 hari", recommended: false },
    ];

    const calculated = services.map((s) => {
      const total = Math.ceil((s.base + chargeKg * s.priceKg * multiplier + distKm * s.priceKm) / 500) * 500;
      return { ...s, total, distKm, chargeKg, volWeight, multiplier };
    });

    setResults(calculated);
  };

  const haversine = (lat1: number, lng1: number, lat2: number, lng2: number) => {
    const R = 6371;
    const dLat = ((lat2 - lat1) * Math.PI) / 180;
    const dLng = ((lng2 - lng1) * Math.PI) / 180;
    const a = Math.sin(dLat / 2) ** 2 + Math.cos((lat1 * Math.PI) / 180) * Math.cos((lat2 * Math.PI) / 180) * Math.sin(dLng / 2) ** 2;
    return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  };

  const cities = [
    { label: "Bandung", value: "-6.917,107.619" },
    { label: "Jakarta", value: "-6.175,106.865" },
    { label: "Surabaya", value: "-7.257,112.752" },
    { label: "Semarang", value: "-6.966,110.419" },
    { label: "Yogyakarta", value: "-7.795,110.369" },
    { label: "Medan", value: "3.595,98.672" },
    { label: "Makassar", value: "-5.147,119.432" },
    { label: "Denpasar", value: "-8.670,115.212" },
  ];

  return (
    <section className="page page--active">
      <div className="page__header">
        <h1 className="page__title">Layanan Pengiriman</h1>
        <p className="page__subtitle">Pilih layanan terbaik sesuai kebutuhan Anda.</p>
      </div>
      <div className="pricing-calculator">
        <div className="card card--glass">
          <div className="card__header">
            <h2 className="card__title">🧮 Kalkulator Ongkos Kirim</h2>
          </div>
          <div className="card__body">
            <div className="calc-form">
              <div className="calc-form__row">
                <div className="form-group">
                  <label className="label">Kota Asal</label>
                  <select className="select" value={origin} onChange={(e) => setOrigin(e.target.value)}>
                    {cities.map((c) => (
                      <option key={c.label} value={c.value}>
                        {c.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label className="label">Kota Tujuan</label>
                  <select className="select" value={dest} onChange={(e) => setDest(e.target.value)}>
                    {cities.map((c) => (
                      <option key={c.label} value={c.value}>
                        {c.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="calc-form__row">
                <div className="form-group">
                  <label className="label">Berat (kg)</label>
                  <input
                    type="number"
                    className="input"
                    value={weight}
                    onChange={(e) => setWeight(parseFloat(e.target.value))}
                    min="0.1"
                    step="0.1"
                  />
                </div>
                <div className="form-group">
                  <label className="label">Panjang (cm)</label>
                  <input
                    type="number"
                    className="input"
                    value={length}
                    onChange={(e) => setLength(parseFloat(e.target.value))}
                    min="1"
                  />
                </div>
                <div className="form-group">
                  <label className="label">Lebar (cm)</label>
                  <input
                    type="number"
                    className="input"
                    value={width}
                    onChange={(e) => setWidth(parseFloat(e.target.value))}
                    min="1"
                  />
                </div>
                <div className="form-group">
                  <label className="label">Tinggi (cm)</label>
                  <input
                    type="number"
                    className="input"
                    value={height}
                    onChange={(e) => setHeight(parseFloat(e.target.value))}
                    min="1"
                  />
                </div>
              </div>
              <button className="btn btn--primary btn--lg btn--full" onClick={calculatePrice}>
                Hitung Ongkos Kirim
              </button>
            </div>
          </div>
        </div>
        <div className="price-results">
          {results.map((s, i) => (
            <div className={`price-card ${s.recommended ? "price-card--recommended" : ""}`} key={i}>
              <div className="price-card__name">{s.name}</div>
              <div className="price-card__est">⏱️ {s.est}</div>
              <div className="price-card__price">Rp {s.total.toLocaleString("id-ID")}</div>
              <div className="price-card__details">
                <div>
                  <span>Jarak</span>
                  <span>{s.distKm.toFixed(1)} km</span>
                </div>
                <div>
                  <span>Berat aktual</span>
                  <span>{weight} kg</span>
                </div>
                <div>
                  <span>Berat volumetrik</span>
                  <span>{s.volWeight.toFixed(1)} kg</span>
                </div>
                <div>
                  <span>Dikenakan</span>
                  <span>{s.chargeKg.toFixed(1)} kg</span>
                </div>
                <div>
                  <span>Zone</span>
                  <span>×{s.multiplier}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};
