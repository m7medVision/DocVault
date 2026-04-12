import { NextRequest, NextResponse } from 'next/server';
import {
  BackendAuthError,
  applyAuthCookies,
  clearAuthCookies,
  getRefreshTokenFromRequest,
  getRememberMe,
  refreshAuthSession,
} from '@/lib/server-auth';

export async function POST(request: NextRequest) {
  const refreshToken = getRefreshTokenFromRequest(request);

  if (!refreshToken) {
    return NextResponse.json({ error: 'Not authenticated' }, { status: 401 });
  }

  try {
    const session = await refreshAuthSession(refreshToken);
    const response = NextResponse.json({ session });
    applyAuthCookies(response, session, getRememberMe(request));
    return response;
  } catch (error) {
    const response = NextResponse.json(
      {
        error:
          error instanceof BackendAuthError
            ? error.message
            : 'Unable to refresh the session',
      },
      {
        status: error instanceof BackendAuthError ? error.status : 500,
      }
    );
    clearAuthCookies(response);
    return response;
  }
}
