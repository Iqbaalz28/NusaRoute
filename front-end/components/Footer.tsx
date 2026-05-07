import React from "react";

export const Footer = () => {
  return (
    <footer className="footer">
      <div className="footer__container">
        <div className="footer__brand">
          <div className="nav__brand">
            <div className="nav__logo-icon"></div>
            <span className="nav__name">Lumina</span>
          </div>
          <p className="footer__tagline">Shipping simplified for the Nusantara. Reliable, transparent, and frictionless.</p>
        </div>
        <div className="footer__links">
          <div className="footer__col">
            <h4>Layanan</h4>
            <a href="#">Standard</a>
            <a href="#">Express</a>
            <a href="#">Cargo</a>
            <a href="#">Overnight</a>
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
        <p>&copy; 2026 Lumina Courier. All rights reserved.</p>
      </div>
    </footer>
  );
};
