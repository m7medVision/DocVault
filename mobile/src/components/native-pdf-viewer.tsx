import { useState } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';
import { PdfView } from '@kishannareshpal/expo-pdf';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useCachedFile } from '@/features/documents/use-cached-file';

export interface PageChangePayload {
  pageIndex: number;
  pageCount: number;
}

export interface NativePdfViewerProps {
  url: string | null;
  documentId: string;
  extension?: string;
  onPageChange?: (payload: PageChangePayload) => void;
}

export function NativePdfViewer({
  url,
  documentId,
  extension = 'pdf',
  onPageChange,
}: NativePdfViewerProps) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { localUri, loading, error, reload } = useCachedFile({
    documentId,
    remoteUrl: url,
    extension,
  });
  const [renderError, setRenderError] = useState<string | null>(null);

  if (!localUri) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        {error || renderError ? (
          <>
            <ThemedText type="small" themeColor="danger">
              {error ?? renderError}
            </ThemedText>
            <ThemedText
              type="code"
              themeColor="accent"
              onPress={reload}
              accessibilityRole="button"
            >
              {t('common.retry')}
            </ThemedText>
          </>
        ) : (
          <>
            <ActivityIndicator />
            <ThemedText type="code" themeColor="textSecondary">
              {loading ? t('viewer.downloadFirst') : t('viewer.loading')}
            </ThemedText>
          </>
        )}
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <PdfView
        uri={localUri}
        style={styles.pdf}
        doubleTapToZoom
        pagingEnabled={false}
        fitMode="width"
        onLoadComplete={(e) => {
          setRenderError(null);
          onPageChange?.({ pageIndex: 0, pageCount: e.pageCount });
        }}
        onPageChanged={(e) => {
          onPageChange?.({ pageIndex: e.pageIndex, pageCount: e.pageCount });
        }}
        onError={(e) => {
          setRenderError(e.message || t('viewer.loadFailed'));
        }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  pdf: {
    flex: 1,
  },
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: Spacing.two,
    padding: Spacing.four,
    borderRadius: Spacing.two,
  },
});
