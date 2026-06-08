import Link from "next/link";
import { useTranslations } from "next-intl";
import { Mail } from "lucide-react";

interface FooterProps {
  locale: string;
}

export function Footer({ locale }: FooterProps) {
  const t = useTranslations("landing.footer");
  const tCommon = useTranslations("common");
  const isRtl = locale === "ar";

  return (
    <footer className="border-t border-border bg-secondary/40">
      <div className="container-wide py-16 lg:py-20">
        <div className="grid gap-12 lg:grid-cols-[1.4fr_2fr] lg:gap-20">
          <div className="space-y-6">
            <div className="flex items-center gap-2 text-base font-semibold tracking-tight text-foreground">
              <span
                aria-hidden="true"
                className="inline-block h-7 w-7 rounded-sm bg-primary"
                style={{
                  backgroundImage:
                    "linear-gradient(135deg, oklch(38% 0.08 250) 0%, oklch(50% 0.10 250) 100%)",
                }}
              />
              {tCommon("appName")}
            </div>
            <p className="max-w-md text-sm leading-relaxed text-muted-foreground">
              {t("tagline")}
            </p>
            <a
              href={`mailto:${t("contactEmail")}`}
              className="inline-flex items-center gap-2 text-sm font-medium text-foreground underline-offset-4 transition-colors hover:text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded-sm"
            >
              <Mail className="h-4 w-4" aria-hidden="true" />
              <span className="tabular-nums">{t("contactEmail")}</span>
              <span className="sr-only">. {t("contactLabel")}</span>
            </a>
          </div>

          <div className="grid grid-cols-2 gap-8 sm:grid-cols-3">
            {(["product", "company", "legal"] as const).map((key) => {
              const column = t.raw(`columns.${key}`) as {
                title: string;
                links: { label: string; href: string }[];
              };
              return (
                <div key={key}>
                  <h3 className="kicker mb-4">{column.title}</h3>
                  <ul className="space-y-3">
                    {column.links.map((link) => (
                      <li key={link.label}>
                        <Link
                          href={link.href}
                          className="text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"
                        >
                          {link.label}
                        </Link>
                      </li>
                    ))}
                  </ul>
                </div>
              );
            })}
          </div>
        </div>

        <div className="mt-16 flex flex-col items-start justify-between gap-3 border-t border-border pt-8 text-sm text-muted-foreground sm:flex-row sm:items-center">
          <p>{t("copyright")}</p>
          <p className="tabular-nums">
            {isRtl ? "صُنع بحرص، في مسقط والقاهرة" : "Made with care, in Muscat and Cairo."}
          </p>
        </div>
      </div>
    </footer>
  );
}
