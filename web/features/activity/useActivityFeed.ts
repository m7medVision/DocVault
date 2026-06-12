import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useInfiniteNotifications } from '@/features/notifications/useInfiniteNotifications';
import { getAuditLog } from '@/lib/api/admin';
import { useAuth } from '@/lib/useAuth';

export interface ActivityItem {
  id: string;
  kind: 'notification' | 'audit';
  title: string;
  subtitle?: string;
  actor?: string;
  action?: string;
  created_at: string;
  link?: string;
}

export function useActivityFeed() {
  const { user, isAuthenticated } = useAuth();
  const isAdmin = user?.role === 'admin' || user?.role === 'owner';

  const notificationsQuery = useInfiniteNotifications();

  const auditQuery = useQuery({
    queryKey: ['activity', 'audit'],
    queryFn: () => getAuditLog({}),
    enabled: isAuthenticated && isAdmin,
  });

  const items = useMemo<ActivityItem[]>(() => {
    const notifications =
      notificationsQuery.data?.pages.flatMap((p) => p.notifications) ?? [];
    const auditEvents = auditQuery.data?.events ?? [];

    const merged: ActivityItem[] = [
      ...notifications.map((n) => ({
        id: `n-${n.id}`,
        kind: 'notification' as const,
        title: n.title,
        subtitle: n.body,
        created_at: n.created_at,
        link: n.link,
      })),
      ...auditEvents.map((e) => ({
        id: `a-${e.id}`,
        kind: 'audit' as const,
        title: `${e.action} · ${e.entity_type}`,
        subtitle: e.entity_id,
        actor: e.actor_id,
        action: e.action,
        created_at: e.created_at,
      })),
    ];

    return merged.sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );
  }, [notificationsQuery.data, auditQuery.data]);

  return {
    items,
    isLoading: notificationsQuery.isLoading || (isAdmin && auditQuery.isLoading),
    isAdmin,
    fetchNextPage: notificationsQuery.fetchNextPage,
    hasNextPage: notificationsQuery.hasNextPage,
    isFetchingNextPage: notificationsQuery.isFetchingNextPage,
  };
}
