import * as SecureStore from 'expo-secure-store';
import { MobileUser, TokenStorageKeys } from './types';

export async function getAccessToken(): Promise<string | null> {
  try {
    return await SecureStore.getItemAsync(TokenStorageKeys.ACCESS_TOKEN);
  } catch {
    return null;
  }
}

export async function setAccessToken(token: string): Promise<void> {
  await SecureStore.setItemAsync(TokenStorageKeys.ACCESS_TOKEN, token);
}

export async function deleteAccessToken(): Promise<void> {
  await SecureStore.deleteItemAsync(TokenStorageKeys.ACCESS_TOKEN);
}

export async function getRefreshToken(): Promise<string | null> {
  try {
    return await SecureStore.getItemAsync(TokenStorageKeys.REFRESH_TOKEN);
  } catch {
    return null;
  }
}

export async function setRefreshToken(token: string): Promise<void> {
  await SecureStore.setItemAsync(TokenStorageKeys.REFRESH_TOKEN, token);
}

export async function deleteRefreshToken(): Promise<void> {
  await SecureStore.deleteItemAsync(TokenStorageKeys.REFRESH_TOKEN);
}

export async function getExpiresAt(): Promise<number | null> {
  try {
    const value = await SecureStore.getItemAsync(TokenStorageKeys.EXPIRES_AT);
    return value ? parseInt(value, 10) : null;
  } catch {
    return null;
  }
}

export async function setExpiresAt(expiresAt: number): Promise<void> {
  await SecureStore.setItemAsync(TokenStorageKeys.EXPIRES_AT, expiresAt.toString());
}

export async function deleteExpiresAt(): Promise<void> {
  await SecureStore.deleteItemAsync(TokenStorageKeys.EXPIRES_AT);
}

export async function getStoredUser(): Promise<MobileUser | null> {
  try {
    const value = await SecureStore.getItemAsync(TokenStorageKeys.USER);
    return value ? (JSON.parse(value) as MobileUser) : null;
  } catch {
    return null;
  }
}

export async function setStoredUser(user: MobileUser): Promise<void> {
  await SecureStore.setItemAsync(TokenStorageKeys.USER, JSON.stringify(user));
}

export async function deleteStoredUser(): Promise<void> {
  await SecureStore.deleteItemAsync(TokenStorageKeys.USER);
}

export async function getRememberMe(): Promise<boolean> {
  try {
    return (await SecureStore.getItemAsync(TokenStorageKeys.REMEMBER_ME)) === 'true';
  } catch {
    return false;
  }
}

export async function setRememberMe(rememberMe: boolean): Promise<void> {
  await SecureStore.setItemAsync(
    TokenStorageKeys.REMEMBER_ME,
    rememberMe ? 'true' : 'false'
  );
}

export async function clearAuthStorage(): Promise<void> {
  await Promise.all([
    SecureStore.deleteItemAsync(TokenStorageKeys.ACCESS_TOKEN),
    SecureStore.deleteItemAsync(TokenStorageKeys.REFRESH_TOKEN),
    SecureStore.deleteItemAsync(TokenStorageKeys.EXPIRES_AT),
    SecureStore.deleteItemAsync(TokenStorageKeys.USER),
    SecureStore.deleteItemAsync(TokenStorageKeys.REMEMBER_ME),
  ]);
}
