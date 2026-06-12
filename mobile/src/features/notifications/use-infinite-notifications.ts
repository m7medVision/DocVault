import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { getNotifications, markNotificationRead } from './api';

export function useInfiniteNotifications(status?: 'unread' | 'read') {
  return useInfiniteQuery({
    queryKey: ['notifications', 'infinite', status ?? 'all'],
    queryFn: ({ pageParam }) =>
      getNotifications({ status, cursor: pageParam, limit: 20 }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.cursor || undefined,
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
