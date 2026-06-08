# Design

## Direction

**Editorial Bilingual.** Warm paper background, deep ink-blue accent, date-amber secondary. Arabic and English treated as parallel, not stacked. A reference: an academic press catalog that happens to also be a working product.

## Color strategy

**Restrained with one committed accent** on the landing. Product UI stays Restrained.

The brand picks **deep ink-blue** (`oklch(38% 0.08 250)`) as primary. Date-amber (`oklch(62% 0.12 65)`) is reserved for dates, reminders, and expiry signals so the user learns to scan for amber. Neutrals are tinted toward the brand hue (chroma 0.005–0.01), never pure gray.

### Tokens (light theme default)

| Role | OKLCH | Use |
| --- | --- | --- |
| `--background` | `oklch(98% 0.005 60)` | Page background, warm paper |
| `--foreground` | `oklch(20% 0.015 60)` | Body text, ink |
| `--card` | `oklch(100% 0 0)` | Elevated surface |
| `--card-foreground` | `oklch(20% 0.015 60)` | Card text |
| `--primary` | `oklch(38% 0.08 250)` | Brand, primary CTAs |
| `--primary-foreground` | `oklch(98% 0.005 60)` | Text on primary |
| `--secondary` | `oklch(94% 0.01 60)` | Subtle blocks |
| `--secondary-foreground` | `oklch(20% 0.015 60)` | Text on secondary |
| `--muted` | `oklch(94% 0.01 60)` | Disabled, subtle fills |
| `--muted-foreground` | `oklch(48% 0.02 60)` | Captions, helper |
| `--accent` | `oklch(94% 0.02 60)` | Hover surfaces |
| `--accent-foreground` | `oklch(20% 0.015 60)` | Text on accent |
| `--destructive` | `oklch(58% 0.20 25)` | Destructive actions |
| `--border` | `oklch(88% 0.01 60)` | Hairlines |
| `--input` | `oklch(88% 0.01 60)` | Form borders |
| `--ring` | `oklch(50% 0.10 250)` | Focus ring |
| `--brand-amber` | `oklch(62% 0.12 65)` | Dates, reminders, expiry |
| `--brand-amber-soft` | `oklch(94% 0.04 75)` | Tinted background for date chips |

Dark theme: surfaces at `oklch(15% 0.008 60)`, ink `oklch(94% 0.005 60)`, primary shifted to `oklch(72% 0.10 250)`, amber shifted to `oklch(78% 0.14 70)`.

## Typography

One family pair: **Noto Sans** for Latin, **Noto Sans Arabic** for Arabic. The Latin leg is loaded via `next/font/google` if needed; the Arabic leg is already in `app/layout.tsx`. Both are free, both support the full bilingual range, both are outside the reflex-reject list.

We use one family, two scripts. No display serif, no second weight family. Hierarchy comes from size and weight contrast inside the single family.

### Scale

| Role | rem | Weight | Use |
| --- | --- | --- | --- |
| Display | `clamp(2.5rem, 5vw + 1rem, 4.5rem)` | 700 | Landing hero h1 only |
| h1 | `1.875rem` | 700 | Page titles |
| h2 | `1.5rem` | 600 | Section headings |
| h3 | `1.25rem` | 600 | Card / sub-headings |
| Body | `1rem` (16px floor) | 400 | Default |
| Small | `0.875rem` | 400 | Helper, meta |
| Caption | `0.75rem` | 500 (tracked) | Labels, kickers, file IDs |
| Mono | `0.875rem` | 500 | File IDs, dates, code |

Line-height on body: 1.55. Headings: 1.15. Light type on dark backgrounds: add 0.05 to line-height and `letter-spacing: 0.01em` for body.

Cap body line length at `65ch` on the landing prose. Product UI is unconstrained.

## Spacing

4pt base. Use: `4 / 8 / 12 / 16 / 24 / 32 / 48 / 64 / 96`. Vary for rhythm: tight groups (`8 / 12`), section separations (`64 / 96`), and the hero gets `clamp(4rem, 8vw, 7rem)` of vertical padding.

## Elevation

Subtle. Use background tint to separate surfaces, not shadow. One soft shadow reserved for the floating mobile nav drawer. No glassmorphism.

## Layout

- **Landing:** asymmetric editorial. Hero is two-column on desktop, stacked on mobile. The document mockup is the visual anchor on the right.
- **Features section:** NOT a uniform 3×2 grid. Vary tile sizes: a wide tile, two narrow tiles, a code/console tile, a quote tile.
- **Pricing:** 3 columns, the middle one (Pro) gets the date-amber accent strip. Pricing is per-team with OMR currency.
- **Testimonials:** monogram avatars, no photos. Names and roles only.
- **FAQ:** simple disclosure using `<details>` semantics, custom-styled.
- **Dashboard:** unchanged. Existing sidebar shell.

## Iconography

`lucide-react`, single weight. Icons appear at 16px in nav, 20px in features, 24px in hero. Never as headings.

## Motion

- Landing hero: one staggered reveal on first paint (600ms, ease-out-quart). Reduced motion users get a static reveal.
- FAQ disclosure: `grid-template-rows` transition (not `height`).
- Pricing toggle: opacity crossfade.
- Dashboard: no entrance motion. Hover and focus transitions only, 150ms.

## Components reused

`components/ui/*` (button, card, badge, sidebar, table, dialog, dropdown-menu, input, label, sheet, separator, tabs, tooltip, skeleton, popover, select, checkbox, progress, avatar).

## Bans on this project

- No side-stripe borders (`border-l-4`) on cards, callouts, alerts.
- No `background-clip: text` gradient text.
- No glassmorphism (no `backdrop-blur` decoration).
- No hero-metric template on the landing.
- No identical card grids in the features section.
- No em dashes in any copy.
- No "Get started for free" or "AI-powered" or "magic" in marketing copy.
- No stock photos. No emoji. No fake testimonials presented as real.
