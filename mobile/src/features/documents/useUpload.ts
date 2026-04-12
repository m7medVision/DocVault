import { useState, useCallback } from 'react';
import * as DocumentPicker from 'expo-document-picker';
import { API_URL } from '@/lib/config';
import type { UploadProgress, SelectedFile, UploadMode } from './types';

interface UseUploadResult {
  mode: UploadMode;
  selectedFile: SelectedFile | null;
  uploadProgress: UploadProgress | null;
  uploadError: string | null;
  uploadResult: { id: string; status: string } | null;
  title: string;
  docType: string;
  titleError: string | null;
  setTitle: (title: string) => void;
  setDocType: (docType: string) => void;
  setTitleError: (error: string | null) => void;
  pickDocument: () => Promise<void>;
  clearSelection: () => void;
  uploadDocument: () => Promise<void>;
  retryUpload: () => void;
}

const SUPPORTED_TYPES = [
  'application/pdf',
  'image/jpeg',
  'image/png',
  'image/heic',
  'image/heif',
];

const DOCUMENT_TYPES = [
  { value: 'invoice', label: 'Invoice', icon: '💰' },
  { value: 'contract', label: 'Contract', icon: '📄' },
  { value: 'warranty', label: 'Warranty', icon: '🛡️' },
  { value: 'identity', label: 'Identity', icon: '🪪' },
  { value: 'receipt', label: 'Receipt', icon: '🧾' },
  { value: 'other', label: 'Other', icon: '📎' },
];

export { DOCUMENT_TYPES };

function getFileTypeCategory(mimeType: string): 'pdf' | 'image' {
  if (mimeType === 'application/pdf') return 'pdf';
  return 'image';
}

function validateTitle(title: string): string | null {
  if (!title.trim()) return 'Please enter a document title';
  if (title.trim().length < 2) return 'Title must be at least 2 characters';
  if (title.trim().length > 200) return 'Title must be less than 200 characters';
  return null;
}

export function useUpload(accessToken: string | null): UseUploadResult {
  const [mode, setMode] = useState<UploadMode>('select');
  const [selectedFile, setSelectedFile] = useState<SelectedFile | null>(null);
  const [uploadProgress, setUploadProgress] = useState<UploadProgress | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploadResult, setUploadResult] = useState<{ id: string; status: string } | null>(null);
  const [title, setTitle] = useState('');
  const [docType, setDocType] = useState('other');
  const [titleError, setTitleError] = useState<string | null>(null);

  const clearSelection = useCallback(() => {
    setSelectedFile(null);
    setTitle('');
    setDocType('other');
    setUploadProgress(null);
    setUploadError(null);
    setUploadResult(null);
    setMode('select');
    setTitleError(null);
  }, []);

  const pickDocument = useCallback(async () => {
    try {
      const result = await DocumentPicker.getDocumentAsync({
        type: SUPPORTED_TYPES as string[],
        copyToCacheDirectory: true,
        multiple: false,
      });

      if (result.canceled) return;

      const asset = result.assets[0];
      if (!asset) return;

      if (asset.size && asset.size > 50 * 1024 * 1024) {
        throw new Error('Maximum file size is 50MB');
      }

      const file: SelectedFile = {
        uri: asset.uri,
        name: asset.name,
        size: asset.size || 0,
        mimeType: asset.mimeType || 'application/octet-stream',
        type: getFileTypeCategory(asset.mimeType || ''),
      };

      setSelectedFile(file);
      const nameWithoutExt = asset.name.replace(/\.[^/.]+$/, '');
      setTitle(nameWithoutExt);
      setMode('preview');
      setUploadError(null);
      setTitleError(null);
    } catch (error) {
      console.error('Document picker error:', error);
      setUploadError(error instanceof Error ? error.message : 'Failed to select document');
    }
  }, []);

  const uploadDocument = useCallback(async () => {
    if (!selectedFile || !accessToken) return;

    const error = validateTitle(title);
    if (error) {
      setTitleError(error);
      return;
    }

    setMode('uploading');
    setUploadError(null);
    setUploadResult(null);

    try {
      const formData = new FormData();
      const ext = selectedFile.name.split('.').pop()?.toLowerCase() || '';
      let fileType = selectedFile.mimeType;

      if (ext === 'pdf') fileType = 'application/pdf';
      else if (['jpg', 'jpeg'].includes(ext)) fileType = 'image/jpeg';
      else if (ext === 'png') fileType = 'image/png';
      else if (['heic', 'heif'].includes(ext)) fileType = 'image/heic';

      formData.append('file', {
        uri: selectedFile.uri,
        name: selectedFile.name,
        type: fileType,
      } as unknown as Blob);
      formData.append('title', title.trim());
      formData.append('doc_type', docType);

      const result = await uploadWithProgress(accessToken, formData, (progress) => {
        setUploadProgress(progress);
      });

      setUploadResult(result);
      setMode('success');
    } catch (error) {
      console.error('Upload error:', error);
      setUploadError(error instanceof Error ? error.message : 'Upload failed');
      setMode('error');
    }
  }, [selectedFile, accessToken, title, docType]);

  const retryUpload = useCallback(() => {
    if (selectedFile) {
      setUploadError(null);
      uploadDocument();
    }
  }, [selectedFile, uploadDocument]);

  return {
    mode,
    selectedFile,
    uploadProgress,
    uploadError,
    uploadResult,
    title,
    docType,
    titleError,
    setTitle,
    setDocType,
    setTitleError,
    pickDocument,
    clearSelection,
    uploadDocument,
    retryUpload,
  };
}

async function uploadWithProgress(
  accessToken: string,
  formData: FormData,
  onProgress: (progress: UploadProgress) => void
): Promise<{ id: string; status: string }> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();

    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        onProgress({
          bytesUploaded: event.loaded,
          totalBytes: event.total,
          percentage: Math.round((event.loaded / event.total) * 100),
        });
      }
    });

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const response = JSON.parse(xhr.responseText);
          resolve({ id: response.id, status: response.status });
        } catch {
          resolve({ id: 'unknown', status: 'uploaded' });
        }
      } else {
        try {
          const error = JSON.parse(xhr.responseText);
          reject(new Error(error.error || 'Upload failed'));
        } catch {
          reject(new Error(`Upload failed with status ${xhr.status}`));
        }
      }
    };

    xhr.onerror = () => reject(new Error('Network error'));
    xhr.ontimeout = () => reject(new Error('Upload timed out'));
    xhr.timeout = 120000;

    xhr.open('POST', `${API_URL}/documents/upload`);
    xhr.setRequestHeader('Authorization', `Bearer ${accessToken}`);
    xhr.send(formData);
  });
}
