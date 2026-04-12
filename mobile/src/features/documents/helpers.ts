import { colors, tokens } from '@/theme/tokens';
import type { DocumentStatus } from './api';

interface StatusStyle {
  backgroundColor: string;
  textColor: string;
}

export function getStatusStyle(status: string): StatusStyle {
  switch (status) {
    case 'pending':
      return { backgroundColor: `${colors.warning[500]}33`, textColor: colors.warning[600] };
    case 'processing':
      return { backgroundColor: `${colors.primary[400]}33`, textColor: colors.primary[400] };
    case 'processed':
      return { backgroundColor: `${colors.success[500]}33`, textColor: colors.success[600] };
    case 'failed':
      return { backgroundColor: `${colors.error[500]}33`, textColor: colors.error[500] };
    default:
      return { backgroundColor: `${colors.warning[500]}33`, textColor: colors.warning[600] };
  }
}

export function getDocTypeIcon(docType: string): string {
  switch (docType?.toLowerCase()) {
    case 'contract':
      return '📄';
    case 'invoice':
      return '💰';
    case 'warranty':
      return '🛡️';
    case 'identity':
      return '🪪';
    case 'receipt':
      return '🧾';
    default:
      return '📎';
  }
}

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function getFileTypeIcon(mimeType: string): string {
  if (mimeType === 'application/pdf') return '📄';
  if (mimeType.startsWith('image/')) return '🖼️';
  return '📎';
}
