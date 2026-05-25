import { ScrollView, StyleSheet, View, type ViewProps } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { BottomTabInset, MaxContentWidth, Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';

interface DocVaultScreenProps extends ViewProps {
  scroll?: boolean;
}

export function DocVaultScreen({ children, scroll = true, style, ...props }: DocVaultScreenProps) {
  const theme = useTheme();
  const content = (
    <SafeAreaView style={[styles.safeArea, !scroll && styles.safeAreaFull]}>
      <View style={[styles.content, !scroll && styles.contentFull, style]} {...props}>
        {children}
      </View>
    </SafeAreaView>
  );

  if (!scroll) {
    return <View style={[styles.root, { backgroundColor: theme.background }]}>{content}</View>;
  }

  return (
    <ScrollView
      style={[styles.root, { backgroundColor: theme.background }]}
      contentContainerStyle={styles.scrollContent}>
      {content}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    alignItems: 'center',
  },
  safeArea: {
    width: '100%',
    maxWidth: MaxContentWidth,
    paddingHorizontal: Spacing.three,
    paddingBottom: BottomTabInset + Spacing.four,
  },
  safeAreaFull: {
    flex: 1,
  },
  content: {
    gap: Spacing.three,
    paddingTop: Spacing.three,
  },
  contentFull: {
    flex: 1,
  },
});
