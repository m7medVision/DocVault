import { useTranslations } from "next-intl";
import { ScanText, CalendarClock, BellRing, Search, MessagesSquare, Lock } from "lucide-react";

const ICONS: Record<string, typeof ScanText> = {
  ocr: ScanText,
  dates: CalendarClock,
  reminders: BellRing,
  search: Search,
  chat: MessagesSquare,
  tenancy: Lock,
};

const SLUGS = ["ocr", "dates", "reminders", "search", "chat", "tenancy"] as const;

export function Features() {
  const t = useTranslations("landing.features");

  return (
    <section
      id="features"
      aria-labelledby="features-title"
      className="border-t border-border bg-secondary/30"
    >
      <div className="container-wide py-20 sm:py-28">
        <div className="max-w-3xl">
          <p className="kicker mb-4">{t("kicker")}</p>
          <h2
            id="features-title"
            className="text-balance text-3xl font-semibold tracking-tight text-foreground sm:text-4xl"
          >
            {t("title")}
          </h2>
          <p className="mt-4 max-w-2xl text-base leading-relaxed text-muted-foreground sm:text-lg">
            {t("subtitle")}
          </p>
        </div>

        {/* Asymmetric tile grid: not a uniform 3x2. Wide row, then a mixed row. */}
        <div className="mt-14 grid gap-4 sm:gap-6 md:grid-cols-3 lg:gap-8">
          {/* Wide tile: OCR */}
          <FeatureCard
            slug="ocr"
            className="md:col-span-2"
            variant="wide"
          />

          {/* Compact tile: Dates */}
          <FeatureCard slug="dates" className="md:col-span-1" />

          {/* Tall tile: Search */}
          <FeatureCard
            slug="search"
            className="md:col-span-1"
            variant="tall"
          />

          {/* Standard: Reminders */}
          <FeatureCard slug="reminders" className="md:col-span-1" />

          {/* Standard: Chat */}
          <FeatureCard slug="chat" className="md:col-span-1" />
        </div>

        {/* Pull-quote tile spans full width on the second band */}
        <div className="mt-4 sm:mt-6 lg:mt-8">
          <TenancyBanner slug="tenancy" />
        </div>
      </div>
    </section>
  );
}

function FeatureCard({
  slug,
  className = "",
  variant = "default",
}: {
  slug: (typeof SLUGS)[number];
  className?: string;
  variant?: "default" | "wide" | "tall";
}) {
  const t = useTranslations(`landing.features.items.${slug}`);
  const Icon = ICONS[slug];

  const baseSurface =
    "group relative flex h-full flex-col rounded-2xl border border-border bg-card p-6 transition-colors hover:border-foreground/20 sm:p-8";
  const variantSurface =
    variant === "wide"
      ? "sm:p-10"
      : variant === "tall"
      ? "min-h-[18rem]"
      : "";

  return (
    <div className={`${baseSurface} ${variantSurface} ${className}`}>
      <div className="mb-5 inline-flex h-10 w-10 items-center justify-center rounded-md bg-primary/10 text-primary">
        <Icon className="h-5 w-5" aria-hidden="true" />
      </div>
      <h3 className="text-xl font-semibold tracking-tight text-foreground">
        {t("title")}
      </h3>
      <p className="mt-2 text-sm leading-relaxed text-muted-foreground sm:text-base">
        {t("body")}
      </p>

      {variant === "wide" ? <OcrInline /> : null}
      {variant === "tall" ? <SearchInline /> : null}
    </div>
  );
}

function OcrInline() {
  return (
    <div className="mt-6 grid gap-3 sm:grid-cols-2">
      <div className="rounded-md border border-border bg-secondary/60 p-4">
        <p className="kicker mb-2">AR</p>
        <p className="text-sm leading-relaxed" dir="rtl">
          العقد ساري لمدة اثني عشر شهرًا من تاريخ التوقيع.
        </p>
      </div>
      <div className="rounded-md border border-border bg-secondary/60 p-4">
        <p className="kicker mb-2">EN</p>
        <p className="text-sm leading-relaxed">
          The agreement is valid for twelve months from signing.
        </p>
      </div>
    </div>
  );
}

function SearchInline() {
  return (
    <div className="mt-6 space-y-2">
      <div className="flex items-center justify-between rounded-md border border-border bg-secondary/60 px-3 py-2 text-xs">
        <span className="text-muted-foreground">grace period</span>
        <span className="tabular-nums text-foreground">3 matches</span>
      </div>
      <div className="flex items-center justify-between rounded-md border border-border bg-secondary/60 px-3 py-2 text-xs">
        <span className="text-muted-foreground">مدة السماح</span>
        <span className="tabular-nums text-foreground">٢ نتيجة</span>
      </div>
      <div className="flex items-center justify-between rounded-md border border-border bg-secondary/60 px-3 py-2 text-xs">
        <span className="text-muted-foreground">grace period</span>
        <span className="tabular-nums text-foreground">1 match</span>
      </div>
    </div>
  );
}

function TenancyBanner({ slug }: { slug: "tenancy" }) {
  const t = useTranslations(`landing.features.items.${slug}`);
  const Icon = ICONS[slug];
  return (
    <div className="flex flex-col items-start gap-6 rounded-2xl border border-border bg-foreground p-8 text-background sm:p-10 md:flex-row md:items-center md:gap-10">
      <div className="inline-flex h-12 w-12 shrink-0 items-center justify-center rounded-md bg-background/10">
        <Icon className="h-6 w-6" aria-hidden="true" />
      </div>
      <div className="flex-1">
        <h3 className="text-xl font-semibold tracking-tight sm:text-2xl">
          {t("title")}
        </h3>
        <p className="mt-2 max-w-2xl text-sm leading-relaxed text-background/70 sm:text-base">
          {t("body")}
        </p>
      </div>
    </div>
  );
}
