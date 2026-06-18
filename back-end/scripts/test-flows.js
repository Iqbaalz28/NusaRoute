// test-flows.js — functional (end-to-end) tests for the two business flows.
// Requires all microservices running via docker-compose. Run: node test-flows.js
//
// Flow 1: admin buat pesanan → generate resi (AWB) → notif ke user → sampai selesai (DELIVERED)
// Flow 2: paket dijemput → jika sekota + layanan instant → kirim langsung (tanpa hub);
//         jika tidak → lewat hub dulu.

const GW = "http://localhost:8080/api/v1";
const ORDER = "http://localhost:8004/api/v1";
const PAYMENT = "http://localhost:8002/api/v1";
const DISPATCH = "http://localhost:8006/api/v1";
const HUB = "http://localhost:8007/api/v1";
const TRACKING = "http://localhost:8008/api/v1";
const NOTIF = "http://localhost:8009/api/v1";

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
let pass = 0, fail = 0;
function check(name, cond) {
  if (cond) { pass++; console.log(`   ✓ ${name}`); }
  else { fail++; console.log(`   ✗ FAIL: ${name}`); }
}

async function req(method, url, body, headers = {}) {
  const opts = { method, headers: { ...headers } };
  if (body !== undefined) { opts.headers["Content-Type"] = "application/json"; opts.body = JSON.stringify(body); }
  const res = await fetch(url, opts);
  const text = await res.text();
  let data; try { data = JSON.parse(text); } catch { data = text; }
  return { ok: res.ok, status: res.status, data };
}
const post = (u, b, h) => req("POST", u, b, h);
const get = (u, h) => req("GET", u, undefined, h);

async function getEvents(awb) {
  const t = await get(`${TRACKING}/tracking/${awb}`);
  const tl = t.ok && t.data ? t.data.data : null;
  if (!tl) return [];
  return Array.isArray(tl) ? tl : (tl.events || []);
}
async function pollUntil(fn, predicate, timeoutMs = 18000, stepMs = 1500) {
  const start = Date.now();
  let last = await fn();
  while (!predicate(last) && Date.now() - start < timeoutMs) {
    await sleep(stepMs);
    last = await fn();
  }
  return last;
}

let token, userId;

async function login() {
  const r = await post(`${GW}/auth/login`, { email: "demo@nusaroute.com", password: "password123" });
  if (!r.ok) throw new Error("login failed: " + JSON.stringify(r.data));
  token = r.data.data.token;
  userId = r.data.data.user.id;
}

function authHeaders() { return { Authorization: `Bearer ${token}`, "X-User-ID": userId }; }

async function createOrder(over) {
  const base = {
    sender_name: "Budi", sender_phone: "0812", sender_address: "Jakarta", sender_lat: -6.2088, sender_lng: 106.8456,
    receiver_name: "Siti", receiver_phone: "0813", receiver_address: "Bandung", receiver_lat: -6.9175, receiver_lng: 107.6191,
    item_description: "Barang Uji", weight_kg: 2, length_cm: 10, width_cm: 10, height_cm: 10,
    service_type: "REGULAR", is_insured: false, insured_value: 0, shipping_cost: 25000, insurance_cost: 0, total_cost: 25000,
  };
  const r = await post(`${ORDER}/orders`, { ...base, ...over }, authHeaders());
  if (!r.ok) throw new Error("create order failed: " + JSON.stringify(r.data));
  return r.data.data;
}

async function pay(orderId, amount) {
  await post(`${PAYMENT}/payments/initiate`, { order_id: orderId, amount, method: "VA" });
  await post(`${PAYMENT}/payments/webhook`, { order_id: orderId, status: "PAID", amount, idempotency_key: `pay_${orderId}_VA` });
}

async function pickupWithRetry(orderId) {
  for (let i = 0; i < 12; i++) {
    await sleep(1500);
    const r = await post(`${DISPATCH}/dispatch/pickup`, { order_id: orderId, lat: -6.2, lng: 106.8 });
    if (r.ok) return true;
  }
  return false;
}

// ---------------------------------------------------------------------------
async function flow1() {
  console.log("\n=== FLOW 1: admin buat pesanan → resi → notif → selesai ===");
  const order = await createOrder({ service_type: "REGULAR" });
  check("AWB ter-generate berformat NR...", /^NR/.test(order.awb));
  check("status awal PENDING_PAYMENT", order.status === "PENDING_PAYMENT");

  await pay(order.id, order.total_cost);
  const picked = await pickupWithRetry(order.id);
  check("kurir ter-assign & paket dijemput", picked);

  // VIA_HUB order → scan hub
  const hubs = (await get(`${HUB}/hub/list`)).data.data || [];
  const h = hubs[0];
  await post(`${HUB}/hub/scan/inbound`, { awb: order.awb, hub_id: h.id, operator_id: "OP" });
  await post(`${HUB}/hub/scan/outbound`, { awb: order.awb, hub_id: h.id, operator_id: "OP" });
  await post(`${DISPATCH}/dispatch/deliver`, { order_id: order.id, receiver_name: "Siti", lat: -6.9, lng: 107.6 });

  const events = await pollUntil(() => getEvents(order.awb), (ev) => ev.some((e) => e.status === "DELIVERED"));
  const statuses = events.map((e) => e.status);
  check("tracking mencapai DELIVERED", statuses.includes("DELIVERED"));
  check("timeline punya PICKED_UP", statuses.includes("PICKED_UP"));

  // notifications (sent on courier-assigned / picked-up / delivered)
  const notifs = await pollUntil(
    async () => (await get(`${NOTIF}/notifications`)).data?.data || [],
    (list) => list.some((n) => n.awb === order.awb && /diterima|Terkirim/i.test(n.message || "")),
    12000
  );
  check("notifikasi terkirim untuk AWB ini", notifs.some((n) => n.awb === order.awb));
  check("ada notif 'Paket Terkirim' ke user", notifs.some((n) => n.awb === order.awb && /Terkirim|diterima/i.test(n.message || "")));
}

// ---------------------------------------------------------------------------
async function flow2() {
  console.log("\n=== FLOW 2a: sekota + instant → kirim LANGSUNG (tanpa hub) ===");
  // Intra-Jakarta, SAMEDAY → DIRECT
  const direct = await createOrder({
    service_type: "SAMEDAY",
    receiver_address: "Jakarta", receiver_lat: -6.1751, receiver_lng: 106.8650,
  });
  check("delivery_mode = DIRECT (sekota + instant)", direct.delivery_mode === "DIRECT");
  await pay(direct.id, direct.total_cost);
  const pickedD = await pickupWithRetry(direct.id);
  check("paket dijemput", pickedD);
  // Direct → langsung deliver, TANPA scan hub
  await post(`${DISPATCH}/dispatch/deliver`, { order_id: direct.id, receiver_name: "Siti", lat: -6.175, lng: 106.865 });
  const evD = await pollUntil(() => getEvents(direct.awb), (ev) => ev.some((e) => e.status === "DELIVERED"));
  const stD = evD.map((e) => e.status);
  check("DIRECT: tidak ada IN_TRANSIT (tidak lewat hub)", !stD.includes("IN_TRANSIT"));
  check("DIRECT: PICKED_UP lalu DELIVERED", stD.includes("PICKED_UP") && stD.includes("DELIVERED"));

  console.log("\n=== FLOW 2b: beda kota / non-instant → lewat HUB dulu ===");
  const viaHub = await createOrder({ service_type: "REGULAR" }); // Jakarta→Bandung
  check("delivery_mode = VIA_HUB (beda kota)", viaHub.delivery_mode === "VIA_HUB");
  await pay(viaHub.id, viaHub.total_cost);
  const pickedH = await pickupWithRetry(viaHub.id);
  check("paket dijemput", pickedH);
  const hubs = (await get(`${HUB}/hub/list`)).data.data || [];
  const h = hubs[0];
  await post(`${HUB}/hub/scan/inbound`, { awb: viaHub.awb, hub_id: h.id, operator_id: "OP" });
  await post(`${HUB}/hub/scan/outbound`, { awb: viaHub.awb, hub_id: h.id, operator_id: "OP" });
  await post(`${DISPATCH}/dispatch/deliver`, { order_id: viaHub.id, receiver_name: "Siti", lat: -6.9, lng: 107.6 });
  const evH = await pollUntil(() => getEvents(viaHub.awb), (ev) => ev.some((e) => e.status === "DELIVERED"));
  const stH = evH.map((e) => e.status);
  check("VIA_HUB: ada IN_TRANSIT (lewat hub)", stH.includes("IN_TRANSIT"));
  check("VIA_HUB: berakhir DELIVERED", stH.includes("DELIVERED"));
}

async function main() {
  await login();
  console.log(`🔑 Login OK (${userId})`);
  await flow1();
  await flow2();
  console.log(`\n================ HASIL: ${pass} lulus, ${fail} gagal ================`);
  process.exit(fail > 0 ? 1 : 0);
}

main().catch((e) => { console.error("ERROR:", e.message); process.exit(1); });
