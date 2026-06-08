import { useState } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { MoveDocumentSheet } from '@/components/move-document-sheet';
import { Spacing } from '@/constants/theme';
import { useTranslation } from '@/lib/i18n';
import { useDocuments } from '@/features/documents/use-documents';
import { useFolders } from '@/features/folders/use-folders';

export default function MoveDocumentScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { t } = useTranslation();
  const { documents, loading } = useDocuments({});
  const { moveDoc, isMutating } = useFolders();
  const [error, setError] = useState<string | null>(null);

  const document = documents.find((d) => d.id === id);
  const documentTitle = document?.title ?? t('folders.moveTitle');

  async function handleConfirm(targetFolderId?: string) {
    if (!id) return;
    setError(null);
    try {
      await moveDoc(id, targetFolderId);
      router.back();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('folders.moveFailed'));
    }
  }

  if (loading && !document) {
    return (
      <DocVaultScreen>
        <View style={styles.loading}>
          <ActivityIndicator />
        </View>
      </DocVaultScreen>
    );
  }

  return (
    <DocVaultScreen>
      {error ? (
        <ThemedText type="small" themeColor="error">
          {error}
        </ThemedText>
      ) : null}
      <MoveDocumentSheet
        visible
        documentTitle={documentTitle}
        currentFolderId={document?.folder_id}
        onClose={() => router.back()}
        onConfirm={handleConfirm}
        busy={isMutating}
      />
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  loading: {
    alignItems: 'center',
    padding: Spacing.four,
  },
});
