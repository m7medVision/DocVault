import { Linking, View, StyleSheet, Pressable } from 'react-native';

import { PdfViewer } from '@/components/pdf-viewer';
import { ImageViewer } from '@/components/image-viewer';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { formatFileSize } from '@/constants/document';

type PreviewMode = 'pdf' | 'image' | 'fallback';

function getPreviewMode(mimeType: string): PreviewMode {
  if (mimeType === 'application/pdf') return 'pdf';
  if (mimeType.startsWith('image/jpeg') || mimeType.startsWith('image/png')) return 'image';
  return 'fallback';
}

interface FilePreviewProps {
  url: string | null;
  mimeType: string;
  fileSize?: number;
}

export function FilePreview({ url, mimeType, fileSize }: FilePreviewProps) {
  const theme = useTheme();

  if (!url) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        <ThemedText type="small" themeColor="textSecondary">
          Loading preview...
        </ThemedText>
      </View>
    );
  }

  const mode = getPreviewMode(mimeType);

  if (mode === 'pdf') {
    return <PdfViewer url={url} />;
  }

  if (mode === 'image') {
    return <ImageViewer url={url} />;
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
