import { useQuery } from '@tanstack/react-query';

import { listDocuments } from './api';
import type { ListDocumentsOptions } from './types';

export function useDocuments(filters: ListDocumentsOptions) {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['documents', filters],
    queryFn: () => listDocuments({ ...filters, limit: 50 }),
  });

  return {
    documents: data?.documents ?? [],
    loading: isLoading,
    error: error ? (error instanceof Error ? error.message : 'Failed to load documents') : null,
    reload: refetch,
  };
}
