import React from "react";

export const Footer = () => {
  return (
    <footer className="bg-surface px-8 pt-24 pb-8 border-t border-border mt-16">
      <div className="max-w-[1200px] mx-auto grid grid-cols-1 md:grid-cols-[1.5fr_repeat(3,1fr)] gap-16 mb-24">
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-3">
            <div className="w-6 h-6 bg-primary rounded-[4px] rotate-45"></div>
            <span className="font-heading font-bold text-xl text-text">NusaRoute</span>
          </div>
          <p className="text-muted mt-4 max-w-[240px] leading-relaxed">
            Pengiriman yang disederhanakan untuk Nusantara. Andal, transparan, dan tanpa hambatan.
          </p>
        </div>
        <div className="flex flex-col gap-6">
          <h4 className="font-heading font-bold text-base text-text">Layanan</h4>
          <div className="flex flex-col gap-3">
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Standard</a>
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Express</a>
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Cargo</a>
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Overnight</a>
          </div>
        </div>
        <div className="flex flex-col gap-6">
          <h4 className="font-heading font-bold text-base text-text">Perusahaan</h4>
          <div className="flex flex-col gap-3">
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Tentang Kami</a>
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Karir</a>
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Blog</a>
          </div>
        </div>
        <div className="flex flex-col gap-6">
          <h4 className="font-heading font-bold text-base text-text">Bantuan</h4>
          <div className="flex flex-col gap-3">
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">FAQ</a>
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Hubungi Kami</a>
            <a href="#" className="text-muted hover:text-primary transition-colors text-[0.95rem]">Syarat & Ketentuan</a>
          </div>
        </div>
      </div>
      <div className="border-t border-border text-center py-6 text-[0.8rem] text-muted">
        <p>&copy; 2026 NusaRoute. Hak cipta dilindungi undang-undang.</p>
      </div>
    </footer>
  );
};
