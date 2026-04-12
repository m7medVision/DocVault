import { apiFetch } from './core';

export interface Notification {
  id: string;
  tenant_id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  link?: string;
  status: string;
  created_at: string;
  read_at?: string;
}

export interface NotificationListResponse {
  notifications: Notification[];
  total: number;
  unread_count: number;
}

export interface Reminder {
  id: string;
  document_id: string;
  tenant_id: string;
  rule_type: string;
  trigger_date: string;
  notify_days_before: number[];
  source: string;
  active: boolean;
  created_at: string;
}

export interface ReminderListResponse {
  reminders: Reminder[];
}

export async function listNotifications(): Promise<NotificationListResponse> {
  const response = await apiFetch<NotificationListResponse>('/notifications');

  return {
    ...response,
    notifications: response?.notifications ?? [],
    total: response?.total ?? 0,
    unread_count: response?.unread_count ?? 0,
  };
}

export async function markNotificationRead(id: string): Promise<void> {
  await apiFetch(`/notifications/${id}/read`, { method: 'PATCH' });
}

export async function listReminders(): Promise<ReminderListResponse> {
  const response = await apiFetch<ReminderListResponse>('/reminders');

  return {
    ...response,
    reminders: response?.reminders ?? [],
  };
}

export async function updateReminder(
  id: string,
  active: boolean
): Promise<void> {
  await apiFetch(`/reminders/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ active }),
  });
}