import { useTranslations } from "next-intl";

import { Reveal } from "@/components/ui/reveal";

const STEP_KEYS = ["upload", "process", "act"] as const;

export function HowItWorks() {
  const t = useTranslations("landing.howItWorks");

  return (
    <section
      id="how-it-works"
      aria-labelledby="how-title"
      className="border-t border-border bg-background"
    >
      <div className="container-wide py-20 sm:py-28">
        <div className="grid gap-12 lg:grid-cols-[1fr_1.6fr] lg:gap-20">
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

          <ol className="space-y-12">
            {STEP_KEYS.map((key, idx) => {
              const step = t.raw(`steps.${key}`) as {
                label: string;
                title: string;
                body: string;
              };
              return (
                <li key={key} className="grid grid-cols-[auto_1fr] gap-6">
                  <div className="flex flex-col items-center">
                    <span className="tabular-nums text-2xl font-semibold tracking-tight text-amber">
                      {String(idx + 1).padStart(2, "0")}
                    </span>
                    <span
                      aria-hidden="true"
                      className="mt-3 h-12 w-px bg-border"
                    />
                  </div>
                  <Reveal delay={idx * 0.1}>
                    <p className="kicker mb-2">{step.label}</p>
                    <h3 className="text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
                      {step.title}
                    </h3>
                    <p className="mt-3 max-w-prose text-base leading-relaxed text-muted-foreground">
                      {step.body}
                    </p>
                  </Reveal>
                </li>
              );
            })}
          </ol>
        </div>
      </div>
    </section>
  );
}
