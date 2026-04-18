import { useQuery } from '@tanstack/react-query';
import { listNotifications, type ListNotificationsOptions } from './api';
import { useAuth } from '@/lib/useAuth';

export function useNotifications(options: ListNotificationsOptions = {}) {
  const { isAuthenticated } = useAuth();
  
  return useQuery({
    queryKey: ['notifications', options],
    queryFn: () => listNotifications(options),
    enabled: isAuthenticated,
    refetchInterval: 60000, // Refetch every minute
  });
}
