import type { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { routing } from "@/routing";

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  return locale === "ar"
    ? {
        title: "دوك فولت | مساحتك الهادئة لكل مستند يهمّك",
        description:
          "مساحة عمل ثنائية اللغة لإدارة العقود والفواتير والتذكيرات، مدعومة بالتعرف الضوئي على الحروف والبحث الذكي.",
      }
    : {
        title: "DocVault | A calm workspace for every document that matters",
        description:
          "A bilingual workspace for contracts, invoices, and reminders, powered by OCR and intelligent search.",
      };
}

export default async function AuthLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  return <>{children}</>;
}
