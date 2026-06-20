"use client";

import { useTranslations } from "next-intl";
import { useReducedMotion } from "motion/react";

import { Reveal } from "@/components/ui/reveal";

const TESTIMONIAL_KEYS = ["salim", "mona", "khaled"] as const;

function initialsOf(name: string): string {
  const parts = name.split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "·";
  if (parts.length === 1) return parts[0].slice(0, 2);
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

type Testimonial = { quote: string; name: string; role: string };

function TestimonialCard({
  item,
  className = "",
}: {
  item: Testimonial;
  className?: string;
}) {
  return (
    <figure
      className={`flex h-full flex-col justify-between rounded-2xl border border-border bg-card p-6 sm:p-8 ${className}`}
    >
      <blockquote className="text-balance text-lg leading-relaxed text-foreground">
        <span aria-hidden="true" className="select-none text-amber">
          &ldquo;
        </span>
        {item.quote}
        <span aria-hidden="true" className="select-none text-amber">
          &rdquo;
        </span>
      </blockquote>
      <figcaption className="mt-6 flex items-center gap-3">
        <span
          aria-hidden="true"
          className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-foreground text-sm font-semibold text-background"
        >
          {initialsOf(item.name)}
        </span>
        <div>
          <p className="text-sm font-medium text-foreground">{item.name}</p>
          <p className="text-xs text-muted-foreground">{item.role}</p>
        </div>
      </figcaption>
    </figure>
  );
}

export function Testimonials() {
  const t = useTranslations("landing.testimonials");
  const reduce = useReducedMotion();

  const items = TESTIMONIAL_KEYS.map(
    (key) => t.raw(`items.${key}`) as Testimonial,
  );

  return (
    <section
      aria-labelledby="testimonials-title"
      className="border-t border-border bg-background"
    >
      <div className="container-wide py-20 sm:py-28">
        <Reveal>
          <div className="max-w-2xl">
            <p className="kicker mb-4">{t("kicker")}</p>
            <h2
              id="testimonials-title"
              className="text-balance text-3xl font-semibold tracking-tight text-foreground sm:text-4xl"
            >
              {t("title")}
            </h2>
            <p className="mt-4 max-w-xl text-base leading-relaxed text-muted-foreground sm:text-lg">
              {t("subtitle")}
            </p>
          </div>
        </Reveal>

        {reduce ? (
          <div className="mt-14 grid gap-4 md:grid-cols-2 lg:grid-cols-3 lg:gap-6">
            {items.map((item, idx) => (
              <TestimonialCard
                key={idx}
                item={item}
                className={idx === 1 ? "md:col-span-2 lg:col-span-1" : ""}
              />
            ))}
          </div>
        ) : (
          <div className="marquee mt-14 overflow-hidden">
            <div className="marquee-track flex w-max">
              <ul className="flex gap-4 pe-4 sm:gap-6 sm:pe-6">
                {items.map((item, idx) => (
                  <li key={idx} className="w-[20rem] shrink-0 sm:w-[24rem]">
                    <TestimonialCard item={item} className="h-full" />
                  </li>
                ))}
              </ul>
              <ul
                className="flex gap-4 pe-4 sm:gap-6 sm:pe-6"
                aria-hidden="true"
              >
                {items.map((item, idx) => (
                  <li key={idx} className="w-[20rem] shrink-0 sm:w-[24rem]">
                    <TestimonialCard item={item} className="h-full" />
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
