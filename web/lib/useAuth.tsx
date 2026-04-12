'use client';

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import { useLocale } from 'next-intl';
import { usePathname, useRouter } from 'next/navigation';
import { setClientAccessToken, setRefreshHandler } from '@/lib/client-auth';
import {
  type AuthRouteResponse,
  type AuthSession,
  type LoginInput,
  type RegisterInput,
  type User,
} from './auth';

interface AuthContextType {
  user: User | null;
  accessToken: string | null;
  expiresAt: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (input: LoginInput) => Promise<void>;
  register: (input: RegisterInput) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<string | null>;
}

const AuthContext = createContext<AuthContextType | null>(null);
const protectedSegments = ['/documents', '/search', '/reminders', '/settings', '/admin'];

function isProtectedPath(pathname: string): boolean {
  return protectedSegments.some((segment) => pathname.includes(segment));
}

async function parseAuthResponse(response: Response): Promise<AuthSession> {
  const payload = (await response.json()) as
    | AuthRouteResponse
    | { error?: string; session?: AuthSession };
  const errorMessage = 'error' in payload ? payload.error : undefined;

  if (!response.ok || !payload.session) {
    throw new Error(errorMessage || 'Authentication failed');
  }

  return payload.session;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const locale = useLocale();
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessTokenState] = useState<string | null>(null);
  const [expiresAt, setExpiresAt] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const setSession = useCallback((session: AuthSession | null) => {
    setUser(session?.user || null);
    setAccessTokenState(session?.accessToken || null);
    setExpiresAt(session?.expiresAt || null);
    setClientAccessToken(session?.accessToken || null);
  }, []);

  const clearSession = useCallback(() => {
    setSession(null);
  }, [setSession]);

  const refresh = useCallback(async (): Promise<string | null> => {
    try {
      const response = await fetch('/api/auth/refresh', {
        method: 'POST',
        cache: 'no-store',
      });

      if (!response.ok) {
        clearSession();
        return null;
      }

      const session = await parseAuthResponse(response);
      setSession(session);
      return session.accessToken;
    } catch {
      clearSession();
      return null;
    }
  }, [clearSession, setSession]);

  useEffect(() => {
    setRefreshHandler(refresh);
    return () => setRefreshHandler(null);
  }, [refresh]);

  useEffect(() => {
    let active = true;

    const loadSession = async () => {
      try {
        const response = await fetch('/api/auth/session', {
          method: 'GET',
          cache: 'no-store',
        });

        const payload = (await response.json()) as {
          isAuthenticated?: boolean;
          session?: AuthSession;
        };

        if (!active) {
          return;
        }

        if (payload.isAuthenticated && payload.session) {
          setSession(payload.session);
        } else {
          clearSession();
        }
      } catch {
        if (active) {
          clearSession();
        }
      } finally {
        if (active) {
          setIsLoading(false);
        }
      }
    };

    void loadSession();

    return () => {
      active = false;
    };
  }, [clearSession, setSession]);

  useEffect(() => {
    if (!expiresAt || !accessToken) {
      return;
    }

    const refreshAt = new Date(expiresAt).getTime() - Date.now() - 60_000;

    if (refreshAt <= 0) {
      void refresh();
      return;
    }

    const timeoutId = window.setTimeout(() => {
      void refresh();
    }, refreshAt);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [accessToken, expiresAt, refresh]);

  useEffect(() => {
    if (isLoading || user || !isProtectedPath(pathname)) {
      return;
    }

    const loginUrl = new URL(`/${locale}/auth/login`, window.location.origin);
    loginUrl.searchParams.set('redirect', pathname);
    router.replace(`${loginUrl.pathname}${loginUrl.search}`);
  }, [isLoading, locale, pathname, router, user]);

  const login = useCallback(
    async (input: LoginInput) => {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(input),
      });

      const session = await parseAuthResponse(response);
      setSession(session);
    },
    [setSession]
  );

  const register = useCallback(
    async (input: RegisterInput) => {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(input),
      });

      const session = await parseAuthResponse(response);
      setSession(session);
    },
    [setSession]
  );

  const logout = useCallback(async () => {
    await fetch('/api/auth/logout', {
      method: 'POST',
      cache: 'no-store',
    }).catch(() => null);

    clearSession();
  }, [clearSession]);

  const value = useMemo<AuthContextType>(
    () => ({
      user,
      accessToken,
      expiresAt,
      isAuthenticated: Boolean(user && accessToken),
      isLoading,
      login,
      register,
      logout,
      refresh,
    }),
    [accessToken, expiresAt, isLoading, login, logout, refresh, register, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);

  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return context;
}
