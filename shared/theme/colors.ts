// Color palette tokens for DocVault (light + dark theme support)
// Consumed by both /web and /mobile applications

export const colors = {
  // Primary brand colors
  primary: {
    50: '#eef2ff',
    100: '#e0e7ff',
    200: '#c7d2fe',
    300: '#a5b4fc',
    400: '#818cf8',
    500: '#6366f1', // Main primary (indigo)
    600: '#4f46e5', // Primary action
    700: '#4338ca',
    800: '#3730a3',
    900: '#312e81',
  },

  // Neutral grays
  gray: {
    50: '#f9fafb',
    100: '#f3f4f6',
    200: '#e5e7eb',
    300: '#d1d5db',
    400: '#9ca3af', // Placeholder text
    500: '#6b7280', // Muted text
    600: '#4b5563',
    700: '#374151',
    800: '#1f2937',
    900: '#111827',
  },

  // Semantic colors
  success: {
    50: '#ecfdf5',
    100: '#d1fae5',
    200: '#a7f3d0',
    300: '#6ee7b7',
    400: '#34d399',
    500: '#10b981',
    600: '#059669',
    700: '#047857',
  },

  warning: {
    50: '#fffbeb',
    100: '#fef3c7',
    200: '#fde68a',
    300: '#fcd34d',
    400: '#fbbf24',
    500: '#f59e0b',
    600: '#d97706',
    700: '#b45309',
  },

  error: {
    50: '#fef2f2',
    100: '#fee2e2',
    200: '#fecaca',
    300: '#fca5a5',
    400: '#f87171',
    500: '#ef4444', // Error state
    600: '#dc2626',
    700: '#b91c1c',
  },

  // Dark theme backgrounds
  dark: {
    bg: {
      primary: '#1a1a2e',   // Main dark background
      secondary: '#0f0f1a', // Darker surface
      tertiary: '#16213e',  // Card backgrounds
    },
    text: {
      primary: '#ffffff',
      secondary: '#9ca3af',
      muted: '#6b7280',
    },
  },

  // Light theme backgrounds
  light: {
    bg: {
      primary: '#ffffff',
      secondary: '#f9fafb',
      tertiary: '#f3f4f6',
    },
    text: {
      primary: '#111827',
      secondary: '#4b5563',
      muted: '#9ca3af',
    },
  },

  // Transparent overlays
  overlay: {
    light: 'rgba(255, 255, 255, 0.8)',
    dark: 'rgba(0, 0, 0, 0.5)',
    primary: 'rgba(99, 102, 241, 0.1)', // 10% primary
  },
} as const;

export type Colors = typeof colors;
