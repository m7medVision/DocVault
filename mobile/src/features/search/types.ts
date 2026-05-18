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
