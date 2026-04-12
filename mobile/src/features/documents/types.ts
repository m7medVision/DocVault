import type {
  Document,
  DocumentDetail,
  DocumentMetadata,
  DocumentPage,
  DocumentVersion,
  DocumentStatus,
} from '@/lib/api/types';

export type {
  Document,
  DocumentDetail,
  DocumentMetadata,
  DocumentPage,
  DocumentVersion,
  DocumentStatus,
};

export interface FilterState {
  type: string;
  folder_id: string;
  status: string;
}

export interface DocumentListItem {
  id: string;
  title: string;
  doc_type: string;
  status: string;
  folder_id?: string;
  created_at: string;
  thumbnail_url?: string;
}

export interface UploadProgress {
  bytesUploaded: number;
  totalBytes: number;
  percentage: number;
}

export interface SelectedFile {
  uri: string;
  name: string;
  size: number;
  mimeType: string;
  type: 'pdf' | 'image';
}

export type UploadMode = 'select' | 'preview' | 'uploading' | 'success' | 'error';
