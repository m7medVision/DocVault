import { apiFetch } from '@/lib/api/client';
import type { SearchResponse } from './types';

export interface SearchOptions {
  query: string;
  doc_type?: string;
  language?: string;
  folder_id?: string;
  limit?: number;
}

export async function searchDocuments(options: SearchOptions): Promise<SearchResponse> {
  return apiFetch<SearchResponse>('/search', {
    method: 'POST',
    body: JSON.stringify(options),
  });
}
