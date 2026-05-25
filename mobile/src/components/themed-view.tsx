import { View, type ViewProps } from 'react-native';

import { useThemeColor } from 'heroui-native';
import { ThemeColor } from '@/constants/theme';

export type ThemedViewProps = ViewProps & {
  lightColor?: string;
  darkColor?: string;
  type?: ThemeColor;
};

export function ThemedView({ style, type, ...otherProps }: ThemedViewProps) {
  const [background, surface, accent, muted, border, danger, warning, success] = useThemeColor([
    'background',
    'surface',
    'accent',
    'muted',
    'border',
    'danger',
    'warning',
    'success',
  ]);

  const colorValues = {
    background,
    surface,
    accent,
    muted,
    border,
    danger,
    warning,
    success,
    foreground: muted,
    textSecondary: muted,
    error: danger,
    errorMuted: `${danger}20`,
    warningMuted: `${warning}20`,
    successMuted: `${success}20`,
    backgroundElement: surface,
    backgroundSelected: muted,
  };
  const backgroundColor = type ? (colorValues[type as keyof typeof colorValues] ?? background) : background;

  return <View style={[{ backgroundColor }, style]} {...otherProps} />;
}
