// Typography tokens for DocVault (Arabic + English bilingual support)
// Consumed by both /web and /mobile applications

export const typography = {
  // Font families
  fontFamily: {
    // English fonts
    sans: {
      regular: 'System',
      medium: 'System',
      semibold: 'System',
      bold: 'System',
    },
    // Arabic fonts (right-to-left)
    arabic: {
      regular: 'System',
      medium: 'System',
      semibold: 'System',
      bold: 'System',
    },
    // Monospace for code/data
    mono: {
      regular: 'System',
    },
  },

  // Font sizes
  fontSize: {
    xs: 12,
    sm: 14,
    base: 16,
    lg: 18,
    xl: 20,
    '2xl': 24,
    '3xl': 30,
    '4xl': 36,
    '5xl': 48,
  },

  // Font weights
  fontWeight: {
    regular: '400' as const,
    medium: '500' as const,
    semibold: '600' as const,
    bold: '700' as const,
  },

  // Line heights
  lineHeight: {
    tight: 1.25,
    normal: 1.5,
    relaxed: 1.75,
  },

  // Letter spacing
  letterSpacing: {
    tighter: -0.8,
    tight: -0.4,
    normal: 0,
    wide: 0.4,
    wider: 0.8,
  },

  // Common text styles
  styles: {
    h1: {
      fontSize: 36,
      fontWeight: '700' as const,
      lineHeight: 1.2,
      letterSpacing: -0.8,
    },
    h2: {
      fontSize: 30,
      fontWeight: '700' as const,
      lineHeight: 1.25,
      letterSpacing: -0.4,
    },
    h3: {
      fontSize: 24,
      fontWeight: '600' as const,
      lineHeight: 1.3,
      letterSpacing: -0.2,
    },
    h4: {
      fontSize: 20,
      fontWeight: '600' as const,
      lineHeight: 1.35,
    },
    body: {
      fontSize: 16,
      fontWeight: '400' as const,
      lineHeight: 1.5,
    },
    bodySmall: {
      fontSize: 14,
      fontWeight: '400' as const,
      lineHeight: 1.5,
    },
    caption: {
      fontSize: 12,
      fontWeight: '400' as const,
      lineHeight: 1.4,
    },
    button: {
      fontSize: 16,
      fontWeight: '600' as const,
      lineHeight: 1,
    },
    buttonSmall: {
      fontSize: 14,
      fontWeight: '600' as const,
      lineHeight: 1,
    },
  },
} as const;

export type Typography = typeof typography;
