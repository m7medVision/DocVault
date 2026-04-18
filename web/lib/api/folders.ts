import { apiFetch } from './core';
import type {
  Folder,
  FolderListResponse,
  MoveDocumentResponse,
} from './types';

export async function listFolders(
  parentId?: string
): Promise<FolderListResponse> {
  const query = parentId ? `?parent_id=${parentId}` : '';
  return apiFetch<FolderListResponse>(`/folders${query}`);
}

export async function listAllFolders(): Promise<FolderListResponse> {
  return apiFetch<FolderListResponse>('/folders/all');
}

export async function createFolder(
  name: string,
  parentId?: string
): Promise<{ folder: Folder }> {
  const body: { name: string; parent_id?: string } = { name };
  if (parentId) {
    body.parent_id = parentId;
  }
  return apiFetch<{ folder: Folder }>('/folders', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function renameFolder(
  id: string,
  name: string
): Promise<{ id: string; name: string }> {
  return apiFetch<{ id: string; name: string }>(`/folders/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ name }),
  });
}

export async function deleteFolder(id: string): Promise<void> {
  await apiFetch(`/folders/${id}`, { method: 'DELETE' });
}

export async function moveDocument(
  id: string,
  folderId?: string
): Promise<MoveDocumentResponse> {
  return apiFetch<MoveDocumentResponse>(`/documents/${id}/move`, {
    method: 'PATCH',
    body: JSON.stringify({ folder_id: folderId }),
  });
}

export async function updateDocumentTitle(
  id: string,
  title: string
): Promise<{ id: string; title: string }> {
  return apiFetch<{ id: string; title: string }>(`/documents/${id}/title`, {
    method: 'PATCH',
    body: JSON.stringify({ title }),
  });
}
