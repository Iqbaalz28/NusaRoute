// simulate-qr-chain.js
// Simulates the FULL two-leg hub-and-spoke QR/AWB chain end to end:
//
//   create order (AWB = the QR) -> admin READY_FOR_PICKUP
//   LEG 1 (first-mile): courier claim -> scan@sender -> scan@origin-hub (handover)
//   LINE-HAUL: operator sort + depart origin -> arrive dest hub (opens last-mile)
//   LEG 2 (last-mile): courier claim -> scan@dest-hub -> scan@receiver -> DELIVERED
//
// The same AWB is validated at every courier scan. Prereq: docker-compose up +
// role accounts seeded. Run: node simulate-qr-chain.js
//
// Pass "self" as an arg to simulate sender drop-off at the hub (no first-mile job):
//   node simulate-qr-chain.js self

const GW = "http://localhost:8080/api/v1";
const PASS = "password123";
const SELF_DROPOFF = process.argv[2] === "self";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function req(method, path, token, body) {
  const headers = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`${GW}${path}`, { method, headers, body: body ? JSON.stringify(body) : undefined });
  const text = await res.text();
  let data; try { data = JSON.parse(text); } catch { data = text; }
  return { ok: res.ok, status: res.status, data };
}
const login = async (email) => {
  const r = await req("POST", "/auth/login", null, { email, password: PASS });
  if (!r.ok) throw new Error(`login ${email} failed: ${r.status}`);
  return r.data.data.token;
};
const ok = (m) => console.log(`   ✅ ${m}`);
const fail = (m) => { console.log(`   ❌ ${m}`); process.exitCode = 1; };
const step = (m) => console.log(`\n${m}`);

// Poll the courier job board until a job for orderId with the given leg appears.
async function waitForJob(courier, orderId, leg) {
  for (let i = 0; i < 15; i++) {
    await sleep(2000);
    const r = await req("GET", "/dispatch/jobs", courier);
    const jobs = (r.ok && r.data.data) || [];
    const j = jobs.find((x) => x.order_id === orderId && x.leg === leg);
    if (j) return j;
  }
  return null;
}

(async () => {
  console.log(`🔗 Simulasi rantai QR dua-leg ${SELF_DROPOFF ? "(antar sendiri ke hub)" : "(dijemput kurir)"}\n`);

  const demo = await login("demo@nusaroute.com");
  const admin = await login("admin@nusaroute.com");
  const courier = await login("courier@nusaroute.com");
  const operator = await login("operator@nusaroute.com");
  await req("POST", "/couriers/ensure", courier, { full_name: "Kurir NusaRoute" });
  ok("login: demo, admin, courier, operator");

  const hubsRes = await req("GET", "/hub/list", demo);
  const hubs = (hubsRes.data && hubsRes.data.data) || [];

  // 1) Order (Jakarta → Bandung, via-hub)
  step("1) Pelanggan membuat pesanan");
  const ord = await req("POST", "/orders", demo, {
    sender_name: "Budi Santoso", sender_phone: "0812", sender_city: "Jakarta",
    sender_address: "Jl. Merdeka 10", sender_lat: -6.2, sender_lng: 106.8,
    receiver_name: "Siti Aminah", receiver_phone: "0899", receiver_city: "Bandung",
    receiver_address: "Jl. Asia Afrika 5", receiver_lat: -6.9175, receiver_lng: 107.6191,
    item_description: "Paket simulasi", weight_kg: 2, length_cm: 20, width_cm: 15, height_cm: 10,
    service_type: "REGULAR", pickup_mode: SELF_DROPOFF ? "SELF_DROPOFF" : "COURIER",
    is_insured: false, insured_value: 0, shipping_cost: 25000, insurance_cost: 0, total_cost: 25000,
  });
  if (!ord.ok) return fail(`create order: ${ord.status} ${JSON.stringify(ord.data)}`);
  const o = ord.data.data, orderId = o.id, AWB = o.awb;
  const hubOrigin = hubs.find((h) => h.code === o.origin_hub_code);
  const hubDest = hubs.find((h) => h.code === o.dest_hub_code);
  ok(`AWB ${AWB} | ${o.sender_city} → ${o.receiver_city} | hub ${o.origin_hub_name} → ${o.dest_hub_name}`);
  if (!hubOrigin || !hubDest) return fail("hub asal/tujuan tidak ditemukan di daftar hub");

  // 2) Ready
  step("2) Admin menandai READY_FOR_PICKUP");
  await req("PUT", "/orders/status", admin, { order_id: orderId, status: "READY_FOR_PICKUP" });
  ok("status → READY_FOR_PICKUP");

  // 3) LEG 1 — first-mile (sender → origin hub). Handover INTO the hub is always
  // scanned by the HUB OPERATOR (the receiving party scans), whether the package
  // arrived via courier or the sender dropped it off.
  if (SELF_DROPOFF) {
    step("3) LEG 1 dilewati — pengirim antar sendiri ke hub asal");
  } else {
    step("3) LEG 1 first-mile (kurir jemput pengirim → antar ke hub asal)");
    const job = await waitForJob(courier, orderId, "FIRST_MILE");
    if (!job) return fail("first-mile job tidak muncul");
    ok("first-mile job tersedia");
    await req("POST", "/dispatch/claim", courier, { order_id: orderId, courier_name: "Kurir NusaRoute" });
    ok("kurir klaim first-mile");
    const p = await req("POST", "/dispatch/pickup", courier, { order_id: orderId, awb: AWB });
    p.ok ? ok(`📷 kurir scan@pengirim → PICKED_UP`) : fail(`pickup: ${p.status} ${JSON.stringify(p.data)}`);
  }
  // Operator scans the package IN at the origin hub (serah ke hub → hub yang scan).
  // For courier first-mile this also closes the courier's first-mile leg.
  const inOrigin = await req("POST", "/hub/scan/inbound", operator, { awb: AWB, order_id: orderId, hub_id: hubOrigin.id, operator_id: "OP-SIM" });
  inOrigin.ok ? ok(`📷 operator scan masuk ${hubOrigin.name} (serah ke hub)`) : fail(`scan inbound asal: ${inOrigin.status}`);

  // 4) LINE-HAUL — operator sortir + keluar hub asal, lalu masuk hub tujuan
  step("4) Line-haul antar-hub (operator)");
  await req("POST", "/hub/scan/sort", operator, { awb: AWB, order_id: orderId, hub_id: hubOrigin.id, operator_id: "OP-SIM" });
  await req("POST", "/hub/scan/outbound", operator, { awb: AWB, order_id: orderId, hub_id: hubOrigin.id, operator_id: "OP-SIM" });
  ok(`📷 sortir + keluar ${hubOrigin.name}`);
  const arr = await req("POST", "/hub/scan/inbound", operator, { awb: AWB, order_id: orderId, hub_id: hubDest.id, operator_id: "OP-SIM" });
  arr.ok ? ok(`📷 masuk ${hubDest.name} → memicu last-mile`) : fail(`scan dest: ${arr.status}`);

  // 5) LEG 2 — last-mile (dest hub → receiver)
  step("5) LEG 2 last-mile (hub tujuan → penerima)");
  const job2 = await waitForJob(courier, orderId, "LAST_MILE");
  if (!job2) return fail("last-mile job tidak muncul (cek pemicu di order/dispatch)");
  ok("last-mile job otomatis tersedia");
  await req("POST", "/dispatch/claim", courier, { order_id: orderId, courier_name: "Kurir NusaRoute" });
  ok("kurir klaim last-mile");
  const p2 = await req("POST", "/dispatch/pickup", courier, { order_id: orderId, awb: AWB });
  p2.ok ? ok(`📷 scan@hub tujuan → keluar untuk diantar`) : fail(`pickup2: ${p2.status} ${JSON.stringify(p2.data)}`);
  const d2 = await req("POST", "/dispatch/deliver", courier, { order_id: orderId, awb: AWB, receiver_name: "Siti Aminah" });
  d2.ok ? ok(`📷 scan@penerima → DELIVERED`) : fail(`deliver2: ${d2.status} ${JSON.stringify(d2.data)}`);

  // 6) Timeline
  step("6) Timeline tracking (menunggu Kafka 5s)");
  await sleep(5000);
  const t = await req("GET", `/tracking/${AWB}`, demo);
  const tl = t.ok && t.data ? t.data.data : null;
  const evs = tl ? (Array.isArray(tl) ? tl : tl.events || []) : [];
  evs.slice().sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
    .forEach((e) => console.log(`   • ${String(e.status).padEnd(16)} @ ${e.location || "-"}  ${e.detail || ""}`));

  console.log(`\n${process.exitCode ? "⚠️ Selesai dengan error" : "🎉 Rantai dua-leg lengkap"} — AWB ${AWB}`);
})().catch((e) => { console.error("❌", e.message); process.exit(1); });
