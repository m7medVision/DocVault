// Shared API types used across all modules

export interface ApiError {
  error: string;
  code: string;
  request_id: string;
}

export interface Document {
  id: string;
  tenant_id: string;
  org_id: string;
  folder_id?: string;
  owner_id: string;
  title: string;
  doc_type: string;
  status: DocumentStatus;
  language?: string;
  created_at: string;
}

export type DocumentStatus = 'pending' | 'processing' | 'processed' | 'failed';

export interface DocumentListResponse {
  documents: Document[];
  cursor?: string;
  total?: number;
}

export interface ListDocumentsOptions {
  type?: string;
  folder_id?: string;
  status?: string;
  language?: string;
  cursor?: string;
  limit?: number;
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

export interface DocumentPage {
  id: string;
  document_id: string;
  version_id: string;
  page_number: number;
  ocr_text?: string;
  confidence?: number;
  ocr_model: string;
  created_at: string;
}

export interface DocumentMetadata {
  id: string;
  document_id: string;
  key: string;
  extracted_value?: string;
  corrected_value?: string;
  corrected_by?: string;
  corrected_at?: string;
  created_at: string;
}

export interface DocumentDetail {
  document: Document;
  versions: DocumentVersion[];
  metadata: DocumentMetadata[];
}

export interface DownloadURL {
  presigned_url: string;
  expires_at: string;
  storage_key: string;
}

export interface MetadataUpdate {
  key: string;
  value: string;
}

export interface SearchResult {
  document_id: string;
  document_title?: string;
  doc_type?: string;
  chunk_id: string;
  chunk_text: string;
  page_number: number;
  score: number;
  language: string;
  is_translation: boolean;
  confidence?: number;
  match_context?: string;
}

export interface SearchOutput {
  results: SearchResult[];
  query: string;
  total: number;
  page: number;
  page_size: number;
  has_more: boolean;
  next_cursor?: string;
}

export interface SearchFilters {
  doc_type?: string;
  language?: string;
  folder_id?: string;
  start_date?: string;
  end_date?: string;
}

export interface Reminder {
  id: string;
  tenant_id: string;
  document_id: string;
  document_title: string;
  user_id: string;
  due_date: string;
  reminder_time: string;
  status: ReminderStatus;
  snoozed_count: number;
  snoozed_until?: string;
  notification_sent: boolean;
  notification_id?: string;
  created_at: string;
  updated_at: string;
}

export type ReminderStatus = 'pending' | 'sent' | 'dismissed' | 'completed';

export interface CreateReminderRequest {
  document_id: string;
  reminder_time: string;
  due_date?: string;
}

export interface SnoozeReminderRequest {
  snooze_minutes: number;
}

export interface RegisterPushTokenRequest {
  push_token: string;
  platform: 'ios' | 'android';
  device_id?: string;
}
