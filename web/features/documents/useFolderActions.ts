'use client';

import { useState, useCallback } from 'react';
import {
  listAllFolders,
  createFolder,
  renameFolder,
  deleteFolder,
  moveDocument,
  updateDocumentTitle,
  suggestFolder,
} from '@/lib/api/folders';
import type { Folder, SuggestFolderResponse } from '@/lib/api/types';
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
  renameDoc: (documentId: string, title: string) => Promise<boolean>;
  suggest: (documentId: string) => Promise<SuggestFolderResponse | null>;
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

  const suggest = useCallback(
    async (documentId: string): Promise<SuggestFolderResponse | null> => {
      try {
        const response = await suggestFolder(documentId);
        return response.suggestion;
      } catch {
        toast.error('Failed to get suggestion');
        return null;
      }
    },
    []
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
    renameDoc,
    suggest,
  };
}
