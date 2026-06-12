"use client";

import { useRouter, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { formatDistanceToNow } from "date-fns";
import { Activity, Bell, Shield, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useActivityFeed } from "@/features/activity/useActivityFeed";

export default function ActivityPage() {
  const t = useTranslations("activity");
  const router = useRouter();
  const pathname = usePathname();
  const localePrefix = pathname.split("/")[1] || "en";

  const { items, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useActivityFeed();

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-3xl font-bold">
          <Activity className="size-7" />
          {t("title")}
        </h1>
        <p className="text-muted-foreground">{t("subtitle")}</p>
      </div>

      {isLoading ? (
        <div className="flex h-40 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : items.length === 0 ? (
        <Card className="py-16 text-center text-muted-foreground">{t("empty")}</Card>
      ) : (
        <div className="relative space-y-0 ps-4">
          <div className="absolute bottom-2 start-[1.45rem] top-2 w-px bg-border" aria-hidden />
          {items.map((item) => {
            const clickable = item.kind === "notification" && !!item.link;
            const Wrapper = clickable ? "button" : "div";
            return (
              <Wrapper
                key={item.id}
                onClick={
                  clickable
                    ? () =>
                        router.push(
                          item.link!.startsWith("/")
                            ? `/${localePrefix}${item.link}`
                            : item.link!
                        )
                    : undefined
                }
                className={`relative flex w-full items-start gap-3 py-3 text-start ${
                  clickable ? "cursor-pointer" : ""
                }`}
              >
                <div
                  className={`z-10 flex size-9 shrink-0 items-center justify-center rounded-full ${
                    item.kind === "audit"
                      ? "bg-amber-100 text-amber-700"
                      : "bg-primary/10 text-primary"
                  }`}
                >
                  {item.kind === "audit" ? (
                    <Shield className="size-4" />
                  ) : (
                    <Bell className="size-4" />
                  )}
                </div>
                <div className="min-w-0 flex-1 space-y-0.5 pt-1">
                  <p className="truncate font-medium capitalize">{item.title}</p>
                  {item.subtitle && (
                    <p className="truncate text-sm text-muted-foreground">{item.subtitle}</p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    {formatDistanceToNow(new Date(item.created_at), { addSuffix: true })}
                  </p>
                </div>
              </Wrapper>
            );
          })}

          {hasNextPage && (
            <div className="flex justify-center pt-4">
              <Button
                variant="outline"
                onClick={() => fetchNextPage()}
                disabled={isFetchingNextPage}
                className="gap-1.5"
              >
                {isFetchingNextPage && <Loader2 className="size-4 animate-spin" />}
                {t("loadMore")}
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
