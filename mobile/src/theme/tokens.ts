// Design tokens for DocVault Mobile
// These tokens mirror the shared theme from /shared/theme
// The shared/theme is the source of truth

// Import shared theme tokens (resolved via metro.config.js watchFolders)
import { colors as sharedColors, typography as sharedTypography, spacing as sharedSpacing, shadows as sharedShadows } from '../../../shared/theme';

// Re-export shared tokens for mobile consumption
export const colors = sharedColors;
export const typography = sharedTypography;
export const spacing = sharedSpacing;
export const shadows = sharedShadows;

// Mobile-specific token aliases (consuming shared theme structure)
export const tokens = {
  backgroundColor: {
    primary: '#1a1a2e',
    secondary: '#0f0f1a',
    tertiary: '#16213e',
  },
  textColor: {
    primary: '#ffffff',
    secondary: '#9ca3af',
    muted: '#6b7280',
  },
  accentColor: '#4f46e5',
  errorColor: '#ef4444',
  screenPadding: 24,
  cardPadding: 16,
  cardRadius: 12,
  cardShadow: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
} as const;

export type Tokens = typeof tokens;
