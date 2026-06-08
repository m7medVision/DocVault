import { NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale } from "next-intl/server";
import { routing } from "@/routing";
import { AuthProvider } from "@/lib/useAuth";
import { HtmlAttributes } from "@/components/HtmlAttributes";
import { ReactQueryProvider } from "@/components/ReactQueryProvider";
import { Toaster } from "sonner";

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

  const messages = await getMessages();

  return (
    <>
      <HtmlAttributes locale={locale} />
      <NextIntlClientProvider messages={messages}>
        <AuthProvider>
          <ReactQueryProvider>
            {children}
            <Toaster position="bottom-right" />
          </ReactQueryProvider>
        </AuthProvider>
      </NextIntlClientProvider>
    </>
  );
}
