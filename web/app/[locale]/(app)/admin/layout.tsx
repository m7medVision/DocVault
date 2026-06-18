import Link from "next/link";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { Shield } from "lucide-react";
import { Button } from "@/components/ui/button";

export default async function AdminLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("admin");

  return (
    <div className="grid gap-6 lg:grid-cols-[200px_1fr]">
      <aside className="space-y-4">
        <div className="flex items-center gap-2 font-semibold">
          <Shield className="h-5 w-5" />
          {t("title")}
        </div>
        <nav className="space-y-1">
          <Button variant="ghost" className="w-full justify-start" asChild>
            <Link href={`/${locale}/admin/audit`}>{t("auditLog")}</Link>
          </Button>
          <Button variant="ghost" className="w-full justify-start" asChild>
            <Link href={`/${locale}/admin/members`}>{t("members")}</Link>
          </Button>
          <Button variant="ghost" className="w-full justify-start" asChild>
            <Link href={`/${locale}/admin/groups`}>{t("groups")}</Link>
          </Button>
          <Button variant="ghost" className="w-full justify-start" asChild>
            <Link href={`/${locale}/admin/casbin`}>{t("policies")}</Link>
          </Button>
        </nav>
      </aside>
      <main>{children}</main>
    </div>
  );
}
