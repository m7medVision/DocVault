import { useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, TextInput, View } from 'react-native';
import { useRouter } from 'expo-router';
import { useQuery } from '@tanstack/react-query';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTags } from '@/features/tags/use-tags';
import type { Tag } from '@/features/tags/api';
import { searchDocuments } from '@/features/search/api';
import { useTranslation } from '@/lib/i18n';

export default function TagsScreen() {
  const { t } = useTranslation();
  const theme = useTheme();
  const router = useRouter();

  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<Tag | null>(null);

  const { data: tagsData, isLoading } = useTags(query);
  const tags = tagsData?.tags ?? [];

  const { data: results, isLoading: searching } = useQuery({
    queryKey: ['tagSearch', selected?.id],
    queryFn: () =>
      searchDocuments({ query: selected!.name, tags: [selected!.id], limit: 50 }),
    enabled: !!selected,
  });

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <ThemedText type="subtitle">{t('tags.title')}</ThemedText>
        <ThemedText type="small" themeColor="muted">
          {t('tags.subtitle')}
        </ThemedText>
      </View>

      <TextInput
        value={query}
        onChangeText={setQuery}
        placeholder={t('tags.searchPlaceholder')}
        placeholderTextColor={theme.textSecondary}
        style={[
          styles.input,
          { backgroundColor: theme.surface, color: theme.text, borderColor: theme.border },
        ]}
      />

      {isLoading ? (
        <ActivityIndicator style={styles.loader} />
      ) : tags.length === 0 ? (
        <ThemedText type="small" themeColor="muted">
          {t('tags.empty')}
        </ThemedText>
      ) : (
        <View style={styles.chips}>
          {tags.map((tag) => (
            <Pressable
              key={tag.id}
              onPress={() => setSelected(tag)}
              style={[
                styles.chip,
                {
                  backgroundColor: selected?.id === tag.id ? theme.accent : theme.surface,
                  borderColor: theme.border,
                },
              ]}
            >
              <ThemedText
                type="small"
                style={selected?.id === tag.id ? { color: '#fff' } : undefined}
              >
                {tag.name}
              </ThemedText>
            </Pressable>
          ))}
        </View>
      )}

      {selected && (
        <View style={styles.results}>
          <ThemedText type="smallBold">
            {t('tags.documentsWithTag', { tag: selected.name })}
          </ThemedText>
          {searching ? (
            <ActivityIndicator style={styles.loader} />
          ) : !results || results.results.length === 0 ? (
            <ThemedText type="small" themeColor="muted">
              {t('tags.noResults')}
            </ThemedText>
          ) : (
            results.results.map((r) => (
              <Pressable
                key={r.document_id}
                onPress={() => router.push(`/documents/${r.document_id}`)}
                style={[styles.resultRow, { backgroundColor: theme.surface, borderColor: theme.border }]}
              >
                <ThemedText type="small" numberOfLines={1}>
                  {r.file}
                </ThemedText>
              </Pressable>
            ))
          )}
        </View>
      )}
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: { gap: Spacing.one },
  input: {
    borderRadius: Spacing.two,
    borderWidth: StyleSheet.hairlineWidth,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
  },
  loader: { marginTop: Spacing.three },
  chips: { flexDirection: 'row', flexWrap: 'wrap', gap: Spacing.two },
  chip: {
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one + 2,
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
  },
  results: { gap: Spacing.two },
  resultRow: {
    borderRadius: Spacing.two,
    borderWidth: StyleSheet.hairlineWidth,
    padding: Spacing.three,
  },
});
