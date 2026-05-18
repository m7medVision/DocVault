import { apiFetch } from '@/lib/api/client';
import type { Reminder } from './types';

export interface ReminderListResponse {
  reminders: Reminder[];
}

export async function listReminders(): Promise<ReminderListResponse> {
  const response = await apiFetch<ReminderListResponse>('/reminders');
  return {
    ...response,
    reminders: response?.reminders ?? [],
  };
}

export async function dismissReminder(id: string): Promise<void> {
  await apiFetch(`/reminders/${id}/dismiss`, { method: 'PATCH' });
}

export async function snoozeReminder(id: string, minutes: number): Promise<void> {
  await apiFetch(`/reminders/${id}/snooze`, {
    method: 'PATCH',
    body: JSON.stringify({ snooze_minutes: minutes }),
  });
}
