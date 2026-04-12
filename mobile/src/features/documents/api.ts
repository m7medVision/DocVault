export {
  listDocuments,
  getDocument,
  getDocumentDownloadURL,
  getDocumentPages,
  getDocumentVersions,
  updateDocumentMetadata,
  formatFileSize,
  formatConfidence,
  getStatusColor,
  getDocTypeLabel,
} from '@/lib/api/documents';

export type {
  Document,
  DocumentStatus,
  DocumentListResponse,
  ListDocumentsOptions,
  DocumentVersion,
  DocumentPage,
  DocumentDetail,
  DownloadURL,
  MetadataUpdate,
} from '@/lib/api/types';
