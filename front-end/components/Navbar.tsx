"use client";

import React, { useState, useEffect } from "react";
import { useAuth } from "@/lib/AuthContext";
import { canAccessPage } from "@/lib/access";

interface NavbarProps {
  activePage: string;
  setActivePage: (page: string) => void;
  onLoginClick: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activePage, setActivePage, onLoginClick }) => {
  const { isAuthenticated, user, logout } = useAuth();
  const [scrolled, setScrolled] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const displayName = user?.full_name || user?.email || "Akun";

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const navLinks = [
    { id: "dashboard", label: "Dashboard" },
    { id: "tracking", label: "Lacak" },
    { id: "orders", label: "Pesanan" },
    { id: "services", label: "Layanan" },
    { id: "courier", label: "Kurir" },
    { id: "hubs", label: "Hub" },
    { id: "admin", label: "Admin" },
  ].filter((link) => canAccessPage(link.id, isAuthenticated, user?.role));

  return (
    <nav className={`fixed top-0 left-0 right-0 z-[100] transition-all duration-300 ${scrolled ? "bg-surface border-b border-border shadow-soft" : "bg-background/80 backdrop-blur-md"}`} id="mainNav">
      <div className="max-w-[1200px] mx-auto px-8 flex items-center justify-between h-20">
        <div 
          className="flex items-center gap-3 cursor-pointer" 
          onClick={() => setActivePage("dashboard")}
        >
          <div className="w-6 h-6 bg-primary rounded-[4px] rotate-45"></div>
          <span className="font-heading font-bold text-xl text-text">NusaRoute</span>
        </div>

        <div className={`flex gap-8 max-md:hidden ${isMenuOpen ? "max-md:flex max-md:absolute max-md:top-20 max-md:left-0 max-md:right-0 max-md:bg-surface max-md:flex-col max-md:p-8 max-md:border-b max-md:border-border max-md:shadow-soft" : ""}`} id="navLinks">
          {navLinks.map((link) => (
            <button
              key={link.id}
              className={`font-body text-[0.95rem] font-medium transition-all duration-300 px-4 py-2 rounded-lg ${activePage === link.id ? "text-primary bg-primary/5 opacity-100" : "text-text opacity-80 hover:opacity-100 hover:text-primary hover:bg-primary/5"}`}
              onClick={() => {
                setActivePage(link.id);
                setIsMenuOpen(false);
              }}
            >
              {link.label}
            </button>
          ))}
          {isAuthenticated ? (
            <div className="flex items-center gap-3 max-md:flex-col max-md:items-stretch">
              <span className="flex items-center gap-2 font-body text-[0.9rem] font-semibold text-text whitespace-nowrap">
                <span className="w-8 h-8 rounded-full bg-primary/15 text-primary flex items-center justify-center font-heading font-bold uppercase">
                  {displayName.charAt(0)}
                </span>
                {displayName}
              </span>
              <button
                className="bg-primary/10 text-primary font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary/20 transition-all duration-300 whitespace-nowrap"
                onClick={() => { logout(); setIsMenuOpen(false); }}
              >
                Keluar
              </button>
            </div>
          ) : (
            <button className="bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover hover:-translate-y-0.5 hover:shadow-[0_8px_16px_rgba(255,107,74,0.2)] transition-all duration-300 whitespace-nowrap" onClick={onLoginClick}>
              Masuk
            </button>
          )}
        </div>

        <div className="flex items-center">
          <button className="p-2 text-text opacity-80 hover:opacity-100 transition-opacity" id="btnTheme" aria-label="Toggle theme">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="10" cy="10" r="4" />
              <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.93 4.93l1.41 1.41M13.66 13.66l1.41 1.41M4.93 15.07l1.41-1.41M13.66 6.34l1.41-1.41" />
            </svg>
          </button>
        </div>

        <button className="hidden max-md:flex flex-col gap-1.5 p-2 bg-none border-none cursor-pointer" id="navBurger" aria-label="Menu" onClick={() => setIsMenuOpen(!isMenuOpen)}>
          <span className={`block w-[22px] h-[0.5] bg-text transition-all duration-300 rounded-[2px] ${isMenuOpen ? "rotate-45 translate-y-2" : ""}`}></span>
          <span className={`block w-[22px] h-[0.5] bg-text transition-all duration-300 rounded-[2px] ${isMenuOpen ? "opacity-0" : ""}`}></span>
          <span className={`block w-[22px] h-[0.5] bg-text transition-all duration-300 rounded-[2px] ${isMenuOpen ? "-rotate-45 -translate-y-2" : ""}`}></span>
        </button>
      </div>
    </nav>
  );
};
