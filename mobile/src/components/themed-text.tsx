import { Platform, StyleSheet, Text, type TextProps } from 'react-native';

import { Fonts, ThemeColor } from '@/constants/theme';
import { useThemeColor } from 'heroui-native';

export type ThemedTextProps = TextProps & {
  type?: 'default' | 'title' | 'small' | 'smallBold' | 'subtitle' | 'link' | 'linkPrimary' | 'code';
  themeColor?: ThemeColor;
};

export function ThemedText({ style, type = 'default', themeColor, ...rest }: ThemedTextProps) {
  const [text, accent, background, surface, muted, border, danger, warning, success] = useThemeColor([
    'foreground',
    'accent',
    'background',
    'surface',
    'muted',
    'border',
    'danger',
    'warning',
    'success',
  ]);

  const colorValues = {
    foreground: text,
    accent,
    background,
    surface,
    muted,
    border,
    danger,
    warning,
    success,
    textSecondary: muted,
    error: danger,
    errorMuted: `${danger}20`,
    warningMuted: `${warning}20`,
    successMuted: `${success}20`,
    backgroundElement: surface,
    backgroundSelected: accent,
  };
  const mappedColor = themeColor ? colorValues[themeColor as keyof typeof colorValues] ?? text : text;

  return (
    <Text
      style={[
        { color: mappedColor },
        type === 'default' && styles.default,
        type === 'title' && styles.title,
        type === 'small' && styles.small,
        type === 'smallBold' && styles.smallBold,
        type === 'subtitle' && styles.subtitle,
        type === 'link' && styles.link,
        type === 'linkPrimary' && { color: accent },
        type === 'code' && styles.code,
        style,
      ]}
      {...rest}
    />
  );
}

const styles = StyleSheet.create({
  small: {
    fontSize: 14,
    lineHeight: 20,
    fontWeight: 500,
  },
  smallBold: {
    fontSize: 14,
    lineHeight: 20,
    fontWeight: 700,
  },
  default: {
    fontSize: 16,
    lineHeight: 24,
    fontWeight: 500,
  },
  title: {
    fontSize: 48,
    fontWeight: 600,
    lineHeight: 52,
  },
  subtitle: {
    fontSize: 32,
    lineHeight: 44,
    fontWeight: 600,
  },
  link: {
    lineHeight: 30,
    fontSize: 14,
  },
  code: {
    fontFamily: Fonts.mono,
    fontWeight: Platform.select({ android: 700 }) ?? 500,
    fontSize: 12,
  },
});
