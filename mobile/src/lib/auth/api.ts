import { API_BASE_URL } from '@/lib/config';
import type { AuthResponse, BackendUser } from './types';

async function parseJson<T>(response: Response): Promise<T> {
  const data = await response.json().catch(() => null);

  if (!response.ok) {
    const message = data && typeof data.error === 'string' ? data.error : `Request failed (${response.status})`;
    throw new Error(message);
  }

  return data as T;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  return parseJson<AuthResponse>(response);
}

export interface RegisterParams {
  email: string;
  password: string;
  displayName: string;
  tenantName?: string;
  orgName?: string;
  locale?: 'en' | 'ar';
}

export async function register(params: RegisterParams): Promise<AuthResponse> {
  const response = await fetch(`${API_BASE_URL}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: params.email,
      password: params.password,
      display_name: params.displayName,
      tenant_name: params.tenantName || `${params.displayName}'s Workspace`,
      org_name: params.orgName || 'Default Organization',
      locale: params.locale || 'en',
    }),
  });

  return parseJson<AuthResponse>(response);
}

export async function refreshAuth(refreshToken: string): Promise<{ tokens: { access_token: string; refresh_token: string; expires_at: string } }> {
  const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  return parseJson(response);
}

export async function fetchCurrentUser(accessToken: string): Promise<BackendUser> {
  const response = await fetch(`${API_BASE_URL}/auth/me`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${accessToken}`,
    },
  });

  return parseJson<BackendUser>(response);
}

export async function logout(refreshToken: string): Promise<void> {
  try {
    await fetch(`${API_BASE_URL}/auth/logout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  } catch {}
}
