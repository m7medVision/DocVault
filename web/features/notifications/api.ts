import { apiFetch } from '@/lib/api/core';

export interface Notification {
  id: string;
  tenant_id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  link?: string;
  status: 'unread' | 'read';
  metadata?: Record<string, unknown>;
  created_at: string;
  read_at?: string;
}

export interface ListNotificationsResponse {
  notifications: Notification[];
  cursor?: string;
  unread_count: number;
}

export interface ListNotificationsOptions {
  status?: 'unread' | 'read';
  cursor?: string;
  limit?: number;
}

export async function listNotifications(options: ListNotificationsOptions = {}): Promise<ListNotificationsResponse> {
  const query = new URLSearchParams();
  if (options.status) query.append('status', options.status);
  if (options.cursor) query.append('cursor', options.cursor);
  if (options.limit) query.append('limit', options.limit.toString());

  const queryString = query.toString();
  const url = queryString ? `/notifications?${queryString}` : '/notifications';

  return apiFetch<ListNotificationsResponse>(url);
}

export async function markNotificationRead(id: string): Promise<void> {
  return apiFetch<void>(`/notifications/${id}/read`, {
    method: 'PATCH',
  });
}
