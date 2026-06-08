import { useCallback } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import {
  createFolder,
  deleteFolder,
  listAllFolders,
  moveDocument,
  renameFolder,
} from './api';
import type { Folder, FolderListResponse } from './types';

const FOLDERS_KEY = ['folders'] as const;

export function useFolders() {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: FOLDERS_KEY,
    queryFn: listAllFolders,
  });

  const invalidate = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: FOLDERS_KEY });
    queryClient.invalidateQueries({ queryKey: ['documents'] });
  }, [queryClient]);

  const createMutation = useMutation({
    mutationFn: ({ name, parentId }: { name: string; parentId?: string }) =>
      createFolder(name, parentId),
    onMutate: async ({ name, parentId }) => {
      await queryClient.cancelQueries({ queryKey: FOLDERS_KEY });
      const previous = queryClient.getQueryData<FolderListResponse>(FOLDERS_KEY);
      queryClient.setQueryData<FolderListResponse>(FOLDERS_KEY, (old) => {
        const folders = old?.folders ?? [];
        const tempId = `temp-${Date.now()}`;
        const placeholder: Folder = {
          id: tempId,
          tenant_id: '',
          org_id: '',
          parent_id: parentId,
          name,
          created_at: new Date().toISOString(),
        };
        return { folders: [...folders, placeholder] };
      });
      return { previous };
    },
    onError: (_err, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(FOLDERS_KEY, context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: FOLDERS_KEY });
    },
  });

  const renameMutation = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      renameFolder(id, name),
    onMutate: async ({ id, name }) => {
      await queryClient.cancelQueries({ queryKey: FOLDERS_KEY });
      const previous = queryClient.getQueryData<FolderListResponse>(FOLDERS_KEY);
      queryClient.setQueryData<FolderListResponse>(FOLDERS_KEY, (old) => {
        const folders = old?.folders ?? [];
        return {
          folders: folders.map((f) => (f.id === id ? { ...f, name } : f)),
        };
      });
      return { previous };
    },
    onError: (_err, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(FOLDERS_KEY, context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: FOLDERS_KEY });
    },
  });

  const removeMutation = useMutation({
    mutationFn: (id: string) => deleteFolder(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: FOLDERS_KEY });
      const previous = queryClient.getQueryData<FolderListResponse>(FOLDERS_KEY);
      queryClient.setQueryData<FolderListResponse>(FOLDERS_KEY, (old) => {
        const folders = old?.folders ?? [];
        return { folders: folders.filter((f) => f.id !== id) };
      });
      return { previous };
    },
    onError: (_err, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(FOLDERS_KEY, context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: FOLDERS_KEY });
      queryClient.invalidateQueries({ queryKey: ['documents'] });
    },
  });

  const moveMutation = useMutation({
    mutationFn: ({ id, folderId }: { id: string; folderId?: string }) =>
      moveDocument(id, folderId),
    onSettled: () => {
      invalidate();
    },
  });

  return {
    folders: query.data?.folders ?? [],
    loading: query.isLoading,
    error: query.error
      ? query.error instanceof Error
        ? query.error.message
        : 'Failed to load folders'
      : null,
    reload: query.refetch,
    create: (name: string, parentId?: string) =>
      createMutation.mutateAsync({ name, parentId }),
    rename: (id: string, name: string) =>
      renameMutation.mutateAsync({ id, name }),
    remove: (id: string) => removeMutation.mutateAsync(id),
    moveDoc: (id: string, folderId?: string) =>
      moveMutation.mutateAsync({ id, folderId }),
    isMutating:
      createMutation.isPending ||
      renameMutation.isPending ||
      removeMutation.isPending ||
      moveMutation.isPending,
  };
}
