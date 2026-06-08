import { useState } from 'react';
import { ActivityIndicator, Linking, Pressable, StyleSheet, View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { FilePreview } from '@/components/file-preview';
import { DocumentViewer } from '@/components/document-viewer';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useDocumentDetail } from '@/features/documents/use-document-detail';
import { useTranslation } from '@/lib/i18n';

type Tab = 'document' | 'ocr' | 'translated';

interface PdfPageInfo {
  pageIndex: number;
  pageCount: number;
}

export default function DocumentDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const {
    document,
    pages,
    downloadUrl,
    mimeType,
    fileSize,
    loading,
    error,
  } = useDocumentDetail(id);
  const [activeTab, setActiveTab] = useState<Tab>('document');
  const [pdfPage, setPdfPage] = useState<PdfPageInfo | null>(null);
  const { t } = useTranslation();

  const hasTranslation = document?.language && document.language !== 'en' && pages.some((p) => p.translated_text);
  const tabs: { key: Tab; label: string }[] = [
    { key: 'document', label: 'Document' },
    { key: 'ocr', label: 'OCR Text' },
  ];
  if (hasTranslation) {
    tabs.push({ key: 'translated', label: 'Translated' });
  }

  if (loading) {
    return (
      <DocVaultScreen>
        <ActivityIndicator style={styles.loader} />
      </DocVaultScreen>
    );
  }

  if (error || !document) {
    return (
      <DocVaultScreen>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">Back</ThemedText>
        </Pressable>
        <View style={styles.errorBox}>
          <ThemedText type="small" style={{ color: '#DC2626' }}>
            {error ?? 'Document not found'}
          </ThemedText>
        </View>
      </DocVaultScreen>
    );
  }

  return (
    <DocVaultScreen scroll={false} style={styles.screen}>
      <View style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">Back</ThemedText>
        </Pressable>
        <View style={styles.titleRow}>
          <ThemedText type="subtitle" style={styles.flexText} numberOfLines={2}>
            {document.title}
          </ThemedText>
          {pdfPage ? (
            <View style={styles.pagePill}>
              <ThemedText type="code" themeColor="textSecondary">
                {t('viewer.pageOf', { current: pdfPage.pageIndex + 1, total: pdfPage.pageCount })}
              </ThemedText>
            </View>
          ) : null}
          {downloadUrl && (
            <Pressable
              style={styles.downloadBtn}
              onPress={() => Linking.openURL(downloadUrl)}
            >
              <ThemedText type="smallBold" style={{ color: '#fff' }}>
                Download
              </ThemedText>
            </Pressable>
          )}
        </View>
        <View style={styles.metaRow}>
          <ThemedText type="small" themeColor="textSecondary">
            {document.doc_type} · {document.language || 'unknown'} ·{' '}
            {new Date(document.created_at).toLocaleDateString()}
          </ThemedText>
          <StatusBadge status={document.status} />
        </View>
      </View>

      <View style={styles.tabBar}>
        {tabs.map((tab) => (
          <Pressable
            key={tab.key}
            onPress={() => setActiveTab(tab.key)}
            style={[styles.tab, activeTab === tab.key && styles.tabActive]}
          >
            <ThemedText
              type="smallBold"
              style={activeTab === tab.key ? styles.tabTextActive : undefined}
            >
              {tab.label}
            </ThemedText>
          </Pressable>
        ))}
      </View>

      <View style={styles.content}>
        {activeTab === 'document' && (
          <FilePreview
            url={downloadUrl}
            mimeType={mimeType}
            fileSize={fileSize}
            documentId={id}
            onPdfPageChange={setPdfPage}
          />
        )}
        {activeTab === 'ocr' && <DocumentViewer pages={pages} />}
        {activeTab === 'translated' && (
          <DocumentViewer pages={pages} translated />
        )}
      </View>
    </DocVaultScreen>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, { bg: string; text: string }> = {
    pending: { bg: '#FEF3C7', text: '#92400E' },
    processed: { bg: '#DCFCE7', text: '#166534' },
    failed: { bg: '#FEE2E2', text: '#991B1B' },
  };
  const c = colors[status] || { bg: '#F4F4F5', text: '#3F3F46' };

  return (
    <View style={[styles.badge, { backgroundColor: c.bg }]}>
      <ThemedText type="code" style={{ color: c.text, fontSize: 11 }}>{status}</ThemedText>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
  },
  loader: {
    marginTop: Spacing.five,
  },
  errorBox: {
    borderRadius: Spacing.two,
    padding: Spacing.three,
    backgroundColor: 'rgba(220,38,38,0.08)',
  },
  header: {
    gap: Spacing.two,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  flexText: {
    flex: 1,
  },
  downloadBtn: {
    backgroundColor: '#208AEF',
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one + 2,
    borderRadius: Spacing.two,
  },
  pagePill: {
    paddingHorizontal: Spacing.two,
    paddingVertical: Spacing.half,
    borderRadius: 999,
    backgroundColor: '#F4F4F5',
  },
  metaRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  badge: {
    borderRadius: 999,
    paddingHorizontal: Spacing.two,
    paddingVertical: Spacing.half,
  },
  tabBar: {
    flexDirection: 'row',
    gap: Spacing.one,
  },
  tab: {
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one + 2,
    borderRadius: 999,
    backgroundColor: '#F4F4F5',
  },
  tabActive: {
    backgroundColor: '#208AEF',
  },
  tabTextActive: {
    color: '#fff',
  },
  content: {
    flex: 1,
  },
});
