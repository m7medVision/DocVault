import { useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Button, Card, Input, useThemeColor } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { DOCUMENT_STATUSES, DOCUMENT_TYPES } from '@/constants/document';
import { Spacing } from '@/constants/theme';
import { useSearch } from '@/features/search/use-search';
import type { SearchResult } from '@/features/search/types';

export default function SearchScreen() {
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [docType, setDocType] = useState('');
  const [status, setStatus] = useState('');
  const [language, setLanguage] = useState('');
  const { results, total, loading, error, hasSearched, search, reset } = useSearch();

  function handleQueryChange(value: string) {
    setQuery(value);
    if (!value.trim()) reset();
  }

  function handleSearch() {
    const trimmedQuery = query.trim();
    if (!trimmedQuery || loading) return;

    void search({
      query: trimmedQuery,
      doc_type: docType || undefined,
      status: status || undefined,
      language: language || undefined,
      limit: 20,
    });
  }

  function openDocument(result: SearchResult) {
    router.push({ pathname: '/documents/[id]', params: { id: result.document_id } });
  }

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <ThemedText type="subtitle">Search</ThemedText>
        <ThemedText type="small" themeColor="textSecondary">
          Find documents by content, OCR text, type, and language.
        </ThemedText>
      </View>

      <Card className="rounded-3xl border border-divider bg-content1 p-4">
        <View style={styles.searchCard}>
          <Input
            placeholder="Search documents..."
            value={query}
            onChangeText={handleQueryChange}
            returnKeyType="search"
            onSubmitEditing={handleSearch}
          />
          <Button variant="primary" isDisabled={!query.trim() || loading} onPress={handleSearch}>
            <Button.Label>{loading ? 'Searching...' : 'Search'}</Button.Label>
          </Button>
        </View>
      </Card>

      <View style={styles.filterSection}>
        <ThemedText type="smallBold">Type</ThemedText>
        <View style={styles.chips}>
          <FilterChip label="All" selected={!docType} onPress={() => setDocType('')} />
          {DOCUMENT_TYPES.map((item) => (
            <FilterChip key={item} label={item} selected={docType === item} onPress={() => setDocType(item)} />
          ))}
        </View>
      </View>

      <View style={styles.filterSection}>
        <ThemedText type="smallBold">Status</ThemedText>
        <View style={styles.chips}>
          <FilterChip label="All" selected={!status} onPress={() => setStatus('')} />
          {DOCUMENT_STATUSES.map((item) => (
            <FilterChip key={item} label={item} selected={status === item} onPress={() => setStatus(item)} />
          ))}
        </View>
      </View>

      <View style={styles.filterSection}>
        <ThemedText type="smallBold">Language</ThemedText>
        <View style={styles.chips}>
          <FilterChip label="All" selected={!language} onPress={() => setLanguage('')} />
          <FilterChip label="English" selected={language === 'en'} onPress={() => setLanguage('en')} />
          <FilterChip label="Arabic" selected={language === 'ar'} onPress={() => setLanguage('ar')} />
        </View>
      </View>

      {loading && <ActivityIndicator />}

      {error && (
        <Card className="rounded-3xl border border-divider bg-content1 p-4">
          <ThemedText type="small" themeColor="error">
            {error}
          </ThemedText>
        </Card>
      )}

      {!loading && !error && !hasSearched && (
        <EmptyCard
          title="Search inside your vault"
          description="Enter a name, phrase, OCR text, or translated text to find matching files."
        />
      )}

      {!loading && !error && hasSearched && results.length === 0 && (
        <EmptyCard title="No matches" description="Try fewer filters or search for a phrase from the document text." />
      )}

      {!loading && !error && results.length > 0 && (
        <View style={styles.resultsHeader}>
          <ThemedText type="smallBold">
            {total} result{total === 1 ? '' : 's'}
          </ThemedText>
          <ThemedText type="small" themeColor="textSecondary" style={styles.flexText}>
            Sorted by strongest match
          </ThemedText>
        </View>
      )}

      {!loading && results.map((result) => (
        <Pressable key={result.document_id} onPress={() => openDocument(result)}>
          <Card className="rounded-3xl border border-divider bg-content1 p-4">
            <View style={styles.resultCard}>
              <ScoreBadge score={result.max_score} />
              <View style={styles.resultText}>
                <ThemedText type="smallBold" numberOfLines={2}>
                  {result.file}
                </ThemedText>
                <ThemedText type="small" themeColor="textSecondary">
                  Tap to open document
                </ThemedText>
              </View>
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

function ScoreBadge({ score }: { score: number }) {
  const normalizedScore = Math.max(0, Math.min(score, 1));
  const percent = Math.round(normalizedScore * 100);
  const [success, warning, danger] = useThemeColor(['success', 'warning', 'danger']);
  const color = normalizedScore >= 0.9 ? success : normalizedScore >= 0.7 ? warning : danger;

  return (
    <View style={[styles.scoreBadge, { backgroundColor: color }]}>
      <ThemedText type="code" style={styles.scoreText}>
        {percent}%
      </ThemedText>
    </View>
  );
}

function EmptyCard({ title, description }: { title: string; description: string }) {
  return (
    <Card className="rounded-3xl border border-divider bg-content1 p-6">
      <View style={styles.emptyCard}>
        <ThemedText type="smallBold">{title}</ThemedText>
        <ThemedText type="small" themeColor="textSecondary" style={styles.emptyDescription}>
          {description}
        </ThemedText>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  header: {
    gap: Spacing.one,
  },
  searchCard: {
    gap: Spacing.three,
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
  resultsHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  resultCard: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
  },
  resultText: {
    flex: 1,
    gap: Spacing.one,
  },
  scoreBadge: {
    width: 56,
    height: 56,
    borderRadius: 999,
    alignItems: 'center',
    justifyContent: 'center',
  },
  scoreText: {
    color: '#fff',
    fontSize: 13,
  },
  flexText: {
    flex: 1,
    textAlign: 'right',
  },
  emptyCard: {
    gap: Spacing.one,
    alignItems: 'center',
  },
  emptyDescription: {
    textAlign: 'center',
  },
});
