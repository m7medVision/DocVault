import { useQuery } from '@tanstack/react-query';
import { useAuth } from '@/lib/auth/auth-context';
import { getDocumentStats } from './api';

export function useDocumentStats() {
  const { isAuthenticated } = useAuth();

  return useQuery({
    queryKey: ['documentStats'],
    queryFn: getDocumentStats,
    enabled: isAuthenticated,
    refetchInterval: 30_000,
  });
}
