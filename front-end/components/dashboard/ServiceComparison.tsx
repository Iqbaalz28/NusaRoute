import React from "react";

export const ServiceComparison = () => {
  const services = [
    { 
      icon: "🚚", 
      name: "Standar", 
      desc: "Pengiriman andal dan hemat biaya untuk kebutuhan sehari-hari. Pengiriman dalam 3-5 hari kerja.",
      isPopular: false,
      iconType: "soft"
    },
    { 
      icon: "⚡", 
      name: "Ekspres", 
      desc: "Pengiriman tercepat untuk paket mendesak dan peka waktu. Dijamin sampai esok hari.",
      isPopular: true,
      iconType: "solid"
    },
    { 
      icon: "📦", 
      name: "Kargo", 
      desc: "Solusi pengiriman kargo berat dan volume besar untuk bisnis dan pengirim massal.",
      isPopular: false,
      iconType: "soft"
    },
  ];

  return (
    <>
      <h2 className="text-center font-heading font-bold text-[2.5rem] mb-2 leading-tight">Pilih kecepatan Anda.</h2>
      <p className="text-center text-muted mb-16">Tingkatan pengiriman fleksibel yang dirancang untuk memenuhi kebutuhan pengiriman spesifik Anda.</p>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-10 max-w-[1200px] mx-auto px-8 mb-24">
        {services.map((s, i) => (
          <div className="bg-surface rounded-[32px] p-12 shadow-[0_10px_40px_rgba(45,49,66,0.04)] transition-all duration-300 hover:-translate-y-2 hover:shadow-[0_20px_60px_rgba(45,49,66,0.08)] flex flex-col relative border border-black/5" key={i}>
            {s.isPopular && (
              <span className="absolute -top-3 right-6 bg-primary text-white px-4 py-1.5 rounded-xl text-[0.75rem] font-extrabold uppercase tracking-widest shadow-[0_4px_12px_rgba(255,107,74,0.3)]">
                Populer
              </span>
            )}
            <div className={`w-16 h-16 rounded-full flex items-center justify-center text-2xl mb-10 ${s.iconType === 'solid' ? 'bg-primary text-white' : 'bg-[#FFF0ED] text-primary'}`}>
              {s.icon}
            </div>
            <h3 className="text-[1.75rem] font-bold mb-5">{s.name}</h3>
            <p className="text-muted text-base mb-10 flex-1 leading-relaxed">{s.desc}</p>
            <a href="#" className="font-heading font-semibold text-primary text-[1.1rem] flex items-center gap-2 group">
              Pelajari selengkapnya <span className="transition-transform group-hover:translate-x-1">→</span>
            </a>
          </div>
        ))}
      </div>
    </>
  );
};
