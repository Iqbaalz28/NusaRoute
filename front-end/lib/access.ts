// Central role-based access rules for client-side navigation gating.
//
// NOTE: This is UX gating only — it hides what a user can't use. Real
// enforcement lives at the API gateway / services (JWT + role checks).

export type Role = "USER" | "ADMIN" | "COURIER" | "HUB_OPERATOR";

interface PageAccess {
  requiresAuth?: boolean; // any logged-in user
  roles?: string[];       // only these roles (implies requiresAuth)
}

// Pages NOT listed here are public (visible to everyone, logged in or not):
// dashboard, tracking, services, hubs (hub list is public; the scan console
// inside it is gated separately via canScanHub).
const PAGE_ACCESS: Record<string, PageAccess> = {
  orders: { requiresAuth: true },
  courier: { roles: ["COURIER", "ADMIN"] },
  admin: { roles: ["ADMIN"] },
};

export function canAccessPage(pageId: string, isAuthenticated: boolean, role?: string): boolean {
  const rule = PAGE_ACCESS[pageId];
  if (!rule) return true; // public page
  if (rule.roles) return isAuthenticated && !!role && rule.roles.includes(role);
  if (rule.requiresAuth) return isAuthenticated;
  return true;
}

// Whether the current user may operate the Hub scan console. Regular users can
// still view the hub list; only operators/admins record scans.
export function canScanHub(isAuthenticated: boolean, role?: string): boolean {
  return isAuthenticated && (role === "HUB_OPERATOR" || role === "ADMIN");
}
