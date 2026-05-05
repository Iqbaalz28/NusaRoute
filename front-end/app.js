/* ================================================
   NusaRoute Frontend — Application Logic
   ================================================ */

const API_BASE = 'http://localhost:8080';

// ================================================
// Navigation & Routing
// ================================================
document.addEventListener('DOMContentLoaded', () => {
  initNavigation();
  initCounters();
  initChart();
  initEventFeed();
  initServiceComparison();
  initHubGrid();
  initOrders();
  initTracking();
  initPricingCalc();
  initModal();
  initScrollEffects();
});

function initNavigation() {
  const links = document.querySelectorAll('.nav__link');
  const pages = document.querySelectorAll('.page');
  const burger = document.getElementById('navBurger');
  const navLinks = document.getElementById('navLinks');

  links.forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      const target = link.dataset.page;
      links.forEach(l => l.classList.remove('nav__link--active'));
      link.classList.add('nav__link--active');
      pages.forEach(p => p.classList.remove('page--active'));
      document.getElementById('page' + capitalize(target)).classList.add('page--active');
      navLinks.classList.remove('nav__links--open');
    });
  });

  burger.addEventListener('click', () => navLinks.classList.toggle('nav__links--open'));

  document.getElementById('btnTrackHero').addEventListener('click', () => {
    document.querySelector('[data-page="tracking"]').click();
  });
  document.getElementById('btnStartShipping').addEventListener('click', () => {
    document.querySelector('[data-page="services"]').click();
  });
}

function capitalize(s) { return s.charAt(0).toUpperCase() + s.slice(1); }

// ================================================
// Scroll Effects
// ================================================
function initScrollEffects() {
  window.addEventListener('scroll', () => {
    const nav = document.getElementById('mainNav');
    nav.classList.toggle('nav--scrolled', window.scrollY > 20);
  });
}

// ================================================
// Animated Counters
// ================================================
function initCounters() {
  const counters = document.querySelectorAll('.stat-card__value');
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        animateCounter(entry.target);
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.5 });
  counters.forEach(c => observer.observe(c));
}

function animateCounter(el) {
  const target = parseInt(el.dataset.count);
  const suffix = el.dataset.suffix || '';
  const duration = 2000;
  const start = performance.now();

  function update(now) {
    const elapsed = now - start;
    const progress = Math.min(elapsed / duration, 1);
    const eased = 1 - Math.pow(1 - progress, 3);
    el.textContent = Math.floor(target * eased).toLocaleString('id-ID') + suffix;
    if (progress < 1) requestAnimationFrame(update);
  }
  requestAnimationFrame(update);
}

// ================================================
// Chart (Canvas-based)
// ================================================
function initChart() {
  const canvas = document.getElementById('chartVolume');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  canvas.width = canvas.offsetWidth * dpr;
  canvas.height = canvas.offsetHeight * dpr;
  ctx.scale(dpr, dpr);

  const W = canvas.offsetWidth, H = canvas.offsetHeight;
  const data = [12400, 14200, 11800, 15600, 16800, 14900, 15847];
  const labels = ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min'];
  const maxVal = Math.max(...data) * 1.15;
  const pad = { top: 20, right: 20, bottom: 40, left: 60 };
  const chartW = W - pad.left - pad.right;
  const chartH = H - pad.top - pad.bottom;

  // Grid
  ctx.strokeStyle = 'rgba(255,255,255,0.06)';
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const y = pad.top + (chartH / 4) * i;
    ctx.beginPath(); ctx.moveTo(pad.left, y); ctx.lineTo(W - pad.right, y); ctx.stroke();
    const val = Math.round(maxVal - (maxVal / 4) * i);
    ctx.fillStyle = '#94a3b8'; ctx.font = '11px Inter'; ctx.textAlign = 'right';
    ctx.fillText(val.toLocaleString('id-ID'), pad.left - 8, y + 4);
  }

  // Area gradient
  const gradient = ctx.createLinearGradient(0, pad.top, 0, H - pad.bottom);
  gradient.addColorStop(0, 'rgba(99,102,241,0.3)');
  gradient.addColorStop(1, 'rgba(99,102,241,0)');

  const points = data.map((v, i) => ({
    x: pad.left + (chartW / (data.length - 1)) * i,
    y: pad.top + chartH - (v / maxVal) * chartH
  }));

  // Fill area
  ctx.beginPath();
  ctx.moveTo(points[0].x, H - pad.bottom);
  points.forEach(p => ctx.lineTo(p.x, p.y));
  ctx.lineTo(points[points.length - 1].x, H - pad.bottom);
  ctx.closePath();
  ctx.fillStyle = gradient; ctx.fill();

  // Line
  ctx.beginPath();
  ctx.moveTo(points[0].x, points[0].y);
  for (let i = 1; i < points.length; i++) {
    const cp1x = (points[i - 1].x + points[i].x) / 2;
    ctx.bezierCurveTo(cp1x, points[i - 1].y, cp1x, points[i].y, points[i].x, points[i].y);
  }
  ctx.strokeStyle = '#6366f1'; ctx.lineWidth = 2.5; ctx.stroke();

  // Dots
  points.forEach((p, i) => {
    ctx.beginPath(); ctx.arc(p.x, p.y, 4, 0, Math.PI * 2);
    ctx.fillStyle = i === points.length - 1 ? '#06b6d4' : '#6366f1';
    ctx.fill();
    ctx.strokeStyle = 'rgba(10,10,26,0.8)'; ctx.lineWidth = 2; ctx.stroke();
  });

  // Labels
  labels.forEach((l, i) => {
    ctx.fillStyle = '#94a3b8'; ctx.font = '12px Inter'; ctx.textAlign = 'center';
    ctx.fillText(l, points[i].x, H - pad.bottom + 20);
  });
}

// ================================================
// Live Event Feed
// ================================================
function initEventFeed() {
  const feed = document.getElementById('eventFeed');
  const events = [
    { icon: '✅', text: 'Paket NR8847291034 telah diterima oleh Siti Aminah di Jakarta Selatan', time: '2 detik lalu' },
    { icon: '🛵', text: 'Kurir Rudi Hartono ditugaskan untuk menjemput paket NR7762830192', time: '15 detik lalu' },
    { icon: '📦', text: 'Paket NR6621940283 tiba di Hub Surabaya (SBY-01) — scan inbound', time: '32 detik lalu' },
    { icon: '💳', text: 'Pembayaran Rp 45.500 dikonfirmasi untuk order NR5529174620', time: '1 menit lalu' },
    { icon: '🔄', text: 'Paket NR4418293710 disortir di Hub Bandung (BDG-01)', time: '2 menit lalu' },
    { icon: '⚠️', text: 'Pengiriman NR3307182649 gagal — percobaan ke-2 dari 3', time: '3 menit lalu' },
    { icon: '🚚', text: 'Paket NR2296071538 berangkat dari Hub Semarang ke Hub Jakarta', time: '5 menit lalu' },
  ];

  function render() {
    feed.innerHTML = events.map(e =>
      `<div class="event-item">
        <span class="event-item__icon">${e.icon}</span>
        <div>
          <div class="event-item__text">${e.text}</div>
          <div class="event-item__time">${e.time}</div>
        </div>
      </div>`
    ).join('');
  }
  render();

  // Simulate live updates
  const newEvents = [
    { icon: '✅', text: 'Paket NR9934827162 telah diterima oleh Ahmad di Medan', time: 'baru saja' },
    { icon: '📦', text: 'Paket NR1123948271 tiba di Hub Makassar (MKS-01)', time: 'baru saja' },
    { icon: '🛵', text: 'Kurir Diana ditugaskan pickup NR8812736451 di Bandung', time: 'baru saja' },
  ];
  let idx = 0;
  setInterval(() => {
    events.unshift(newEvents[idx % newEvents.length]);
    events.pop();
    idx++;
    render();
  }, 5000);
}

// ================================================
// Service Comparison
// ================================================
function initServiceComparison() {
  const container = document.getElementById('serviceComparison');
  const services = [
    { icon: '📦', name: 'Reguler (REG)', desc: '2-4 hari kerja', price: 'Mulai Rp 8.000' },
    { icon: '⚡', name: 'YES', desc: 'Yakin Esok Sampai', price: 'Mulai Rp 15.000' },
    { icon: '🚀', name: 'Same Day', desc: '< 12 jam', price: 'Mulai Rp 25.000' },
    { icon: '🏗️', name: 'Kargo', desc: '5-7 hari, barang besar', price: 'Mulai Rp 5.000' },
  ];
  container.innerHTML = services.map(s =>
    `<div class="svc-card">
      <div class="svc-card__icon">${s.icon}</div>
      <div class="svc-card__name">${s.name}</div>
      <div class="svc-card__desc">${s.desc}</div>
      <div class="svc-card__price">${s.price}</div>
    </div>`
  ).join('');
}

// ================================================
// Hub Network Grid
// ================================================
function initHubGrid() {
  const grid = document.getElementById('hubGrid');
  const hubs = [
    { name: 'Hub Jakarta Pusat', code: 'JKT-01', city: 'Jakarta, DKI Jakarta', type: 'SORTATION', lat: -6.1751, lng: 106.8650 },
    { name: 'Hub Bandung', code: 'BDG-01', city: 'Bandung, Jawa Barat', type: 'SORTATION', lat: -6.9175, lng: 107.6191 },
    { name: 'Hub Surabaya', code: 'SBY-01', city: 'Surabaya, Jawa Timur', type: 'SORTATION', lat: -7.2575, lng: 112.7521 },
    { name: 'Hub Semarang', code: 'SMG-01', city: 'Semarang, Jawa Tengah', type: 'TRANSIT', lat: -6.9666, lng: 110.4196 },
    { name: 'Hub Medan', code: 'MDN-01', city: 'Medan, Sumatera Utara', type: 'SORTATION', lat: 3.5952, lng: 98.6722 },
    { name: 'Hub Makassar', code: 'MKS-01', city: 'Makassar, Sulawesi Selatan', type: 'SORTATION', lat: -5.1477, lng: 119.4327 },
    { name: 'Hub Denpasar', code: 'DPS-01', city: 'Denpasar, Bali', type: 'TRANSIT', lat: -8.6705, lng: 115.2126 },
    { name: 'Hub Yogyakarta', code: 'YK-01', city: 'Yogyakarta, DI Yogyakarta', type: 'DISTRIBUTION', lat: -7.7956, lng: 110.3695 },
  ];
  grid.innerHTML = hubs.map(h =>
    `<div class="hub-card">
      <div class="hub-card__name">${h.name}</div>
      <div class="hub-card__city">${h.city}</div>
      <span class="hub-card__type">${h.type}</span>
      <div class="hub-card__coords">${h.lat.toFixed(4)}, ${h.lng.toFixed(4)}</div>
    </div>`
  ).join('');
}

// ================================================
// Orders Table (Demo Data)
// ================================================
function initOrders() {
  const tbody = document.getElementById('ordersTableBody');
  const orders = [
    { awb: 'NR8847291034', sender: 'Budi S.', receiver: 'Siti A.', service: 'YES', status: 'DELIVERED', cost: 47500, date: '05 Mei 2026' },
    { awb: 'NR7762830192', sender: 'Rina W.', receiver: 'Ahmad K.', service: 'REG', status: 'IN_TRANSIT', cost: 18000, date: '04 Mei 2026' },
    { awb: 'NR6621940283', sender: 'Deni P.', receiver: 'Maya L.', service: 'SAME', status: 'IN_TRANSIT', cost: 85000, date: '04 Mei 2026' },
    { awb: 'NR5529174620', sender: 'Farhan R.', receiver: 'Lisa N.', service: 'REG', status: 'PENDING_PAYMENT', cost: 22500, date: '04 Mei 2026' },
    { awb: 'NR4418293710', sender: 'Galih M.', receiver: 'Putri H.', service: 'CARGO', status: 'IN_TRANSIT', cost: 125000, date: '03 Mei 2026' },
    { awb: 'NR3307182649', sender: 'Hadi S.', receiver: 'Wulan D.', service: 'YES', status: 'DELIVERY_FAILED', cost: 35000, date: '03 Mei 2026' },
    { awb: 'NR2296071538', sender: 'Ivan T.', receiver: 'Joko P.', service: 'REG', status: 'DELIVERED', cost: 15500, date: '02 Mei 2026' },
    { awb: 'NR1185960427', sender: 'Kiki A.', receiver: 'Lani B.', service: 'SAME', status: 'CANCELLED', cost: 92000, date: '01 Mei 2026' },
  ];

  function renderOrders(filter) {
    const filtered = filter === 'all' ? orders : orders.filter(o => o.status === filter);
    tbody.innerHTML = filtered.map(o => {
      const badge = statusBadge(o.status);
      return `<tr>
        <td><strong>${o.awb}</strong></td>
        <td>${o.sender}</td><td>${o.receiver}</td>
        <td>${o.service}</td>
        <td>${badge}</td>
        <td>Rp ${o.cost.toLocaleString('id-ID')}</td>
        <td>${o.date}</td>
      </tr>`;
    }).join('');
  }

  renderOrders('all');

  document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('filter-btn--active'));
      btn.classList.add('filter-btn--active');
      renderOrders(btn.dataset.status);
    });
  });
}

function statusBadge(status) {
  const map = {
    PENDING_PAYMENT: ['badge--pending', 'Menunggu Bayar'],
    IN_TRANSIT: ['badge--transit', 'Dalam Perjalanan'],
    DELIVERED: ['badge--delivered', 'Terkirim'],
    CANCELLED: ['badge--cancelled', 'Dibatalkan'],
    DELIVERY_FAILED: ['badge--cancelled', 'Gagal Kirim'],
  };
  const [cls, label] = map[status] || ['', status];
  return `<span class="badge badge--status ${cls}">${label}</span>`;
}

// ================================================
// Tracking
// ================================================
function initTracking() {
  document.getElementById('btnTrack').addEventListener('click', doTrack);
  document.getElementById('trackingInput').addEventListener('keydown', e => { if (e.key === 'Enter') doTrack(); });
}

function doTrack() {
  const awb = document.getElementById('trackingInput').value.trim();
  if (!awb) return;

  const result = document.getElementById('trackingResult');
  result.style.display = 'block';

  // Demo tracking data
  const events = [
    { status: 'Paket Diterima', detail: 'Paket telah diterima oleh Siti Aminah', location: 'Jakarta Selatan', time: '05 Mei 2026, 14:32 WIB' },
    { status: 'Sedang Diantar', detail: 'Kurir Rudi sedang menuju alamat penerima', location: 'Jakarta Selatan', time: '05 Mei 2026, 13:45 WIB' },
    { status: 'Paket Tiba di Hub Jakarta', detail: 'Paket disortir di Hub Jakarta Pusat (JKT-01)', location: 'Hub Jakarta Pusat', time: '05 Mei 2026, 08:12 WIB' },
    { status: 'Berangkat dari Hub Bandung', detail: 'Paket berangkat menuju Hub Jakarta', location: 'Hub Bandung', time: '04 Mei 2026, 22:30 WIB' },
    { status: 'Paket Tiba di Hub Bandung', detail: 'Paket tiba dan discan di Hub Bandung (BDG-01)', location: 'Hub Bandung', time: '04 Mei 2026, 18:15 WIB' },
    { status: 'Paket Dijemput', detail: 'Paket dijemput oleh kurir dari alamat pengirim', location: 'Bandung', time: '04 Mei 2026, 15:20 WIB' },
    { status: 'Kurir Ditugaskan', detail: 'Kurir Ahmad ditugaskan untuk penjemputan', location: '-', time: '04 Mei 2026, 14:55 WIB' },
    { status: 'Pembayaran Dikonfirmasi', detail: 'Pembayaran Rp 47.500 via VA berhasil', location: '-', time: '04 Mei 2026, 14:30 WIB' },
  ];

  result.innerHTML = `
    <div class="card card--glass">
      <div class="card__header">
        <h2 class="card__title">📦 ${awb}</h2>
        <span class="badge badge--status badge--delivered">Terkirim</span>
      </div>
      <div class="card__body">
        <div class="timeline">
          ${events.map((e, i) => `
            <div class="timeline__item" style="animation-delay: ${i * 0.1}s">
              <div class="timeline__dot"></div>
              <div class="timeline__status">${e.status}</div>
              <div class="timeline__detail">${e.detail}</div>
              <div class="timeline__time">📍 ${e.location} · ${e.time}</div>
            </div>
          `).join('')}
        </div>
      </div>
    </div>`;
}

// ================================================
// Pricing Calculator
// ================================================
function initPricingCalc() {
  document.getElementById('btnCalcPrice').addEventListener('click', calculatePrice);
}

function calculatePrice() {
  const origin = document.getElementById('calcOrigin').value.split(',').map(Number);
  const dest = document.getElementById('calcDest').value.split(',').map(Number);
  const weight = parseFloat(document.getElementById('calcWeight').value) || 1;
  const length = parseFloat(document.getElementById('calcLength').value) || 1;
  const width = parseFloat(document.getElementById('calcWidth').value) || 1;
  const height = parseFloat(document.getElementById('calcHeight').value) || 1;

  const distKm = haversine(origin[0], origin[1], dest[0], dest[1]);
  const volWeight = (length * width * height) / 6000;
  const chargeKg = Math.max(weight, volWeight);

  // Zone multiplier
  let multiplier = 1.0;
  if (distKm > 5000) multiplier = 2.5;
  else if (distKm > 800) multiplier = 2.0;
  else if (distKm > 150) multiplier = 1.5;
  else if (distKm > 30) multiplier = 1.2;

  const services = [
    { code: 'REG', name: 'Reguler', priceKm: 30, priceKg: 2500, base: 8000, est: '2-4 hari', recommended: true },
    { code: 'YES', name: 'Yakin Esok Sampai', priceKm: 80, priceKg: 5000, base: 15000, est: '1 hari', recommended: false },
    { code: 'SAME', name: 'Same Day', priceKm: 150, priceKg: 8000, base: 25000, est: '< 12 jam', recommended: false },
    { code: 'CARGO', name: 'Kargo', priceKm: 15, priceKg: 1500, base: 5000, est: '5-7 hari', recommended: false },
  ];

  const container = document.getElementById('priceResults');
  container.innerHTML = services.map(s => {
    const total = Math.ceil((s.base + chargeKg * s.priceKg * multiplier + distKm * s.priceKm) / 500) * 500;
    return `
      <div class="price-card ${s.recommended ? 'price-card--recommended' : ''}">
        <div class="price-card__name">${s.name}</div>
        <div class="price-card__est">⏱️ ${s.est}</div>
        <div class="price-card__price">Rp ${total.toLocaleString('id-ID')}</div>
        <div class="price-card__details">
          <div><span>Jarak</span><span>${distKm.toFixed(1)} km</span></div>
          <div><span>Berat aktual</span><span>${weight} kg</span></div>
          <div><span>Berat volumetrik</span><span>${volWeight.toFixed(1)} kg</span></div>
          <div><span>Dikenakan</span><span>${chargeKg.toFixed(1)} kg</span></div>
          <div><span>Zone</span><span>×${multiplier}</span></div>
        </div>
      </div>`;
  }).join('');
}

function haversine(lat1, lng1, lat2, lng2) {
  const R = 6371;
  const dLat = (lat2 - lat1) * Math.PI / 180;
  const dLng = (lng2 - lng1) * Math.PI / 180;
  const a = Math.sin(dLat / 2) ** 2 + Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) * Math.sin(dLng / 2) ** 2;
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
}

// ================================================
// Modal
// ================================================
function initModal() {
  const modal = document.getElementById('loginModal');
  document.getElementById('btnLogin').addEventListener('click', () => modal.classList.add('modal--open'));
  document.getElementById('modalClose').addEventListener('click', () => modal.classList.remove('modal--open'));
  document.getElementById('modalBackdrop').addEventListener('click', () => modal.classList.remove('modal--open'));

  document.getElementById('btnSubmitLogin').addEventListener('click', () => {
    const email = document.getElementById('loginEmail').value;
    const pass = document.getElementById('loginPassword').value;
    if (!email || !pass) return alert('Email dan password harus diisi');
    // In production: call API Gateway → User Service
    alert('Login berhasil! (Demo Mode)');
    modal.classList.remove('modal--open');
  });
}
