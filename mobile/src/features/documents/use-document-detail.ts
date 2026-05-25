import { useState, useEffect, useCallback } from 'react';
import {
  getDocument,
  getDocumentPages,
  getDocumentVersions,
  downloadDocument,
} from './api';
import type {
  Document,
  DocumentPage,
  DocumentVersion,
  DocumentMetadata,
} from './types';

interface DocumentDetailState {
  document: Document | null;
  pages: DocumentPage[];
  versions: DocumentVersion[];
  metadata: DocumentMetadata[];
  downloadUrl: string | null;
  mimeType: string;
  fileSize: number;
  loading: boolean;
  error: string | null;
}

export function useDocumentDetail(id: string) {
  const [state, setState] = useState<DocumentDetailState>({
    document: null,
    pages: [],
    versions: [],
    metadata: [],
    downloadUrl: null,
    mimeType: '',
    fileSize: 0,
    loading: true,
    error: null,
  });

  const load = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
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
      } catch {}

      setState({
        document: detailRes.document,
        pages: pagesRes.pages ?? [],
        versions: versionsRes.versions ?? [],
        metadata: detailRes.metadata ?? [],
        downloadUrl,
        mimeType: latestVersion?.mime_type ?? '',
        fileSize: latestVersion?.file_size ?? 0,
        loading: false,
        error: null,
      });
    } catch (err: any) {
      setState((prev) => ({
        ...prev,
        loading: false,
        error: err?.message ?? 'Failed to load document',
      }));
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  return { ...state, reload: load };
}
