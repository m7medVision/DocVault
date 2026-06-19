# web/AGENTS.md — Next.js app

Part of the docvault monorepo. Read the root **[../AGENTS.md](../AGENTS.md)** first (commit
identity, conventions). Brand/UX truth: **[../DESIGN.md](../DESIGN.md)** +
**[../PRODUCT.md](../PRODUCT.md)** — follow them, including the **bans**.

## Stack

Next.js 15 (App Router) · React 19 · TypeScript · **Tailwind v4** · **next-intl** (en/ar,
RTL) · TanStack Query + `@tanstack/ai-react` (chat streaming) · **Radix** (`radix-ui`) ·
`react-markdown` + `remark-gfm` · **lucide-react**. Package `docvault-web`. **pnpm** for
deps, `npm run` for scripts.

## Gate

```sh
cd web && npx tsc --noEmit      # THE web gate (exit 0)
```

> ⚠️ `npm run lint` is **broken at baseline** (eslint 9 ↔ `eslint-config-next` patch
> incompatibility) — do **not** use it as a gate. Use `tsc`.

## Layout

```
app/[locale]/(marketing|app|auth)/   route groups (en unprefixed, ar under /ar)
app/api/*/route.ts                   proxy routes to the Go backend (auth/cookies)
app/globals.css                      OKLCH design tokens (light+dark)
app/layout.tsx                       fonts (Noto Sans + Noto Sans Arabic), <html dir/lang>
components/                          feature components
components/ui/                       shadcn-style primitives (reuse these)
components/layout/                   shell (AppSidebar, NotificationBell, …)
features/                            hooks wrapping lib/api
lib/api/                             typed API clients
lib/utils.ts                         cn() = clsx + tailwind-merge
messages/{en,ar}.json               i18n catalogs
```

## Design system & shadcn — how we structure UI

- **Reuse `components/ui/*`** — the installed primitives: `avatar badge button card checkbox
  collapsible dialog dropdown-menu input label popover progress select separator sheet sidebar
  skeleton table tabs textarea tooltip`. **Do not** reinvent these or add a second component
  library. lucide-react is the **only** icon set (single weight; never emoji as icons).
- **Component pattern:** Radix primitive + **`class-variance-authority` (`cva`)** for variants
  + **`cn()`** (`lib/utils.ts`, `clsx`+`tailwind-merge`) to compose classes. To add a missing
  primitive, follow this exact pattern (or `npx shadcn@latest add …`, then align it to our
  tokens) — match `components/ui/input.tsx` / `button.tsx`.
- **Tokens, never raw colors.** All color/spacing comes from the **OKLCH CSS variables** in
  `app/globals.css` (light + dark themes): warm-paper `--background`, ink `--foreground`,
  **deep ink-blue `--primary`**, and **`--brand-amber`** reserved for dates / reminders /
  expiry signals (teach the eye to scan for amber). Use semantic utilities
  (`bg-background text-foreground border-border bg-primary …`), never hex.
- **Typography:** one family pair — **Noto Sans** (Latin) + **Noto Sans Arabic** — set in
  `app/layout.tsx`. Hierarchy from size/weight, not new families. Markdown/prose via
  `@tailwindcss/typography` (`prose prose-sm dark:prose-invert …`; see `ChatPanel.tsx`).
- **Spacing/elevation/motion:** 4pt scale; separate surfaces with background tint not shadow;
  motion only conveys state. See [../DESIGN.md](../DESIGN.md) for the full scale and the
  **bans** (no glassmorphism / `backdrop-blur`, no `background-clip:text` gradient text, no
  `border-l-4` side stripes, no em dashes in copy, no "AI magic"/"magic" marketing).

## i18n / RTL

- All user-facing strings live in `messages/en.json` **and** `messages/ar.json` —
  `eslint-plugin-i18next` forbids hardcoded copy. Arabic must be natural and calm (no
  exclamation marks). Add keys to **both** files; keep JSON valid.
- `localePrefix: 'as-needed'` (English unprefixed, Arabic `/ar/...`). RTL is driven by `dir`
  on `<html>`; use **logical** Tailwind utilities (`ms-/me-`, `ps-/pe-`, `start/end`) and
  `rtl:` variants, and flip directional icons/side props for `ar`.

## Talking to the backend

The web calls the Go API through `app/api/*` proxy route handlers (forwarding the session
cookie/bearer) or a server base URL — match the existing `lib/api/*` + proxy patterns; do not
invent a new transport. Chat uses SSE (`@tanstack/ai-react` + the `/api/chat` proxy).
