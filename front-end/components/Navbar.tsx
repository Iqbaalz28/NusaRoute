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
    { id: "tracking", label: "Track" },
    { id: "orders", label: "Orders" },
    { id: "services", label: "Services" },
    { id: "hubs", label: "Hubs" },
  ];

  return (
    <nav className={`nav ${scrolled ? "nav--scrolled" : ""}`} id="mainNav">
      <div className="nav__container">
        <div 
          className="nav__brand" 
          style={{ cursor: "pointer" }} 
          onClick={() => setActivePage("dashboard")}
        >
          <div className="nav__logo-icon"></div>
          <span className="nav__name">Lumina</span>
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
          <button className="btn btn--primary" onClick={onLoginClick}>
            Sign In
          </button>
        </div>

        <div className="nav__actions">
          <button className="btn btn--ghost" id="btnTheme" aria-label="Toggle theme">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="10" cy="10" r="4" />
              <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.93 4.93l1.41 1.41M13.66 13.66l1.41 1.41M4.93 15.07l1.41-1.41M13.66 6.34l1.41-1.41" />
            </svg>
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
