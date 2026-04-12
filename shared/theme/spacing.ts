// Spacing scale tokens for DocVault
// Consumed by both /web and /mobile applications

export const spacing = {
  // Base spacing scale (4px base unit) - using string keys for numeric values
  '0': 0,
  '0.5': 2,
  '1': 4,
  '2': 8,
  '3': 12,
  '4': 16,
  '5': 20,
  '6': 24,
  '8': 32,
  '10': 40,
  '12': 48,
  '16': 64,
  '20': 80,
  '24': 96,

  // Semantic spacing names
  xs: 4,    // Extra small
  sm: 8,    // Small
  md: 16,   // Medium (default)
  lg: 24,   // Large
  xl: 32,   // Extra large
  '2xl': 48,  // 2x extra large
  '3xl': 64,  // 3x extra large

  // Component-specific spacing
  buttonPadding: {
    vertical: 16,
    horizontal: 24,
  },
  inputPadding: {
    vertical: 12,
    horizontal: 16,
  },
  cardPadding: 16,
  cardMargin: 16,
  screenPadding: 24,
  listItemPadding: 12,
  sectionGap: 24,

  // Border radius
  radius: {
    none: 0,
    sm: 4,
    md: 8,
    lg: 12,
    xl: 16,
    '2xl': 24,
    full: 9999,
  },
} as const;

export type Spacing = typeof spacing;
