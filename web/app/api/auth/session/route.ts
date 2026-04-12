import { NextRequest, NextResponse } from 'next/server';
import {
  type AuthSession,
  type BackendUserResponse,
  getTokenExpiresAt,
  normalizeUser,
} from '@/lib/auth';
import {
  BackendAuthError,
  applyAuthCookies,
  clearAuthCookies,
  getAccessTokenFromRequest,
  getRefreshTokenFromRequest,
  getRememberMe,
  refreshAuthSession,
} from '@/lib/server-auth';

const apiBaseUrl =
  process.env.API_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:8080/api/v1';

async function fetchCurrentUser(accessToken: string): Promise<AuthSession | null> {
  const response = await fetch(`${apiBaseUrl}/auth/me`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    cache: 'no-store',
  });

  if (!response.ok) {
    return null;
  }

  const user = (await response.json()) as BackendUserResponse;

  return {
    user: normalizeUser(user, accessToken),
    accessToken,
    expiresAt:
      getTokenExpiresAt(accessToken) ||
      new Date(Date.now() + 15 * 60 * 1000).toISOString(),
  };
}

export async function GET(request: NextRequest) {
  const accessToken = getAccessTokenFromRequest(request);

  if (accessToken) {
    const session = await fetchCurrentUser(accessToken);
    if (session) {
      return NextResponse.json({ isAuthenticated: true, session });
    }
  }

  const refreshToken = getRefreshTokenFromRequest(request);
  if (!refreshToken) {
    const response = NextResponse.json({ isAuthenticated: false });
    clearAuthCookies(response);
    return response;
  }

  try {
    const session = await refreshAuthSession(refreshToken);
    const response = NextResponse.json({ isAuthenticated: true, session });
    applyAuthCookies(response, session, getRememberMe(request));
    return response;
  } catch (error) {
    const response = NextResponse.json(
      {
        isAuthenticated: false,
        ...(error instanceof BackendAuthError ? { error: error.message } : {}),
      },
      {
        status: error instanceof BackendAuthError ? error.status : 200,
      }
    );
    clearAuthCookies(response);
    return response;
  }
}
