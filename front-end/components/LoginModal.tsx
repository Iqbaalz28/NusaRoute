"use client";

import React, { useState } from "react";
import { apiPost } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import { AuthResponse } from "@/lib/types";

interface LoginModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const LoginModal: React.FC<LoginModalProps> = ({ isOpen, onClose }) => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const { login } = useAuth();

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) {
      setError("Email dan kata sandi harus diisi.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const data = await apiPost<AuthResponse>('/api/v1/auth/login', { email, password });
      login(data.token, data.user);
      onClose(); // Close modal on success
      
      // Optionally reset form
      setEmail("");
      setPassword("");
    } catch (err: any) {
      setError(err.message || "Gagal masuk. Periksa email dan kata sandi Anda.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center p-6">
      <div className="absolute inset-0 bg-text/40 backdrop-blur-md" onClick={onClose}></div>
      
      <div className="bg-surface w-full max-w-[440px] rounded-[32px] p-10 relative z-10 shadow-hover animate-fade-up">
        <button className="absolute top-8 right-8 text-muted hover:text-text transition-colors" onClick={onClose} disabled={loading}>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>

        <div className="mb-10">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-6 h-6 bg-primary rounded-[4px] rotate-45"></div>
            <span className="font-heading font-bold text-xl text-text">NusaRoute</span>
          </div>
          <h2 className="text-3xl font-bold mb-2">Selamat Datang</h2>
          <p className="text-muted">Masuk untuk mengelola pengiriman Anda.</p>
        </div>

        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <div>
            <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Email</label>
            <input 
              type="email" 
              className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all" 
              placeholder="nama@email.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={loading}
            />
          </div>
          <div>
            <label className="block text-[0.8rem] font-bold uppercase tracking-wider text-muted mb-2 ml-1">Kata Sandi</label>
            <input 
              type="password" 
              className="w-full h-14 bg-background border border-border rounded-2xl px-6 text-base font-body outline-none focus:border-primary focus:ring-4 focus:ring-primary/5 transition-all" 
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loading}
            />
          </div>
          
          <div className="flex justify-between items-center text-[0.9rem] mb-2">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" className="w-4 h-4 rounded border-border text-primary focus:ring-primary" disabled={loading} />
              <span className="text-muted">Ingat saya</span>
            </label>
            <a href="#" className="text-primary font-semibold hover:underline">Lupa sandi?</a>
          </div>

          {error && (
            <div className="text-red-500 text-sm font-bold bg-red-50 p-3 rounded-xl border border-red-100">
              {error}
            </div>
          )}

          <button 
            type="submit"
            className="bg-primary text-white font-heading font-semibold h-14 rounded-pill hover:bg-primary-hover hover:-translate-y-0.5 transition-all text-base mt-2 shadow-lg shadow-primary/20 flex items-center justify-center disabled:opacity-50"
            disabled={loading}
          >
            {loading ? "Memproses..." : "Masuk Sekarang"}
          </button>

          <p className="text-center text-muted text-[0.9rem] mt-4">
            Belum punya akun? <a href="#" className="text-primary font-semibold hover:underline">Daftar Gratis</a>
          </p>
        </form>
      </div>
    </div>
  );
};
