export interface BackendUser {
  id: string;
  email: string;
  display_name: string;
  locale: 'en' | 'ar';
  email_verified: boolean;
  tenant_id: string;
  created_at: string;
}

export interface BackendTokens {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  token_type: string;
}

export interface AuthResponse {
  user: BackendUser;
  tokens: BackendTokens;
}

export interface User {
  id: string;
  email: string;
  displayName: string;
  tenantId: string;
  locale: 'en' | 'ar';
  emailVerified: boolean;
  createdAt: string;
  role?: string;
  orgId?: string;
}

export function normalizeUser(raw: BackendUser, accessToken?: string): User {
  let role: string | undefined;
  let orgId: string | undefined;

  if (accessToken) {
    try {
      const payload = JSON.parse(atob(accessToken.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')));
      role = typeof payload.role === 'string' ? payload.role : undefined;
      orgId = typeof payload.org_id === 'string' ? payload.org_id : undefined;
    } catch {}
  }

  return {
    id: raw.id,
    email: raw.email,
    displayName: raw.display_name,
    tenantId: raw.tenant_id,
    locale: raw.locale || 'en',
    emailVerified: raw.email_verified,
    createdAt: raw.created_at,
    role,
    orgId,
  };
}
