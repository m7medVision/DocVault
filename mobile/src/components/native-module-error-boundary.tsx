import { Component, type ErrorInfo, type ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';

interface Props {
  children: ReactNode;
  fallbackTitle?: string;
}

interface State {
  error: Error | null;
}

export class NativeModuleErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    if (__DEV__) {
      console.warn('[NativeModuleErrorBoundary]', error.message, info.componentStack);
    }
  }

  reset = () => this.setState({ error: null });

  render() {
    if (this.state.error) {
      return <NativeModuleFallback error={this.state.error} onRetry={this.reset} title={this.props.fallbackTitle} />;
    }
    return this.props.children;
  }
}

function NativeModuleFallback({
  error,
  onRetry,
  title,
}: {
  error: Error;
  onRetry: () => void;
  title?: string;
}) {
  const theme = useTheme();
  const looksLikeMissingNativeModule =
    error.message.includes('ViewManager') ||
    error.message.includes('requireNativeComponent') ||
    error.message.includes('KJExpoPdf') ||
    error.message.includes('Galeria');

  return (
    <View style={[styles.container, { backgroundColor: theme.surface }]}>
      <ThemedText type="smallBold" themeColor="danger">
        {title ?? 'Viewer unavailable'}
      </ThemedText>
      <ThemedText type="code" themeColor="textSecondary">
        {looksLikeMissingNativeModule
          ? 'The native module is not registered in the dev client. Run `npx expo prebuild` and `npx expo run:ios` (or `:android`) to rebuild.'
          : error.message}
      </ThemedText>
      <ThemedText
        type="code"
        themeColor="accent"
        onPress={onRetry}
        accessibilityRole="button"
      >
        Retry
      </ThemedText>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: Spacing.two,
    padding: Spacing.four,
    borderRadius: Spacing.two,
  },
});
