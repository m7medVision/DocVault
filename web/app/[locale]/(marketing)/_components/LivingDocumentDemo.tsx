"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { useAnimate, stagger, useReducedMotion } from "motion/react";
import {
  ScanText,
  CalendarClock,
  Building2,
  RefreshCw,
  Languages,
} from "lucide-react";

import { cn } from "@/lib/utils";

const wait = (ms: number) => new Promise((r) => setTimeout(r, ms));

export function LivingDocumentDemo() {
  const t = useTranslations("landing.features.demo");
  const reduce = useReducedMotion();
  const [scope, animate] = useAnimate();

  useEffect(() => {
    if (reduce) return;
    let cancelled = false;

    const loop = async () => {
      while (!cancelled) {
        // reset
        await animate(".dv-chip", { opacity: 0, x: -12 }, { duration: 0 });
        await animate(".dv-clause", { backgroundColor: "rgba(0,0,0,0)" }, { duration: 0 });
        await animate(".dv-scan", { top: "-12%", opacity: 0 }, { duration: 0 });
        if (cancelled) break;

        // scan sweep down the document
        await animate(".dv-scan", { opacity: 1 }, { duration: 0.25 });
        await animate(".dv-scan", { top: "108%" }, { duration: 2, ease: "linear" });
        await animate(".dv-scan", { opacity: 0 }, { duration: 0.25 });
        if (cancelled) break;

        // light up the key clause, then fly the extracted chips in
        await animate(
          ".dv-clause",
          { backgroundColor: "var(--brand-amber-soft)" },
          { duration: 0.4 },
        );
        await animate(
          ".dv-chip",
          { opacity: 1, x: 0 },
          { duration: 0.45, delay: stagger(0.16) },
        );

        await wait(2200);
        if (cancelled) break;

        // clear for the next pass
        await animate(".dv-clause", { backgroundColor: "rgba(0,0,0,0)" }, { duration: 0.4 });
        await animate(".dv-chip", { opacity: 0, x: -12 }, { duration: 0.3, delay: stagger(0.05) });
        await wait(500);
      }
    };

    loop();
    return () => {
      cancelled = true;
    };
  }, [reduce, animate]);

  return (
    <div
      ref={scope}
      dir="ltr"
      className="grid gap-6 lg:grid-cols-[1.2fr_1fr] lg:items-center lg:gap-10"
    >
      {/* ── The document being read ── */}
      <div className="relative overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        {/* title bar */}
        <div className="flex items-center gap-3 border-b border-border bg-secondary/40 px-5 py-3">
          <span className="inline-flex h-7 w-7 items-center justify-center rounded-md bg-primary/10 text-primary">
            <ScanText className="h-4 w-4" aria-hidden="true" />
          </span>
          <span className="text-sm font-medium text-foreground">
            lease-2026.pdf
          </span>
          <span className="ms-auto inline-flex items-center gap-1.5 text-xs text-muted-foreground">
            <span className="h-1.5 w-1.5 rounded-full bg-primary" aria-hidden="true" />
            {t("scanning")}
          </span>
        </div>

        {/* document body */}
        <div className="relative min-h-[17rem] px-6 py-6">
          {/* scan band (animated by JS; hidden by default for reduced-motion / no-JS) */}
          <div className="pointer-events-none absolute inset-0 overflow-hidden">
            <div
              className="dv-scan absolute inset-x-0 h-20 opacity-0"
              style={{ top: "-12%" }}
            >
              <div className="h-full w-full bg-gradient-to-b from-transparent via-primary/15 to-transparent" />
              <div className="h-px w-full bg-gradient-to-r from-transparent via-primary to-transparent" />
            </div>
          </div>

          <p className="text-base font-semibold text-foreground">
            عقد إيجار / Lease Agreement
          </p>

          <div className="mt-5 space-y-3">
            <Bar w="92%" />
            <Bar w="78%" />
            {/* the key clause that gets extracted */}
            <p className="dv-clause rounded-md px-2 py-1 text-sm leading-relaxed text-foreground">
              The agreement is valid for twelve months from signing
              <span dir="rtl" className="mt-1 block text-muted-foreground">
                صالح لمدة اثني عشر شهرًا من تاريخ التوقيع
              </span>
            </p>
            <Bar w="85%" />
            <Bar w="64%" />
            <Bar w="73%" />
          </div>
        </div>
      </div>

      {/* ── What DocVault pulled out ── */}
      <div>
        <p className="kicker mb-4">{t("extracted")}</p>
        <div className="space-y-3">
          <Chip
            icon={CalendarClock}
            label={t("expiry")}
            value="01 Dec 2026"
            accent
          />
          <Chip icon={Building2} label={t("parties")} value="Party A · Party B" />
          <Chip icon={RefreshCw} label={t("renewal")} value={t("renewalValue")} />
          <Chip icon={Languages} label={t("translation")} value="AR ↔ EN" />
        </div>
      </div>
    </div>
  );
}

function Bar({ w }: { w: string }) {
  return (
    <div
      aria-hidden="true"
      className="h-2.5 rounded-full bg-secondary"
      style={{ width: w }}
    />
  );
}

function Chip({
  icon: Icon,
  label,
  value,
  accent = false,
}: {
  icon: typeof CalendarClock;
  label: string;
  value: string;
  accent?: boolean;
}) {
  return (
    <div
      className={cn(
        "dv-chip flex items-center gap-3 rounded-xl border bg-card p-3 sm:p-4",
        accent ? "border-amber/40" : "border-border",
      )}
    >
      <span
        className={cn(
          "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md",
          accent ? "bg-amber-soft text-amber" : "bg-primary/10 text-primary",
        )}
      >
        <Icon className="h-4 w-4" aria-hidden="true" />
      </span>
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p
          className={cn(
            "truncate text-sm font-semibold",
            accent ? "text-amber" : "text-foreground",
          )}
        >
          {value}
        </p>
      </div>
    </div>
  );
}
