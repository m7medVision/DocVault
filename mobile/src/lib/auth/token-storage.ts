import { Platform } from 'react-native';
import * as SecureStore from 'expo-secure-store';

const KEYS = {
  ACCESS_TOKEN: 'dv_access_token',
  REFRESH_TOKEN: 'dv_refresh_token',
  USER: 'dv_user',
} as const;

function webSetItem(key: string, value: string | null) {
  if (value === null) {
    localStorage.removeItem(key);
  } else {
    localStorage.setItem(key, value);
  }
}

function webGetItem(key: string): string | null {
  return localStorage.getItem(key);
}

function isSecureStoreAvailable(): boolean {
  if (Platform.OS === 'web') return false;
  return true;
}

export async function getStoredAccessToken(): Promise<string | null> {
  if (!isSecureStoreAvailable()) return webGetItem(KEYS.ACCESS_TOKEN);
  try {
    return await SecureStore.getItemAsync(KEYS.ACCESS_TOKEN);
  } catch {
    return null;
  }
}

export async function setStoredAccessToken(token: string | null): Promise<void> {
  if (!isSecureStoreAvailable()) {
    webSetItem(KEYS.ACCESS_TOKEN, token);
    return;
  }
  try {
    if (!token) {
      await SecureStore.deleteItemAsync(KEYS.ACCESS_TOKEN);
    } else {
      await SecureStore.setItemAsync(KEYS.ACCESS_TOKEN, token);
    }
  } catch {
    webSetItem(KEYS.ACCESS_TOKEN, token);
  }
}

export async function getStoredRefreshToken(): Promise<string | null> {
  if (!isSecureStoreAvailable()) return webGetItem(KEYS.REFRESH_TOKEN);
  try {
    return await SecureStore.getItemAsync(KEYS.REFRESH_TOKEN);
  } catch {
    return null;
  }
}

export async function setStoredRefreshToken(token: string | null): Promise<void> {
  if (!isSecureStoreAvailable()) {
    webSetItem(KEYS.REFRESH_TOKEN, token);
    return;
  }
  try {
    if (!token) {
      await SecureStore.deleteItemAsync(KEYS.REFRESH_TOKEN);
    } else {
      await SecureStore.setItemAsync(KEYS.REFRESH_TOKEN, token);
    }
  } catch {
    webSetItem(KEYS.REFRESH_TOKEN, token);
  }
}

export async function getStoredUser(): Promise<string | null> {
  if (!isSecureStoreAvailable()) return webGetItem(KEYS.USER);
  try {
    return await SecureStore.getItemAsync(KEYS.USER);
  } catch {
    return null;
  }
}

export async function setStoredUser(json: string | null): Promise<void> {
  if (!isSecureStoreAvailable()) {
    webSetItem(KEYS.USER, json);
    return;
  }
  try {
    if (!json) {
      await SecureStore.deleteItemAsync(KEYS.USER);
    } else {
      await SecureStore.setItemAsync(KEYS.USER, json);
    }
  } catch {
    webSetItem(KEYS.USER, json);
  }
}

export async function clearStoredSession(): Promise<void> {
  if (!isSecureStoreAvailable()) {
    localStorage.removeItem(KEYS.ACCESS_TOKEN);
    localStorage.removeItem(KEYS.REFRESH_TOKEN);
    localStorage.removeItem(KEYS.USER);
    return;
  }
  try {
    await Promise.all([
      SecureStore.deleteItemAsync(KEYS.ACCESS_TOKEN),
      SecureStore.deleteItemAsync(KEYS.REFRESH_TOKEN),
      SecureStore.deleteItemAsync(KEYS.USER),
    ]);
  } catch {
    localStorage.removeItem(KEYS.ACCESS_TOKEN);
    localStorage.removeItem(KEYS.REFRESH_TOKEN);
    localStorage.removeItem(KEYS.USER);
  }
}