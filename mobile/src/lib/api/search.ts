// Search API calls to the backend

import { CONFIG } from '../config';
import { authorizedFetch } from '../auth';
import { handleResponse } from './client';
import { SearchResult, SearchOutput, SearchFilters } from './types';

export { SearchResult, SearchOutput, SearchFilters };

// POST /api/v1/search - Perform semantic search with vector + filters
export async function searchDocuments(
  accessToken: string,
  query: string,
  filters?: SearchFilters,
  limit: number = 20,
  cursor?: string
): Promise<SearchOutput> {
  const body: Record<string, unknown> = {
    query,
    limit,
  };

  if (filters) {
    if (filters.doc_type) body.doc_type = filters.doc_type;
    if (filters.language) body.language = filters.language;
    if (filters.folder_id) body.folder_id = filters.folder_id;
    if (filters.start_date) body.start_date = filters.start_date;
    if (filters.end_date) body.end_date = filters.end_date;
  }

  if (cursor) {
    body.cursor = cursor;
  }

  const response = await authorizedFetch(accessToken, `${CONFIG.apiBaseUrl}/search`, {
    method: 'POST',
    body: JSON.stringify(body),
  });

  return handleResponse<SearchOutput>(response);
}

// Format relevance score as percentage
export function formatRelevance(score: number): string {
  const relevance = Math.max(0, Math.min(100, (1 - score) * 100));
  return `${Math.round(relevance)}%`;
}

// Highlight matching text in snippet
export function highlightMatch(text: string, query: string): { text: string; isMatch: boolean }[] {
  if (!query.trim()) return [{ text, isMatch: false }];

  const parts: { text: string; isMatch: boolean }[] = [];
  const lowerText = text.toLowerCase();
  const lowerQuery = query.toLowerCase();
  let lastIndex = 0;

  const words = lowerQuery.split(/\s+/).filter(w => w.length > 2);

  for (const word of words) {
    let index = lowerText.indexOf(word);
    while (index !== -1) {
      if (index > lastIndex) {
        parts.push({ text: text.slice(lastIndex, index), isMatch: false });
      }
      parts.push({ text: text.slice(index, index + word.length), isMatch: true });
      lastIndex = index + word.length;
      index = lowerText.indexOf(word, lastIndex);
    }
  }

  if (lastIndex < text.length) {
    parts.push({ text: text.slice(lastIndex), isMatch: false });
  }

  return parts.length > 0 ? parts : [{ text, isMatch: false }];
}
