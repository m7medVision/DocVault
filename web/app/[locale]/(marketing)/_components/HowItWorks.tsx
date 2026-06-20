import { useTranslations } from "next-intl";

import { Reveal } from "@/components/ui/reveal";

import { ProcessingDiagram } from "./ProcessingDiagram";

export function HowItWorks() {
  const t = useTranslations("landing.howItWorks");

  return (
    <section
      id="how-it-works"
      aria-labelledby="how-title"
      className="border-t border-border bg-background"
    >
      <div className="container-wide py-20 sm:py-28">
        <div className="grid items-center gap-12 lg:grid-cols-[1fr_1.4fr] lg:gap-20">
          <Reveal className="lg:sticky lg:top-28 lg:self-start">
            <p className="kicker mb-4">{t("kicker")}</p>
            <h2
              id="how-title"
              className="text-balance text-3xl font-semibold tracking-tight text-foreground sm:text-4xl"
            >
              {t("title")}
            </h2>
            <p className="mt-4 max-w-md text-base leading-relaxed text-muted-foreground sm:text-lg">
              {t("subtitle")}
            </p>
          </Reveal>

          <Reveal delay={0.1}>
            <ProcessingDiagram />
          </Reveal>
        </div>
      </div>
    </section>
  );
}
