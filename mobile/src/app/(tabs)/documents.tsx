import { useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button, Card , useThemeColor } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { FolderIcon, ChevronRightIcon } from '@/components/icons';
import { DOCUMENT_STATUSES, DOCUMENT_TYPES, formatFileSize } from '@/constants/document';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useDocumentPicker } from '@/features/documents/use-document-picker';
import { useDocuments } from '@/features/documents/use-documents';
import { useUpload } from '@/features/documents/use-upload';
import { useFolders } from '@/features/folders/use-folders';

export default function DocumentsScreen() {
  const router = useRouter();
  const theme = useTheme();
  const { t, isRTL } = useTranslation();
  const { folders } = useFolders();
  const [type, setType] = useState('');
  const [status, setStatus] = useState('');
  const filters = { type, status };
  const { documents, loading, error, reload } = useDocuments(filters);
  const { files, error: pickerError, pickFiles, removeFile, clearFiles } = useDocumentPicker();
  const { uploading, progress, error: uploadError, upload } = useUpload();

  const rootFolders = folders.filter((f) => !f.parent_id).slice(0, 4);

  async function uploadSelectedFiles() {
    if (uploading || files.length === 0) return;

    try {
      for (const file of files) {
        await upload(file.uri, file.name, file.mimeType || 'application/octet-stream', {
          title: file.name,
          doc_type: type || 'other',
        });
      }

      clearFiles();
      await reload();
    } catch {
      // useUpload exposes the error message for rendering.
    }
  }

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <View style={styles.headerText}>
          <ThemedText type="subtitle">Documents</ThemedText>
          <ThemedText type="small" themeColor="textSecondary">
            Browse OCR-ready files, statuses, and languages.
          </ThemedText>
        </View>
        <Button variant="primary" onPress={() => void pickFiles()}>
          <Button.Label>Upload</Button.Label>
        </Button>
      </View>

      {pickerError && <ThemedText themeColor="textSecondary">{pickerError}</ThemedText>}
      {uploadError && <ThemedText themeColor="textSecondary">{uploadError}</ThemedText>}
      {progress && <ThemedText themeColor="textSecondary">{progress}</ThemedText>}

      {rootFolders.length > 0 ? (
        <Pressable
          onPress={() => router.push('/folders' as never)}
          accessibilityRole="button"
          accessibilityLabel={t('folders.title')}
        >
          <Card className="rounded-2xl border border-divider bg-content1 p-3">
            <View style={[styles.folderEntry, isRTL && styles.rtlRow]}>
              <View style={[styles.folderEntryIcon, { backgroundColor: `${theme.accent}1A` }]}>
                <FolderIcon size={18} color={theme.accent} strokeWidth={1.5} />
              </View>
              <View style={styles.folderEntryText}>
                <ThemedText type="smallBold">{t('folders.title')}</ThemedText>
                <ThemedText type="code" themeColor="textSecondary">
                  {rootFolders.length}
                </ThemedText>
              </View>
              <ChevronRightIcon size={18} color={theme.muted} strokeWidth={1.5} />
            </View>
          </Card>
        </Pressable>
      ) : null}
      {files.length > 0 && (
        <Card className="rounded-3xl border border-divider bg-content1 p-4">
          <View style={styles.documentCard}>
            <ThemedText type="smallBold">Selected files</ThemedText>
            {files.map((file) => (
              <Pressable key={file.uri} onPress={() => removeFile(file.uri)}>
                <View style={styles.selectedFileRow}>
                  <ThemedText type="small" style={styles.flexText}>
                    {file.name}
                  </ThemedText>
                  <ThemedText type="code">
                    {file.size ? formatFileSize(file.size) : 'tap to remove'}
                  </ThemedText>
                </View>
              </Pressable>
            ))}
            <Button variant="secondary" onPress={() => void uploadSelectedFiles()} isDisabled={uploading}>
              <Button.Label>{uploading ? 'Uploading...' : 'Upload selected'}</Button.Label>
            </Button>
          </View>
        </Card>
      )}

      <View style={styles.filterSection}>
        <ThemedText type="smallBold">Type</ThemedText>
        <View style={styles.chips}>
          <FilterChip label="All" selected={!type} onPress={() => setType('')} />
          {DOCUMENT_TYPES.map((item) => (
            <FilterChip key={item} label={item} selected={type === item} onPress={() => setType(item)} />
          ))}
        </View>
      </View>

      <View style={styles.filterSection}>
        <ThemedText type="smallBold">Status</ThemedText>
        <View style={styles.chips}>
          <FilterChip label="All" selected={!status} onPress={() => setStatus('')} />
          {DOCUMENT_STATUSES.map((item) => (
            <FilterChip
              key={item}
              label={item}
              selected={status === item}
              onPress={() => setStatus(item)}
            />
          ))}
        </View>
      </View>

      {loading && <ActivityIndicator />}
      {error && <ThemedText themeColor="textSecondary">{error}</ThemedText>}
      {!loading && !error && documents.length === 0 && (
        <EmptyCard title="No documents yet" description="Scan or upload your first document." />
      )}

      {documents.map((document) => (
        <Pressable key={document.id} onPress={() => router.push({ pathname: '/documents/[id]', params: { id: document.id } })}>
          <Card className="rounded-3xl border border-divider bg-content1 p-4">
            <View style={styles.documentCard}>
              <View style={styles.documentTitleRow}>
                <ThemedText type="smallBold" style={styles.flexText}>
                  {document.title}
                </ThemedText>
                <ThemedText type="code">{document.status}</ThemedText>
              </View>
              <ThemedText type="small" themeColor="textSecondary">
                {document.doc_type} · {document.language || 'unknown'} ·{' '}
                {new Date(document.created_at).toLocaleDateString()}
              </ThemedText>
            </View>
          </Card>
        </Pressable>
      ))}
    </DocVaultScreen>
  );
}

function FilterChip({ label, selected, onPress }: { label: string; selected: boolean; onPress: () => void }) {
  const [accent, background, foreground] = useThemeColor(['accent', 'background', 'foreground']);

  return (
    <Pressable onPress={onPress}>
      <View style={[styles.chip, { backgroundColor: selected ? accent : background }]}>
        <ThemedText type="small" style={{ color: selected ? '#fff' : foreground }}>
          {label}
        </ThemedText>
      </View>
    </Pressable>
  );
}

function EmptyCard({ title, description }: { title: string; description: string }) {
  return (
    <Card className="rounded-3xl border border-divider bg-content1 p-6">
      <View style={styles.emptyCard}>
        <ThemedText type="smallBold">{title}</ThemedText>
        <ThemedText type="small" themeColor="textSecondary">
          {description}
        </ThemedText>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
  },
  headerText: {
    flex: 1,
    gap: Spacing.one,
  },
  filterSection: {
    gap: Spacing.two,
  },
  chips: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: Spacing.two,
  },
  chip: {
    borderRadius: 999,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
  },
  documentCard: {
    gap: Spacing.one,
  },
  documentTitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  flexText: {
    flex: 1,
  },
  selectedFileRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
    paddingVertical: Spacing.one,
  },
  emptyCard: {
    gap: Spacing.one,
    alignItems: 'center',
  },
  folderEntry: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
  },
  rtlRow: {
    flexDirection: 'row-reverse',
  },
  folderEntryIcon: {
    width: 36,
    height: 36,
    borderRadius: 10,
    alignItems: 'center',
    justifyContent: 'center',
  },
  folderEntryText: {
    flex: 1,
    gap: 1,
  },
});
