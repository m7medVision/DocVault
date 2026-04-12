import { NextRequest, NextResponse } from 'next/server';
import {
  AUTH_COOKIE_NAMES,
  SERVER_API_BASE_URL,
  type AuthSession,
  type BackendAuthResponse,
  type BackendTokenPair,
  type BackendUserResponse,
  type LoginInput,
  type RegisterInput,
  getTokenExpiresAt,
  toAuthSession,
  toAuthSessionFromRefresh,
} from './auth';

const REFRESH_COOKIE_MAX_AGE = 30 * 24 * 60 * 60;

export class BackendAuthError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'BackendAuthError';
    this.status = status;
  }
}

async function parseBackendResponse<T>(response: Response): Promise<T> {
  const data = await response.json().catch(() => null);

  if (!response.ok) {
    const message =
      data && typeof data.error === 'string'
        ? data.error
        : 'Authentication request failed';
    throw new BackendAuthError(message, response.status);
  }

  return data as T;
}

async function backendFetch<T>(
  path: string,
  init: RequestInit
): Promise<T> {
  const response = await fetch(`${SERVER_API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers || {}),
    },
    cache: 'no-store',
  });

  return parseBackendResponse<T>(response);
}

export async function loginWithPassword(
  input: LoginInput
): Promise<AuthSession & { refreshToken: string }> {
  const response = await backendFetch<BackendAuthResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({
      email: input.email,
      password: input.password,
    }),
  });

  return {
    ...toAuthSession(response),
    refreshToken: response.tokens.refresh_token,
  };
}

export async function registerWithPassword(
  input: RegisterInput
): Promise<AuthSession & { refreshToken: string }> {
  const response = await backendFetch<BackendAuthResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({
      email: input.email,
      password: input.password,
      display_name: input.displayName,
      locale: input.locale,
    }),
  });

  return {
    ...toAuthSession(response),
    refreshToken: response.tokens.refresh_token,
  };
}

async function fetchCurrentUser(accessToken: string): Promise<BackendUserResponse> {
  return backendFetch<BackendUserResponse>('/auth/me', {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function refreshAuthSession(
  refreshToken: string
): Promise<AuthSession & { refreshToken: string }> {
  const response = await backendFetch<{ tokens: BackendTokenPair }>('/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  const user = await fetchCurrentUser(response.tokens.access_token);

  return {
    ...toAuthSessionFromRefresh(user, response.tokens),
    refreshToken: response.tokens.refresh_token,
  };
}

function baseCookieOptions() {
  return {
    httpOnly: true as const,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict' as const,
    path: '/',
  };
}

export function applyAuthCookies(
  response: NextResponse,
  session: { accessToken: string; refreshToken: string },
  rememberMe: boolean
): void {
  const accessExpiresAt = getTokenExpiresAt(session.accessToken);

  response.cookies.set(AUTH_COOKIE_NAMES.accessToken, session.accessToken, {
    ...baseCookieOptions(),
    ...(accessExpiresAt ? { expires: new Date(accessExpiresAt) } : {}),
  });

  response.cookies.set(AUTH_COOKIE_NAMES.refreshToken, session.refreshToken, {
    ...baseCookieOptions(),
    ...(rememberMe ? { maxAge: REFRESH_COOKIE_MAX_AGE } : {}),
  });

  response.cookies.set(AUTH_COOKIE_NAMES.rememberSession, rememberMe ? '1' : '0', {
    ...baseCookieOptions(),
    ...(rememberMe ? { maxAge: REFRESH_COOKIE_MAX_AGE } : {}),
  });
}

export function clearAuthCookies(response: NextResponse): void {
  response.cookies.delete(AUTH_COOKIE_NAMES.accessToken);
  response.cookies.delete(AUTH_COOKIE_NAMES.refreshToken);
  response.cookies.delete(AUTH_COOKIE_NAMES.rememberSession);
}

export function getRememberMe(request: NextRequest): boolean {
  return request.cookies.get(AUTH_COOKIE_NAMES.rememberSession)?.value === '1';
}

export function getAccessTokenFromRequest(request: NextRequest): string | null {
  return request.cookies.get(AUTH_COOKIE_NAMES.accessToken)?.value || null;
}

export function getRefreshTokenFromRequest(request: NextRequest): string | null {
  return request.cookies.get(AUTH_COOKIE_NAMES.refreshToken)?.value || null;
}
