"use client";

import React, { useState, useEffect } from "react";

interface NavbarProps {
  activePage: string;
  setActivePage: (page: string) => void;
  onLoginClick: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activePage, setActivePage, onLoginClick }) => {
  const [scrolled, setScrolled] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const navLinks = [
    { id: "dashboard", label: "Dashboard" },
    { id: "tracking", label: "Lacak Paket" },
    { id: "orders", label: "Pesanan" },
    { id: "services", label: "Layanan" },
    { id: "hubs", label: "Hub Network" },
  ];

  return (
    <nav className={`nav ${scrolled ? "nav--scrolled" : ""}`} id="mainNav">
      <div className="nav__container">
        <div className="nav__brand">
          <div className="nav__logo">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
              <circle cx="16" cy="16" r="14" stroke="url(#grad)" strokeWidth="2.5" fill="none" />
              <path d="M10 16L14 20L22 12" stroke="url(#grad)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
              <defs>
                <linearGradient id="grad" x1="0" y1="0" x2="32" y2="32">
                  <stop stopColor="#6366f1" />
                  <stop offset="1" stopColor="#06b6d4" />
                </linearGradient>
              </defs>
            </svg>
          </div>
          <span className="nav__name">NusaRoute</span>
        </div>

        <div className={`nav__links ${isMenuOpen ? "nav__links--open" : ""}`} id="navLinks">
          {navLinks.map((link) => (
            <button
              key={link.id}
              className={`nav__link ${activePage === link.id ? "nav__link--active" : ""}`}
              onClick={() => {
                setActivePage(link.id);
                setIsMenuOpen(false);
              }}
            >
              {link.label}
            </button>
          ))}
        </div>

        <div className="nav__actions">
          <button className="btn btn--ghost" id="btnTheme" aria-label="Toggle theme">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="10" cy="10" r="4" />
              <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.93 4.93l1.41 1.41M13.66 13.66l1.41 1.41M4.93 15.07l1.41-1.41M13.66 6.34l1.41-1.41" />
            </svg>
          </button>
          <button className="btn btn--primary" id="btnLogin" onClick={onLoginClick}>
            Masuk
          </button>
        </div>

        <button className="nav__burger" id="navBurger" aria-label="Menu" onClick={() => setIsMenuOpen(!isMenuOpen)}>
          <span></span>
          <span></span>
          <span></span>
        </button>
      </div>
    </nav>
  );
};
