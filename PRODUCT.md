# Product

## Users

Knowledge workers handling critical documents in mixed Arabic and English environments: legal, accounting, administration, freelance consultants, and small operations teams in the GCC. They care about losing nothing, finding things fast, and not being embarrassed by an expired visa or a missed contract renewal. They are not technical power users. They open the app three to seven times a week, scan, sign, and close it.

## Product purpose

A calm, bilingual workspace for every document that matters: contracts, invoices, IDs, warranties, receipts, certifications. The app accepts a scan or upload, reads it with OCR, extracts dates and entities, translates between Arabic and English, surfaces what is expiring soon, and answers questions about the document's contents. Tenant-scoped so teams and freelancers keep their work private.

## Brand voice

Quiet, archival, considered. A reference: a well-run records office that respects paper, mixed with a contemporary bilingual editorial. The app should feel like a careful librarian, not a hyped SaaS dashboard.

Three voice words: **calm, precise, bilingual**.

## Tone

- **Voice (always):** composed, dry-warm, never salesy. The app does not congratulate the user for opening it.
- **Tone (shifts by moment):** reassuring on errors, brief on success, plain on empty states, deliberate on confirmations. No exclamation marks in product UI. Marketing copy is allowed one or two per page, max.

## Register

Brand for the public landing (`/`, marketing surfaces). Product for everything authenticated (`/dashboard`, `/documents`, `/search`, `/reminders`, `/settings`, `/admin`).

## Strategic principles

1. **Bilingual parity is structural, not decorative.** Arabic and English are first-class. The Arabic script is a design feature, not an afterthought.
2. **Document first.** The artifact matters more than the metadata. Every screen earns its place by serving the document.
3. **OCR is the differentiator.** Search, dates, and chat are downstream of the OCR pipeline. Show its confidence; do not hide it.
4. **Tenant scoping is non-negotiable.** Documents are private by default. The product never suggests sharing; it allows it deliberately.
5. **Calm over clever.** No confetti, no playful empty states, no gamification. Empty states teach; they do not entertain.
6. **Honest placeholders.** When real data is not available, the placeholder is clearly a placeholder, not a fabricated testimonial.

## Anti-references

- Notion-cream minimalism (too soft, hides the document's gravity).
- Linear-stark dark (too cold, alienates the legal/finance user).
- Stripe purple-on-white (overused; not distinctive in 2026).
- Display-serif editorial-typographic landing pages with italic drop caps and ruled three-column layouts (saturated by 2026).
- Stock photo "team high-five" hero imagery (dishonest, template-feeling).
- Emoji as icon system.
- Marketing copy that promises "AI magic."

## Brand permissions

- Asymmetric editorial layouts, but **not** magazine aesthetics.
- A single committed accent color (deep ink-blue) carrying primary actions.
- A second accent (date-amber) reserved for reminders, dates, expiry signals.
- Real monogram avatars in testimonials (no fake stock faces).
- Generous vertical rhythm with deliberate contrast between tight groupings and open space.
- Fluid `clamp()` type only on the landing hero. Product UI is fixed `rem`.

## Product permissions

- Standard shadcn-style components from `components/ui/`. Reuse them.
- Sidebar app shell, top header, system-aware theme (light + dark).
- Density is fine; tables and lists may run wide.
- `lucide-react` for icons. Do not introduce a second icon set.
- No motion that does not convey state. No hero choreography in the dashboard.
