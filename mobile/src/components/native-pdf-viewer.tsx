import { useState } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useCachedFile } from '@/features/documents/use-cached-file';
import { NativeModuleErrorBoundary } from '@/components/native-module-error-boundary';

type PdfViewComponent = React.ComponentType<{
  uri: string;
  style?: object;
  doubleTapToZoom?: boolean;
  pagingEnabled?: boolean;
  fitMode?: 'width' | 'height' | 'both';
  onLoadComplete?: (e: { nativeEvent: { pageCount: number } }) => void;
  onPageChanged?: (e: { nativeEvent: { pageIndex: number; pageCount: number } }) => void;
  onError?: (e: { nativeEvent: { message?: string } }) => void;
}>;

let cachedPdfView: PdfViewComponent | null = null;
let cachedPdfViewError: string | null = null;

function loadPdfView(): PdfViewComponent | null {
  if (cachedPdfView) return cachedPdfView;
  if (cachedPdfViewError) return null;
  try {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const mod = require('@kishannareshpal/expo-pdf');
    cachedPdfView = mod.PdfView as PdfViewComponent;
    return cachedPdfView;
  } catch (err) {
    cachedPdfViewError =
      err instanceof Error ? err.message : 'PDF native module not available';
    return null;
  }
}

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
  const PdfView = loadPdfView();

  if (!PdfView) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        <ThemedText type="small" themeColor="danger">
          PDF viewer not available
        </ThemedText>
        <ThemedText type="code" themeColor="textSecondary">
          {cachedPdfViewError ?? 'Run `npx expo prebuild` and rebuild the dev client.'}
        </ThemedText>
      </View>
    );
  }

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
      <NativeModuleErrorBoundary fallbackTitle="PDF viewer unavailable">
        <PdfView
          uri={localUri}
          style={styles.pdf}
          doubleTapToZoom
          pagingEnabled={false}
          fitMode="width"
          onLoadComplete={(e) => {
            setRenderError(null);
            onPageChange?.({ pageIndex: 0, pageCount: e.nativeEvent.pageCount });
          }}
          onPageChanged={(e) => {
            onPageChange?.({
              pageIndex: e.nativeEvent.pageIndex,
              pageCount: e.nativeEvent.pageCount,
            });
          }}
          onError={(e) => {
            setRenderError(e.nativeEvent.message || t('viewer.loadFailed'));
          }}
        />
      </NativeModuleErrorBoundary>
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
