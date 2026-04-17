'use client';

import { useState, useCallback, useRef } from 'react';
import { toast } from 'sonner';
import { API_BASE_URL } from '@/lib/auth/constants';
import { getClientAccessToken } from '@/lib/client-auth';

export type UploadStatus = 'idle' | 'uploading' | 'processing' | 'completed' | 'error';

export interface UploadedFile {
  id: string;
  name: string;
  size: number;
  status: UploadStatus;
  progress: number;
  message: string;
  error?: string;
}

export interface UseUploadWithProgressOptions {
  docType?: string;
  onComplete?: (fileId: string, fileName: string) => void;
}

export interface UseUploadWithProgressResult {
  files: File[];
  uploads: UploadedFile[];
  status: UploadStatus;
  error: string | null;
  isDragging: boolean;
  addFiles: (newFiles: File[]) => void;
  removeFile: (index: number) => void;
  upload: () => Promise<void>;
  reset: () => void;
  setDragging: (dragging: boolean) => void;
}

export const MAX_FILE_SIZE = 50 * 1024 * 1024;

export const ALLOWED_FILE_TYPES = [
  'application/pdf',
  'image/jpeg',
  'image/png',
  'image/tiff',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
] as const;

const POLL_INTERVAL_MS = 2000;

function pollDocument(documentId: string, uploadIndex: number, fileName: string, setUploads: React.Dispatch<React.SetStateAction<UploadedFile[]>>, options: UseUploadWithProgressOptions) {
  const intervalId = setInterval(async () => {
    try {
      const token = getClientAccessToken();
      const res = await fetch(`${API_BASE_URL}/documents/${documentId}`, {
        headers: token ? { 'Authorization': `Bearer ${token}` } : {},
      });
      if (!res.ok) return;

      const data = await res.json();
      const stage = data.document?.processing_stage;
      const docStatus = data.document?.status;

      if (stage === 'completed' || docStatus === 'processed') {
        clearInterval(intervalId);
        setUploads((prev) =>
          prev.map((u, i) =>
            i === uploadIndex
              ? { ...u, status: 'completed', progress: 100, message: 'Processing complete' }
              : u
          )
        );
        options.onComplete?.(documentId, fileName);
      } else if (stage === 'ocr_failed' || stage === 'processing_failed' || docStatus === 'failed') {
        clearInterval(intervalId);
        const errMsg = data.document?.processing_error || 'Processing failed';
        setUploads((prev) =>
          prev.map((u, i) =>
            i === uploadIndex
              ? { ...u, status: 'error', error: errMsg }
              : u
          )
        );
      }
    } catch {
      // polling error — keep trying
    }
  }, POLL_INTERVAL_MS);
  return intervalId;
}

export function useUploadWithProgress(
  options: UseUploadWithProgressOptions = {}
): UseUploadWithProgressResult {
  const [files, setFiles] = useState<File[]>([]);
  const [uploads, setUploads] = useState<UploadedFile[]>([]);
  const [status, setStatus] = useState<UploadStatus>('idle');
  const [error, setError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const abortControllers = useRef<Map<string, AbortController>>(new Map());
  const wsConnections = useRef<WebSocket[]>([]);
  const pollIntervals = useRef<ReturnType<typeof setInterval>[]>([]);

  const addFiles = useCallback((newFiles: File[]) => {
    setFiles((prev) => [...prev, ...newFiles]);
  }, []);

  const removeFile = useCallback((index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const reset = useCallback(() => {
    setFiles([]);
    setUploads([]);
    setStatus('idle');
    setError(null);
    abortControllers.current.forEach((ac) => ac.abort());
    abortControllers.current.clear();
    wsConnections.current.forEach((ws) => ws.close());
    wsConnections.current = [];
    pollIntervals.current.forEach(clearInterval);
    pollIntervals.current = [];
  }, []);

  const setDragging = useCallback((dragging: boolean) => {
    setIsDragging(dragging);
  }, []);

  const connectProgressStream = useCallback((documentId: string, uploadIndex: number, fileName: string) => {
    const token = getClientAccessToken();
    if (!token) {
      const intervalId = pollDocument(documentId, uploadIndex, fileName, setUploads, options);
      pollIntervals.current.push(intervalId);
      return null;
    }

    const wsUrl = `${API_BASE_URL.replace('http', 'ws')}/documents/${documentId}/progress/ws?token=${token}`;

    const ws = new WebSocket(wsUrl);
    wsConnections.current.push(ws);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.error) {
          console.error('WebSocket error:', data.error);
          const intervalId = pollDocument(documentId, uploadIndex, fileName, setUploads, options);
          pollIntervals.current.push(intervalId);
          ws.close();
          return;
        }

        setUploads((prev) =>
          prev.map((u, i) =>
            i === uploadIndex
              ? {
                  ...u,
                  status: data.status === 'failed' ? 'error' : (data.status === 'completed' ? 'completed' : 'processing') as UploadStatus,
                  progress: data.progress ?? u.progress,
                  message: data.message ?? u.message,
                  error: data.status === 'failed' ? (data.message ?? 'Processing failed') : undefined,
                }
              : u
          )
        );

        if (data.status === 'completed') {
          ws.close();
          options.onComplete?.(documentId, fileName);
        } else if (data.status === 'failed') {
          ws.close();
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };

    ws.onerror = () => {
      const intervalId = pollDocument(documentId, uploadIndex, fileName, setUploads, options);
      pollIntervals.current.push(intervalId);
    };

    ws.onclose = () => {
      wsConnections.current = wsConnections.current.filter((w) => w !== ws);
    };

    return ws;
  }, [options]);

  const upload = useCallback(async () => {
    if (files.length === 0) return;

    setStatus('uploading');
    setError(null);

    const newUploads: UploadedFile[] = files.map((file) => ({
      id: '',
      name: file.name,
      size: file.size,
      status: 'uploading' as UploadStatus,
      progress: 0,
      message: 'Starting upload...',
    }));
    setUploads(newUploads);

    const failures: string[] = [];

    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const formData = new FormData();
      formData.set('file', file);
      formData.set('title', file.name);
      formData.set('doc_type', options.docType || 'other');

      setUploads((prev) =>
        prev.map((u, idx) =>
          idx === i
            ? { ...u, progress: 10, message: 'Uploading...' }
            : u
        )
      );

      try {
        const token = getClientAccessToken();
        const response = await fetch(`${API_BASE_URL}/documents/upload`, {
          method: 'POST',
          headers: token ? { 'Authorization': `Bearer ${token}` } : {},
          body: formData,
        });

        if (!response.ok) {
          const err = await response.json().catch(() => ({ error: 'Upload failed' }));
          throw new Error(err.error || 'Upload failed');
        }

        const result = await response.json();
        const documentId = result.id;

        setUploads((prev) =>
          prev.map((u, idx) =>
            idx === i
              ? {
                  ...u,
                  id: documentId,
                  status: 'processing',
                  progress: 20,
                  message: 'Queued for processing...',
                }
              : u
          )
        );

        connectProgressStream(documentId, i, file.name);
      } catch (err) {
        failures.push(`${file.name}: ${err instanceof Error ? err.message : 'Upload failed'}`);
        setUploads((prev) =>
          prev.map((u, idx) =>
            idx === i
              ? {
                  ...u,
                  status: 'error',
                  error: err instanceof Error ? err.message : 'Upload failed',
                  progress: 0,
                }
              : u
          )
        );
      }
    }

    if (failures.length > 0) {
      setStatus('error');
      setError(failures.join('\n'));
      toast.error('Some files failed to upload');
    }
  }, [files, options, connectProgressStream]);

  return {
    files,
    uploads,
    status,
    error,
    isDragging,
    addFiles,
    removeFile,
    upload,
    reset,
    setDragging,
  };
}
