"use client";

import { useRef } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { CheckCircle2, Calendar, ShieldCheck } from "lucide-react";
import {
  motion,
  useScroll,
  useTransform,
  useReducedMotion,
} from "motion/react";

import { Reveal } from "@/components/ui/reveal";
import { NoiseTexture } from "@/components/magicui/noise-texture";

export function Hero() {
  const t = useTranslations("landing.hero");
  const sectionRef = useRef<HTMLElement>(null);
  const reduce = useReducedMotion();

  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ["start start", "end start"],
  });

  const imageY = useTransform(
    scrollYProgress,
    [0, 1],
    [0, reduce ? 0 : -28],
  );

  return (
    <section
      ref={sectionRef}
      aria-labelledby="hero-title"
      className="relative overflow-hidden bg-background"
    >
      <div className="grain absolute inset-0 -z-10" aria-hidden="true" />
      <NoiseTexture className="absolute inset-0 -z-10 opacity-[0.04]" />
      <div className="container-wide py-20 sm:py-24 lg:py-32">
        <div className="grid items-center gap-12 lg:grid-cols-[1.05fr_1fr] lg:gap-16">
          <div className="flex flex-col items-start">
            <Reveal inView={false} blur={6} delay={0}>
              <p className="kicker mb-6">{t("eyebrow")}</p>
            </Reveal>
            <Reveal inView={false} blur={6} delay={0.08}>
              <h1
                id="hero-title"
                className="text-display text-balance text-foreground"
              >
                {t("title")}
              </h1>
            </Reveal>
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
                  {t("secondaryCta")} →
                </a>
              </div>
            </Reveal>
            <Reveal inView={false} blur={6} delay={0.32}>
              <p className="mt-3 text-xs text-muted-foreground">
                {t("primaryCtaHint")}
              </p>
            </Reveal>
            <Reveal inView={false} blur={6} delay={0.32}>
              <div className="mt-12 flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-muted-foreground">
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
            </Reveal>
          </div>

          <Reveal
            inView={false}
            blur={8}
            delay={0.2}
            offset={12}
            className="lg:ps-4"
          >
            <motion.div style={reduce ? undefined : { y: imageY }}>
              <figure className="relative overflow-hidden rounded-2xl border border-border bg-card">
                <Image
                  src="/screenshot.png"
                  alt={t("screenshotAlt")}
                  width={1567}
                  height={1029}
                  priority
                  sizes="(max-width: 1024px) 100vw, 50vw"
                  className="block h-auto w-full"
                />
              </figure>
            </motion.div>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
