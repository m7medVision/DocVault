// Document Detail Screen
// Shows document metadata, OCR text with confidence scores, version history,
// provides download action, and allows metadata correction.

import { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  Linking,
  RefreshControl,
  TextInput,
  Modal,
} from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { tokens } from '../../src/theme/tokens';
import {
  DocumentDetail,
  DocumentPage,
  DocumentVersion,
  DocumentMetadata,
  MetadataUpdate,
  getDocument,
  getDocumentDownloadURL,
  getDocumentPages,
  getDocumentVersions,
  updateDocumentMetadata,
  formatFileSize,
  formatConfidence,
  getStatusColor,
  getDocTypeLabel,
} from '../../src/lib/api';
import i18n from '../../src/i18n';

// Translation helpers
const t = (key: string) => i18n.t(key);

export default function DocumentDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { accessToken, user } = useAuth();

  const [document, setDocument] = useState<DocumentDetail | null>(null);
  const [pages, setPages] = useState<DocumentPage[]>([]);
  const [versions, setVersions] = useState<DocumentVersion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState<'details' | 'ocr' | 'versions'>('details');

  // Metadata correction modal
  const [showCorrectionModal, setShowCorrectionModal] = useState(false);
  const [editingMetadata, setEditingMetadata] = useState<DocumentMetadata | null>(null);
  const [correctedValue, setCorrectedValue] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  // Fetch document data
  const fetchData = useCallback(async (showRefresh = false) => {
    if (!accessToken || !id) return;

    if (showRefresh) setIsRefreshing(true);
    else setIsLoading(true);

    try {
      const [docData, pagesData, versionsData] = await Promise.all([
        getDocument(accessToken, id),
        getDocumentPages(accessToken, id),
        getDocumentVersions(accessToken, id),
      ]);

      setDocument(docData);
      setPages(pagesData);
      setVersions(versionsData);
    } catch (error) {
      console.error('Failed to fetch document:', error);
      Alert.alert('Error', 'Failed to load document details');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, [accessToken, id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Handle download
  const handleDownload = async () => {
    if (!accessToken || !id) return;

    try {
      const { presigned_url, expires_at } = await getDocumentDownloadURL(accessToken, id);
      const expires = new Date(expires_at);
      const timeLeft = Math.round((expires.getTime() - Date.now()) / 1000 / 60);

      Alert.alert(
        'Download Document',
        `This link expires in ${timeLeft} minutes. Open the download URL?`,
        [
          { text: 'Cancel', style: 'cancel' },
          {
            text: 'Open',
            onPress: () => Linking.openURL(presigned_url),
          },
        ]
      );
    } catch (error) {
      console.error('Failed to get download URL:', error);
      Alert.alert('Error', 'Failed to get download link');
    }
  };

  // Handle metadata correction
  const handleCorrectMetadata = (metadata: DocumentMetadata) => {
    setEditingMetadata(metadata);
    setCorrectedValue(metadata.corrected_value || metadata.extracted_value || '');
    setShowCorrectionModal(true);
  };

  const handleSaveCorrection = async () => {
    if (!accessToken || !id || !editingMetadata) return;

    setIsSaving(true);
    try {
      const update: MetadataUpdate = {
        key: editingMetadata.key,
        value: correctedValue,
      };
      await updateDocumentMetadata(accessToken, id, [update]);

      // Refresh data
      await fetchData();
      setShowCorrectionModal(false);
      Alert.alert('Success', 'Metadata corrected successfully');
    } catch (error) {
      console.error('Failed to save correction:', error);
      Alert.alert('Error', 'Failed to save correction');
    } finally {
      setIsSaving(false);
    }
  };

  // Format date for display
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString(user?.locale === 'ar' ? 'ar' : 'en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Get metadata display key
  const getMetadataLabel = (key: string) => {
    const labels: Record<string, string> = {
      issuer: 'Issuer',
      amount: 'Amount',
      currency: 'Currency',
      issue_date: 'Issue Date',
      expiry_date: 'Expiry Date',
      document_number: 'Document Number',
      language: 'Language',
    };
    return labels[key] || key;
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={tokens.accentColor} />
        <Text style={styles.loadingText}>Loading document...</Text>
      </View>
    );
  }

  if (!document) {
    return (
      <View style={styles.errorContainer}>
        <Text style={styles.errorText}>Document not found</Text>
        <TouchableOpacity style={styles.backButton} onPress={() => router.back()}>
          <Text style={styles.backButtonText}>Go Back</Text>
        </TouchableOpacity>
      </View>
    );
  }

  const { document: doc, metadata } = document;

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => router.back()} style={styles.backIcon}>
          <Text style={styles.backIconText}>←</Text>
        </TouchableOpacity>
        <View style={styles.headerContent}>
          <Text style={styles.title} numberOfLines={1}>{doc.title}</Text>
          <View style={styles.statusBadge}>
            <View style={[styles.statusDot, { backgroundColor: getStatusColor(doc.status) }]} />
            <Text style={styles.statusText}>{doc.status.toUpperCase()}</Text>
          </View>
        </View>
      </View>

      {/* Tab Bar */}
      <View style={styles.tabBar}>
        {(['details', 'ocr', 'versions'] as const).map((tab) => (
          <TouchableOpacity
            key={tab}
            style={[styles.tab, activeTab === tab && styles.activeTab]}
            onPress={() => setActiveTab(tab)}
          >
            <Text style={[styles.tabText, activeTab === tab && styles.activeTabText]}>
              {tab === 'details' ? 'Details' : tab === 'ocr' ? 'OCR Text' : 'Versions'}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Content */}
      <ScrollView
        style={styles.content}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={() => fetchData(true)}
            tintColor={tokens.accentColor}
          />
        }
      >
        {/* Details Tab */}
        {activeTab === 'details' && (
          <View style={styles.section}>
            {/* Document Info Card */}
            <View style={styles.card}>
              <Text style={styles.cardTitle}>Document Information</Text>

              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>Type</Text>
                <Text style={styles.infoValue}>{getDocTypeLabel(doc.doc_type)}</Text>
              </View>

              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>Language</Text>
                <Text style={styles.infoValue}>{doc.language || 'Not specified'}</Text>
              </View>

              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>Created</Text>
                <Text style={styles.infoValue}>{formatDate(doc.created_at)}</Text>
              </View>

              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>Tenant ID</Text>
                <Text style={styles.infoValue} numberOfLines={1}>{doc.tenant_id}</Text>
              </View>
            </View>

            {/* Metadata Card */}
            {metadata.length > 0 && (
              <View style={styles.card}>
                <Text style={styles.cardTitle}>Extracted Metadata</Text>
                {metadata.map((item) => (
                  <TouchableOpacity
                    key={item.id}
                    style={styles.metadataRow}
                    onPress={() => handleCorrectMetadata(item)}
                  >
                    <View style={styles.metadataInfo}>
                      <Text style={styles.metadataLabel}>{getMetadataLabel(item.key)}</Text>
                      {item.corrected_value ? (
                        <>
                          <Text style={styles.extractedValue}>
                            {item.extracted_value}
                          </Text>
                          <Text style={styles.correctedValue}>
                            → {item.corrected_value}
                          </Text>
                        </>
                      ) : (
                        <Text style={styles.metadataValue}>
                          {item.extracted_value || 'N/A'}
                        </Text>
                      )}
                      {item.corrected_by && (
                        <Text style={styles.correctedBy}>
                          Corrected by {item.corrected_by}
                        </Text>
                      )}
                    </View>
                    <Text style={styles.editIcon}>✎</Text>
                  </TouchableOpacity>
                ))}
                <Text style={styles.tapHint}>Tap any field to correct</Text>
              </View>
            )}

            {/* Download Button */}
            <TouchableOpacity style={styles.downloadButton} onPress={handleDownload}>
              <Text style={styles.downloadButtonText}>↓ Download Document</Text>
            </TouchableOpacity>
          </View>
        )}

        {/* OCR Text Tab */}
        {activeTab === 'ocr' && (
          <View style={styles.section}>
            {pages.length === 0 ? (
              <View style={styles.emptyState}>
                <Text style={styles.emptyStateText}>No OCR text available</Text>
                <Text style={styles.emptyStateSubtext}>
                  OCR processing may still be in progress or failed
                </Text>
              </View>
            ) : (
              pages.map((page) => (
                <View key={page.id} style={styles.card}>
                  <View style={styles.pageHeader}>
                    <Text style={styles.pageTitle}>Page {page.page_number}</Text>
                    <View style={styles.confidenceBadge}>
                      <Text style={styles.confidenceText}>
                        {formatConfidence(page.confidence)}
                      </Text>
                    </View>
                  </View>
                  <Text style={styles.ocrText}>
                    {page.ocr_text || 'No text extracted'}
                  </Text>
                  <Text style={styles.ocrModel}>via {page.ocr_model}</Text>
                </View>
              ))
            )}
          </View>
        )}

        {/* Versions Tab */}
        {activeTab === 'versions' && (
          <View style={styles.section}>
            {versions.length === 0 ? (
              <View style={styles.emptyState}>
                <Text style={styles.emptyStateText}>No version history</Text>
              </View>
            ) : (
              versions
                .sort((a, b) => b.version_number - a.version_number)
                .map((version, index) => (
                  <View key={version.id} style={styles.card}>
                    <View style={styles.versionHeader}>
                      <Text style={styles.versionTitle}>
                        Version {version.version_number}
                      </Text>
                      {index === 0 && (
                        <View style={styles.latestBadge}>
                          <Text style={styles.latestBadgeText}>Latest</Text>
                        </View>
                      )}
                    </View>
                    <View style={styles.infoRow}>
                      <Text style={styles.infoLabel}>Size</Text>
                      <Text style={styles.infoValue}>{formatFileSize(version.file_size)}</Text>
                    </View>
                    <View style={styles.infoRow}>
                      <Text style={styles.infoLabel}>Type</Text>
                      <Text style={styles.infoValue}>{version.mime_type}</Text>
                    </View>
                    <View style={styles.infoRow}>
                      <Text style={styles.infoLabel}>Uploaded</Text>
                      <Text style={styles.infoValue}>{formatDate(version.created_at)}</Text>
                    </View>
                    {version.uploaded_by && (
                      <View style={styles.infoRow}>
                        <Text style={styles.infoLabel}>By</Text>
                        <Text style={styles.infoValue}>{version.uploaded_by}</Text>
                      </View>
                    )}
                  </View>
                ))
            )}
          </View>
        )}
      </ScrollView>

      {/* Metadata Correction Modal */}
      <Modal
        visible={showCorrectionModal}
        animationType="slide"
        transparent
        onRequestClose={() => setShowCorrectionModal(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>
              Correct {editingMetadata ? getMetadataLabel(editingMetadata.key) : ''}
            </Text>

            {editingMetadata?.extracted_value && (
              <View style={styles.originalValue}>
                <Text style={styles.originalLabel}>Original:</Text>
                <Text style={styles.originalText}>{editingMetadata.extracted_value}</Text>
              </View>
            )}

            <TextInput
              style={styles.correctionInput}
              value={correctedValue}
              onChangeText={setCorrectedValue}
              placeholder="Enter corrected value"
              placeholderTextColor={tokens.textColor.muted}
              autoFocus
            />

            <View style={styles.modalActions}>
              <TouchableOpacity
                style={styles.cancelButton}
                onPress={() => setShowCorrectionModal(false)}
              >
                <Text style={styles.cancelButtonText}>Cancel</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.saveButton, isSaving && styles.saveButtonDisabled]}
                onPress={handleSaveCorrection}
                disabled={isSaving}
              >
                {isSaving ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text style={styles.saveButtonText}>Save</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
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
  loadingText: {
    color: tokens.textColor.secondary,
    marginTop: 12,
    fontSize: 16,
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: tokens.backgroundColor.secondary,
    padding: 24,
  },
  errorText: {
    color: tokens.errorColor,
    fontSize: 18,
    marginBottom: 16,
  },
  backButton: {
    backgroundColor: tokens.accentColor,
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: tokens.cardRadius,
  },
  backButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: tokens.backgroundColor.primary,
    paddingTop: 48,
    paddingBottom: 16,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: tokens.backgroundColor.tertiary,
  },
  backIcon: {
    width: 40,
    height: 40,
    justifyContent: 'center',
    alignItems: 'center',
  },
  backIconText: {
    color: tokens.textColor.primary,
    fontSize: 24,
  },
  headerContent: {
    flex: 1,
    marginLeft: 8,
  },
  title: {
    color: tokens.textColor.primary,
    fontSize: 20,
    fontWeight: '600',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 4,
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 6,
  },
  statusText: {
    color: tokens.textColor.secondary,
    fontSize: 12,
    fontWeight: '500',
  },
  tabBar: {
    flexDirection: 'row',
    backgroundColor: tokens.backgroundColor.primary,
    borderBottomWidth: 1,
    borderBottomColor: tokens.backgroundColor.tertiary,
  },
  tab: {
    flex: 1,
    paddingVertical: 14,
    alignItems: 'center',
    borderBottomWidth: 2,
    borderBottomColor: 'transparent',
  },
  activeTab: {
    borderBottomColor: tokens.accentColor,
  },
  tabText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    fontWeight: '500',
  },
  activeTabText: {
    color: tokens.accentColor,
  },
  content: {
    flex: 1,
  },
  section: {
    padding: tokens.screenPadding,
  },
  card: {
    backgroundColor: tokens.backgroundColor.primary,
    borderRadius: tokens.cardRadius,
    padding: tokens.cardPadding,
    marginBottom: 16,
    ...tokens.cardShadow,
  },
  cardTitle: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 16,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: tokens.backgroundColor.tertiary,
  },
  infoLabel: {
    color: tokens.textColor.secondary,
    fontSize: 14,
  },
  infoValue: {
    color: tokens.textColor.primary,
    fontSize: 14,
    fontWeight: '500',
    flex: 1,
    textAlign: 'right',
    marginLeft: 12,
  },
  metadataRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: tokens.backgroundColor.tertiary,
  },
  metadataInfo: {
    flex: 1,
  },
  metadataLabel: {
    color: tokens.textColor.secondary,
    fontSize: 12,
    marginBottom: 2,
  },
  metadataValue: {
    color: tokens.textColor.primary,
    fontSize: 15,
    fontWeight: '500',
  },
  extractedValue: {
    color: tokens.textColor.muted,
    fontSize: 13,
    textDecorationLine: 'line-through',
  },
  correctedValue: {
    color: tokens.textColor.primary,
    fontSize: 15,
    fontWeight: '500',
  },
  correctedBy: {
    color: tokens.textColor.muted,
    fontSize: 11,
    marginTop: 2,
  },
  editIcon: {
    color: tokens.accentColor,
    fontSize: 18,
    marginLeft: 12,
  },
  tapHint: {
    color: tokens.textColor.muted,
    fontSize: 12,
    textAlign: 'center',
    marginTop: 12,
  },
  downloadButton: {
    backgroundColor: tokens.accentColor,
    paddingVertical: 16,
    borderRadius: tokens.cardRadius,
    alignItems: 'center',
    marginTop: 8,
  },
  downloadButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: 48,
  },
  emptyStateText: {
    color: tokens.textColor.secondary,
    fontSize: 16,
    fontWeight: '500',
  },
  emptyStateSubtext: {
    color: tokens.textColor.muted,
    fontSize: 14,
    marginTop: 8,
    textAlign: 'center',
  },
  pageHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  pageTitle: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  confidenceBadge: {
    backgroundColor: tokens.backgroundColor.tertiary,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  confidenceText: {
    color: tokens.textColor.secondary,
    fontSize: 12,
    fontWeight: '500',
  },
  ocrText: {
    color: tokens.textColor.primary,
    fontSize: 14,
    lineHeight: 22,
    backgroundColor: tokens.backgroundColor.tertiary,
    padding: 12,
    borderRadius: 8,
  },
  ocrModel: {
    color: tokens.textColor.muted,
    fontSize: 11,
    marginTop: 8,
    textAlign: 'right',
  },
  versionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  versionTitle: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  latestBadge: {
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  latestBadgeText: {
    color: tokens.textColor.primary,
    fontSize: 11,
    fontWeight: '600',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: tokens.backgroundColor.primary,
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 24,
    paddingBottom: 40,
  },
  modalTitle: {
    color: tokens.textColor.primary,
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 20,
  },
  originalValue: {
    backgroundColor: tokens.backgroundColor.tertiary,
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  originalLabel: {
    color: tokens.textColor.secondary,
    fontSize: 12,
    marginBottom: 4,
  },
  originalText: {
    color: tokens.textColor.muted,
    fontSize: 15,
    textDecorationLine: 'line-through',
  },
  correctionInput: {
    backgroundColor: tokens.backgroundColor.tertiary,
    borderRadius: 8,
    padding: 16,
    color: tokens.textColor.primary,
    fontSize: 16,
    marginBottom: 20,
  },
  modalActions: {
    flexDirection: 'row',
    gap: 12,
  },
  cancelButton: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.tertiary,
    paddingVertical: 14,
    borderRadius: tokens.cardRadius,
    alignItems: 'center',
  },
  cancelButtonText: {
    color: tokens.textColor.secondary,
    fontSize: 16,
    fontWeight: '600',
  },
  saveButton: {
    flex: 1,
    backgroundColor: tokens.accentColor,
    paddingVertical: 14,
    borderRadius: tokens.cardRadius,
    alignItems: 'center',
  },
  saveButtonDisabled: {
    opacity: 0.7,
  },
  saveButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
});
