import createMiddleware from 'next-intl/middleware';
import { NextRequest, NextResponse } from 'next/server';
import { AUTH_COOKIE_NAMES } from '@/lib/auth';
import { routing } from './routing';

const protectedRoutes = ['/documents', '/search', '/settings', '/reminders', '/admin'];
const guestOnlyRoutes = ['/auth/login', '/auth/register'];
const intlMiddleware = createMiddleware(routing);

function hasSessionCookie(request: NextRequest): boolean {
  return Boolean(
    request.cookies.get(AUTH_COOKIE_NAMES.accessToken)?.value ||
      request.cookies.get(AUTH_COOKIE_NAMES.refreshToken)?.value
  );
}

export default function middleware(request: NextRequest) {
  const response = intlMiddleware(request);
  const { pathname, search } = request.nextUrl;
  const locale = pathname.split('/')[1] || routing.defaultLocale;
  const isAuthenticated = hasSessionCookie(request);
  const isProtected = protectedRoutes.some((route) => pathname.includes(route));
  const isGuestOnly = guestOnlyRoutes.some((route) => pathname.endsWith(route));

  if (isProtected && !isAuthenticated) {
    const loginUrl = new URL(`/${locale}/auth/login`, request.url);
    loginUrl.searchParams.set('redirect', `${pathname}${search}`);
    return NextResponse.redirect(loginUrl);
  }

  if (isGuestOnly && isAuthenticated) {
    const redirectTarget = request.nextUrl.searchParams.get('redirect');
    const target =
      redirectTarget && redirectTarget.startsWith('/')
        ? redirectTarget
        : `/${locale}`;

    return NextResponse.redirect(new URL(target, request.url));
  }

  return response;
}

export const config = {
  matcher: ['/((?!api|_next|favicon.ico|.*\\..*).*)'],
};
