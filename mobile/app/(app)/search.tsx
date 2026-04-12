// Search screen with semantic retrieval, bilingual support, and chunk-level grounding
// Uses shared design tokens from /shared/theme
// i18next integration for translations
import { useState, useCallback } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  FlatList,
  ActivityIndicator,
  Modal,
  ScrollView,
  Pressable,
  Keyboard,
  RefreshControl,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { tokens, colors } from '../../src/theme/tokens';
import { spacing } from '../../../shared/theme/spacing';
import { useSearch } from '../../src/features/search/useSearch';
import { formatRelevance, highlightMatch, SearchFilters } from '../../src/features/search/api';

interface FilterState {
  doc_type: string;
  language: string;
  folder_id: string;
  start_date: string;
  end_date: string;
}

const getDocTypeIcon = (docType?: string) => {
  switch (docType?.toLowerCase()) {
    case 'contract': return '📄';
    case 'invoice': return '💰';
    case 'warranty': return '🛡️';
    case 'identity': return '🪪';
    case 'receipt': return '🧾';
    default: return '📎';
  }
};

const getLanguageFlag = (lang?: string) => {
  switch (lang?.toLowerCase()) {
    case 'en': return '🇬🇧';
    case 'ar': return '🇸🇦';
    case 'fr': return '🇫🇷';
    case 'es': return '🇪🇸';
    case 'de': return '🇩🇪';
    default: return '🌐';
  }
};

export default function SearchScreen() {
  const router = useRouter();
  const { t } = useTranslation();
  const {
    query,
    setQuery,
    results,
    loading,
    refreshing,
    error,
    hasMore,
    total,
    searchTime,
    search,
    loadMore,
  } = useSearch();

  const [filterModalVisible, setFilterModalVisible] = useState(false);
  const [filters, setFilters] = useState<FilterState>({
    doc_type: '',
    language: '',
    folder_id: '',
    start_date: '',
    end_date: '',
  });
  const [activeFilterCount, setActiveFilterCount] = useState(0);

  const DOCUMENT_TYPES = [
    { value: '', label: t('allTypes') },
    { value: 'contract', label: t('contract') },
    { value: 'invoice', label: t('invoice') },
    { value: 'warranty', label: t('warranty') },
    { value: 'identity', label: t('identity') },
    { value: 'receipt', label: t('receipt') },
    { value: 'other', label: t('other') },
  ];

  const LANGUAGES = [
    { value: '', label: t('allLanguages') },
    { value: 'en', label: `English` },
    { value: 'ar', label: `العربية / Arabic` },
    { value: 'fr', label: `Français / French` },
    { value: 'es', label: `Español / Spanish` },
    { value: 'de', label: `Deutsch / German` },
  ];

  const FOLDERS = [
    { value: '', label: t('allFolders') },
    { value: 'personal', label: t('personal') },
    { value: 'work', label: t('work') },
    { value: 'financial', label: t('financial') },
    { value: 'legal', label: t('legal') },
  ];

  const updateFilterCount = useCallback((newFilters: FilterState) => {
    const count = [
      newFilters.doc_type,
      newFilters.language,
      newFilters.folder_id,
      newFilters.start_date,
      newFilters.end_date,
    ].filter(Boolean).length;
    setActiveFilterCount(count);
  }, []);

  const buildSearchFilters = useCallback((): SearchFilters => ({
    doc_type: filters.doc_type || undefined,
    language: filters.language || undefined,
    folder_id: filters.folder_id || undefined,
    start_date: filters.start_date || undefined,
    end_date: filters.end_date || undefined,
  }), [filters]);

  const handleSearch = useCallback(() => {
    Keyboard.dismiss();
    search(query, buildSearchFilters());
  }, [query, search, buildSearchFilters]);

  const handleLoadMore = useCallback(() => {
    loadMore(buildSearchFilters());
  }, [loadMore, buildSearchFilters]);

  const onRefresh = useCallback(() => {
    search(query, buildSearchFilters(), undefined, true);
  }, [query, search, buildSearchFilters]);

  const handleFilterChange = (key: keyof FilterState, value: string) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    updateFilterCount(newFilters);
  };

  const applyFilters = () => {
    setFilterModalVisible(false);
    if (query.trim()) {
      handleSearch();
    }
  };

  const clearFilters = () => {
    const cleared = { doc_type: '', language: '', folder_id: '', start_date: '', end_date: '' };
    setFilters(cleared);
    updateFilterCount(cleared);
  };

  const handleResultPress = (result: { document_id: string }) => {
    router.push(`/documents/${result.document_id}`);
  };

  const renderHighlightedText = (text: string, queryStr: string) => {
    const parts = highlightMatch(text, queryStr);
    return (
      <Text style={styles.snippetText}>
        {parts.map((part, idx) => (
          <Text
            key={idx}
            style={part.isMatch ? styles.highlightMatch : undefined}
          >
            {part.text}
          </Text>
        ))}
      </Text>
    );
  };

  const renderResult = ({ item }: { item: typeof results[0] }) => {
    const relevance = formatRelevance(item.score);

    return (
      <TouchableOpacity
        style={styles.resultCard}
        activeOpacity={0.7}
        onPress={() => handleResultPress(item)}
      >
        <View style={styles.resultHeader}>
          <View style={styles.resultIcon}>
            <Text style={styles.resultIconText}>
              {getDocTypeIcon(item.doc_type)}
            </Text>
          </View>
          <View style={styles.resultTitleContainer}>
            <Text style={styles.resultTitle} numberOfLines={1}>
              {item.document_title || 'Untitled Document'}
            </Text>
            <View style={styles.resultMeta}>
              <Text style={styles.resultMetaText}>
                {getLanguageFlag(item.language)} {item.language?.toUpperCase() || 'N/A'}
              </Text>
              {item.is_translation && (
                <View style={styles.translationBadge}>
                  <Text style={styles.translationText}>Translated</Text>
                </View>
              )}
              {item.page_number > 0 && (
                <Text style={styles.resultMetaText}>Page {item.page_number}</Text>
              )}
            </View>
          </View>
          <View style={styles.relevanceContainer}>
            <Text style={styles.relevanceLabel}>{t('relevance')}</Text>
            <Text style={styles.relevanceValue}>{relevance}</Text>
          </View>
        </View>

        <View style={styles.snippetContainer}>
          {renderHighlightedText(item.chunk_text, query)}
        </View>

        {item.match_context && (
          <View style={styles.contextContainer}>
            <Text style={styles.contextLabel}>Grounding context:</Text>
            <Text style={styles.contextText} numberOfLines={2}>
              {item.match_context}
            </Text>
          </View>
        )}

        {item.confidence !== undefined && (
          <View style={styles.confidenceContainer}>
            <Text style={styles.confidenceLabel}>Confidence:</Text>
            <View style={styles.confidenceBar}>
              <View
                style={[
                  styles.confidenceFill,
                  { width: `${item.confidence * 100}%` },
                ]}
              />
            </View>
            <Text style={styles.confidenceValue}>
              {Math.round(item.confidence * 100)}%
            </Text>
          </View>
        )}
      </TouchableOpacity>
    );
  };

  const renderFilterModal = () => (
    <Modal
      animationType="slide"
      transparent={true}
      visible={filterModalVisible}
      onRequestClose={() => setFilterModalVisible(false)}
    >
      <View style={styles.modalOverlay}>
        <Pressable style={styles.modalBackdrop} onPress={() => setFilterModalVisible(false)} />
        <View style={styles.modalContent}>
          <View style={styles.modalHeader}>
            <Text style={styles.modalTitle}>Search Filters</Text>
            <TouchableOpacity onPress={() => setFilterModalVisible(false)}>
              <Text style={styles.modalClose}>✕</Text>
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.modalBody}>
            <View style={styles.filterSection}>
              <Text style={styles.filterLabel}>Document Type</Text>
              <View style={styles.chipContainer}>
                {DOCUMENT_TYPES.map((type) => (
                  <TouchableOpacity
                    key={type.value}
                    style={[
                      styles.chip,
                      filters.doc_type === type.value && styles.chipActive,
                    ]}
                    onPress={() => handleFilterChange('doc_type', type.value)}
                  >
                    <Text
                      style={[
                        styles.chipText,
                        filters.doc_type === type.value && styles.chipTextActive,
                      ]}
                    >
                      {type.label}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            <View style={styles.filterSection}>
              <Text style={styles.filterLabel}>Language</Text>
              <View style={styles.chipContainer}>
                {LANGUAGES.map((lang) => (
                  <TouchableOpacity
                    key={lang.value}
                    style={[
                      styles.chip,
                      filters.language === lang.value && styles.chipActive,
                    ]}
                    onPress={() => handleFilterChange('language', lang.value)}
                  >
                    <Text
                      style={[
                        styles.chipText,
                        filters.language === lang.value && styles.chipTextActive,
                      ]}
                    >
                      {lang.label}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            <View style={styles.filterSection}>
              <Text style={styles.filterLabel}>Folder</Text>
              <View style={styles.chipContainer}>
                {FOLDERS.map((folder) => (
                  <TouchableOpacity
                    key={folder.value}
                    style={[
                      styles.chip,
                      filters.folder_id === folder.value && styles.chipActive,
                    ]}
                    onPress={() => handleFilterChange('folder_id', folder.value)}
                  >
                    <Text
                      style={[
                        styles.chipText,
                        filters.folder_id === folder.value && styles.chipTextActive,
                      ]}
                    >
                      {folder.label}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            <View style={styles.filterSection}>
              <Text style={styles.filterLabel}>Date Range</Text>
              <View style={styles.dateRangeContainer}>
                <View style={styles.dateInputContainer}>
                  <Text style={styles.dateLabel}>From</Text>
                  <TextInput
                    style={styles.dateInput}
                    placeholder="YYYY-MM-DD"
                    placeholderTextColor={tokens.textColor.muted}
                    value={filters.start_date}
                    onChangeText={(value) => handleFilterChange('start_date', value)}
                  />
                </View>
                <View style={styles.dateInputContainer}>
                  <Text style={styles.dateLabel}>To</Text>
                  <TextInput
                    style={styles.dateInput}
                    placeholder="YYYY-MM-DD"
                    placeholderTextColor={tokens.textColor.muted}
                    value={filters.end_date}
                    onChangeText={(value) => handleFilterChange('end_date', value)}
                  />
                </View>
              </View>
            </View>
          </ScrollView>

          <View style={styles.modalFooter}>
            <TouchableOpacity style={styles.modalClearButton} onPress={clearFilters}>
              <Text style={styles.modalClearButtonText}>{t('clearAll')}</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.applyButton} onPress={applyFilters}>
              <Text style={styles.applyButtonText}>Apply Filters</Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </Modal>
  );

  const renderEmptyState = () => {
    if (loading) return null;

    if (query.trim() === '') {
      return (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyIcon}>🔍</Text>
          <Text style={styles.emptyTitle}>Search Your Documents</Text>
          <Text style={styles.emptySubtext}>
            Enter a search query to find documents using semantic search.{'\n'}
            Supports English and Arabic text.
          </Text>
        </View>
      );
    }

    if (error) {
      return (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyIcon}>⚠️</Text>
          <Text style={styles.emptyTitle}>Search Failed</Text>
          <Text style={styles.emptySubtext}>{error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={handleSearch}>
            <Text style={styles.retryButtonText}>Retry</Text>
          </TouchableOpacity>
        </View>
      );
    }

    return (
      <View style={styles.emptyContainer}>
        <Text style={styles.emptyIcon}>📭</Text>
        <Text style={styles.emptyTitle}>{t('noResults')}</Text>
        <Text style={styles.emptySubtext}>
          {t('tryAdjustingFilters')}{'\n'}
        </Text>
      </View>
    );
  };

  const renderFooter = () => {
    if (!hasMore || results.length === 0) return null;
    return (
      <View style={styles.loadingFooter}>
        <ActivityIndicator size="small" color={tokens.accentColor} />
        <Text style={styles.loadingFooterText}>Loading more results...</Text>
      </View>
    );
  };

  return (
    <View style={styles.container}>
      <View style={styles.searchHeader}>
        <View style={styles.searchContainer}>
          <View style={styles.searchInputWrapper}>
            <Text style={styles.searchIcon}>🔍</Text>
            <TextInput
              style={styles.searchInput}
              placeholder="Search documents..."
              placeholderTextColor={tokens.textColor.muted}
              value={query}
              onChangeText={setQuery}
              onSubmitEditing={handleSearch}
              returnKeyType="search"
              autoCorrect={false}
              autoCapitalize="none"
              multiline={false}
            />
            {query.length > 0 && (
              <TouchableOpacity
                style={styles.clearButton}
                onPress={() => setQuery('')}
              >
                <Text style={styles.clearIcon}>✕</Text>
              </TouchableOpacity>
            )}
          </View>
          <TouchableOpacity
            style={[styles.searchButton, !query.trim() && styles.searchButtonDisabled]}
            onPress={handleSearch}
            disabled={!query.trim() || loading}
          >
            {loading ? (
              <ActivityIndicator size="small" color={colors.gray[50]} />
            ) : (
              <Text style={styles.searchButtonText}>Search</Text>
            )}
          </TouchableOpacity>
        </View>

        {activeFilterCount > 0 && (
          <View style={styles.activeFilters}>
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
              {filters.doc_type && (
                <TouchableOpacity
                  style={styles.activeFilterChip}
                  onPress={() => handleFilterChange('doc_type', '')}
                >
                  <Text style={styles.activeFilterText}>
                    Type: {DOCUMENT_TYPES.find(t => t.value === filters.doc_type)?.label}
                  </Text>
                  <Text style={styles.activeFilterRemove}>✕</Text>
                </TouchableOpacity>
              )}
              {filters.language && (
                <TouchableOpacity
                  style={styles.activeFilterChip}
                  onPress={() => handleFilterChange('language', '')}
                >
                  <Text style={styles.activeFilterText}>
                    Lang: {LANGUAGES.find(l => l.value === filters.language)?.label.split(' / ')[0]}
                  </Text>
                  <Text style={styles.activeFilterRemove}>✕</Text>
                </TouchableOpacity>
              )}
              {filters.folder_id && (
                <TouchableOpacity
                  style={styles.activeFilterChip}
                  onPress={() => handleFilterChange('folder_id', '')}
                >
                  <Text style={styles.activeFilterText}>
                    Folder: {FOLDERS.find(f => f.value === filters.folder_id)?.label}
                  </Text>
                  <Text style={styles.activeFilterRemove}>✕</Text>
                </TouchableOpacity>
              )}
            </ScrollView>
          </View>
        )}

        <View style={styles.filterBar}>
          <TouchableOpacity
            style={styles.filterButton}
            onPress={() => setFilterModalVisible(true)}
          >
            <Text style={styles.filterButtonIcon}>⚙️</Text>
            <Text style={styles.filterButtonText}>Filters</Text>
            {activeFilterCount > 0 && (
              <View style={styles.filterBadge}>
                <Text style={styles.filterBadgeText}>{activeFilterCount}</Text>
              </View>
            )}
          </TouchableOpacity>

          {searchTime !== null && !loading && results.length > 0 && (
            <Text style={styles.searchStats}>
              {total} results • {searchTime}ms
            </Text>
          )}
        </View>
      </View>

      <FlatList
        data={results}
        renderItem={renderResult}
        keyExtractor={(item) => `${item.document_id}-${item.chunk_id}`}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor={tokens.accentColor}
          />
        }
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
        ListEmptyComponent={renderEmptyState}
        ListFooterComponent={renderFooter}
        keyboardShouldPersistTaps="handled"
      />

      {renderFilterModal()}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.secondary,
  },
  searchHeader: {
    backgroundColor: tokens.backgroundColor.primary,
    borderBottomWidth: 1,
    borderBottomColor: colors.gray[800],
    paddingTop: 8,
  },
  searchContainer: {
    flexDirection: 'row',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: spacing.sm,
  },
  searchInputWrapper: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.gray[800],
    borderRadius: spacing.radius.lg,
    paddingHorizontal: spacing.md,
  },
  searchIcon: {
    fontSize: 16,
    marginRight: spacing.sm,
  },
  searchInput: {
    flex: 1,
    paddingVertical: 12,
    color: tokens.textColor.primary,
    fontSize: 16,
  },
  clearButton: {
    padding: spacing.xs,
  },
  clearIcon: {
    color: tokens.textColor.muted,
    fontSize: 14,
  },
  searchButton: {
    backgroundColor: tokens.accentColor,
    borderRadius: spacing.radius.lg,
    paddingHorizontal: 20,
    justifyContent: 'center',
    alignItems: 'center',
    minWidth: 80,
  },
  searchButtonDisabled: {
    opacity: 0.5,
  },
  searchButtonText: {
    color: colors.gray[50],
    fontSize: 14,
    fontWeight: '600',
  },
  activeFilters: {
    paddingHorizontal: spacing.md,
    paddingBottom: spacing.sm,
    gap: spacing.xs,
  },
  activeFilterChip: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 16,
    marginRight: 8,
    gap: 6,
  },
  activeFilterText: {
    color: colors.gray[50],
    fontSize: 12,
  },
  activeFilterRemove: {
    color: colors.gray[50],
    fontSize: 10,
    fontWeight: 'bold',
  },
  filterBar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
  },
  filterButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.gray[800],
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 20,
    gap: 6,
  },
  filterButtonIcon: {
    fontSize: 14,
  },
  filterButtonText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
  },
  filterBadge: {
    backgroundColor: tokens.accentColor,
    borderRadius: 10,
    minWidth: 20,
    height: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },
  filterBadgeText: {
    color: colors.gray[50],
    fontSize: 12,
    fontWeight: 'bold',
  },
  searchStats: {
    color: tokens.textColor.muted,
    fontSize: 12,
  },
  listContent: {
    padding: spacing.md,
    paddingBottom: 100,
  },
  resultCard: {
    backgroundColor: colors.gray[800],
    borderRadius: tokens.cardRadius,
    padding: spacing.cardPadding,
    marginBottom: 12,
  },
  resultHeader: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  resultIcon: {
    width: 40,
    height: 40,
    borderRadius: 8,
    backgroundColor: tokens.accentColor,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  resultIconText: {
    fontSize: 20,
  },
  resultTitleContainer: {
    flex: 1,
    marginRight: 12,
  },
  resultTitle: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 4,
  },
  resultMeta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    flexWrap: 'wrap',
  },
  resultMetaText: {
    color: tokens.textColor.muted,
    fontSize: 12,
  },
  translationBadge: {
    backgroundColor: colors.primary[500] + '33',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
  },
  translationText: {
    color: colors.primary[400],
    fontSize: 10,
    fontWeight: '500',
  },
  relevanceContainer: {
    alignItems: 'flex-end',
  },
  relevanceLabel: {
    color: tokens.textColor.muted,
    fontSize: 10,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  relevanceValue: {
    color: colors.success[500],
    fontSize: 16,
    fontWeight: '700',
  },
  snippetContainer: {
    backgroundColor: colors.gray[900],
    borderRadius: spacing.radius.md,
    padding: spacing.sm,
    marginBottom: 8,
  },
  snippetText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    lineHeight: 20,
  },
  highlightMatch: {
    backgroundColor: colors.warning[500] + '66',
    color: colors.warning[300],
    fontWeight: '600',
  },
  contextContainer: {
    marginTop: 4,
  },
  contextLabel: {
    color: tokens.textColor.muted,
    fontSize: 11,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: 2,
  },
  contextText: {
    color: tokens.textColor.muted,
    fontSize: 12,
    fontStyle: 'italic',
  },
  confidenceContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 8,
    gap: 8,
  },
  confidenceLabel: {
    color: tokens.textColor.muted,
    fontSize: 11,
  },
  confidenceBar: {
    flex: 1,
    height: 4,
    backgroundColor: colors.gray[700],
    borderRadius: 2,
    overflow: 'hidden',
  },
  confidenceFill: {
    height: '100%',
    backgroundColor: colors.success[500],
    borderRadius: 2,
  },
  confidenceValue: {
    color: colors.success[500],
    fontSize: 12,
    fontWeight: '600',
    minWidth: 40,
    textAlign: 'right',
  },
  loadingFooter: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: spacing.lg,
    gap: spacing.sm,
  },
  loadingFooterText: {
    color: tokens.textColor.muted,
    fontSize: 14,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingTop: 100,
    paddingHorizontal: spacing.xl,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: spacing.md,
  },
  emptyTitle: {
    color: tokens.textColor.primary,
    fontSize: 20,
    fontWeight: '600',
    marginBottom: spacing.sm,
  },
  emptySubtext: {
    color: tokens.textColor.muted,
    fontSize: 14,
    textAlign: 'center',
    lineHeight: 20,
  },
  retryButton: {
    marginTop: spacing.lg,
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: spacing.radius.md,
  },
  retryButtonText: {
    color: colors.gray[50],
    fontSize: 14,
    fontWeight: '600',
  },
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-end',
  },
  modalBackdrop: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
  },
  modalContent: {
    backgroundColor: colors.gray[900],
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    maxHeight: '80%',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: colors.gray[800],
  },
  modalTitle: {
    color: tokens.textColor.primary,
    fontSize: 20,
    fontWeight: 'bold',
  },
  modalClose: {
    color: tokens.textColor.secondary,
    fontSize: 24,
    padding: 4,
  },
  modalBody: {
    padding: spacing.lg,
  },
  filterSection: {
    marginBottom: spacing.xl,
  },
  filterLabel: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  chipContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  chip: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: colors.gray[800],
    borderWidth: 1,
    borderColor: colors.gray[700],
  },
  chipActive: {
    backgroundColor: tokens.accentColor,
    borderColor: tokens.accentColor,
  },
  chipText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
  },
  chipTextActive: {
    color: colors.gray[50],
    fontWeight: '600',
  },
  dateRangeContainer: {
    flexDirection: 'row',
    gap: spacing.md,
  },
  dateInputContainer: {
    flex: 1,
  },
  dateLabel: {
    color: tokens.textColor.muted,
    fontSize: 12,
    marginBottom: 4,
  },
  dateInput: {
    backgroundColor: colors.gray[800],
    borderRadius: spacing.radius.md,
    paddingHorizontal: spacing.md,
    paddingVertical: 10,
    color: tokens.textColor.primary,
    fontSize: 14,
    borderWidth: 1,
    borderColor: colors.gray[700],
  },
  modalFooter: {
    flexDirection: 'row',
    padding: spacing.lg,
    gap: 12,
    borderTopWidth: 1,
    borderTopColor: colors.gray[800],
  },
  modalClearButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: spacing.radius.md,
    backgroundColor: colors.gray[800],
    alignItems: 'center',
  },
  modalClearButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  applyButton: {
    flex: 2,
    paddingVertical: 14,
    borderRadius: spacing.radius.md,
    backgroundColor: tokens.accentColor,
    alignItems: 'center',
  },
  applyButtonText: {
    color: colors.gray[50],
    fontSize: 16,
    fontWeight: '600',
  },
});
