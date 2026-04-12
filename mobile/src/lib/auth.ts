import { CONFIG } from './config';
import { MobileUser, RegisterInput } from './types';

export { SessionManager, sessionManager } from './auth/session';
export { TokenRefresher, tokenRefresher } from './auth/refresh';

interface BackendUserResponse {
  id: string;
  email: string;
  display_name: string;
  locale: string;
  email_verified: boolean;
  tenant_id: string;
  created_at: string;
}

interface BackendTokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  token_type: string;
}

interface BackendAuthResponse {
  user: BackendUserResponse;
  tokens: BackendTokenPair;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}

export interface AuthSessionPayload {
  user: MobileUser;
  tokens: AuthTokens;
}

let refreshHandler: (() => Promise<string | null>) | null = null;
let inflightRefresh: Promise<string | null> | null = null;

function parseExpiry(value: string): number {
  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) {
    return Date.now() + 15 * 60 * 1000;
  }
  return parsed;
}

function normalizeUser(user: BackendUserResponse): MobileUser {
  return {
    id: user.id,
    email: user.email,
    name: user.display_name,
    displayName: user.display_name,
    tenantId: user.tenant_id,
    locale: user.locale,
    emailVerified: user.email_verified,
  };
}

async function parseJson<T>(response: Response): Promise<T> {
  const payload = await response.json().catch(() => null);

  if (!response.ok) {
    const message =
      payload && typeof payload.error === 'string'
        ? payload.error
        : 'Request failed';
    throw new Error(message);
  }

  return payload as T;
}

export function setAuthRefreshHandler(
  handler: (() => Promise<string | null>) | null
): void {
  refreshHandler = handler;
}

export async function requestTokenRefresh(): Promise<string | null> {
  if (!refreshHandler) {
    return null;
  }

  if (!inflightRefresh) {
    inflightRefresh = refreshHandler().finally(() => {
      inflightRefresh = null;
    });
  }

  return inflightRefresh;
}

export async function authorizedFetch(
  accessToken: string,
  input: string,
  init: RequestInit = {},
  allowRetry = true
): Promise<Response> {
  const headers = new Headers(init.headers || {});

  if (!headers.has('Content-Type') && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  headers.set('Authorization', `Bearer ${accessToken}`);

  const response = await fetch(input, {
    ...init,
    headers,
  });

  if (response.status === 401 && allowRetry) {
    const refreshedToken = await requestTokenRefresh();
    if (refreshedToken) {
      return authorizedFetch(refreshedToken, input, init, false);
    }
  }

  return response;
}

export async function loginWithPassword(
  email: string,
  password: string
): Promise<AuthSessionPayload> {
  const response = await fetch(`${CONFIG.apiBaseUrl}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
  });

  const payload = await parseJson<BackendAuthResponse>(response);

  return {
    user: normalizeUser(payload.user),
    tokens: {
      accessToken: payload.tokens.access_token,
      refreshToken: payload.tokens.refresh_token,
      expiresAt: parseExpiry(payload.tokens.expires_at),
    },
  };
}

export async function registerWithPassword(
  input: RegisterInput
): Promise<AuthSessionPayload> {
  const response = await fetch(`${CONFIG.apiBaseUrl}/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email: input.email,
      password: input.password,
      display_name: input.displayName,
      locale: input.locale || 'en',
    }),
  });

  const payload = await parseJson<BackendAuthResponse>(response);

  return {
    user: normalizeUser(payload.user),
    tokens: {
      accessToken: payload.tokens.access_token,
      refreshToken: payload.tokens.refresh_token,
      expiresAt: parseExpiry(payload.tokens.expires_at),
    },
  };
}

async function fetchCurrentUser(accessToken: string): Promise<MobileUser> {
  const response = await fetch(`${CONFIG.apiBaseUrl}/auth/me`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });

  const payload = await parseJson<BackendUserResponse>(response);
  return normalizeUser(payload);
}

export async function refreshWithToken(
  refreshToken: string
): Promise<AuthSessionPayload> {
  const response = await fetch(`${CONFIG.apiBaseUrl}/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  const payload = await parseJson<{ tokens: BackendTokenPair }>(response);
  const user = await fetchCurrentUser(payload.tokens.access_token);

  return {
    user,
    tokens: {
      accessToken: payload.tokens.access_token,
      refreshToken: payload.tokens.refresh_token,
      expiresAt: parseExpiry(payload.tokens.expires_at),
    },
  };
}

export async function logoutWithToken(refreshToken: string): Promise<void> {
  const response = await fetch(`${CONFIG.apiBaseUrl}/auth/logout`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!response.ok && response.status !== 400) {
    await parseJson(response);
  }
}
