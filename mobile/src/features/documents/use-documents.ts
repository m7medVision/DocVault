import { useEffect, useState } from 'react';

import { listDocuments } from './api';
import type { Document, ListDocumentsOptions } from './types';

export function useDocuments(filters: ListDocumentsOptions) {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    async function load() {
      try {
        setLoading(true);
        setError(null);
        const response = await listDocuments({ ...filters, limit: 50 });
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

    return () => {
      active = false;
    };
  }, [filters]);

  return { documents, loading, error };
}
