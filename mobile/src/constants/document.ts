export const MAX_FILE_SIZE = 50 * 1024 * 1024;

export const DOCUMENT_TYPES = ['contract', 'invoice', 'warranty', 'identity', 'receipt', 'other'] as const;
export const DOCUMENT_STATUSES = ['pending', 'processed', 'failed'] as const;

export const ALLOWED_FILE_TYPES = [
  'application/pdf',
  'image/jpeg',
  'image/png',
  'image/tiff',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
] as const;

export type DocumentType = (typeof DOCUMENT_TYPES)[number];
export type DocumentStatus = (typeof DOCUMENT_STATUSES)[number];

export function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}
