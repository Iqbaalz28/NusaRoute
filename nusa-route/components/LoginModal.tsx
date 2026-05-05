"use client";

import React, { useState } from "react";

interface LoginModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const LoginModal: React.FC<LoginModalProps> = ({ isOpen, onClose }) => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = () => {
    if (!email || !password) {
      return alert("Email dan password harus diisi");
    }
    alert("Login berhasil! (Demo Mode)");
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className={`modal ${isOpen ? "modal--open" : ""}`} id="loginModal">
      <div className="modal__backdrop" onClick={onClose}></div>
      <div className="modal__content card card--glass">
        <div className="modal__header">
          <h2 className="card__title">Masuk ke NusaRoute</h2>
          <button className="modal__close" onClick={onClose}>
            &times;
          </button>
        </div>
        <div className="modal__body">
          <div className="form-group">
            <label className="label">Email</label>
            <input
              type="email"
              className="input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="nama@email.com"
            />
          </div>
          <div className="form-group">
            <label className="label">Password</label>
            <input
              type="password"
              className="input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
            />
          </div>
          <button className="btn btn--primary btn--full" onClick={handleSubmit}>
            Masuk
          </button>
          <p className="modal__footer-text">
            Belum punya akun? <a href="#">Daftar sekarang</a>
          </p>
        </div>
      </div>
    </div>
  );
};
