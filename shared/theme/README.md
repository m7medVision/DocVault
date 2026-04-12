# Shared Theme

Shared design tokens consumed by both `/web` and `/mobile` applications.

## Contents

- `colors.ts` — Color palette (light + dark theme)
- `typography.ts` — Font families, sizes, weights (Arabic + English)
- `spacing.ts` — Spacing scale (margins, paddings)
- `shadows.ts` — Shadow/elevation tokens
- `breakpoints.ts` — Responsive breakpoint values

## Usage

```typescript
import { colors, typography } from '@docvault/theme';
```

## Building

```bash
cd shared/theme
npm install
npm run build
```
