import { apiFetch } from './core';
import type { BackendUserResponse } from '@/lib/auth/types';

export interface UpdateProfileInput {
  display_name?: string;
  locale?: string;
}

export interface UpdateEmailInput {
  email: string;
  current_password: string;
}

export interface UpdatePasswordInput {
  current_password: string;
  new_password: string;
}

export async function updateProfile(
  input: UpdateProfileInput
): Promise<BackendUserResponse> {
  return apiFetch<BackendUserResponse>('/profile', {
    method: 'PATCH',
    body: JSON.stringify(input),
  });
}

export async function updateEmail(
  input: UpdateEmailInput
): Promise<BackendUserResponse> {
  return apiFetch<BackendUserResponse>('/profile/email', {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export async function updatePassword(input: UpdatePasswordInput): Promise<void> {
  await apiFetch('/profile/password', {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}