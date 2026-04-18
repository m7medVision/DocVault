import { useQuery } from '@tanstack/react-query';
import { getDocumentStats } from '@/lib/api/documents';
import { useAuth } from '@/lib/useAuth';

export function useDocumentStats() {
  const { isAuthenticated } = useAuth();
  
  return useQuery({
    queryKey: ['documentStats'],
    queryFn: getDocumentStats,
    enabled: isAuthenticated,
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}
