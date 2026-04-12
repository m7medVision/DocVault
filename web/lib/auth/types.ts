export interface User {
  id: string;
  email: string;
  displayName: string;
  tenantId: string;
  locale: 'en' | 'ar';
  emailVerified: boolean;
  createdAt: string;
  role?: string;
  orgId?: string;
}

export interface AuthSession {
  user: User;
  accessToken: string;
  expiresAt: string;
}

export interface BackendUserResponse {
  id: string;
  email: string;
  display_name: string;
  locale: 'en' | 'ar';
  email_verified: boolean;
  tenant_id: string;
  created_at: string;
}

export interface BackendTokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  token_type: string;
}

export interface BackendAuthResponse {
  user: BackendUserResponse;
  tokens: BackendTokenPair;
}

export interface LoginInput {
  email: string;
  password: string;
  rememberMe: boolean;
}

export interface RegisterInput extends LoginInput {
  displayName: string;
  locale: 'en' | 'ar';
}

export interface AuthRouteResponse {
  session: AuthSession;
}
