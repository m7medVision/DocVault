import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react';
import { AppState } from 'react-native';
import {
  loginWithPassword,
  logoutWithToken,
  registerWithPassword,
  setAuthRefreshHandler,
} from '../lib/auth';
import { sessionManager } from '../lib/auth/session';
import { tokenRefresher } from '../lib/auth/refresh';
import { notificationService } from '../lib/notifications/service';
import {
  AuthContextType,
  LoginInput,
  MobileUser,
  RegisterInput,
} from '../lib/types';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<MobileUser | null>(null);
  const [accessToken, setAccessTokenState] = useState<string | null>(null);
  const [rememberMeState, setRememberMeState] = useState(true);
  const [isLoading, setIsLoading] = useState(true);
  const expiresAtRef = useRef<number | null>(null);
  const refreshTokenRef = useRef<string | null>(null);
  const refreshTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearRefreshTimeout = useCallback(() => {
    if (refreshTimeoutRef.current) {
      clearTimeout(refreshTimeoutRef.current);
      refreshTimeoutRef.current = null;
    }
  }, []);

  const clearSession = useCallback(async () => {
    clearRefreshTimeout();
    refreshTokenRef.current = null;
    expiresAtRef.current = null;
    setUser(null);
    setAccessTokenState(null);
    setRememberMeState(true);
    await sessionManager.clearSession();
  }, [clearRefreshTimeout]);

  const applySession = useCallback(
    async (
      nextUser: MobileUser,
      nextAccessToken: string,
      nextRefreshToken: string,
      expiresAt: number,
      rememberMe: boolean
    ) => {
      setUser(nextUser);
      setAccessTokenState(nextAccessToken);
      setRememberMeState(rememberMe);
      expiresAtRef.current = expiresAt;
      refreshTokenRef.current = nextRefreshToken;

      await sessionManager.saveSession(
        nextUser,
        { accessToken: nextAccessToken, refreshToken: nextRefreshToken, expiresAt },
        rememberMe
      );
    },
    []
  );

  const refreshSession = useCallback(async (): Promise<string | null> => {
    const storedSession = await sessionManager.getSession();
    const refreshToken = refreshTokenRef.current || storedSession?.tokens.refreshToken;

    if (!refreshToken) {
      await clearSession();
      return null;
    }

    try {
      const session = await tokenRefresher.refresh(refreshToken);
      const rememberMe = storedSession ? true : false;
      await applySession(
        session.user,
        session.tokens.accessToken,
        session.tokens.refreshToken,
        session.tokens.expiresAt,
        rememberMe
      );
      return session.tokens.accessToken;
    } catch (error) {
      console.error('Failed to refresh session:', error);
      await clearSession();
      return null;
    }
  }, [applySession, clearSession]);

  useEffect(() => {
    setAuthRefreshHandler(refreshSession);
    return () => setAuthRefreshHandler(null);
  }, [refreshSession]);

  useEffect(() => {
    const loadStoredAuth = async () => {
      try {
        const session = await sessionManager.getSession();

        if (!session) {
          await clearSession();
          setIsLoading(false);
          return;
        }

        refreshTokenRef.current = session.tokens.refreshToken;
        setRememberMeState(true);

        if (Date.now() < session.tokens.expiresAt - 60_000) {
          setUser(session.user);
          setAccessTokenState(session.tokens.accessToken);
          expiresAtRef.current = session.tokens.expiresAt;
          await notificationService.register();
          setIsLoading(false);
          return;
        }

        if (session.tokens.refreshToken) {
          const refreshedToken = await refreshSession();
          if (refreshedToken) {
            await notificationService.register();
          }
          setIsLoading(false);
          return;
        }

        await clearSession();
      } catch (error) {
        console.error('Failed to load stored auth:', error);
        await clearSession();
      } finally {
        setIsLoading(false);
      }
    };

    void loadStoredAuth();
  }, [clearSession, refreshSession]);

  useEffect(() => {
    if (!accessToken || !expiresAtRef.current) {
      clearRefreshTimeout();
      return;
    }

    const refreshDelay = expiresAtRef.current - Date.now() - 60_000;

    if (refreshDelay <= 0) {
      void refreshSession();
      return;
    }

    clearRefreshTimeout();
    refreshTimeoutRef.current = setTimeout(() => {
      void refreshSession();
    }, refreshDelay);

    return clearRefreshTimeout;
  }, [accessToken, clearRefreshTimeout, refreshSession]);

  useEffect(() => {
    const subscription = AppState.addEventListener('change', (state) => {
      if (
        state === 'active' &&
        accessToken &&
        expiresAtRef.current &&
        Date.now() >= expiresAtRef.current - 60_000
      ) {
        void refreshSession();
      }
    });

    return () => {
      subscription.remove();
    };
  }, [accessToken, refreshSession]);

  const login = useCallback(
    async (input: LoginInput) => {
      const session = await loginWithPassword(input.email.trim().toLowerCase(), input.password);
      await applySession(
        session.user,
        session.tokens.accessToken,
        session.tokens.refreshToken,
        session.tokens.expiresAt,
        input.rememberMe
      );
      await notificationService.register();
    },
    [applySession]
  );

  const register = useCallback(
    async (input: RegisterInput) => {
      const session = await registerWithPassword({
        ...input,
        email: input.email.trim().toLowerCase(),
        displayName: input.displayName.trim(),
      });
      await applySession(
        session.user,
        session.tokens.accessToken,
        session.tokens.refreshToken,
        session.tokens.expiresAt,
        input.rememberMe
      );
      await notificationService.register();
    },
    [applySession]
  );

  const logout = useCallback(async () => {
    const session = await sessionManager.getSession();
    const refreshToken = refreshTokenRef.current || session?.tokens.refreshToken;

    if (refreshToken) {
      try {
        await logoutWithToken(refreshToken);
      } catch (error) {
        console.error('Failed to notify backend about logout:', error);
      }
    }

    await clearSession();
  }, [clearSession]);

  const value: AuthContextType = {
    isAuthenticated: Boolean(user && accessToken),
    user,
    accessToken,
    isLoading,
    rememberMe: rememberMeState,
    login,
    register,
    logout,
    refreshSession,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);

  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return context;
}
