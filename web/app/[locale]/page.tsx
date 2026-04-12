import { getTranslations, setRequestLocale } from "next-intl/server";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import Link from "next/link";

export default async function HomePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("home");
  const tCommon = await getTranslations("common");

  return (
    <div className="mx-auto max-w-4xl space-y-8 text-center">
      <div className="space-y-4">
        <h1 className="text-4xl font-bold tracking-tight">
          {t("welcome", { appName: tCommon("appName") })}
        </h1>
        <p className="text-xl text-muted-foreground">{t("subtitle")}</p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>{t("uploadTitle")}</CardTitle>
            <CardDescription>{t("uploadDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" asChild>
              <Link href={`/${locale}/documents/upload`}>{t("uploadLink")}</Link>
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("searchTitle")}</CardTitle>
            <CardDescription>{t("searchDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" asChild>
              <Link href={`/${locale}/search`}>{t("searchLink")}</Link>
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("remindersTitle")}</CardTitle>
            <CardDescription>{t("remindersDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" asChild>
              <Link href={`/${locale}/reminders`}>{t("remindersLink")}</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
