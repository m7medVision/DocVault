export interface SearchFilters {
  doc_type: string;
  language: string;
}

export interface NormalizedSearchResult {
  document_id: string;
  file: string;
  max_score: number;
}

export { type SearchResult, type SearchResponse, type SearchOptions } from '@/lib/api/search';
