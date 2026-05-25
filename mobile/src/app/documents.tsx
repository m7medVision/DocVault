import { useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { Button, Card , useThemeColor } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { DOCUMENT_STATUSES, DOCUMENT_TYPES, formatFileSize } from '@/constants/document';
import { Spacing } from '@/constants/theme';
import { useDocumentPicker } from '@/features/documents/use-document-picker';
import { useDocuments } from '@/features/documents/use-documents';

export default function DocumentsScreen() {
  const [type, setType] = useState('');
  const [status, setStatus] = useState('');
  const filters = { type, status };
  const { documents, loading, error } = useDocuments(filters);
  const { files, error: pickerError, pickFiles, removeFile } = useDocumentPicker();

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
            <Button variant="secondary">
              <Button.Label>Upload selected</Button.Label>
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
        <Card key={document.id} className="rounded-3xl border border-divider bg-content1 p-4">
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
});
