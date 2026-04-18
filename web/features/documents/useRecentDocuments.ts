import { useQuery } from '@tanstack/react-query';
import { listDocuments } from '@/lib/api/documents';
import { useAuth } from '@/lib/useAuth';

export function useRecentDocuments(limit: number = 10) {
  const { isAuthenticated } = useAuth();
  
  return useQuery({
    queryKey: ['recentDocuments', limit],
    queryFn: () => listDocuments({ limit }),
    enabled: isAuthenticated,
    refetchInterval: 30000,
  });
}
