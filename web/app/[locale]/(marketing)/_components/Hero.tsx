"use client";

import { useTranslations } from "next-intl";
import { CheckCircle2, Calendar, ShieldCheck } from "lucide-react";

export function Hero() {
  const t = useTranslations("landing.hero");

  return (
    <section
      aria-labelledby="hero-title"
      className="relative overflow-hidden bg-background"
    >
      <div className="grain absolute inset-0 -z-10" aria-hidden="true" />
      <div className="container-wide py-20 sm:py-24 lg:py-32">
        <div className="mx-auto flex max-w-3xl flex-col items-start">
          <p className="kicker mb-6 reveal reveal-1">{t("eyebrow")}</p>
          <h1
            id="hero-title"
            className="text-display text-balance text-foreground reveal reveal-2"
          >
            {t("title")}
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-relaxed text-muted-foreground reveal reveal-3">
            {t("subtitle")}
          </p>
          <div className="mt-10 flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-4 reveal reveal-4">
            <a
              href="#pricing"
              className="inline-flex items-center justify-center gap-2 rounded-md bg-primary px-5 py-3 text-sm font-semibold text-primary-foreground shadow-sm transition-all hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            >
              {t("primaryCta")}
            </a>
            <a
              href="#how-it-works"
              className="inline-flex items-center justify-center gap-2 rounded-md border border-border bg-background px-5 py-3 text-sm font-medium text-foreground transition-all hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            >
              {t("secondaryCta")} →
            </a>
          </div>
          <p className="mt-3 text-xs text-muted-foreground reveal reveal-5">
            {t("primaryCtaHint")}
          </p>

          <div className="mt-12 flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-muted-foreground reveal reveal-5">
            <span className="inline-flex items-center gap-1.5">
              <CheckCircle2 className="h-3.5 w-3.5 text-primary" aria-hidden="true" />
              <span>OCR · EN + AR</span>
            </span>
            <span className="inline-flex items-center gap-1.5">
              <ShieldCheck className="h-3.5 w-3.5 text-primary" aria-hidden="true" />
              <span>Tenant-scoped</span>
            </span>
            <span className="inline-flex items-center gap-1.5">
              <Calendar className="h-3.5 w-3.5 text-amber" aria-hidden="true" />
              <span>Expiry aware</span>
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}
