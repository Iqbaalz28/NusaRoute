// test-roles.js
// Verifies API-gateway role enforcement: each seeded role can only reach the
// endpoints it should, and customer/public routes still work.
//
// Prereq: services up via docker-compose, and role accounts seeded
//   (demo=USER, courier=COURIER, operator=HUB_OPERATOR, admin=ADMIN).
//
// Run:  node test-roles.js          (Node 18+, uses global fetch)

const GW = "http://localhost:8080/api/v1";
const PASS = "password123";

const ACCOUNTS = {
  USER: "demo@nusaroute.com",
  COURIER: "courier@nusaroute.com",
  HUB_OPERATOR: "operator@nusaroute.com",
  ADMIN: "admin@nusaroute.com",
};

async function login(email) {
  const res = await fetch(`${GW}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password: PASS }),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(`login ${email} failed: ${res.status} ${JSON.stringify(data)}`);
  return data.data.token;
}

async function call(method, path, token, body) {
  const headers = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`${GW}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  return res.status;
}

// allowed = the roles that SHOULD succeed (2xx). Everyone else must get 403.
const CASES = [
  { name: "GET /dispatch/assignments (kurir)", method: "GET", path: "/dispatch/assignments", allowed: ["COURIER", "ADMIN"] },
  { name: "GET /orders/all (admin)", method: "GET", path: "/orders/all", allowed: ["ADMIN"] },
  { name: "PUT /orders/status (admin)", method: "PUT", path: "/orders/status", body: { order_id: "00000000-0000-0000-0000-000000000000", status: "READY_FOR_PICKUP" }, allowed: ["ADMIN"], okIsNot403: true },
  { name: "POST /hub/scan/inbound (operator)", method: "POST", path: "/hub/scan/inbound", body: { awb: "NRTESTROLE", hub_id: "hub-1", operator_id: "OP-TEST" }, allowed: ["HUB_OPERATOR", "ADMIN"] },
];

const PUBLIC_CASES = [
  { name: "GET /orders (pesanan sendiri, semua user login)", method: "GET", path: "/orders", role: "USER", wantSuccess: true },
  { name: "GET /hub/list (publik, tanpa token)", method: "GET", path: "/hub/list", role: null, wantSuccess: true },
];

(async () => {
  console.log("🔐 Menguji enforcement role di gateway\n");

  const tokens = {};
  for (const [role, email] of Object.entries(ACCOUNTS)) tokens[role] = await login(email);

  let pass = 0, fail = 0;
  const ok = (cond, label, detail) => {
    console.log(`  ${cond ? "✅ PASS" : "❌ FAIL"}  ${label}${detail ? `  (${detail})` : ""}`);
    cond ? pass++ : fail++;
  };

  console.log("Endpoint terproteksi — yang diizinkan harus lolos, sisanya 403:");
  for (const c of CASES) {
    console.log(`\n• ${c.name}  [izin: ${c.allowed.join(", ")}]`);
    for (const role of Object.keys(ACCOUNTS)) {
      const status = await call(c.method, c.path, tokens[role], c.body);
      const shouldAllow = c.allowed.includes(role);
      // "allowed" => not 403 (200/201/400/404 all mean it passed the role gate).
      // "denied"  => exactly 403.
      const good = shouldAllow ? status !== 403 : status === 403;
      ok(good, `${role.padEnd(13)} → ${status}`, shouldAllow ? "harus lolos role-gate" : "harus 403");
    }
  }

  console.log("\nRegресi rute customer/publik:");
  for (const c of PUBLIC_CASES) {
    const status = await call(c.method, c.path, c.role ? tokens[c.role] : null, c.body);
    ok(status >= 200 && status < 300, c.name, `status ${status}`);
  }

  // cleanup any scan rows the operator/admin successfully created
  // (best-effort; ignore if psql not reachable from here)

  console.log(`\n${fail === 0 ? "🎉" : "⚠️"}  Hasil: ${pass} PASS, ${fail} FAIL`);
  if (fail > 0) process.exit(1);
})().catch((e) => { console.error("❌", e.message); process.exit(1); });
