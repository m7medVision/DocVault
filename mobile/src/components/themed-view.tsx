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

  const colorValues = { background, surface, accent, muted, border, danger, warning, success };
  const backgroundColor = type ? (colorValues[type as keyof typeof colorValues] ?? background) : background;

  return <View style={[{ backgroundColor }, style]} {...otherProps} />;
}
