import { apiFetch } from '@/lib/api/client';
import type { SearchResponse } from './types';

export interface SearchOptions {
  query: string;
  doc_type?: string;
  language?: string;
  status?: string;
  folder_id?: string;
  limit?: number;
}

export async function searchDocuments(options: SearchOptions): Promise<SearchResponse> {
  const response = await apiFetch<SearchResponse>('/search', {
    method: 'POST',
    body: JSON.stringify(options),
  });

  return {
    ...response,
    results: response?.results ?? [],
    total: response?.total ?? 0,
  };
}
