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
