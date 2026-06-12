import { useMutation, useQueryClient } from '@tanstack/react-query';

import { updateDocumentMetadata } from './api';

export function useUpdateMetadata(documentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (updates: Record<string, string>) =>
      updateDocumentMetadata(documentId, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['document', documentId] });
    },
  });
}
