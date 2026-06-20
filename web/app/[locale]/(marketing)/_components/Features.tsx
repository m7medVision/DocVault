import { useTranslations } from "next-intl";
import { BellRing, Search, MessagesSquare } from "lucide-react";

import { Reveal } from "@/components/ui/reveal";
import { NoiseTexture } from "@/components/ui/noise-texture";

import { LivingDocumentDemo } from "./LivingDocumentDemo";

const SUPPORTING = [
  { slug: "reminders", icon: BellRing },
  { slug: "search", icon: Search },
  { slug: "chat", icon: MessagesSquare },
] as const;

export function Features() {
  const t = useTranslations("landing.features");

  return (
    <section
      id="features"
      aria-labelledby="features-title"
      className="relative overflow-hidden border-t border-border bg-secondary/30"
    >
      <NoiseTexture className="pointer-events-none absolute inset-0 opacity-[0.12]" />
      <div className="container-wide relative py-20 sm:py-28">
        <Reveal>
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
        </Reveal>

        {/* The product, doing its thing — upload a document, watch it read and extract. */}
        <Reveal delay={0.1} className="mt-14">
          <LivingDocumentDemo />
        </Reveal>

        {/* Everything that happens after extraction. */}
        <div className="mt-12 grid gap-4 border-t border-border pt-12 sm:grid-cols-3 sm:gap-6">
          {SUPPORTING.map(({ slug, icon: Icon }, i) => (
            <Reveal key={slug} delay={0.1 + i * 0.08}>
              <div className="flex items-start gap-3">
                <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
                  <Icon className="h-4 w-4" aria-hidden="true" />
                </span>
                <div>
                  <h3 className="text-base font-semibold tracking-tight text-foreground">
                    {t(`items.${slug}.title`)}
                  </h3>
                  <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                    {t(`items.${slug}.body`)}
                  </p>
                </div>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}
