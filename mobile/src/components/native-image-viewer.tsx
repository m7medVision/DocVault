import { ActivityIndicator, Dimensions, StyleSheet, View } from 'react-native';
import { Image } from 'expo-image';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useCachedFile } from '@/features/documents/use-cached-file';
import { NativeModuleErrorBoundary } from '@/components/native-module-error-boundary';

export interface NativeImageViewerProps {
  url: string | null;
  documentId: string;
  extension?: string;
}

type GaleriaModuleShape = {
  Galeria: React.ComponentType<{
    urls: string[];
    theme?: 'light' | 'dark';
    children: React.ReactNode;
  }> & {
    Image: (props: { index?: number; children: React.ReactNode }) => React.ReactNode;
  };
};

let cachedGaleria: GaleriaModuleShape['Galeria'] | null = null;
let cachedGaleriaError: string | null = null;

function loadGaleria(): GaleriaModuleShape['Galeria'] | null {
  if (cachedGaleria) return cachedGaleria;
  if (cachedGaleriaError) return null;
  try {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const mod = require('@nandorojo/galeria');
    cachedGaleria = mod.Galeria as GaleriaModuleShape['Galeria'];
    return cachedGaleria;
  } catch (err) {
    cachedGaleriaError =
      err instanceof Error ? err.message : 'Image viewer native module not available';
    return null;
  }
}

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
  const Galeria = loadGaleria();

  if (!Galeria) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        <ThemedText type="small" themeColor="danger">
          Image viewer not available
        </ThemedText>
        <ThemedText type="code" themeColor="textSecondary">
          {cachedGaleriaError ?? 'Run `npx expo prebuild` and rebuild the dev client.'}
        </ThemedText>
      </View>
    );
  }

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

  const GaleriaImage = (Galeria as unknown as {
    Image: (props: { index?: number; children: React.ReactNode }) => React.ReactNode;
  }).Image;

  return (
    <View style={[styles.container, { backgroundColor: theme.background }]}>
      <NativeModuleErrorBoundary fallbackTitle="Image viewer unavailable">
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
      </NativeModuleErrorBoundary>
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
