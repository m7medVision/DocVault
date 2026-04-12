'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  listDocuments,
  type Document,
  type ListDocumentsOptions,
} from '@/features/documents/api';
import { useAuth } from '@/lib/useAuth';

export interface DocumentFilters {
  type?: string;
  status?: string;
}

export interface UseDocumentListResult {
  documents: Document[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useDocumentList(
  filters: DocumentFilters = {}
): UseDocumentListResult {
  const { isLoading: authLoading, isAuthenticated } = useAuth();
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const refetch = useCallback(() => {
    setRefreshKey((k) => k + 1);
  }, []);

  useEffect(() => {
    if (authLoading) {
      return;
    }

    if (!isAuthenticated) {
      setDocuments([]);
      setError(null);
      setLoading(false);
      return;
    }

    async function loadDocuments() {
      try {
        setLoading(true);
        const options: ListDocumentsOptions = {
          type: filters.type || undefined,
          status: filters.status || undefined,
          limit: 20,
        };
        const response = await listDocuments(options);
        setDocuments(response.documents);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load documents');
      } finally {
        setLoading(false);
      }
    }

    loadDocuments();
  }, [authLoading, isAuthenticated, filters.type, filters.status, refreshKey]);

  return { documents, loading, error, refetch };
}
