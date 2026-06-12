import { useInfiniteQuery } from '@tanstack/react-query';
import { listNotifications } from './api';
import { useAuth } from '@/lib/useAuth';

export function useInfiniteNotifications(status?: 'unread' | 'read') {
  const { isAuthenticated } = useAuth();

  return useInfiniteQuery({
    queryKey: ['notifications', 'infinite', status ?? 'all'],
    queryFn: ({ pageParam }) =>
      listNotifications({ status, cursor: pageParam, limit: 20 }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.cursor || undefined,
    enabled: isAuthenticated,
  });
}
