import createMiddleware from 'next-intl/middleware';
import { NextRequest, NextResponse } from 'next/server';
import { AUTH_COOKIE_NAMES } from '@/lib/auth';
import { routing } from './routing';

const protectedRoutes = ['/dashboard', '/documents', '/search', '/settings', '/reminders', '/admin'];
const guestOnlyRoutes = ['/auth/login', '/auth/register'];
const intlMiddleware = createMiddleware(routing);

function hasSessionCookie(request: NextRequest): boolean {
  return Boolean(
    request.cookies.get(AUTH_COOKIE_NAMES.accessToken)?.value ||
      request.cookies.get(AUTH_COOKIE_NAMES.refreshToken)?.value
  );
}

function isLocale(value: string): boolean {
  return routing.locales.includes(value as 'en' | 'ar');
}

function localePath(locale: string, path: string): string {
  return locale === routing.defaultLocale ? path : `/${locale}${path}`;
}

function resolveLocale(pathname: string): string {
  const first = pathname.split('/')[1] || '';
  return isLocale(first) ? first : routing.defaultLocale;
}

export default function middleware(request: NextRequest) {
  const response = intlMiddleware(request);
  const { pathname, search } = request.nextUrl;
  const locale = resolveLocale(pathname);
  const isAuthenticated = hasSessionCookie(request);
  const isProtected = protectedRoutes.some((route) => pathname.includes(route));
  const isGuestOnly = guestOnlyRoutes.some((route) => pathname.endsWith(route));

  const segments = pathname.split('/').filter(Boolean);
  const firstSegment = segments[0] || '';
  const isMarketingRoot =
    pathname === '/' ||
    (isLocale(firstSegment) && segments.length === 1);

  if (isProtected && !isAuthenticated) {
    const loginUrl = new URL(localePath(locale, '/auth/login'), request.url);
    loginUrl.searchParams.set('redirect', `${pathname}${search}`);
    return NextResponse.redirect(loginUrl);
  }

  if (isMarketingRoot && isAuthenticated) {
    return NextResponse.redirect(new URL(localePath(locale, '/dashboard'), request.url));
  }

  if (isGuestOnly && isAuthenticated) {
    const redirectTarget = request.nextUrl.searchParams.get('redirect');
    const target =
      redirectTarget && redirectTarget.startsWith('/')
        ? redirectTarget
        : localePath(locale, '/dashboard');

    return NextResponse.redirect(new URL(target, request.url));
  }

  return response;
}

export const config = {
  matcher: ['/((?!api|_next|favicon.ico|.*\\..*).*)'],
};
