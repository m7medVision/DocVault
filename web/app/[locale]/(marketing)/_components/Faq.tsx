import { useTranslations } from "next-intl";

const FAQ_KEYS = ["data", "languages", "trial", "migrate", "billing", "support"] as const;

export function Faq() {
  const t = useTranslations("landing.faq");

  return (
    <section
      id="faq"
      aria-labelledby="faq-title"
      className="border-t border-border bg-secondary/30"
    >
      <div className="container-wide py-20 sm:py-28">
        <div className="grid gap-12 lg:grid-cols-[1fr_1.4fr] lg:gap-20">
          <div>
            <p className="kicker mb-4">{t("kicker")}</p>
            <h2
              id="faq-title"
              className="text-balance text-3xl font-semibold tracking-tight text-foreground sm:text-4xl"
            >
              {t("title")}
            </h2>
          </div>

          <div className="divide-y divide-border border-y border-border">
            {FAQ_KEYS.map((key) => {
              const item = t.raw(`items.${key}`) as {
                question: string;
                answer: string;
              };
              return (
                <details
                  key={key}
                  className="group py-2"
                >
                  <summary
                    className="flex cursor-pointer list-none items-center justify-between gap-6 py-4 text-left text-base font-medium text-foreground transition-colors hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded-sm"
                  >
                    <span>{item.question}</span>
                    <span
                      aria-hidden="true"
                      className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-border text-foreground transition-transform duration-200 group-open:rotate-45"
                    >
                      <svg
                        viewBox="0 0 16 16"
                        className="h-3.5 w-3.5"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="1.5"
                      >
                        <path d="M8 3v10M3 8h10" strokeLinecap="round" />
                      </svg>
                    </span>
                  </summary>
                  <div className="grid grid-rows-[0fr] transition-[grid-template-rows] duration-300 ease-out group-open:grid-rows-[1fr]">
                    <div className="overflow-hidden">
                      <p className="pb-4 pr-12 text-sm leading-relaxed text-muted-foreground sm:text-base">
                        {item.answer}
                      </p>
                    </div>
                  </div>
                </details>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}
