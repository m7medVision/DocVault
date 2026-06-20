"use client";

import Image from "next/image";
import { useTranslations } from "next-intl";
import { CheckCircle2, Calendar, ShieldCheck } from "lucide-react";

import { Reveal } from "@/components/ui/reveal";
import { NoiseTexture } from "@/components/ui/noise-texture";
import { Spotlight } from "@/components/ui/spotlight";
import { TextGenerateEffect } from "@/components/ui/text-generate-effect";
import { MacbookScroll } from "@/components/ui/macbook-scroll";

export function Hero() {
  const t = useTranslations("landing.hero");

  return (
    <section
      aria-labelledby="hero-title"
      className="relative overflow-hidden bg-background"
    >
      <Spotlight className="left-1/2 top-0 text-primary" fill="currentColor" />
      <NoiseTexture className="absolute inset-0 -z-10 opacity-[0.04] dark:opacity-[0.05]" />

      <div className="container-wide relative z-10 pt-20 sm:pt-24 lg:pt-32">
        <div className="mx-auto flex max-w-3xl flex-col items-center text-center">
          <Reveal inView={false} blur={6} delay={0}>
            <p className="kicker mb-6">{t("eyebrow")}</p>
          </Reveal>
          <h1
            id="hero-title"
            className="text-display text-balance text-foreground"
          >
            <TextGenerateEffect words={t("title")} />
          </h1>
          <Reveal inView={false} blur={6} delay={0.16}>
            <p className="mt-6 max-w-2xl text-lg leading-relaxed text-muted-foreground">
              {t("subtitle")}
            </p>
          </Reveal>
          <Reveal inView={false} blur={6} delay={0.24}>
            <div className="mt-10 flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-4">
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
                {t("secondaryCta")}
                <span aria-hidden="true" className="rtl:hidden">
                  →
                </span>
                <span aria-hidden="true" className="hidden rtl:inline">
                  ←
                </span>
              </a>
            </div>
          </Reveal>
          <Reveal inView={false} blur={6} delay={0.32}>
            <p className="mt-3 text-xs text-muted-foreground">
              {t("primaryCtaHint")}
            </p>
          </Reveal>
          <Reveal inView={false} blur={6} delay={0.32}>
            <div className="mt-12 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-xs text-muted-foreground">
              <span className="inline-flex items-center gap-1.5">
                <CheckCircle2
                  className="h-3.5 w-3.5 text-primary"
                  aria-hidden="true"
                />
                <span>OCR · EN + AR</span>
              </span>
              <span className="inline-flex items-center gap-1.5">
                <ShieldCheck
                  className="h-3.5 w-3.5 text-primary"
                  aria-hidden="true"
                />
                <span>Tenant-scoped</span>
              </span>
              <span className="inline-flex items-center gap-1.5">
                <Calendar className="h-3.5 w-3.5 text-amber" aria-hidden="true" />
                <span>Expiry aware</span>
              </span>
            </div>
          </Reveal>
        </div>
      </div>

      {/* Animated MacBook reveal (desktop only — it owns a tall scroll track). */}
      <div dir="ltr" className="hidden lg:block">
        <MacbookScroll
          src="/screenshot.png"
          showGradient={false}
          title={t("macbookTitle")}
        />
      </div>

      {/* Static screenshot for small / medium screens. */}
      <div className="container-wide relative z-10 pb-20 pt-12 sm:pb-28 lg:hidden">
        <Reveal blur={8} offset={12}>
          <figure className="relative mx-auto max-w-3xl overflow-hidden rounded-2xl border border-border bg-card">
            <Image
              src="/screenshot.png"
              alt={t("screenshotAlt")}
              width={1567}
              height={1029}
              priority
              sizes="100vw"
              className="block h-auto w-full"
            />
          </figure>
        </Reveal>
      </div>
    </section>
  );
}
