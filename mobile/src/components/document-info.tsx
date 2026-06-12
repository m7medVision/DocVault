import { useState } from 'react';
import { Alert, Pressable, StyleSheet, TextInput, View } from 'react-native';

import { ThemedText } from './themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useUpdateMetadata } from '@/features/documents/use-update-metadata';
import type { DocumentMetadata } from '@/features/documents/types';
import { useTranslation } from '@/lib/i18n';

export function DocumentInfo({
  documentId,
  metadata,
}: {
  documentId: string;
  metadata: DocumentMetadata[];
}) {
  const theme = useTheme();
  const { t } = useTranslation();
  const updateMetadata = useUpdateMetadata(documentId);

  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [draft, setDraft] = useState('');

  if (!metadata || metadata.length === 0) {
    return (
      <ThemedText type="small" themeColor="muted">
        {t('info.noMetadata')}
      </ThemedText>
    );
  }

  const startEdit = (item: DocumentMetadata) => {
    setEditingKey(item.key);
    setDraft(item.corrected_value ?? item.extracted_value ?? '');
  };

  const save = async (key: string) => {
    try {
      await updateMetadata.mutateAsync({ [key]: draft });
      setEditingKey(null);
      Alert.alert('', t('info.saved'));
    } catch (e) {
      Alert.alert('', e instanceof Error ? e.message : t('info.saveFailed'));
    }
  };

  return (
    <View style={styles.list}>
      {metadata.map((item) => (
        <View
          key={item.key}
          style={[styles.row, { backgroundColor: theme.surface, borderColor: theme.border }]}
        >
          <ThemedText type="code" themeColor="textSecondary">
            {item.key.replace(/_/g, ' ').toUpperCase()}
          </ThemedText>
          {editingKey === item.key ? (
            <View style={styles.editRow}>
              <TextInput
                value={draft}
                onChangeText={setDraft}
                style={[styles.input, { color: theme.text, borderColor: theme.border }]}
                autoFocus
              />
              <Pressable
                onPress={() => save(item.key)}
                disabled={updateMetadata.isPending}
                style={[styles.saveBtn, { backgroundColor: theme.accent }]}
              >
                <ThemedText type="code" style={{ color: '#fff' }}>
                  {t('info.save')}
                </ThemedText>
              </Pressable>
            </View>
          ) : (
            <Pressable onPress={() => startEdit(item)} style={styles.valueRow}>
              <ThemedText type="small" style={styles.flex}>
                {item.corrected_value || item.extracted_value || '—'}
              </ThemedText>
              <ThemedText type="code" themeColor="accent">
                {t('info.edit')}
              </ThemedText>
            </Pressable>
          )}
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  list: { gap: Spacing.two },
  row: {
    borderRadius: Spacing.three,
    borderWidth: StyleSheet.hairlineWidth,
    padding: Spacing.three,
    gap: Spacing.one,
  },
  valueRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  flex: { flex: 1 },
  editRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  input: {
    flex: 1,
    borderRadius: Spacing.two,
    borderWidth: StyleSheet.hairlineWidth,
    paddingHorizontal: Spacing.two,
    paddingVertical: Spacing.one,
  },
  saveBtn: {
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one + 2,
    borderRadius: Spacing.two,
  },
});
