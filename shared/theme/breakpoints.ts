// Responsive breakpoint tokens for DocVault
// Consumed by both /web and /mobile applications

export const breakpoints = {
  // Base breakpoints (matching Tailwind defaults)
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,

  // Mobile-first ranges
  mobile: {
    min: 0,
    max: 639,
  },
  tablet: {
    min: 640,
    max: 1023,
  },
  desktop: {
    min: 1024,
    max: Infinity,
  },

  // Orientation
  portrait: 'portrait',
  landscape: 'landscape',
} as const;

export type Breakpoints = typeof breakpoints;
