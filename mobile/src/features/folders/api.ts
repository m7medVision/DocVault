import { apiFetch } from '@/lib/api/client';
import type {
  Folder,
  FolderListResponse,
  MoveDocumentResponse,
} from './types';

export async function listAllFolders(): Promise<FolderListResponse> {
  const response = await apiFetch<FolderListResponse>('/folders/all');
  return {
    ...response,
    folders: response?.folders ?? [],
  };
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
    body: JSON.stringify({ folder_id: folderId ?? null }),
  });
}
