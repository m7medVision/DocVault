export const TokenStorageKeys = {
  ACCESS_TOKEN: 'docvault_access_token',
  REFRESH_TOKEN: 'docvault_refresh_token',
  EXPIRES_AT: 'docvault_expires_at',
  USER: 'docvault_user',
  REMEMBER_ME: 'docvault_remember_me',
} as const;

export interface MobileUser {
  id: string;
  email: string;
  name: string;
  displayName: string;
  tenantId?: string;
  locale?: string;
  emailVerified?: boolean;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: MobileUser | null;
  accessToken: string | null;
  isLoading: boolean;
  rememberMe: boolean;
}

export interface LoginInput {
  email: string;
  password: string;
  rememberMe: boolean;
}

export interface RegisterInput extends LoginInput {
  displayName: string;
  locale?: string;
}

export interface AuthContextType extends AuthState {
  login: (input: LoginInput) => Promise<void>;
  register: (input: RegisterInput) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<string | null>;
}

export interface ApiError {
  error: string;
  code: string;
  request_id: string;
}
