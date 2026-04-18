import { useMutation, useQueryClient } from '@tanstack/react-query';
import { markNotificationRead } from './api';

export function useMarkRead() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => markNotificationRead(id),
    onSuccess: () => {
      // Invalidate queries to refetch unread notifications and the counter
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}
