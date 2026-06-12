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

// The backend exposes only `PATCH /reminders/{id}` with an `active` flag — there
// is no dedicated dismiss/snooze endpoint. Dismissing a reminder deactivates it.
export async function setReminderActive(id: string, active: boolean): Promise<void> {
  await apiFetch(`/reminders/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ active }),
  });
}
