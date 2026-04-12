// Shared theme exports for DocVault
// Consumed by both /web and /mobile applications

export { colors } from './colors';
export type { Colors } from './colors';

export { typography } from './typography';
export type { Typography } from './typography';

export { spacing } from './spacing';
export type { Spacing } from './spacing';

export { shadows } from './shadows';
export type { Shadows } from './shadows';

export { breakpoints } from './breakpoints';
export type { Breakpoints } from './breakpoints';

// Combined theme object
import { colors } from './colors';
import { typography } from './typography';
import { spacing } from './spacing';
import { shadows } from './shadows';
import { breakpoints } from './breakpoints';

export const theme = {
  colors,
  typography,
  spacing,
  shadows,
  breakpoints,
} as const;

export type Theme = typeof theme;
