import { useState, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { searchDocuments, SearchResult, SearchFilters, SearchOutput } from './api';

export interface UseSearchOptions {
  limit?: number;
}

export function useSearch(options: UseSearchOptions = {}) {
  const { limit = 20 } = options;
  const { accessToken } = useAuth();

  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [cursor, setCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [total, setTotal] = useState(0);
  const [searchTime, setSearchTime] = useState<number | null>(null);

  const search = useCallback(
    async (q: string, filters?: SearchFilters, cursorValue?: string, isRefresh = false) => {
      if (!q.trim() || !accessToken) return;

      if (!cursorValue) {
        if (isRefresh) {
          setRefreshing(true);
        } else {
          setLoading(true);
        }
        setError(null);
      }

      const startTime = Date.now();

      try {
        const searchFiltersObj: SearchFilters = {};
        if (filters?.doc_type) searchFiltersObj.doc_type = filters.doc_type;
        if (filters?.language) searchFiltersObj.language = filters.language;
        if (filters?.folder_id) searchFiltersObj.folder_id = filters.folder_id;
        if (filters?.start_date) searchFiltersObj.start_date = filters.start_date;
        if (filters?.end_date) searchFiltersObj.end_date = filters.end_date;

        const output: SearchOutput = await searchDocuments(
          accessToken,
          q,
          Object.keys(searchFiltersObj).length > 0 ? searchFiltersObj : undefined,
          limit,
          cursorValue
        );

        if (cursorValue) {
          setResults((prev) => [...prev, ...output.results]);
        } else {
          setResults(output.results);
        }

        setTotal(output.total);
        setHasMore(output.has_more);
        setCursor(output.next_cursor ?? null);
        setSearchTime(Date.now() - startTime);
      } catch (err) {
        console.error('Search error:', err);
        setError('Search failed. Please try again.');
        if (!cursorValue) setResults([]);
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [accessToken, limit]
  );

  const loadMore = useCallback(
    (filters?: SearchFilters) => {
      if (hasMore && !loading && cursor) {
        search(query, filters, cursor);
      }
    },
    [hasMore, loading, cursor, query, search]
  );

  const clearResults = useCallback(() => {
    setResults([]);
    setCursor(null);
    setHasMore(false);
    setTotal(0);
    setSearchTime(null);
    setError(null);
  }, []);

  return {
    query,
    setQuery,
    results,
    loading,
    refreshing,
    error,
    hasMore,
    total,
    searchTime,
    search,
    loadMore,
    clearResults,
  };
}
