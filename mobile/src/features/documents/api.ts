import { apiFetch } from '@/lib/api/client';
import type {
  DocumentListResponse,
  ListDocumentsOptions,
  DocumentDetailResponse,
  DocumentPagesResponse,
  DocumentVersionsResponse,
  DocumentDownloadResponse,
  DocumentStats,
} from './types';

export async function getDocumentStats(): Promise<DocumentStats> {
  return apiFetch<DocumentStats>('/documents/stats');
}

export async function listDocuments(options: ListDocumentsOptions = {}): Promise<DocumentListResponse> {
  const params = new URLSearchParams();

  if (options.type) params.set('type', options.type);
  if (options.folder_id) params.set('folder_id', options.folder_id);
  if (options.status) params.set('status', options.status);
  if (options.language) params.set('language', options.language);
  if (options.cursor) params.set('cursor', options.cursor);
  if (options.limit) params.set('limit', String(options.limit));

  const query = params.toString();
  const response = await apiFetch<DocumentListResponse>(`/documents${query ? `?${query}` : ''}`);

  return {
    ...response,
    documents: response?.documents ?? [],
    total: response?.total ?? 0,
  };
}

export async function uploadDocument(fileUri: string, fileName: string, mimeType: string, options?: { title?: string; doc_type?: string; folder_id?: string }) {
  const formData = new FormData();

  formData.append('file', {
    uri: fileUri,
    name: fileName,
    type: mimeType,
  } as unknown as Blob);

  formData.append('title', options?.title || fileName);
  formData.append('doc_type', options?.doc_type || 'other');

  if (options?.folder_id) {
    formData.append('folder_id', options.folder_id);
  }

  return apiFetch<{ id: string; status: string; message: string; title: string }>('/documents/upload', {
    method: 'POST',
    body: formData,
  });
}

export async function getDocument(id: string): Promise<DocumentDetailResponse> {
  return apiFetch<DocumentDetailResponse>(`/documents/${id}`);
}

export async function getDocumentPages(id: string): Promise<DocumentPagesResponse> {
  return apiFetch<DocumentPagesResponse>(`/documents/${id}/pages`);
}

export async function getDocumentVersions(id: string): Promise<DocumentVersionsResponse> {
  return apiFetch<DocumentVersionsResponse>(`/documents/${id}/versions`);
}

export async function downloadDocument(id: string): Promise<DocumentDownloadResponse> {
  return apiFetch<DocumentDownloadResponse>(`/documents/${id}/download`);
}

export async function updateDocumentMetadata(
  id: string,
  updates: Record<string, string>,
): Promise<void> {
  await apiFetch(`/documents/${id}/metadata`, {
    method: 'PATCH',
    body: JSON.stringify(updates),
  });
}
