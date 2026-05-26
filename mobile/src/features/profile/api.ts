import { apiFetch } from '@/lib/api/client';
import type { UpdateEmailInput, UpdatePasswordInput, UpdateProfileInput, UserResponse } from './types';

export async function updateProfile(input: UpdateProfileInput): Promise<UserResponse> {
  return apiFetch<UserResponse>('/profile', {
    method: 'PATCH',
    body: JSON.stringify(input),
  });
}

export async function updateEmail(input: UpdateEmailInput): Promise<UserResponse> {
  return apiFetch<UserResponse>('/profile/email', {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export async function updatePassword(input: UpdatePasswordInput): Promise<void> {
  return apiFetch<void>('/profile/password', {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}