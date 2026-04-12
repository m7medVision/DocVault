import { apiFetch } from './core';

export interface SearchResult {
  document_id: string;
  file: string;
  max_score: number;
}

export interface SearchResponse {
  results: SearchResult[];
  query: string;
  total: number;
  page: number;
  page_size: number;
  has_more: boolean;
  next_cursor?: string;
}

export interface SearchOptions {
  query: string;
  doc_type?: string;
  language?: string;
  status?: string;
  folder_id?: string;
  limit?: number;
}

export async function searchDocuments(
  options: SearchOptions
): Promise<SearchResponse> {
  return apiFetch<SearchResponse>('/search', {
    method: 'POST',
    body: JSON.stringify(options),
  });
}