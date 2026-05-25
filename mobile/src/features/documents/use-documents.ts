import { useCallback, useEffect, useRef, useState } from 'react';

import { listDocuments } from './api';
import type { Document, ListDocumentsOptions } from './types';

export function useDocuments(filters: ListDocumentsOptions) {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const filtersRef = useRef(filters);

  filtersRef.current = filters;

  const reload = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await listDocuments({ ...filtersRef.current, limit: 50 });
      setDocuments(response.documents);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Failed to load documents');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let active = true;

    async function load() {
      try {
        setLoading(true);
        setError(null);
        const response = await listDocuments({ ...filtersRef.current, limit: 50 });
        if (active) setDocuments(response.documents);
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : 'Failed to load documents');
        }
      } finally {
        if (active) setLoading(false);
      }
    }

    void load();
    return () => { active = false; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(filters)]);

  return { documents, loading, error, reload };
}