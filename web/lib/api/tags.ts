import { apiFetch } from './core';

export interface Tag {
  id: string;
  tenant_id: string;
  name: string;
  created_at: string;
}

export interface TagListResponse {
  tags: Tag[];
}

export async function listTags(query = '', limit = 50): Promise<TagListResponse> {
  const params = new URLSearchParams();
  if (query) params.set('q', query);
  if (limit) params.set('limit', String(limit));

  const search = params.toString();
  const response = await apiFetch<TagListResponse>(
    `/tags${search ? `?${search}` : ''}`
  );

  return {
    ...response,
    tags: response?.tags ?? [],
  };
}
