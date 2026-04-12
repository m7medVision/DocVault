export {
  listDocuments,
  getDocument,
  uploadDocument,
  deleteDocument,
  getDocumentPages,
  getDocumentVersions,
  downloadDocument,
  updateDocumentMetadata,
} from '@/lib/api/documents';

export type {
  Document,
  DocumentVersion,
  DocumentMetadata,
  DocumentPage,
  DocumentDetailResponse,
  DocumentPagesResponse,
  DocumentVersionsResponse,
  DocumentDownloadResponse,
  UploadDocumentOptions,
  UploadDocumentResponse,
  DocumentListResponse,
  ListDocumentsOptions,
} from '@/lib/api/types';
