const DEFAULT_API_BASE_URL = 'http://localhost:8080/api/v1';

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || DEFAULT_API_BASE_URL;

export const SERVER_API_BASE_URL =
  process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || DEFAULT_API_BASE_URL;

export const AUTH_COOKIE_NAMES = {
  accessToken: 'docvault_access_token',
  refreshToken: 'docvault_refresh_token',
  rememberSession: 'docvault_remember_session',
} as const;
