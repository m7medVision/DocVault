'use client';

import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'next/navigation';
import { searchDocuments, type SearchResult, type SearchOptions } from './api';
import type { SearchFilters } from './types';

export interface UseSearchReturn {
  results: SearchResult[];
  loading: boolean;
  error: string | null;
  query: string;
  setQuery: (query: string) => void;
  filters: SearchFilters;
  setFilter: (key: keyof SearchFilters, value: string) => void;
  performSearch: (q: string, f: SearchFilters) => Promise<void>;
}

export function useSearch(): UseSearchReturn {
  const searchParams = useSearchParams();
  const queryParam = searchParams.get('q') || '';

  const [query, setQuery] = useState(queryParam);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<SearchFilters>({
    doc_type: '',
    language: '',
  });

  const performSearch = useCallback(
    async (searchQuery: string, searchFilters: SearchFilters) => {
      if (!searchQuery.trim()) return;

      setIsLoading(true);
      setError(null);

      try {
        const options: SearchOptions = {
          query: searchQuery,
          doc_type: searchFilters.doc_type || undefined,
          language: searchFilters.language || undefined,
          limit: 20,
        };
        const response = await searchDocuments(options);
        setResults(response.results);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Search failed');
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  useEffect(() => {
    if (queryParam) {
      setQuery(queryParam);
      performSearch(queryParam, filters);
    }
  }, [queryParam, filters, performSearch]);

  const handleFilterChange = (key: keyof SearchFilters, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  return {
    results,
    loading: isLoading,
    error,
    query,
    setQuery,
    filters,
    setFilter: handleFilterChange,
    performSearch,
  };
}
