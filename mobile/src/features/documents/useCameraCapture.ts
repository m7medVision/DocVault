import { useState, useRef, useEffect, useCallback } from 'react';
import { CameraView, CameraType, useCameraPermissions, useMicrophonePermissions } from 'expo-camera';
import * as MediaLibrary from 'expo-media-library';
import { API_URL } from '@/lib/config';
import type { UploadProgress } from './types';

type CameraMode = 'idle' | 'capturing' | 'preview' | 'uploading';

interface UseCameraCaptureResult {
  cameraPermission: ReturnType<typeof useCameraPermissions>[0];
  mediaPermission: ReturnType<typeof MediaLibrary.usePermissions>[0];
  cameraRef: React.RefObject<CameraView | null>;
  facing: CameraType;
  flash: 'off' | 'on';
  mode: CameraMode;
  capturedImage: string | null;
  uploadProgress: UploadProgress | null;
  uploadError: string | null;
  uploadResult: { id: string; status: string } | null;
  title: string;
  docType: string;
  toggleFacing: () => void;
  toggleFlash: () => void;
  capturePhoto: () => Promise<void>;
  retakePhoto: () => void;
  uploadDocument: () => Promise<void>;
  setTitle: (title: string) => void;
  setDocType: (docType: string) => void;
  requestCameraPermission: () => Promise<{ granted: boolean; status: string }>;
}

const DOCUMENT_TYPES = [
  { value: 'invoice', label: 'Invoice', icon: '💰' },
  { value: 'contract', label: 'Contract', icon: '📄' },
  { value: 'warranty', label: 'Warranty', icon: '🛡️' },
  { value: 'identity', label: 'Identity', icon: '🪪' },
  { value: 'receipt', label: 'Receipt', icon: '🧾' },
  { value: 'other', label: 'Other', icon: '📎' },
];

export { DOCUMENT_TYPES };

export function useCameraCapture(accessToken: string | null): UseCameraCaptureResult {
  const cameraRef = useRef<CameraView>(null);

  const [cameraPermission, requestCameraPermission] = useCameraPermissions();
  const [mediaPermission, requestMediaPermission] = MediaLibrary.usePermissions();

  const [facing, setFacing] = useState<CameraType>('back');
  const [flash, setFlash] = useState<'off' | 'on'>('off');

  const [mode, setMode] = useState<CameraMode>('idle');
  const [capturedImage, setCapturedImage] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState<UploadProgress | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploadResult, setUploadResult] = useState<{ id: string; status: string } | null>(null);

  const [title, setTitle] = useState('');
  const [docType, setDocType] = useState('other');

  useEffect(() => {
    if (!cameraPermission?.granted) {
      requestCameraPermission();
    }
    if (!mediaPermission?.granted) {
      requestMediaPermission();
    }
  }, []);

  const toggleFacing = useCallback(() => {
    setFacing((current) => (current === 'back' ? 'front' : 'back'));
  }, []);

  const toggleFlash = useCallback(() => {
    setFlash((current) => (current === 'off' ? 'on' : 'off'));
  }, []);

  const capturePhoto = useCallback(async () => {
    if (!cameraRef.current) return;

    try {
      setMode('capturing');
      const photo = await cameraRef.current.takePictureAsync({
        quality: 0.8,
        skipProcessing: false,
      });

      if (photo?.uri) {
        setCapturedImage(photo.uri);
        setMode('preview');
      } else {
        setMode('idle');
      }
    } catch (error) {
      console.error('Capture error:', error);
      setMode('idle');
    }
  }, []);

  const retakePhoto = useCallback(() => {
    setCapturedImage(null);
    setUploadProgress(null);
    setUploadError(null);
    setUploadResult(null);
    setMode('idle');
    setTitle('');
  }, []);

  const uploadDocument = useCallback(async () => {
    if (!capturedImage || !accessToken) return;

    if (!title.trim()) {
      return;
    }

    setMode('uploading');
    setUploadError(null);
    setUploadResult(null);

    try {
      const formData = new FormData();
      const filename = capturedImage.split('/').pop() || 'photo.jpg';
      const match = /\.(\w+)$/.exec(filename);
      const type = match ? `image/${match[1]}` : 'image/jpeg';

      formData.append('file', {
        uri: capturedImage,
        name: filename,
        type,
      } as unknown as Blob);
      formData.append('title', title.trim());
      formData.append('doc_type', docType);

      const result = await uploadWithProgress(accessToken, formData, (progress) => {
        setUploadProgress(progress);
      });

      setUploadResult(result);
      setMode('preview');
    } catch (error) {
      console.error('Upload error:', error);
      setUploadError(error instanceof Error ? error.message : 'Upload failed');
      setMode('preview');
    }
  }, [capturedImage, accessToken, title, docType]);

  return {
    cameraPermission,
    mediaPermission,
    cameraRef,
    facing,
    flash,
    mode,
    capturedImage,
    uploadProgress,
    uploadError,
    uploadResult,
    title,
    docType,
    toggleFacing,
    toggleFlash,
    capturePhoto,
    retakePhoto,
    uploadDocument,
    setTitle,
    setDocType,
    requestCameraPermission,
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
