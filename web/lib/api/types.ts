export interface Document {
  id: string;
  tenant_id: string;
  org_id: string;
  folder_id?: string;
  owner_id: string;
  title: string;
  doc_type: string;
  status: string;
  language?: string;
  created_at: string;
}

export interface DocumentVersion {
  id: string;
  document_id: string;
  version_number: number;
  storage_key: string;
  mime_type: string;
  file_size: number;
  uploaded_by?: string;
  created_at: string;
}

export interface DocumentMetadata {
  id?: string;
  document_id?: string;
  key: string;
  extracted_value?: string;
  corrected_value?: string;
  corrected_by?: string;
  corrected_at?: string;
  created_at?: string;
}

export interface DocumentPage {
  id: string;
  document_id: string;
  version_id: string;
  page_number: number;
  ocr_text?: string;
  translated_text?: string;
  confidence?: number;
  ocr_model: string;
  created_at: string;
}

export interface DocumentDetailResponse {
  document: Document;
  versions: DocumentVersion[];
  metadata: DocumentMetadata[];
}

export interface DocumentPagesResponse {
  document_id: string;
  pages: DocumentPage[];
}

export interface DocumentVersionsResponse {
  document_id: string;
  versions: DocumentVersion[];
}

export interface DocumentDownloadResponse {
  download_url: string;
  expires_at: string;
  storage_key: string;
}

export interface UploadDocumentOptions {
  title?: string;
  doc_type?: string;
  folder_id?: string;
  language?: string;
}

export interface UploadDocumentResponse {
  id: string;
  status: string;
  message: string;
  title: string;
  doc_type: string;
}

export interface DocumentListResponse {
  documents: Document[];
  cursor?: string;
  total: number;
}

export interface ListDocumentsOptions {
  type?: string;
  folder_id?: string;
  status?: string;
  language?: string;
  cursor?: string;
  limit?: number;
}

export interface Folder {
  id: string;
  tenant_id: string;
  org_id: string;
  parent_id?: string;
  name: string;
  created_by?: string;
  created_at: string;
}

export interface FolderListResponse {
  folders: Folder[];
}

export interface MoveDocumentResponse {
  message: string;
  folder_id?: string;
  folder_name: string;
}