import { useTranslations } from "next-intl";

const TESTIMONIAL_KEYS = ["salim", "mona", "khaled"] as const;

function initialsOf(name: string): string {
  const parts = name.split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "·";
  if (parts.length === 1) return parts[0].slice(0, 2);
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function Testimonials() {
  const t = useTranslations("landing.testimonials");

  return (
    <section
      aria-labelledby="testimonials-title"
      className="border-t border-border bg-background"
    >
      <div className="container-wide py-20 sm:py-28">
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

        <div className="mt-14 grid gap-4 md:grid-cols-2 lg:grid-cols-3 lg:gap-6">
          {TESTIMONIAL_KEYS.map((key, idx) => {
            const item = t.raw(`items.${key}`) as {
              quote: string;
              name: string;
              role: string;
            };
            return (
              <figure
                key={key}
                className={`flex h-full flex-col justify-between rounded-2xl border border-border bg-card p-6 sm:p-8 ${
                  idx === 1 ? "md:col-span-2 lg:col-span-1" : ""
                }`}
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
                    <p className="text-sm font-medium text-foreground">
                      {item.name}
                    </p>
                    <p className="text-xs text-muted-foreground">{item.role}</p>
                  </div>
                </figcaption>
              </figure>
            );
          })}
        </div>
      </div>
    </section>
  );
}
