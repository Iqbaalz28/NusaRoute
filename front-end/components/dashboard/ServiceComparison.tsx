import React from "react";

export const ServiceComparison = () => {
  const services = [
    { 
      icon: "🚚", 
      name: "Standard", 
      desc: "Reliable and cost-effective shipping for everyday needs. Delivery in 3-5 business days.",
      isPopular: false,
      iconType: "soft"
    },
    { 
      icon: "⚡", 
      name: "Express", 
      desc: "Fastest delivery for urgent and time-sensitive packages. Next day guaranteed.",
      isPopular: true,
      iconType: "solid"
    },
    { 
      icon: "📦", 
      name: "Cargo", 
      desc: "Heavy freight and large volume shipping solutions for businesses and bulk senders.",
      isPopular: false,
      iconType: "soft"
    },
  ];

  return (
    <>
      <h2 className="section-title">Choose your speed.</h2>
      <p className="section-subtitle">Flexible shipping tiers designed to meet your specific delivery needs.</p>
      
      <div className="card-grid">
        {services.map((s, i) => (
          <div className="service-card" key={i}>
            {s.isPopular && <span className="badge-popular">Popular</span>}
            <div className={`service-card__icon service-card__icon--${s.iconType}`}>
              {s.icon}
            </div>
            <h3 className="service-card__name">{s.name}</h3>
            <p className="service-card__desc">{s.desc}</p>
            <a href="#" className="service-card__link">
              Learn more <span>→</span>
            </a>
          </div>
        ))}
      </div>
    </>
  );
};
