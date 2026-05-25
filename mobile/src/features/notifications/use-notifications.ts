import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getNotifications, markNotificationRead } from './api';
import type { NotificationsOptions } from './types';

export function useNotifications(options: NotificationsOptions = {}) {
  return useQuery({
    queryKey: ['notifications', options],
    queryFn: () => getNotifications(options),
    select: (data) => ({
      notifications: data.notifications,
      unreadCount: data.unread_count,
      total: data.total,
    }),
  });
}

export function useMarkRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: markNotificationRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}
