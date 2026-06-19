'use client';

import { useState, useEffect, useCallback } from 'react';
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
import { acceptSuggestion, dismissSuggestion } from '@/lib/api/documents';
import { updateDocumentTitle } from '@/lib/api/folders';
import { useTranslations } from 'next-intl';
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
  suggestionLoading: boolean;
  handleDownload: () => void;
  handleUpdateMetadata: (key: string, value: string) => void;
  handleRenameTitle: (title: string) => Promise<boolean>;
  handleAcceptSuggestion: () => Promise<boolean>;
  handleDismissSuggestion: () => Promise<boolean>;
}

export function useDocumentDetail(documentId: string): UseDocumentDetailResult {
  const t = useTranslations('documents');
  const [document, setDocument] = useState<DocumentDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [presignedUrl, setPresignedUrl] = useState<string | null>(null);
  const [downloadLoading, setDownloadLoading] = useState(false);
  const [suggestionLoading, setSuggestionLoading] = useState(false);

  const reloadDocument = useCallback(async () => {
    const [detail, pagesResponse, versionsResponse] = await Promise.all([
      getDocument(documentId),
      getDocumentPages(documentId),
      getDocumentVersions(documentId),
    ]);

    setDocument({
      ...detail.document,
      metadata: detail.metadata,
      pages: pagesResponse.pages,
      versions: versionsResponse.versions,
    });
  }, [documentId]);

  useEffect(() => {
    async function loadDocument() {
      if (!documentId) return;

      try {
        setLoading(true);
        await reloadDocument();

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
  }, [documentId, reloadDocument]);

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

  const handleAcceptSuggestion = async (): Promise<boolean> => {
    try {
      setSuggestionLoading(true);
      await acceptSuggestion(documentId);
      await reloadDocument();
      toast.success(t('suggestionAccepted'));
      return true;
    } catch (err) {
      const message = err instanceof Error ? err.message : t('suggestionFailed');
      toast.error(message);
      return false;
    } finally {
      setSuggestionLoading(false);
    }
  };

  const handleDismissSuggestion = async (): Promise<boolean> => {
    try {
      setSuggestionLoading(true);
      await dismissSuggestion(documentId);
      setDocument((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          suggested_folder_name: null,
          suggested_filename: null,
          suggestion_confidence: null,
          suggestion_create_new: false,
        };
      });
      toast.success(t('suggestionDismissed'));
      return true;
    } catch (err) {
      const message = err instanceof Error ? err.message : t('suggestionFailed');
      toast.error(message);
      return false;
    } finally {
      setSuggestionLoading(false);
    }
  };

  return {
    document,
    loading,
    error,
    presignedUrl,
    downloadLoading,
    suggestionLoading,
    handleDownload,
    handleUpdateMetadata,
    handleRenameTitle,
    handleAcceptSuggestion,
    handleDismissSuggestion,
  };
}
