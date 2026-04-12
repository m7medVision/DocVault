import { useState, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  ActivityIndicator,
  Modal,
  Pressable,
  ScrollView,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../../src/contexts/AuthContext';
import { tokens, colors } from '../../src/theme/tokens';
import { useDocumentList, useActiveFilterCount } from '../../src/features/documents/useDocumentList';
import { getStatusStyle, getDocTypeIcon } from '../../src/features/documents/helpers';
import type { Document } from '../../src/features/documents/api';
import type { FilterState } from '../../src/features/documents/types';

const DOCUMENT_TYPES = [
  { value: '', label: 'allTypes' },
  { value: 'contract', label: 'contract' },
  { value: 'invoice', label: 'invoice' },
  { value: 'warranty', label: 'warranty' },
  { value: 'identity', label: 'identity' },
  { value: 'receipt', label: 'receipt' },
  { value: 'other', label: 'other' },
];

const DOCUMENT_STATUSES = [
  { value: '', label: 'allStatuses' },
  { value: 'pending', label: 'pending' },
  { value: 'processing', label: 'processing' },
  { value: 'processed', label: 'processed' },
  { value: 'failed', label: 'failed' },
];

const FOLDERS = [
  { value: '', label: 'allFolders' },
  { value: 'personal', label: 'personal' },
  { value: 'work', label: 'work' },
  { value: 'financial', label: 'financial' },
  { value: 'legal', label: 'legal' },
];

interface DocumentsScreenProps {
  onFilterChange?: (filters: FilterState) => void;
}

export default function DocumentsScreen({ onFilterChange }: DocumentsScreenProps) {
  const router = useRouter();
  const { t } = useTranslation();
  const { accessToken, user } = useAuth();
  const [filterModalVisible, setFilterModalVisible] = useState(false);
  const [filters, setFilters] = useState<FilterState>({
    type: '',
    folder_id: '',
    status: '',
  });

  const { documents, loading, refreshing, error, refetch } = useDocumentList(accessToken, filters);
  const activeFilterCount = useActiveFilterCount(filters);

  const handleFilterChange = useCallback((key: keyof FilterState, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    onFilterChange?.(filters);
  }, [filters, onFilterChange]);

  const clearFilters = useCallback(() => {
    setFilters({ type: '', folder_id: '', status: '' });
    onFilterChange?.({ type: '', folder_id: '', status: '' });
  }, [onFilterChange]);

  const applyFilters = useCallback(() => {
    setFilterModalVisible(false);
    refetch();
  }, [refetch]);

  const renderDocument = ({ item }: { item: Document }) => {
    const statusStyle = getStatusStyle(item.status);
    return (
      <TouchableOpacity style={styles.documentCard} activeOpacity={0.7}>
        <View style={styles.documentIcon}>
          <Text style={styles.documentIconText}>
            {getDocTypeIcon(item.doc_type)}
          </Text>
        </View>
        <View style={styles.documentInfo}>
          <Text style={styles.documentTitle} numberOfLines={1}>
            {item.title}
          </Text>
          <View style={styles.documentMeta}>
            <View style={[styles.statusBadge, { backgroundColor: statusStyle.backgroundColor }]}>
              <Text style={[styles.statusText, { color: statusStyle.textColor }]}>
                {item.status}
              </Text>
            </View>
            <Text style={styles.documentDate}>
              {new Date(item.created_at).toLocaleDateString()}
            </Text>
          </View>
          {item.doc_type && (
            <Text style={styles.docTypeLabel}>{item.doc_type}</Text>
          )}
        </View>
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
        <View style={styles.modalContent}>
          <View style={styles.modalHeader}>
            <Text style={styles.modalTitle}>{t('filterDocuments')}</Text>
            <TouchableOpacity onPress={() => setFilterModalVisible(false)}>
              <Text style={styles.modalClose}>✕</Text>
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.modalBody}>
            <View style={styles.filterSection}>
              <Text style={styles.filterLabel}>{t('documentType')}</Text>
              <View style={styles.chipContainer}>
                {DOCUMENT_TYPES.map((type) => (
                  <TouchableOpacity
                    key={type.value}
                    style={[
                      styles.chip,
                      filters.type === type.value && styles.chipActive,
                    ]}
                    onPress={() => handleFilterChange('type', type.value)}
                  >
                    <Text
                      style={[
                        styles.chipText,
                        filters.type === type.value && styles.chipTextActive,
                      ]}
                    >
                      {t(type.label)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            <View style={styles.filterSection}>
              <Text style={styles.filterLabel}>{t('folder')}</Text>
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
                      {t(folder.label)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            <View style={styles.filterSection}>
              <Text style={styles.filterLabel}>{t('status')}</Text>
              <View style={styles.chipContainer}>
                {DOCUMENT_STATUSES.map((status) => (
                  <TouchableOpacity
                    key={status.value}
                    style={[
                      styles.chip,
                      filters.status === status.value && styles.chipActive,
                    ]}
                    onPress={() => handleFilterChange('status', status.value)}
                  >
                    <Text
                      style={[
                        styles.chipText,
                        filters.status === status.value && styles.chipTextActive,
                      ]}
                    >
                      {t(status.label)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>
          </ScrollView>

          <View style={styles.modalFooter}>
            <TouchableOpacity style={styles.clearButton} onPress={clearFilters}>
              <Text style={styles.clearButtonText}>{t('clearAll')}</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.applyButton} onPress={applyFilters}>
              <Text style={styles.applyButtonText}>{t('applyFilters')}</Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </Modal>
  );

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={tokens.accentColor} />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <View style={styles.headerTop}>
          <View>
            <Text style={styles.welcomeText}>{t('welcomeBack')}</Text>
            <Text style={styles.userName}>{user?.name || 'User'}</Text>
          </View>
          <View style={styles.headerButtons}>
            <TouchableOpacity
              style={styles.uploadButton}
              onPress={() => router.push('/(app)/upload')}
            >
              <Text style={styles.uploadButtonText}>📁</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={styles.cameraButton}
              onPress={() => router.push('/(app)/camera')}
            >
              <Text style={styles.cameraButtonText}>📷</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={styles.filterButton}
              onPress={() => setFilterModalVisible(true)}
            >
              <Text style={styles.filterButtonText}>⚙️</Text>
              {activeFilterCount > 0 && (
                <View style={styles.filterBadge}>
                  <Text style={styles.filterBadgeText}>{activeFilterCount}</Text>
                </View>
              )}
            </TouchableOpacity>
          </View>
        </View>
        {user?.tenantId && (
          <Text style={styles.tenantInfo}>{t('organizationId')}: {user.tenantId}</Text>
        )}
      </View>

      {activeFilterCount > 0 && (
        <View style={styles.activeFilters}>
          <Text style={styles.activeFiltersLabel}>Active filters:</Text>
          <ScrollView horizontal showsHorizontalScrollIndicator={false}>
            {filters.type && (
              <TouchableOpacity
                style={styles.activeFilterChip}
                onPress={() => handleFilterChange('type', '')}
              >
                <Text style={styles.activeFilterText}>
                  {t('documentType')}: {DOCUMENT_TYPES.find(d => d.value === filters.type)?.label}
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
                  {t('folder')}: {FOLDERS.find(f => f.value === filters.folder_id)?.label}
                </Text>
                <Text style={styles.activeFilterRemove}>✕</Text>
              </TouchableOpacity>
            )}
            {filters.status && (
              <TouchableOpacity
                style={styles.activeFilterChip}
                onPress={() => handleFilterChange('status', '')}
              >
                <Text style={styles.activeFilterText}>
                  {t('status')}: {DOCUMENT_STATUSES.find(s => s.value === filters.status)?.label}
                </Text>
                <Text style={styles.activeFilterRemove}>✕</Text>
              </TouchableOpacity>
            )}
          </ScrollView>
        </View>
      )}

      {error ? (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={refetch}>
            <Text style={styles.retryButtonText}>{t('retry')}</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <FlatList
          data={documents}
          renderItem={renderDocument}
          keyExtractor={(item) => item.id}
          contentContainerStyle={styles.listContent}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={refetch}
              tintColor={tokens.accentColor}
            />
          }
          ListEmptyComponent={
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>{t('noDocumentsFound')}</Text>
              <Text style={styles.emptySubtext}>
                {activeFilterCount > 0
                  ? t('tryAdjustingFilters')
                  : t('uploadYourFirst')}
              </Text>
            </View>
          }
        />
      )}

      {renderFilterModal()}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.secondary,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: tokens.backgroundColor.secondary,
  },
  header: {
    padding: 20,
    paddingTop: 10,
    borderBottomWidth: 1,
    borderBottomColor: colors.gray[800],
  },
  headerTop: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
  },
  welcomeText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
  },
  userName: {
    color: tokens.textColor.primary,
    fontSize: 24,
    fontWeight: 'bold',
    marginTop: 4,
  },
  tenantInfo: {
    color: tokens.textColor.muted,
    fontSize: 12,
    marginTop: 8,
    fontFamily: 'monospace',
  },
  headerButtons: {
    flexDirection: 'row',
    gap: 12,
  },
  uploadButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.gray[800],
    justifyContent: 'center',
    alignItems: 'center',
  },
  uploadButtonText: {
    fontSize: 20,
  },
  cameraButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: tokens.accentColor,
    justifyContent: 'center',
    alignItems: 'center',
  },
  cameraButtonText: {
    fontSize: 20,
  },
  filterButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.gray[800],
    justifyContent: 'center',
    alignItems: 'center',
  },
  filterButtonText: {
    fontSize: 20,
  },
  filterBadge: {
    position: 'absolute',
    top: -4,
    right: -4,
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
  activeFilters: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: colors.gray[800],
    gap: 8,
  },
  activeFiltersLabel: {
    color: tokens.textColor.secondary,
    fontSize: 12,
  },
  activeFilterChip: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 16,
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
  listContent: {
    padding: 16,
    paddingBottom: 100,
  },
  documentCard: {
    flexDirection: 'row',
    backgroundColor: colors.gray[800],
    borderRadius: tokens.cardRadius,
    padding: 16,
    marginBottom: 12,
    alignItems: 'center',
  },
  documentIcon: {
    width: 48,
    height: 48,
    borderRadius: 8,
    backgroundColor: tokens.accentColor,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  documentIconText: {
    fontSize: 24,
  },
  documentInfo: {
    flex: 1,
  },
  documentTitle: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 6,
  },
  documentMeta: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 4,
    marginRight: 12,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '500',
    textTransform: 'capitalize',
  },
  documentDate: {
    color: tokens.textColor.secondary,
    fontSize: 12,
  },
  docTypeLabel: {
    color: tokens.textColor.muted,
    fontSize: 11,
    marginTop: 4,
    textTransform: 'capitalize',
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  errorText: {
    color: tokens.errorColor,
    fontSize: 16,
    marginBottom: 16,
    textAlign: 'center',
  },
  retryButton: {
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  retryButtonText: {
    color: tokens.textColor.primary,
    fontSize: 14,
    fontWeight: '600',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingTop: 100,
  },
  emptyText: {
    color: tokens.textColor.secondary,
    fontSize: 18,
    fontWeight: '600',
  },
  emptySubtext: {
    color: tokens.textColor.muted,
    fontSize: 14,
    marginTop: 8,
    textAlign: 'center',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'flex-end',
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
    padding: 20,
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
    padding: 20,
  },
  filterSection: {
    marginBottom: 24,
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
  modalFooter: {
    flexDirection: 'row',
    padding: 20,
    gap: 12,
    borderTopWidth: 1,
    borderTopColor: colors.gray[800],
  },
  clearButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 8,
    backgroundColor: colors.gray[800],
    alignItems: 'center',
  },
  clearButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  applyButton: {
    flex: 2,
    paddingVertical: 14,
    borderRadius: 8,
    backgroundColor: tokens.accentColor,
    alignItems: 'center',
  },
  applyButtonText: {
    color: colors.gray[50],
    fontSize: 16,
    fontWeight: '600',
  },
});
