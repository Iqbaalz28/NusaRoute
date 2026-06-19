"use client";

import React, { useCallback, useEffect, useState } from "react";
import { apiGet, apiPost, apiPut } from "@/lib/api";
import { Hub } from "@/lib/types";
import { CITIES } from "@/lib/cities";

const HUB_TYPES = ["SORTATION", "TRANSIT", "DISTRIBUTION"];

const emptyForm = {
  name: "", code: "", city: "", province: "",
  lat: 0, lng: 0, type: "SORTATION", is_active: true,
};
type HubForm = typeof emptyForm;

export const AdminHubsManager = () => {
  const [hubs, setHubs] = useState<Hub[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  const [editing, setEditing] = useState<Hub | null>(null); // null = not open
  const [creating, setCreating] = useState(false);
  const [form, setForm] = useState<HubForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [formErr, setFormErr] = useState<string | null>(null);

  const fetchHubs = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setHubs((await apiGet<Hub[]>("/api/v1/hub/manage", { requiresAuth: true })) || []);
    } catch (err: any) {
      setError(err.message || "Gagal memuat hub.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchHubs(); }, [fetchHubs]);

  const openCreate = () => { setForm(emptyForm); setCreating(true); setEditing(null); setFormErr(null); };
  const openEdit = (h: Hub) => {
    setForm({ name: h.name, code: h.code, city: h.city, province: h.province, lat: h.lat, lng: h.lng, type: h.type, is_active: h.is_active });
    setEditing(h); setCreating(false); setFormErr(null);
  };
  const closeForm = () => { setCreating(false); setEditing(null); setFormErr(null); };

  // Quick-fill coordinates + city from the shared city list.
  const pickCity = (name: string) => {
    const c = CITIES.find((x) => x.name === name);
    if (c) setForm((f) => ({ ...f, city: c.name, lat: c.lat, lng: c.lng }));
  };

  const save = async () => {
    if (!form.name.trim() || !form.code.trim() || !form.city.trim()) {
      setFormErr("Nama, kode, dan kota wajib diisi.");
      return;
    }
    setSaving(true);
    setFormErr(null);
    try {
      if (editing) {
        await apiPut(`/api/v1/hub/manage/${editing.id}`, form, { requiresAuth: true });
      } else {
        await apiPost("/api/v1/hub/manage", form, { requiresAuth: true });
      }
      closeForm();
      await fetchHubs();
    } catch (err: any) {
      setFormErr(err.message || "Gagal menyimpan hub.");
    } finally {
      setSaving(false);
    }
  };

  const toggleActive = async (h: Hub) => {
    setBusyId(h.id);
    setError(null);
    try {
      await apiPut(`/api/v1/hub/manage/${h.id}`, { ...h, is_active: !h.is_active }, { requiresAuth: true });
      setHubs((prev) => prev.map((x) => (x.id === h.id ? { ...x, is_active: !x.is_active } : x)));
    } catch (err: any) {
      setError(err.message || "Gagal mengubah status hub.");
    } finally {
      setBusyId(null);
    }
  };

  const inputClass = "w-full h-11 bg-background border border-border rounded-xl px-4 text-base outline-none focus:border-primary transition-all";

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <p className="text-muted">Kelola hub sortir: tambah, ubah lokasi, aktif/nonaktifkan.</p>
        <button className="bg-primary text-white font-heading font-semibold px-6 py-3 rounded-pill hover:bg-primary-hover transition-all shrink-0" onClick={openCreate}>
          + Tambah Hub
        </button>
      </div>

      {error && <div className="text-red-500 text-sm font-bold bg-red-50 p-3 rounded-xl border border-red-100 mb-4">{error}</div>}

      <div className="bg-surface border border-border rounded-card shadow-soft overflow-hidden">
        {loading ? (
          <div className="h-64 flex items-center justify-center text-muted animate-pulse">Memuat hub...</div>
        ) : hubs.length === 0 ? (
          <div className="h-64 flex items-center justify-center text-muted">Belum ada hub.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-[0.9rem]">
              <thead>
                <tr className="bg-background/50 border-b border-border text-left text-[0.7rem] uppercase tracking-wider text-muted">
                  <th className="p-4">Hub</th>
                  <th className="p-4">Kota / Provinsi</th>
                  <th className="p-4">Tipe</th>
                  <th className="p-4">Koordinat</th>
                  <th className="p-4">Status</th>
                  <th className="p-4">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {hubs.map((h) => (
                  <tr key={h.id} className={`border-b border-border last:border-none hover:bg-background/30 ${!h.is_active ? "opacity-60" : ""}`}>
                    <td className="p-4"><span className="font-bold">{h.name}</span> <span className="text-muted">({h.code})</span></td>
                    <td className="p-4 text-muted">{h.city}{h.province ? `, ${h.province}` : ""}</td>
                    <td className="p-4 text-muted">{h.type}</td>
                    <td className="p-4 text-muted tabular-nums text-[0.8rem]">{h.lat.toFixed(3)}, {h.lng.toFixed(3)}</td>
                    <td className="p-4">
                      <span className={`text-[0.75rem] font-bold rounded-full px-3 py-1 ${h.is_active ? "bg-green-100 text-green-700" : "bg-gray-200 text-gray-600"}`}>
                        {h.is_active ? "Aktif" : "Nonaktif"}
                      </span>
                    </td>
                    <td className="p-4">
                      <div className="flex gap-3">
                        <button className="text-primary font-bold hover:underline" onClick={() => openEdit(h)}>Ubah</button>
                        <button
                          className={`font-bold hover:underline ${h.is_active ? "text-red-500" : "text-green-600"}`}
                          onClick={() => toggleActive(h)}
                          disabled={busyId === h.id}
                        >
                          {busyId === h.id ? "..." : h.is_active ? "Nonaktifkan" : "Aktifkan"}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {(creating || editing) && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-text/40 backdrop-blur-md" onClick={closeForm}></div>
          <div className="bg-surface w-full max-w-[560px] max-h-[90vh] overflow-y-auto rounded-[32px] p-10 relative z-10 shadow-hover animate-fade-up">
            <button onClick={closeForm} className="absolute top-6 right-7 text-muted hover:text-text text-2xl leading-none">×</button>
            <h2 className="text-2xl font-bold mb-6">{editing ? "Ubah Hub" : "Tambah Hub"}</h2>
            <div className="grid grid-cols-2 gap-3">
              <input className={inputClass} placeholder="Nama Hub (mis. Hub Jakarta Pusat)" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
              <input className={inputClass} placeholder="Kode (mis. JKT-01)" value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} />
              <div className="col-span-2">
                <label className="block text-[0.7rem] font-bold uppercase tracking-wider text-muted mb-1.5 ml-1">Isi cepat dari kota</label>
                <select className={inputClass} value="" onChange={(e) => pickCity(e.target.value)}>
                  <option value="">— pilih kota untuk isi koordinat —</option>
                  {CITIES.map((c) => <option key={c.name} value={c.name}>{c.name}</option>)}
                </select>
              </div>
              <input className={inputClass} placeholder="Kota" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
              <input className={inputClass} placeholder="Provinsi" value={form.province} onChange={(e) => setForm({ ...form, province: e.target.value })} />
              <input className={inputClass} type="number" step="any" placeholder="Lat" value={form.lat} onChange={(e) => setForm({ ...form, lat: Number(e.target.value) })} />
              <input className={inputClass} type="number" step="any" placeholder="Lng" value={form.lng} onChange={(e) => setForm({ ...form, lng: Number(e.target.value) })} />
              <select className={inputClass} value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}>
                {HUB_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
              </select>
              <label className="flex items-center gap-2 px-2 font-semibold text-[0.9rem]">
                <input type="checkbox" checked={form.is_active} onChange={(e) => setForm({ ...form, is_active: e.target.checked })} />
                Aktif
              </label>
            </div>
            {formErr && <div className="text-red-500 text-sm font-bold mt-4">{formErr}</div>}
            <button className="bg-primary text-white font-heading font-semibold w-full h-12 rounded-pill hover:bg-primary-hover mt-6 disabled:opacity-50" onClick={save} disabled={saving}>
              {saving ? "Menyimpan..." : editing ? "Simpan Perubahan" : "Tambah Hub"}
            </button>
          </div>
        </div>
      )}
    </>
  );
};
