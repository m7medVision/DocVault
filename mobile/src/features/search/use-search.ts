import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';

import { searchDocuments, type SearchOptions } from './api';
import type { SearchResult } from './types';

export function useSearch() {
  const [hasSearched, setHasSearched] = useState(false);

  const mutation = useMutation({
    mutationFn: searchDocuments,
    onSuccess: () => setHasSearched(true),
    onError: () => setHasSearched(true),
  });

  async function search(options: SearchOptions) {
    if (!options.query.trim()) return;
    return mutation.mutateAsync(options);
  }

  function reset() {
    mutation.reset();
    setHasSearched(false);
  }

  return {
    results: (mutation.data?.results ?? []) as SearchResult[],
    total: mutation.data?.total ?? 0,
    loading: mutation.isPending,
    error: mutation.error ? (mutation.error instanceof Error ? mutation.error.message : 'Search failed') : null,
    hasSearched,
    search,
    reset,
  };
}
