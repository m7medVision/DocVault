import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';

import * as authApi from '@/lib/auth/api';
import { normalizeUser, type User } from '@/lib/auth/types';
import {
  clearStoredSession,
  getStoredAccessToken,
  getStoredRefreshToken,
  getStoredUser,
  setStoredAccessToken,
  setStoredRefreshToken,
  setStoredUser,
} from '@/lib/auth/token-storage';
import { setRefreshFunction } from '@/lib/api/client';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (params: authApi.RegisterParams) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<string | null>;
  updateUser: (updates: Partial<User>) => void;
}

const AuthContext = createContext<AuthState | null>(null);

let inMemoryAccessToken: string | null = null;

export function getAccessToken(): string | null {
  return inMemoryAccessToken;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const setSession = useCallback((newUser: User | null, newToken: string | null, newRefreshToken?: string | null) => {
    setUser(newUser);
    setAccessToken(newToken);
    inMemoryAccessToken = newToken;

    const persistOps: Promise<void>[] = [];
    persistOps.push(setStoredAccessToken(newToken));
    persistOps.push(setStoredUser(newUser ? JSON.stringify(newUser) : null));
    if (newRefreshToken !== undefined) {
      persistOps.push(setStoredRefreshToken(newRefreshToken));
    }
    void Promise.all(persistOps);
  }, []);

  const clearSession = useCallback(() => {
    setSession(null, null, null);
    void clearStoredSession();
  }, [setSession]);

  const refreshSession = useCallback(async (): Promise<string | null> => {
    const refreshToken = await getStoredRefreshToken();
    if (!refreshToken) {
      clearSession();
      return null;
    }

    try {
      const result = await authApi.refreshAuth(refreshToken);
      const currentUser = await authApi.fetchCurrentUser(result.tokens.access_token);
      const normalized = normalizeUser(currentUser, result.tokens.access_token);
      setSession(normalized, result.tokens.access_token, result.tokens.refresh_token);
      return result.tokens.access_token;
    } catch {
      clearSession();
      return null;
    }
  }, [clearSession, setSession]);

  useEffect(() => {
    setRefreshFunction(refreshSession);
    return () => setRefreshFunction(null);
  }, [refreshSession]);

  useEffect(() => {
    let active = true;

    async function restore() {
      try {
        const storedToken = await getStoredAccessToken();
        const storedUserJson = await getStoredUser();

        if (!active) return;

        if (storedToken && storedUserJson) {
          try {
            JSON.parse(storedUserJson);
            const normalized = JSON.parse(storedUserJson) as User;
            inMemoryAccessToken = storedToken;
            setUser(normalized);
            setAccessToken(storedToken);
          } catch {
            const newToken = await refreshSession();
            if (!newToken && active) clearSession();
          }
        } else {
          clearSession();
        }
      } catch {
        if (active) clearSession();
      } finally {
        if (active) setIsLoading(false);
      }
    }

    void restore();
    return () => { active = false; };
  }, [clearSession, refreshSession]);

  const loginFn = useCallback(async (email: string, password: string) => {
    const response = await authApi.login(email.trim().toLowerCase(), password);
    const normalized = normalizeUser(response.user, response.tokens.access_token);
    setSession(normalized, response.tokens.access_token, response.tokens.refresh_token);
  }, [setSession]);

  const registerFn = useCallback(async (params: authApi.RegisterParams) => {
    const response = await authApi.register(params);
    const normalized = normalizeUser(response.user, response.tokens.access_token);
    setSession(normalized, response.tokens.access_token, response.tokens.refresh_token);
  }, [setSession]);

  const logoutFn = useCallback(async () => {
    const refreshToken = await getStoredRefreshToken();
    if (refreshToken) {
      await authApi.logout(refreshToken);
    }
    clearSession();
  }, [clearSession]);

  const updateUser = useCallback((updates: Partial<User>) => {
    setUser((prev) => (prev ? { ...prev, ...updates } : null));
  }, []);

  const value = useMemo<AuthState>(() => ({
    user,
    accessToken,
    isAuthenticated: Boolean(user && accessToken),
    isLoading,
    login: loginFn,
    register: registerFn,
    logout: logoutFn,
    refreshSession,
    updateUser,
  }), [user, accessToken, isLoading, loginFn, registerFn, logoutFn, refreshSession, updateUser]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within AuthProvider');
  return context;
}
