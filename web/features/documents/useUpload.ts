'use client';

import { useState, useCallback } from 'react';
import { uploadDocument } from '@/features/documents/api';

export type UploadStatus = 'idle' | 'uploading' | 'success' | 'error';

export interface UseUploadOptions {
  docType?: string;
}

export interface UseUploadResult {
  files: File[];
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

export function useUpload(options: UseUploadOptions = {}): UseUploadResult {
  const [files, setFiles] = useState<File[]>([]);
  const [status, setStatus] = useState<UploadStatus>('idle');
  const [error, setError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const addFiles = useCallback((newFiles: File[]) => {
    setFiles((prev) => [...prev, ...newFiles]);
  }, []);

  const removeFile = useCallback((index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const reset = useCallback(() => {
    setFiles([]);
    setStatus('idle');
    setError(null);
  }, []);

  const setDragging = useCallback((dragging: boolean) => {
    setIsDragging(dragging);
  }, []);

  const upload = useCallback(async () => {
    if (files.length === 0) return;

    setStatus('uploading');
    setError(null);

    try {
      const failures: string[] = [];

      for (const file of files) {
        try {
          await uploadDocument(file, {
            title: file.name,
            doc_type: options.docType || 'other',
          });
        } catch (err) {
          failures.push(
            `${file.name}: ${err instanceof Error ? err.message : 'Upload failed'}`
          );
        }
      }

      if (failures.length > 0) {
        throw new Error(failures.join('\n'));
      }

      setStatus('success');
      setFiles([]);
    } catch (err) {
      setStatus('error');
      setError(err instanceof Error ? err.message : 'Upload failed. Please try again.');
    }
  }, [files, options.docType]);

  return {
    files,
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
