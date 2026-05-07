"use client";

import React, { useState, useEffect } from "react";

interface NavbarProps {
  activePage: string;
  setActivePage: (page: string) => void;
  onLoginClick: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activePage, setActivePage, onLoginClick }) => {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const navLinks = [
    { id: "tracking", label: "Track" },
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

        <div className="nav__links">
          {navLinks.map((link) => (
            <button
              key={link.id}
              className={`nav__link ${activePage === link.id ? "nav__link--active" : ""}`}
              onClick={() => setActivePage(link.id)}
            >
              {link.label}
            </button>
          ))}
          <button className="btn btn--primary" onClick={onLoginClick}>
            Sign In
          </button>
        </div>
      </div>
    </nav>
  );
};
