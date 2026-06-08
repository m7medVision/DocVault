import { useMemo } from 'react';
import { useFolders } from './use-folders';
import { useDocuments } from '@/features/documents/use-documents';
import type { Folder } from './types';

export interface FolderAncestor extends Folder {
  depth: number;
}

export interface UseFolderDetailOptions {
  id?: string;
}

export interface UseFolderDetailResult {
  folder: Folder | null;
  ancestors: FolderAncestor[];
  children: Folder[];
  documents: ReturnType<typeof useDocuments>['documents'];
  loading: boolean;
  documentsLoading: boolean;
  error: string | null;
  reload: () => Promise<unknown>;
}

export function useFolderDetail({ id }: UseFolderDetailOptions): UseFolderDetailResult {
  const foldersQuery = useFolders();
  const documentsQuery = useDocuments({ folder_id: id ?? '' });

  const folder = useMemo(() => {
    if (!id) return null;
    return foldersQuery.folders.find((f) => f.id === id) ?? null;
  }, [foldersQuery.folders, id]);

  const ancestors = useMemo<FolderAncestor[]>(() => {
    if (!folder) return [];
    const chain: FolderAncestor[] = [];
    let current: Folder | undefined = folder;
    const seen = new Set<string>();
    let depth = 0;
    while (current && !seen.has(current.id)) {
      seen.add(current.id);
      chain.unshift({ ...current, depth });
      depth += 1;
      if (!current.parent_id) break;
      current = foldersQuery.folders.find((f) => f.id === current!.parent_id);
    }
    return chain;
  }, [folder, foldersQuery.folders]);

  const children = useMemo(() => {
    if (!id) return [];
    return foldersQuery.folders.filter((f) => f.parent_id === id);
  }, [foldersQuery.folders, id]);

  return {
    folder,
    ancestors,
    children,
    documents: documentsQuery.documents,
    loading: foldersQuery.loading,
    documentsLoading: documentsQuery.loading,
    error: foldersQuery.error,
    reload: async () => {
      await Promise.all([foldersQuery.reload(), documentsQuery.reload()]);
    },
  };
}
