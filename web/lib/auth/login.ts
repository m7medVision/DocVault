import type {
  AuthSession,
  BackendAuthResponse,
  BackendTokenPair,
  BackendUserResponse,
  User,
} from './types';

function decodeBase64Url(value: string): string | null {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/');
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');

  try {
    if (typeof window === 'undefined') {
      return Buffer.from(padded, 'base64').toString('utf8');
    }

    return window.atob(padded);
  } catch {
    return null;
  }
}

function decodeJwtPayload(token: string): Record<string, unknown> | null {
  const payload = token.split('.')[1];
  if (!payload) {
    return null;
  }

  const decoded = decodeBase64Url(payload);
  if (!decoded) {
    return null;
  }

  try {
    return JSON.parse(decoded) as Record<string, unknown>;
  } catch {
    return null;
  }
}

export function getTokenExpiresAt(token: string): string | null {
  const payload = decodeJwtPayload(token);
  const exp = payload?.exp;

  if (typeof exp !== 'number') {
    return null;
  }

  return new Date(exp * 1000).toISOString();
}

export function normalizeUser(
  user: BackendUserResponse,
  accessToken?: string
): User {
  const claims = accessToken ? decodeJwtPayload(accessToken) : null;
  const role = typeof claims?.role === 'string' ? claims.role : undefined;
  const orgId = typeof claims?.org_id === 'string' ? claims.org_id : undefined;
  const tenantId =
    typeof claims?.tenant_id === 'string' ? claims.tenant_id : user.tenant_id;

  return {
    id: user.id,
    email: user.email,
    displayName: user.display_name,
    tenantId,
    locale: user.locale || 'en',
    emailVerified: user.email_verified,
    createdAt: user.created_at,
    role,
    orgId,
  };
}

export function toAuthSession(response: BackendAuthResponse): AuthSession {
  return {
    user: normalizeUser(response.user, response.tokens.access_token),
    accessToken: response.tokens.access_token,
    expiresAt:
      response.tokens.expires_at ||
      getTokenExpiresAt(response.tokens.access_token) ||
      new Date(Date.now() + 15 * 60 * 1000).toISOString(),
  };
}

export function toAuthSessionFromRefresh(
  user: BackendUserResponse,
  tokens: BackendTokenPair
): AuthSession {
  return {
    user: normalizeUser(user, tokens.access_token),
    accessToken: tokens.access_token,
    expiresAt:
      tokens.expires_at ||
      getTokenExpiresAt(tokens.access_token) ||
      new Date(Date.now() + 15 * 60 * 1000).toISOString(),
  };
}
