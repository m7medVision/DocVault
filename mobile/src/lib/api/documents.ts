// Document API calls to the backend

import { CONFIG } from '../config';
import { authorizedFetch } from '../auth';
import { handleResponse } from './client';
import {
  Document,
  DocumentStatus,
  DocumentListResponse,
  ListDocumentsOptions,
  DocumentVersion,
  DocumentPage,
  DocumentDetail,
  DownloadURL,
  MetadataUpdate,
} from './types';

export {
  Document,
  DocumentStatus,
  DocumentListResponse,
  ListDocumentsOptions,
  DocumentVersion,
  DocumentPage,
  DocumentDetail,
  DownloadURL,
  MetadataUpdate,
};

// GET /api/v1/documents - List documents with filters
export async function listDocuments(
  accessToken: string,
  options: ListDocumentsOptions = {}
): Promise<DocumentListResponse> {
  const params = new URLSearchParams();

  if (options.type) params.set('type', options.type);
  if (options.folder_id) params.set('folder_id', options.folder_id);
  if (options.status) params.set('status', options.status);
  if (options.language) params.set('language', options.language);
  if (options.cursor) params.set('cursor', options.cursor);
  if (options.limit) params.set('limit', options.limit.toString());

  const query = params.toString();
  const response = await authorizedFetch(
    accessToken,
    `${CONFIG.apiBaseUrl}/documents${query ? `?${query}` : ''}`,
    {
      method: 'GET',
    }
  );

  return handleResponse<DocumentListResponse>(response);
}

// GET /api/v1/documents/:id - Get document detail
export async function getDocument(
  accessToken: string,
  documentId: string
): Promise<DocumentDetail> {
  const response = await authorizedFetch(
    accessToken,
    `${CONFIG.apiBaseUrl}/documents/${documentId}`,
    {
      method: 'GET',
    }
  );
  return handleResponse<DocumentDetail>(response);
}

// GET /api/v1/documents/:id/download - Get download URL
export async function getDocumentDownloadURL(
  accessToken: string,
  documentId: string
): Promise<DownloadURL> {
  const response = await authorizedFetch(
    accessToken,
    `${CONFIG.apiBaseUrl}/documents/${documentId}/download`,
    {
      method: 'GET',
    }
  );
  return handleResponse<DownloadURL>(response);
}

// GET /api/v1/documents/:id/pages - Get OCR pages with confidence scores
export async function getDocumentPages(
  accessToken: string,
  documentId: string
): Promise<DocumentPage[]> {
  const response = await authorizedFetch(
    accessToken,
    `${CONFIG.apiBaseUrl}/documents/${documentId}/pages`,
    {
      method: 'GET',
    }
  );
  return handleResponse<DocumentPage[]>(response);
}

// GET /api/v1/documents/:id/versions - Get document version history
export async function getDocumentVersions(
  accessToken: string,
  documentId: string
): Promise<DocumentVersion[]> {
  const response = await authorizedFetch(
    accessToken,
    `${CONFIG.apiBaseUrl}/documents/${documentId}/versions`,
    {
      method: 'GET',
    }
  );
  return handleResponse<DocumentVersion[]>(response);
}

// PATCH /api/v1/documents/:id/metadata - Update document metadata
export async function updateDocumentMetadata(
  accessToken: string,
  documentId: string,
  updates: MetadataUpdate[]
): Promise<{ message: string }> {
  const response = await authorizedFetch(
    accessToken,
    `${CONFIG.apiBaseUrl}/documents/${documentId}/metadata`,
    {
      method: 'PATCH',
      body: JSON.stringify({ updates }),
    }
  );
  return handleResponse<{ message: string }>(response);
}

// Format file size for display
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

// Format confidence as percentage
export function formatConfidence(confidence?: number): string {
  if (confidence === undefined) return 'N/A';
  return `${Math.round(confidence * 100)}%`;
}

// Get status color based on document status
export function getStatusColor(status: DocumentStatus): string {
  switch (status) {
    case 'pending':
      return '#f59e0b';
    case 'processing':
      return '#3b82f6';
    case 'processed':
      return '#10b981';
    case 'failed':
      return '#ef4444';
    default:
      return '#6b7280';
  }
}

// Get document type display name
export function getDocTypeLabel(docType: string): string {
  const labels: Record<string, string> = {
    invoice: 'Invoice',
    contract: 'Contract',
    identity: 'Identity',
    warranty: 'Warranty',
    receipt: 'Receipt',
    other: 'Other',
  };
  return labels[docType] || docType;
}
