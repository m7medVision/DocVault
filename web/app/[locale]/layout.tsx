import { NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale } from "next-intl/server";
import { routing } from "@/routing";
import { AuthProvider } from "@/lib/useAuth";
import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/layout/AppSidebar";
import { AppHeader } from "@/components/layout/AppHeader";
import { Toaster } from "sonner";
import { HtmlAttributes } from "@/components/HtmlAttributes";

import { ReactQueryProvider } from "@/components/ReactQueryProvider";

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
            <SidebarProvider>
              <div className="flex min-h-screen w-full">
                <AppSidebar />
                <div className="flex flex-1 flex-col overflow-hidden">
                  <AppHeader />
                  <main className="flex-1 p-8">{children}</main>
                  <Toaster position="bottom-right" />
                </div>
              </div>
            </SidebarProvider>
          </ReactQueryProvider>
        </AuthProvider>
      </NextIntlClientProvider>
    </>
  );
}
