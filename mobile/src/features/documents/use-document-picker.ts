import { useState } from 'react';
import * as DocumentPicker from 'expo-document-picker';

import { ALLOWED_FILE_TYPES, MAX_FILE_SIZE } from '@/constants/document';

export interface PickedDocument {
  name: string;
  uri: string;
  mimeType?: string;
  size?: number;
}

export function useDocumentPicker() {
  const [files, setFiles] = useState<PickedDocument[]>([]);
  const [error, setError] = useState<string | null>(null);

  async function pickFiles() {
    setError(null);
    const result = await DocumentPicker.getDocumentAsync({
      multiple: true,
      type: [...ALLOWED_FILE_TYPES],
      copyToCacheDirectory: true,
    });

    if (result.canceled) return;

    const validFiles: PickedDocument[] = [];
    const errors: string[] = [];

    for (const asset of result.assets) {
      if (asset.mimeType && !ALLOWED_FILE_TYPES.includes(asset.mimeType as never)) {
        errors.push(`${asset.name}: unsupported type`);
        continue;
      }
      if (asset.size && asset.size > MAX_FILE_SIZE) {
        errors.push(`${asset.name}: larger than 50MB`);
        continue;
      }

      validFiles.push({
        name: asset.name,
        uri: asset.uri,
        mimeType: asset.mimeType,
        size: asset.size,
      });
    }

    setFiles((previous) => [...previous, ...validFiles]);
    setError(errors.length > 0 ? errors.join('\n') : null);
  }

  function removeFile(uri: string) {
    setFiles((previous) => previous.filter((file) => file.uri !== uri));
  }

  return { files, error, pickFiles, removeFile };
}
