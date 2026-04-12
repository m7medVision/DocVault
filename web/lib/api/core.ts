import {
  getClientAccessToken,
  refreshClientAccessToken,
} from '@/lib/client-auth';
import { API_BASE_URL } from '@/lib/auth';

interface ApiErrorPayload {
  error?: string;
}

export async function parseJson<T>(response: Response): Promise<T> {
  if (response.status === 204) {
    return undefined as T;
  }

  const contentType = response.headers.get('content-type') || '';
  if (!contentType.includes('application/json')) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {},
  allowRetry = true
): Promise<T> {
  const headers = new Headers(options.headers || {});
  let token = getClientAccessToken();

  if (!token && allowRetry) {
    token = await refreshClientAccessToken();
  }

  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
    cache: 'no-store',
  });

  if (response.status === 401 && allowRetry) {
    const refreshedToken = await refreshClientAccessToken();
    if (refreshedToken) {
      return apiFetch<T>(endpoint, options, false);
    }
  }

  if (!response.ok) {
    const error = await parseJson<ApiErrorPayload>(response).catch(() => null);
    throw new Error(error?.error || `API error: ${response.status}`);
  }

  return parseJson<T>(response);
}
