import React from "react";

export const Footer = () => {
  return (
    <footer className="footer">
      <div className="footer__container">
        <div className="footer__brand">
          <span className="nav__name">NusaRoute</span>
          <p className="footer__tagline">Pengiriman Terpercaya untuk Seluruh Nusantara 🇮🇩</p>
        </div>
        <div className="footer__links">
          <div className="footer__col">
            <h4>Layanan</h4>
            <a href="#">Reguler (REG)</a>
            <a href="#">Yakin Esok Sampai (YES)</a>
            <a href="#">Same Day</a>
            <a href="#">Kargo</a>
          </div>
          <div className="footer__col">
            <h4>Perusahaan</h4>
            <a href="#">Tentang Kami</a>
            <a href="#">Karir</a>
            <a href="#">Blog</a>
          </div>
          <div className="footer__col">
            <h4>Bantuan</h4>
            <a href="#">FAQ</a>
            <a href="#">Hubungi Kami</a>
            <a href="#">Syarat & Ketentuan</a>
          </div>
        </div>
      </div>
      <div className="footer__bottom">
        <p>&copy; 2026 NusaRoute. Hak Cipta Dilindungi.</p>
      </div>
    </footer>
  );
};
