/**
 * Learn more about light and dark modes:
 * https://docs.expo.dev/guides/color-schemes/
 */

import { useThemeColor } from 'heroui-native';

export function useTheme() {
  const [background, foreground, surface, accent, muted, border, danger, warning, success] =
    useThemeColor([
      'background',
      'foreground',
      'surface',
      'accent',
      'muted',
      'border',
      'danger',
      'warning',
      'success',
    ]);

  return {
    background,
    foreground,
    surface,
    accent,
    muted,
    border,
    error: danger,
    errorMuted: `${danger}20`,
    warning,
    warningMuted: `${warning}20`,
    success,
    successMuted: `${success}20`,
    text: foreground,
    textSecondary: muted,
  };
}
