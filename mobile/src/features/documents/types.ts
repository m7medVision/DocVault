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
