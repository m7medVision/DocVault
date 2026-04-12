import { NextRequest, NextResponse } from 'next/server';
import { clearAuthCookies, getRefreshTokenFromRequest } from '@/lib/server-auth';

const apiBaseUrl =
  process.env.API_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:8080/api/v1';

export async function POST(request: NextRequest) {
  const refreshToken = getRefreshTokenFromRequest(request);

  if (refreshToken) {
    await fetch(`${apiBaseUrl}/auth/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
      cache: 'no-store',
    }).catch(() => null);
  }

  const response = NextResponse.json({ success: true });
  clearAuthCookies(response);
  return response;
}
