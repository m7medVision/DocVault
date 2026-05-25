import { apiFetch } from '@/lib/api/client';
import type { NotificationListResponse, NotificationsOptions } from './types';

export async function getNotifications(options: NotificationsOptions = {}): Promise<NotificationListResponse> {
  const params = new URLSearchParams();
  if (options.status) params.set('status', options.status);
  if (options.cursor) params.set('cursor', options.cursor);
  if (options.limit) params.set('limit', String(options.limit));

  const query = params.toString();
  const response = await apiFetch<NotificationListResponse>(`/notifications${query ? `?${query}` : ''}`);

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
