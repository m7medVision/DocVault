import { ActivityIndicator, Dimensions, StyleSheet, View } from 'react-native';
import { Image } from 'expo-image';
import { Galeria } from '@nandorojo/galeria';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useCachedFile } from '@/features/documents/use-cached-file';

export interface NativeImageViewerProps {
  url: string | null;
  documentId: string;
  extension?: string;
}

const GaleriaImage = (Galeria as unknown as {
  Image: (props: { index?: number; children: React.ReactNode }) => React.ReactNode;
}).Image;

export function NativeImageViewer({
  url,
  documentId,
  extension = 'jpg',
}: NativeImageViewerProps) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { localUri, loading, error, reload } = useCachedFile({
    documentId,
    remoteUrl: url,
    extension,
  });

  if (!localUri) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        {error ? (
          <>
            <ThemedText type="small" themeColor="danger">
              {error}
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
    <View style={[styles.container, { backgroundColor: theme.background }]}>
      <Galeria urls={[localUri]} theme="light">
        <GaleriaImage>
          <Image
            source={{ uri: localUri }}
            style={styles.image}
            contentFit="contain"
            recyclingKey={localUri}
          />
        </GaleriaImage>
      </Galeria>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: Spacing.two,
    overflow: 'hidden',
  },
  image: {
    width: Dimensions.get('window').width,
    height: Dimensions.get('window').height * 0.8,
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
