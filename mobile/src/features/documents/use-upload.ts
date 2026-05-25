import { useCallback, useState } from 'react';

import { uploadDocument } from './api';

interface UploadResult {
  id: string;
  status: string;
  message: string;
  title: string;
}

interface UploadState {
  uploading: boolean;
  progress: string | null;
  result: UploadResult | null;
  error: string | null;
}

export function useUpload() {
  const [state, setState] = useState<UploadState>({
    uploading: false,
    progress: null,
    result: null,
    error: null,
  });

  const upload = useCallback(async (fileUri: string, fileName: string, mimeType: string, options?: { title?: string; doc_type?: string; folder_id?: string }) => {
    setState({ uploading: true, progress: 'Uploading...', result: null, error: null });

    try {
      const result = await uploadDocument(fileUri, fileName, mimeType, options);
      setState({ uploading: false, progress: null, result, error: null });
      return result;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Upload failed';
      setState((prev) => ({ ...prev, uploading: false, progress: null, error: message }));
      throw err;
    }
  }, []);

  const reset = useCallback(() => {
    setState({ uploading: false, progress: null, result: null, error: null });
  }, []);

  return { ...state, upload, reset };
}