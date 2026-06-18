'use client';

import { useState, useCallback } from 'react';
import {
  listAllFolders,
  createFolder,
  renameFolder,
  deleteFolder,
  moveFolder,
  moveDocument,
  updateDocumentTitle,
} from '@/lib/api/folders';
import type { Folder } from '@/lib/api/types';
import { toast } from 'sonner';

export interface UseFolderActionsOptions {
  onSuccess?: () => void;
}

export interface UseFolderActionsResult {
  folders: Folder[];
  loading: boolean;
  error: string | null;
  loadFolders: () => Promise<void>;
  create: (name: string, parentId?: string) => Promise<boolean>;
  rename: (folderId: string, newName: string) => Promise<boolean>;
  remove: (folderId: string) => Promise<boolean>;
  moveDoc: (documentId: string, folderId?: string) => Promise<boolean>;
  saveMove: (folderId: string, parentId: string | null) => Promise<boolean>;
  renameDoc: (documentId: string, title: string) => Promise<boolean>;
}

export function useFolderActions(
  options: UseFolderActionsOptions = {}
): UseFolderActionsResult {
  const [folders, setFolders] = useState<Folder[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadFolders = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await listAllFolders();
      setFolders(response.folders);
    } catch {
      setError('Failed to load folders');
    } finally {
      setLoading(false);
    }
  }, []);

  const create = useCallback(
    async (name: string, parentId?: string): Promise<boolean> => {
      try {
        await createFolder(name, parentId);
        toast.success('Folder created');
        await loadFolders();
        options.onSuccess?.();
        return true;
      } catch {
        toast.error('Failed to create folder');
        return false;
      }
    },
    [loadFolders, options]
  );

  const rename = useCallback(
    async (folderId: string, newName: string): Promise<boolean> => {
      try {
        await renameFolder(folderId, newName);
        toast.success('Folder renamed');
        await loadFolders();
        options.onSuccess?.();
        return true;
      } catch {
        toast.error('Failed to rename folder');
        return false;
      }
    },
    [loadFolders, options]
  );

  const remove = useCallback(
    async (folderId: string): Promise<boolean> => {
      try {
        await deleteFolder(folderId);
        toast.success('Folder deleted');
        await loadFolders();
        options.onSuccess?.();
        return true;
      } catch {
        toast.error('Failed to delete folder');
        return false;
      }
    },
    [loadFolders, options]
  );

  const moveDoc = useCallback(
    async (documentId: string, folderId?: string): Promise<boolean> => {
      try {
        const response = await moveDocument(documentId, folderId);
        toast.success(`Moved to ${response.folder_name || 'root'}`);
        options.onSuccess?.();
        return true;
      } catch {
        toast.error('Failed to move document');
        return false;
      }
    },
    [options]
  );

  const saveMove = useCallback(
    async (folderId: string, parentId: string | null): Promise<boolean> => {
      try {
        await moveFolder(folderId, parentId);
        toast.success('Folder moved');
        await loadFolders();
        options.onSuccess?.();
        return true;
      } catch (err) {
        const message =
          err instanceof Error ? err.message : 'Failed to move folder';
        toast.error(message);
        return false;
      }
    },
    [loadFolders, options]
  );

  const renameDoc = useCallback(
    async (documentId: string, title: string): Promise<boolean> => {
      try {
        await updateDocumentTitle(documentId, title);
        toast.success('Document renamed');
        options.onSuccess?.();
        return true;
      } catch {
        toast.error('Failed to rename document');
        return false;
      }
    },
    [options]
  );

  return {
    folders,
    loading,
    error,
    loadFolders,
    create,
    rename,
    remove,
    moveDoc,
    saveMove,
    renameDoc,
  };
}
