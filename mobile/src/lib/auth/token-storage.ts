import * as SecureStore from 'expo-secure-store';

const KEYS = {
  ACCESS_TOKEN: 'dv_access_token',
  REFRESH_TOKEN: 'dv_refresh_token',
  USER: 'dv_user',
} as const;

export async function getStoredAccessToken(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.ACCESS_TOKEN);
}

export async function setStoredAccessToken(token: string | null) {
  if (!token) {
    await SecureStore.deleteItemAsync(KEYS.ACCESS_TOKEN);
    return;
  }
  await SecureStore.setItemAsync(KEYS.ACCESS_TOKEN, token);
}

export async function getStoredRefreshToken(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.REFRESH_TOKEN);
}

export async function setStoredRefreshToken(token: string | null) {
  if (!token) {
    await SecureStore.deleteItemAsync(KEYS.REFRESH_TOKEN);
    return;
  }
  await SecureStore.setItemAsync(KEYS.REFRESH_TOKEN, token);
}

export async function getStoredUser(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.USER);
}

export async function setStoredUser(json: string | null) {
  if (!json) {
    await SecureStore.deleteItemAsync(KEYS.USER);
    return;
  }
  await SecureStore.setItemAsync(KEYS.USER, json);
}

export async function clearStoredSession() {
  await Promise.all([
    SecureStore.deleteItemAsync(KEYS.ACCESS_TOKEN),
    SecureStore.deleteItemAsync(KEYS.REFRESH_TOKEN),
    SecureStore.deleteItemAsync(KEYS.USER),
  ]);
}
