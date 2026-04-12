import * as SecureStore from 'expo-secure-store';
import { MobileUser, TokenStorageKeys } from '../types';
import { AuthTokens, AuthSessionPayload } from '../auth';

export interface AuthSession {
  user: MobileUser;
  tokens: AuthTokens;
}

export class SessionManager {
  async getSession(): Promise<AuthSession | null> {
    try {
      const [accessToken, user, expiresAt, refreshToken] = await Promise.all([
        SecureStore.getItemAsync(TokenStorageKeys.ACCESS_TOKEN),
        SecureStore.getItemAsync(TokenStorageKeys.USER),
        SecureStore.getItemAsync(TokenStorageKeys.EXPIRES_AT),
        SecureStore.getItemAsync(TokenStorageKeys.REFRESH_TOKEN),
      ]);

      if (!accessToken || !user || !expiresAt) {
        return null;
      }

      return {
        user: JSON.parse(user) as MobileUser,
        tokens: {
          accessToken,
          refreshToken: refreshToken ?? '',
          expiresAt: parseInt(expiresAt, 10),
        },
      };
    } catch {
      return null;
    }
  }

  async saveSession(
    user: MobileUser,
    tokens: AuthTokens,
    rememberMe: boolean
  ): Promise<void> {
    await Promise.all([
      SecureStore.setItemAsync(TokenStorageKeys.ACCESS_TOKEN, tokens.accessToken),
      SecureStore.setItemAsync(TokenStorageKeys.EXPIRES_AT, tokens.expiresAt.toString()),
      SecureStore.setItemAsync(TokenStorageKeys.USER, JSON.stringify(user)),
      SecureStore.setItemAsync(TokenStorageKeys.REMEMBER_ME, rememberMe ? 'true' : 'false'),
      rememberMe
        ? SecureStore.setItemAsync(TokenStorageKeys.REFRESH_TOKEN, tokens.refreshToken)
        : SecureStore.deleteItemAsync(TokenStorageKeys.REFRESH_TOKEN),
    ]);
  }

  async clearSession(): Promise<void> {
    await Promise.all([
      SecureStore.deleteItemAsync(TokenStorageKeys.ACCESS_TOKEN),
      SecureStore.deleteItemAsync(TokenStorageKeys.REFRESH_TOKEN),
      SecureStore.deleteItemAsync(TokenStorageKeys.EXPIRES_AT),
      SecureStore.deleteItemAsync(TokenStorageKeys.USER),
      SecureStore.deleteItemAsync(TokenStorageKeys.REMEMBER_ME),
    ]);
  }
}

export const sessionManager = new SessionManager();
