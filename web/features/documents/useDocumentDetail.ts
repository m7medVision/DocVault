'use client';

import { useState, useEffect } from 'react';
import {
  getDocument,
  getDocumentPages,
  getDocumentVersions,
  downloadDocument,
  updateDocumentMetadata,
  type Document,
  type DocumentPage,
  type DocumentMetadata,
  type DocumentVersion,
} from '@/features/documents/api';
import { updateDocumentTitle } from '@/lib/api/folders';
import { toast } from 'sonner';

export type DocumentDetail = Document & {
  pages?: DocumentPage[];
  metadata?: DocumentMetadata[];
  versions?: DocumentVersion[];
};

export interface UseDocumentDetailResult {
  document: DocumentDetail | null;
  loading: boolean;
  error: string | null;
  presignedUrl: string | null;
  downloadLoading: boolean;
  handleDownload: () => void;
  handleUpdateMetadata: (key: string, value: string) => void;
  handleRenameTitle: (title: string) => Promise<boolean>;
}

export function useDocumentDetail(documentId: string): UseDocumentDetailResult {
  const [document, setDocument] = useState<DocumentDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [presignedUrl, setPresignedUrl] = useState<string | null>(null);
  const [downloadLoading, setDownloadLoading] = useState(false);

  useEffect(() => {
    async function loadDocument() {
      if (!documentId) return;

      try {
        setLoading(true);
        const [detail, pagesResponse, versionsResponse] = await Promise.all([
          getDocument(documentId),
          getDocumentPages(documentId),
          getDocumentVersions(documentId),
        ]);

        const doc: DocumentDetail = {
          ...detail.document,
          metadata: detail.metadata,
          pages: pagesResponse.pages,
          versions: versionsResponse.versions,
        };
        setDocument(doc);

        try {
          const dl = await downloadDocument(documentId);
          setPresignedUrl(dl.download_url);
        } catch {
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load document');
      } finally {
        setLoading(false);
      }
    }

    loadDocument();
  }, [documentId]);

  const handleDownload = async () => {
    try {
      setDownloadLoading(true);
      const url = presignedUrl || (await downloadDocument(documentId)).download_url;
      window.open(url, '_blank');
    } catch {
    } finally {
      setDownloadLoading(false);
    }
  };

  const handleUpdateMetadata = async (key: string, value: string) => {
    await updateDocumentMetadata(documentId, { [key]: value });

    setDocument((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        metadata: prev.metadata?.map((item) =>
          item.key === key ? { ...item, corrected_value: value } : item
        ),
      };
    });
  };

  const handleRenameTitle = async (title: string): Promise<boolean> => {
    try {
      await updateDocumentTitle(documentId, title);
      setDocument((prev) => {
        if (!prev) return prev;
        return { ...prev, title };
      });
      toast.success('Document renamed');
      return true;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to rename document';
      toast.error(message);
      return false;
    }
  };

  return {
    document,
    loading,
    error,
    presignedUrl,
    downloadLoading,
    handleDownload,
    handleUpdateMetadata,
    handleRenameTitle,
  };
}
