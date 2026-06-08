"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { Menu, X, Globe } from "lucide-react";
import { Button } from "@/components/ui/button";
import { routing } from "@/routing";

interface LandingNavProps {
  locale: string;
}

export function LandingNav({ locale }: LandingNavProps) {
  const t = useTranslations("landing.nav");
  const tCommon = useTranslations("common");
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);

  const isRtl = locale === "ar";
  const otherLocale = locale === "en" ? "ar" : "en";
  const otherPath = pathname.replace(`/${locale}`, `/${otherLocale}`) || `/${otherLocale}`;

  useEffect(() => {
    if (!open) return;
    const original = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = original;
    };
  }, [open]);

  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  const navLinks = [
    { href: "#features", label: t("features") },
    { href: "#how-it-works", label: t("howItWorks") },
    { href: "#pricing", label: t("pricing") },
    { href: "#faq", label: t("faq") },
  ];

  return (
    <header className="sticky top-0 z-40 border-b border-border/60 bg-background/85 backdrop-blur supports-[backdrop-filter]:bg-background/70">
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:absolute focus:start-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-foreground focus:px-3 focus:py-2 focus:text-sm focus:text-background"
      >
        {t("skipToContent")}
      </a>
      <div className="container-wide flex h-16 items-center justify-between gap-6">
        <Link
          href={`/${locale}`}
          className="flex items-center gap-2 text-base font-semibold tracking-tight text-foreground"
          aria-label={tCommon("appName")}
        >
          <Image
            src="/logo.png"
            alt=""
            aria-hidden="true"
            width={28}
            height={28}
            className="rounded-sm"
          />
          {tCommon("appName")}
        </Link>

        <nav
          aria-label="Primary"
          className="hidden items-center gap-7 md:flex"
        >
          {navLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded-sm"
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-2 md:flex">
          <button
            type="button"
            onClick={() => router.push(otherPath)}
            className="inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            aria-label={t("languageLabel")}
          >
            <Globe className="h-4 w-4" aria-hidden="true" />
            <span className="font-medium">
              {otherLocale === "ar" ? "العربية" : "English"}
            </span>
          </button>
          <Button variant="ghost" size="sm" asChild>
            <Link href={`/${locale}/auth/login`}>{t("signIn")}</Link>
          </Button>
          <Button size="sm" asChild>
            <Link href={`/${locale}/auth/register`}>{t("getStarted")}</Link>
          </Button>
        </div>

        <button
          type="button"
          className="inline-flex h-10 w-10 items-center justify-center rounded-md text-foreground transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring md:hidden"
          onClick={() => setOpen((v) => !v)}
          aria-expanded={open}
          aria-controls="mobile-nav"
          aria-label={open ? t("closeMenu") : t("openMenu")}
        >
          {open ? (
            <X className="h-5 w-5" aria-hidden="true" />
          ) : (
            <Menu className="h-5 w-5" aria-hidden="true" />
          )}
        </button>
      </div>

      {open ? (
        <div
          id="mobile-nav"
          className="absolute inset-x-0 top-16 origin-top border-b border-border bg-background shadow-sm md:hidden"
        >
          <div className="container-wide flex flex-col gap-1 py-4">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                className="rounded-md px-3 py-3 text-base text-foreground transition-colors hover:bg-accent"
                onClick={() => setOpen(false)}
              >
                {link.label}
              </a>
            ))}
            <div className="my-2 hairline" />
            <button
              type="button"
              onClick={() => {
                setOpen(false);
                router.push(otherPath);
              }}
              className="flex items-center gap-2 rounded-md px-3 py-3 text-base text-muted-foreground transition-colors hover:bg-accent"
            >
              <Globe className="h-4 w-4" aria-hidden="true" />
              {otherLocale === "ar" ? "العربية" : "English"}
            </button>
            <Button variant="outline" asChild className="justify-center">
              <Link href={`/${locale}/auth/login`} onClick={() => setOpen(false)}>
                {t("signIn")}
              </Link>
            </Button>
            <Button asChild className="justify-center">
              <Link
                href={`/${locale}/auth/register`}
                onClick={() => setOpen(false)}
              >
                {t("getStarted")}
              </Link>
            </Button>
          </div>
        </div>
      ) : null}
    </header>
  );
}
