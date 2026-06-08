import { Linking, View, StyleSheet, Pressable } from 'react-native';

import { NativePdfViewer } from '@/components/native-pdf-viewer';
import { NativeImageViewer } from '@/components/native-image-viewer';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { formatFileSize } from '@/constants/document';
import { useTranslation } from '@/lib/i18n';

type PreviewMode = 'pdf' | 'image' | 'fallback';

function getPreviewMode(mimeType: string): PreviewMode {
  if (mimeType === 'application/pdf') return 'pdf';
  if (
    mimeType.startsWith('image/jpeg') ||
    mimeType.startsWith('image/png') ||
    mimeType.startsWith('image/webp') ||
    mimeType.startsWith('image/gif')
  ) {
    return 'image';
  }
  return 'fallback';
}

function extensionFromMime(mimeType: string, fallback: string): string {
  if (mimeType === 'application/pdf') return 'pdf';
  if (mimeType.startsWith('image/jpeg')) return 'jpg';
  if (mimeType.startsWith('image/png')) return 'png';
  if (mimeType.startsWith('image/webp')) return 'webp';
  if (mimeType.startsWith('image/gif')) return 'gif';
  return fallback;
}

interface FilePreviewProps {
  url: string | null;
  mimeType: string;
  fileSize?: number;
  documentId: string;
  onPdfPageChange?: (info: { pageIndex: number; pageCount: number }) => void;
}

export function FilePreview({ url, mimeType, fileSize, documentId, onPdfPageChange }: FilePreviewProps) {
  const theme = useTheme();
  const { t } = useTranslation();

  if (!url) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        <ThemedText type="small" themeColor="textSecondary">
          {t('viewer.loading')}
        </ThemedText>
      </View>
    );
  }

  const mode = getPreviewMode(mimeType);
  const extension = extensionFromMime(mimeType, 'bin');

  if (mode === 'pdf') {
    return (
      <NativePdfViewer
        url={url}
        documentId={documentId}
        extension={extension}
        onPageChange={onPdfPageChange}
      />
    );
  }

  if (mode === 'image') {
    return <NativeImageViewer url={url} documentId={documentId} extension={extension} />;
  }

  return (
    <View style={[styles.center, { backgroundColor: theme.surface }]}>
      <ThemedText type="small" themeColor="textSecondary">
        Preview not available for this file type.
      </ThemedText>
      {fileSize ? (
        <ThemedText type="code">{formatFileSize(fileSize)}</ThemedText>
      ) : null}
      <Pressable
        onPress={() => Linking.openURL(url)}
        style={[styles.downloadBtn, { backgroundColor: theme.accent }]}
      >
        <ThemedText type="smallBold" style={{ color: '#fff' }}>
          Download File
        </ThemedText>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: Spacing.two,
    padding: Spacing.four,
    borderRadius: Spacing.two,
  },
  downloadBtn: {
    marginTop: Spacing.two,
    paddingHorizontal: Spacing.four,
    paddingVertical: Spacing.two,
    borderRadius: Spacing.two,
  },
});
