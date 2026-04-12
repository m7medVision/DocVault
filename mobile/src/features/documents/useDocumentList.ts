import { useState, useEffect, useCallback } from 'react';
import { listDocuments } from './api';
import type { Document, ListDocumentsOptions } from './api';
import type { FilterState } from './types';

interface UseDocumentListResult {
  documents: Document[];
  loading: boolean;
  refreshing: boolean;
  error: string | null;
  refetch: () => void;
}

export function useDocumentList(
  accessToken: string | null,
  filters: FilterState
): UseDocumentListResult {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchDocuments = useCallback(async (isRefresh = false) => {
    if (!accessToken) return;

    if (isRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }

    try {
      const options: ListDocumentsOptions = {
        type: filters.type || undefined,
        folder_id: filters.folder_id || undefined,
        status: filters.status || undefined,
        limit: 20,
      };

      const response = await listDocuments(accessToken, options);
      setDocuments(response.documents || []);
      setError(null);
    } catch (err) {
      console.error('Error fetching documents:', err);
      setError('Failed to load documents');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [accessToken, filters]);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  const refetch = useCallback(() => {
    fetchDocuments(true);
  }, [fetchDocuments]);

  return { documents, loading, refreshing, error, refetch };
}

export function useActiveFilterCount(filters: FilterState): number {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const newCount = [filters.type, filters.folder_id, filters.status].filter(Boolean).length;
    setCount(newCount);
  }, [filters]);

  return count;
}
