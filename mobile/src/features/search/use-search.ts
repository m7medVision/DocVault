import { useState } from 'react';

import { searchDocuments, type SearchOptions } from './api';
import type { SearchResult } from './types';

export function useSearch() {
  const [results, setResults] = useState<SearchResult[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);

  async function search(options: SearchOptions) {
    if (!options.query.trim()) return;

    try {
      setLoading(true);
      setError(null);
      setHasSearched(true);
      setResults([]);
      setTotal(0);
      const response = await searchDocuments(options);
      setResults(response.results);
      setTotal(response.total);
    } catch (err) {
      setResults([]);
      setTotal(0);
      setError(err instanceof Error ? err.message : 'Search failed');
    } finally {
      setLoading(false);
    }
  }

  function reset() {
    setResults([]);
    setTotal(0);
    setError(null);
    setHasSearched(false);
  }

  return { results, total, loading, error, hasSearched, search, reset };
}
