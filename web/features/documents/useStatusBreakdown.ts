import { useQuery } from '@tanstack/react-query';
import { listDocuments } from '@/lib/api/documents';
import { useAuth } from '@/lib/useAuth';

const STATUSES = ['processed', 'pending', 'processing', 'failed'] as const;
const TYPES = ['invoice', 'contract', 'identity', 'warranty', 'receipt', 'other'] as const;

export interface Breakdown {
  byStatus: Record<string, number>;
  byType: Record<string, number>;
}

async function fetchBreakdown(): Promise<Breakdown> {
  const [statusCounts, typeCounts] = await Promise.all([
    Promise.all(
      STATUSES.map((status) =>
        listDocuments({ status, limit: 1 })
          .then((r) => [status, r.total ?? 0] as const)
          .catch(() => [status, 0] as const)
      )
    ),
    Promise.all(
      TYPES.map((type) =>
        listDocuments({ type, limit: 1 })
          .then((r) => [type, r.total ?? 0] as const)
          .catch(() => [type, 0] as const)
      )
    ),
  ]);

  return {
    byStatus: Object.fromEntries(statusCounts),
    byType: Object.fromEntries(typeCounts),
  };
}

export function useStatusBreakdown() {
  const { isAuthenticated } = useAuth();
  return useQuery({
    queryKey: ['documentBreakdown'],
    queryFn: fetchBreakdown,
    enabled: isAuthenticated,
  });
}
