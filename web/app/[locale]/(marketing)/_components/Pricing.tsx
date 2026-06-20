"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { Check, Mail } from "lucide-react";
import { AnimatePresence, motion, useReducedMotion } from "motion/react";

import { Reveal } from "@/components/ui/reveal";

const PLAN_KEYS = ["starter", "practice", "office"] as const;

export function Pricing() {
  const t = useTranslations("landing.pricing");
  const [yearly, setYearly] = useState(false);

  return (
    <section
      id="pricing"
      aria-labelledby="pricing-title"
      className="border-t border-border bg-secondary/30"
    >
      <div className="container-wide py-20 sm:py-28">
        <Reveal>
          <div className="flex flex-col gap-8 lg:flex-row lg:items-end lg:justify-between">
            <div className="max-w-2xl">
              <p className="kicker mb-4">{t("kicker")}</p>
              <h2
                id="pricing-title"
                className="text-balance text-3xl font-semibold tracking-tight text-foreground sm:text-4xl"
              >
                {t("title")}
              </h2>
              <p className="mt-4 max-w-xl text-base leading-relaxed text-muted-foreground sm:text-lg">
                {t("subtitle")}
              </p>
            </div>

            <BillingToggle yearly={yearly} onChange={setYearly} />
          </div>
        </Reveal>

        <div className="mt-12 grid gap-4 lg:grid-cols-3 lg:gap-6">
          {PLAN_KEYS.map((key, idx) => (
            <PlanCard
              key={key}
              slug={key}
              yearly={yearly}
              highlighted={key === "practice"}
              delay={idx * 0.08}
            />
          ))}
        </div>

        <Reveal delay={0.1}>
          <p className="mt-8 max-w-3xl text-sm text-muted-foreground">
            {t("footnote")}
          </p>

          <div className="mt-12 flex flex-col items-start gap-3 rounded-2xl border border-border bg-card p-6 sm:flex-row sm:items-center sm:justify-between sm:p-8">
            <div>
              <p className="text-lg font-semibold tracking-tight text-foreground">
                {t("cta")}
              </p>
              <p className="text-sm text-muted-foreground">{t("ctaHint")}</p>
            </div>
            <a
              href="mailto:hello@docvault.app?subject=DocVault%20plan%20enquiry"
              className="inline-flex items-center gap-2 rounded-md bg-primary px-5 py-3 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            >
              <Mail className="h-4 w-4" aria-hidden="true" />
              hello@docvault.app
            </a>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

function BillingToggle({
  yearly,
  onChange,
}: {
  yearly: boolean;
  onChange: (v: boolean) => void;
}) {
  const t = useTranslations("landing.pricing");
  return (
    <div
      role="radiogroup"
      aria-label="Billing period"
      className="inline-flex items-center rounded-full border border-border bg-card p-1 text-sm"
    >
      <button
        type="button"
        role="radio"
        aria-checked={!yearly}
        onClick={() => onChange(false)}
        className={`rounded-full px-4 py-1.5 transition-colors ${
          !yearly
            ? "bg-foreground text-background"
            : "text-muted-foreground hover:text-foreground"
        }`}
      >
        {t("monthly")}
      </button>
      <button
        type="button"
        role="radio"
        aria-checked={yearly}
        onClick={() => onChange(true)}
        className={`flex items-center gap-2 rounded-full px-4 py-1.5 transition-colors ${
          yearly
            ? "bg-foreground text-background"
            : "text-muted-foreground hover:text-foreground"
        }`}
      >
        {t("yearly")}
        <span
          className={`rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider ${
            yearly
              ? "bg-amber text-amber-foreground"
              : "bg-amber-soft text-amber-700"
          }`}
        >
          {t("saveBadge")}
        </span>
      </button>
    </div>
  );
}

function PlanCard({
  slug,
  yearly,
  highlighted,
  delay = 0,
}: {
  slug: (typeof PLAN_KEYS)[number];
  yearly: boolean;
  highlighted: boolean;
  delay?: number;
}) {
  const t = useTranslations(`landing.pricing.plans`);
  const tCommon = useTranslations("landing.pricing");
  const plan = t.raw(slug) as {
    name: string;
    tagline: string;
    price: string;
    priceSuffix: string;
    features: string[];
    badge?: string;
  };

  const isCustom = plan.price === tCommon("custom");
  const reduce = useReducedMotion();

  return (
    <Reveal
      delay={delay}
      className={`relative flex h-full flex-col rounded-2xl border bg-card p-6 sm:p-8 ${
        highlighted
          ? "border-amber shadow-[0_20px_50px_-25px_oklch(62%_0.12_65_/_0.5)]"
          : "border-border"
      }`}
    >
      {highlighted && plan.badge ? (
        <span className="absolute -top-3 start-6 inline-flex items-center gap-1 rounded-full bg-amber px-3 py-1 text-[11px] font-semibold uppercase tracking-wider text-amber-foreground">
          {plan.badge}
        </span>
      ) : null}

      <h3 className="text-xl font-semibold tracking-tight text-foreground">
        {plan.name}
      </h3>
      <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
        {plan.tagline}
      </p>

      <div className="mt-6 flex items-baseline gap-1.5">
        {isCustom ? (
          <span className="text-3xl font-semibold tracking-tight text-foreground">
            {plan.price}
          </span>
        ) : (
          <>
            <span className="text-4xl font-semibold tracking-tight text-foreground tabular-nums">
              {plan.price}
            </span>
            <span className="text-sm font-medium text-muted-foreground">
              {plan.priceSuffix}
            </span>
            <span className="ms-2 text-sm text-muted-foreground">
              {reduce ? (
                <>{yearly ? tCommon("perYear") : tCommon("perMonth")}</>
              ) : (
                <AnimatePresence mode="wait" initial={false}>
                  <motion.span
                    key={yearly ? "yearly" : "monthly"}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.15, ease: "easeOut" }}
                    className="inline-block"
                  >
                    {yearly ? tCommon("perYear") : tCommon("perMonth")}
                  </motion.span>
                </AnimatePresence>
              )}
            </span>
          </>
        )}
      </div>

      <ul className="mt-8 space-y-3 text-sm">
        {plan.features.map((feature) => (
          <li key={feature} className="flex items-start gap-3 text-foreground/90">
            <Check
              className="mt-0.5 h-4 w-4 shrink-0 text-primary"
              aria-hidden="true"
            />
            <span className="leading-relaxed">{feature}</span>
          </li>
        ))}
      </ul>

      <div className="mt-auto pt-8">
        <a
          href={`mailto:hello@docvault.app?subject=DocVault%20plan%20enquiry%20—%20${encodeURIComponent(plan.name)}`}
          className={`inline-flex w-full items-center justify-center gap-2 rounded-md px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background ${
            highlighted
              ? "bg-primary text-primary-foreground hover:bg-primary/90"
              : "border border-border bg-background text-foreground hover:bg-accent"
          }`}
        >
          {tCommon("cta")}
        </a>
      </div>
    </Reveal>
  );
}
