// seed-lifecycle.js
// Drives the FULL package lifecycle through the real service APIs so that Kafka
// events flow naturally and the tracking timeline fills end-to-end:
//
//   create order -> pay (initiate + webhook PAID) -> [auto-assign courier]
//   -> courier pickup -> hub scans (in/sort/out) -> courier deliver
//
// Run (with all microservices up via docker-compose):  node seed-lifecycle.js
//
// Requires Node 18+ (global fetch).

const GW = "http://localhost:8080/api/v1"; // api-gateway (login)
const ORDER = "http://localhost:8004/api/v1";
const PAYMENT = "http://localhost:8002/api/v1";
const DISPATCH = "http://localhost:8006/api/v1";
const HUB = "http://localhost:8007/api/v1";
const TRACKING = "http://localhost:8008/api/v1";

const N = 3; // number of packages to push through the full lifecycle

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function req(method, url, body, headers = {}) {
  const opts = { method, headers: { ...headers } };
  if (body !== undefined) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(url, opts);
  const text = await res.text();
  let data;
  try { data = JSON.parse(text); } catch { data = text; }
  return { ok: res.ok, status: res.status, data };
}
const post = (url, body, headers) => req("POST", url, body, headers);
const get = (url, headers) => req("GET", url, undefined, headers);

async function main() {
  console.log("🌱 Seeding FULL lifecycle (order → bayar → assign → jemput → hub → antar)\n");

  // 1) Login (to obtain user id + token for order creation)
  const login = await post(`${GW}/auth/login`, { email: "demo@nusaroute.com", password: "password123" });
  if (!login.ok) { console.error("❌ Login gagal:", login.status, login.data); process.exit(1); }
  const token = login.data.data.token;
  const userId = login.data.data.user.id;
  console.log(`🔑 Login OK sebagai ${login.data.data.user.full_name} (${userId})`);

  // Hubs for the transit scans
  const hubsRes = await get(`${HUB}/hub/list`);
  const hubs = (hubsRes.data && hubsRes.data.data) || [];
  if (hubs.length === 0) { console.error("❌ Tidak ada hub. Jalankan seeding hub dulu."); process.exit(1); }
  const originHub = hubs[0];
  const destHub = hubs[1] || hubs[0];

  const awbs = [];

  for (let i = 0; i < N; i++) {
    console.log(`\n=== Paket ${i + 1}/${N} ===`);

    // 2) Create order
    const orderBody = {
      sender_name: "Budi Santoso", sender_phone: "08123456789",
      sender_address: "Jl. Merdeka No. 10, Jakarta", sender_lat: -6.2, sender_lng: 106.8,
      receiver_name: "Siti Aminah", receiver_phone: "08987654321",
      receiver_address: "Jl. Asia Afrika No. 5, Bandung", receiver_lat: -6.9175, receiver_lng: 107.6191,
      item_description: `Paket demo ${i + 1}`, weight_kg: 2, length_cm: 20, width_cm: 15, height_cm: 10,
      service_type: "REGULAR", is_insured: false, insured_value: 0,
      shipping_cost: 25000, insurance_cost: 0, total_cost: 25000,
    };
    const ord = await post(`${ORDER}/orders`, orderBody, { Authorization: `Bearer ${token}`, "X-User-ID": userId });
    if (!ord.ok) { console.error("  ❌ create order:", ord.status, ord.data); continue; }
    const order = ord.data.data;
    const orderId = order.id;
    const awb = order.awb;
    console.log(`  ✅ Order dibuat: ${awb}`);

    // 3) Payment initiate
    const method = "VA";
    const init = await post(`${PAYMENT}/payments/initiate`, { order_id: orderId, amount: order.total_cost, method });
    if (!init.ok) { console.error("  ❌ initiate payment:", init.status, init.data); continue; }

    // 4) Payment webhook PAID  → emits payment.success
    const wh = await post(`${PAYMENT}/payments/webhook`, {
      order_id: orderId, status: "PAID", amount: order.total_cost,
      idempotency_key: `pay_${orderId}_${method}`,
    });
    if (!wh.ok) { console.error("  ❌ webhook:", wh.status, wh.data); continue; }
    console.log("  💰 Pembayaran dikonfirmasi (payment.success → order.ready-for-pickup → auto-assign)");

    // 5) Wait for courier auto-assignment, then pickup (retry until assignment exists)
    let picked = false;
    for (let attempt = 1; attempt <= 12; attempt++) {
      await sleep(1500);
      const pu = await post(`${DISPATCH}/dispatch/pickup`, { order_id: orderId, lat: -6.2, lng: 106.8 });
      if (pu.ok) { console.log(`  🛵 Kurir menjemput paket — PICKED_UP (≈${(attempt * 1.5).toFixed(1)}s)`); picked = true; break; }
    }
    if (!picked) { console.error("  ❌ pickup gagal: kurir belum ter-assign (cek log dispatch-service)"); continue; }

    // 6) Hub scans → IN_TRANSIT (origin hub: masuk → sortir → keluar; tujuan: masuk)
    const op = "OP-JKT-01";
    await post(`${HUB}/hub/scan/inbound`, { awb, order_id: orderId, hub_id: originHub.id, operator_id: op });
    await post(`${HUB}/hub/scan/sort`, { awb, order_id: orderId, hub_id: originHub.id, operator_id: op });
    await post(`${HUB}/hub/scan/outbound`, { awb, order_id: orderId, hub_id: originHub.id, operator_id: op });
    await post(`${HUB}/hub/scan/inbound`, { awb, order_id: orderId, hub_id: destHub.id, operator_id: op });
    console.log(`  🏢 Scan hub: ${originHub.name} (masuk→sortir→keluar) → ${destHub.name} (masuk)`);

    // 7) Deliver → DELIVERED
    await sleep(500);
    const dl = await post(`${DISPATCH}/dispatch/deliver`, { order_id: orderId, receiver_name: "Siti Aminah", lat: -6.9175, lng: 107.6191 });
    if (dl.ok) console.log("  📬 Paket diterima penerima — DELIVERED");
    else console.error("  ❌ deliver:", dl.status, dl.data);

    awbs.push(awb);
  }

  // 8) Let the outbox workers + Kafka consumers catch up, then print timelines
  console.log("\n⏳ Menunggu event Kafka diproses (6s)...");
  await sleep(6000);

  for (const awb of awbs) {
    const t = await get(`${TRACKING}/tracking/${awb}`);
    const tl = t.ok && t.data ? t.data.data : null;
    const events = tl ? (Array.isArray(tl) ? tl : tl.events || []) : [];
    if (events.length === 0) {
      console.log(`\n📦 ${awb} — belum ada tracking (HTTP ${t.status})`);
      continue;
    }
    const status = Array.isArray(tl) ? events[events.length - 1].status : (tl.status || tl.current_status);
    console.log(`\n📦 ${awb} — status: ${status}`);
    events
      .slice()
      .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
      .forEach((e) => console.log(`   • ${String(e.status).padEnd(16)} @ ${e.location}  — ${e.detail || ""}`));
  }

  console.log("\n✅ Selesai. Lacak salah satu AWB di atas pada halaman Tracking.");
}

main().catch((e) => { console.error(e); process.exit(1); });
