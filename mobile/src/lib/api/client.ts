// Base authenticated fetch client for API calls
// Provides error handling and response parsing

import { CONFIG } from '../config';
import { authorizedFetch } from '../auth';
import { ApiError } from './types';

export { ApiError };

export { CONFIG } from '../config';

export async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error: ApiError = await response.json().catch(() => ({
      error: 'Request failed',
      code: response.status.toString(),
      request_id: '',
    }));
    throw error;
  }
  return response.json();
}

export { authorizedFetch };
