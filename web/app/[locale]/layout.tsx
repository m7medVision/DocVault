import Link from "next/link";
import Image from "next/image";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations, setRequestLocale } from "next-intl/server";
import { routing } from "@/routing";
import { AuthProvider } from "@/lib/useAuth";
import AuthNav from "@/components/AuthNav";
import LanguageSwitcher from "@/components/LanguageSwitcher";
import { Button } from "@/components/ui/button";

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const dir = locale === "ar" ? "rtl" : "ltr";

  const messages = await getMessages();
  const t = await getTranslations("nav");
  const tCommon = await getTranslations("common");

  return (
    <div dir={dir} lang={locale}>
      <NextIntlClientProvider messages={messages}>
      <AuthProvider>
        <div className="flex min-h-screen flex-col">
          <header className="sticky top-0 z-50 flex h-16 items-center justify-center gap-4 border-b bg-background/95 px-6 backdrop-blur supports-[backdrop-filter]:bg-background/60">
            <nav className="me-auto hidden gap-1 md:flex" aria-label="Primary">
              <Button variant="ghost" asChild>
                <Link href={`/${locale}`}>{t("home")}</Link>
              </Button>
              <Button variant="ghost" asChild>
                <Link href={`/${locale}/documents`}>{t("documents")}</Link>
              </Button>
              <Button variant="ghost" asChild>
                <Link href={`/${locale}/search`}>{t("search")}</Link>
              </Button>
              <Button variant="ghost" asChild>
                <Link href={`/${locale}/reminders`}>{t("reminders")}</Link>
              </Button>
            </nav>

            <div className="absolute left-1/2 -translate-x-1/2">
              <Link
                href={`/${locale}`}
                className="flex items-center gap-2 font-serif text-xl font-bold"
              >
                <Image
                  src="/favicon.svg"
                  alt=""
                  aria-hidden="true"
                  width={32}
                  height={32}
                />
                {tCommon("appName")}
              </Link>
            </div>

            <div className="ms-auto flex items-center gap-2">
              <LanguageSwitcher />
              <AuthNav />
            </div>
          </header>

          <main className="flex-1 p-8">{children}</main>
          <footer className="border-t p-4 text-center text-sm text-muted-foreground">
            {tCommon("footer", { appName: tCommon("appName") })}
          </footer>
        </div>
      </AuthProvider>
    </NextIntlClientProvider>
    </div>
  );
}
