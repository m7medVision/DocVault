import { useQuery } from '@tanstack/react-query';
import { listTags } from '@/lib/api/tags';
import { useAuth } from '@/lib/useAuth';

export function useTags(query = '') {
  const { isAuthenticated } = useAuth();
  return useQuery({
    queryKey: ['tags', query],
    queryFn: () => listTags(query, 100),
    enabled: isAuthenticated,
  });
}
