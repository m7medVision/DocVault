import { apiFetch } from './core';
import type {
  DocumentListResponse,
  ListDocumentsOptions,
  UploadDocumentOptions,
  UploadDocumentResponse,
  DocumentDetailResponse,
  DocumentPagesResponse,
  DocumentVersionsResponse,
  DocumentDownloadResponse,
} from './types';

export async function listDocuments(
  options: ListDocumentsOptions = {}
): Promise<DocumentListResponse> {
  const params = new URLSearchParams();

  if (options.type) params.set('type', options.type);
  if (options.folder_id) params.set('folder_id', options.folder_id);
  if (options.status) params.set('status', options.status);
  if (options.language) params.set('language', options.language);
  if (options.cursor) params.set('cursor', options.cursor);
  if (options.limit) params.set('limit', options.limit.toString());

  const query = params.toString();
  const response = await apiFetch<DocumentListResponse>(
    `/documents${query ? `?${query}` : ''}`
  );

  return {
    ...response,
    documents: response?.documents ?? [],
    total: response?.total ?? 0,
  };
}

export async function uploadDocument(
  file: File,
  options: UploadDocumentOptions = {}
): Promise<UploadDocumentResponse> {
  const formData = new FormData();
  formData.set('file', file);
  formData.set('title', options.title || file.name);
  formData.set('doc_type', options.doc_type || 'other');

  if (options.folder_id) {
    formData.set('folder_id', options.folder_id);
  }
  if (options.language) {
    formData.set('language', options.language);
  }

  return apiFetch<UploadDocumentResponse>('/documents/upload', {
    method: 'POST',
    body: formData,
  });
}

export async function getDocument(id: string): Promise<DocumentDetailResponse> {
  return apiFetch<DocumentDetailResponse>(`/documents/${id}`);
}

export async function getDocumentPages(
  id: string
): Promise<DocumentPagesResponse> {
  return apiFetch<DocumentPagesResponse>(`/documents/${id}/pages`);
}

export async function updateDocumentMetadata(
  id: string,
  updates: Record<string, string>
): Promise<void> {
  await apiFetch(`/documents/${id}/metadata`, {
    method: 'PATCH',
    body: JSON.stringify(updates),
  });
}

export async function deleteDocument(id: string): Promise<void> {
  await apiFetch(`/documents/${id}`, { method: 'DELETE' });
}

export async function getDocumentVersions(
  id: string
): Promise<DocumentVersionsResponse> {
  return apiFetch<DocumentVersionsResponse>(`/documents/${id}/versions`);
}

export async function downloadDocument(
  id: string
): Promise<DocumentDownloadResponse> {
  return apiFetch<DocumentDownloadResponse>(`/documents/${id}/download`);
}