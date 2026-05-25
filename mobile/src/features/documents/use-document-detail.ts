import { useQuery } from '@tanstack/react-query';

import {
  getDocument,
  getDocumentPages,
  getDocumentVersions,
  downloadDocument,
} from './api';

export function useDocumentDetail(id: string) {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['document', id],
    queryFn: async () => {
      const [detailRes, pagesRes, versionsRes] = await Promise.all([
        getDocument(id),
        getDocumentPages(id),
        getDocumentVersions(id),
      ]);

      const latestVersion = versionsRes.versions.length > 0
        ? versionsRes.versions[0]
        : null;

      let downloadUrl: string | null = null;
      try {
        const dlRes = await downloadDocument(id);
        downloadUrl = dlRes.download_url;
      } catch { /* best-effort */ }

      return {
        document: detailRes.document,
        pages: pagesRes.pages ?? [],
        versions: versionsRes.versions ?? [],
        metadata: detailRes.metadata ?? [],
        downloadUrl,
        mimeType: latestVersion?.mime_type ?? '',
        fileSize: latestVersion?.file_size ?? 0,
      };
    },
    enabled: Boolean(id),
  });

  return {
    document: data?.document ?? null,
    pages: data?.pages ?? [],
    versions: data?.versions ?? [],
    metadata: data?.metadata ?? [],
    downloadUrl: data?.downloadUrl ?? null,
    mimeType: data?.mimeType ?? '',
    fileSize: data?.fileSize ?? 0,
    loading: isLoading,
    error: error ? (error instanceof Error ? error.message : 'Failed to load document') : null,
    reload: refetch,
  };
}
