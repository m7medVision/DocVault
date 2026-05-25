import { useCallback } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { uploadDocument } from './api';

export function useUpload() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ fileUri, fileName, mimeType, options }: {
      fileUri: string;
      fileName: string;
      mimeType: string;
      options?: { title?: string; doc_type?: string; folder_id?: string };
    }) => uploadDocument(fileUri, fileName, mimeType, options),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['documents'] });
      queryClient.invalidateQueries({ queryKey: ['documentStats'] });
    },
  });

  const upload = useCallback(
    async (fileUri: string, fileName: string, mimeType: string, options?: { title?: string; doc_type?: string; folder_id?: string }) => {
      return mutation.mutateAsync({ fileUri, fileName, mimeType, options });
    },
    [mutation],
  );

  return {
    uploading: mutation.isPending,
    progress: mutation.isPending ? 'Uploading...' : null,
    result: mutation.data ?? null,
    error: mutation.error ? (mutation.error instanceof Error ? mutation.error.message : 'Upload failed') : null,
    upload,
    reset: mutation.reset,
  };
}
