"use client";

import { useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Bell, Check, Info, Loader2, Inbox } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useInfiniteNotifications } from "@/features/notifications/useInfiniteNotifications";
import { markNotificationRead, type Notification } from "@/features/notifications/api";

export default function NotificationsPage() {
  const t = useTranslations("notifications");
  const router = useRouter();
  const pathname = usePathname();
  const localePrefix = pathname.split("/")[1] || "en";
  const queryClient = useQueryClient();

  const [tab, setTab] = useState<"all" | "unread">("all");
  const [markingAll, setMarkingAll] = useState(false);

  const status = tab === "unread" ? "unread" : undefined;
  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteNotifications(status);

  const notifications = data?.pages.flatMap((p) => p.notifications) ?? [];
  const unreadCount = data?.pages[0]?.unread_count ?? 0;

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ["notifications"] });

  const handleClick = async (n: Notification) => {
    if (n.status === "unread") {
      try {
        await markNotificationRead(n.id);
        await invalidate();
      } catch {
        /* ignore */
      }
    }
    if (n.link) {
      router.push(n.link.startsWith("/") ? `/${localePrefix}${n.link}` : n.link);
    }
  };

  const handleMarkAll = async () => {
    const unread = notifications.filter((n) => n.status === "unread");
    if (unread.length === 0) return;
    setMarkingAll(true);
    try {
      await Promise.all(unread.map((n) => markNotificationRead(n.id).catch(() => null)));
      await invalidate();
    } finally {
      setMarkingAll(false);
    }
  };

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <h1 className="flex items-center gap-2 text-3xl font-bold">
            <Bell className="size-7" />
            {t("title")}
          </h1>
          <p className="text-muted-foreground">{t("subtitle")}</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleMarkAll}
          disabled={markingAll || unreadCount === 0}
          className="gap-1.5"
        >
          {markingAll ? <Loader2 className="size-4 animate-spin" /> : <Check className="size-4" />}
          {t("markAllRead")}
        </Button>
      </div>

      <Tabs value={tab} onValueChange={(v) => setTab(v as "all" | "unread")}>
        <TabsList>
          <TabsTrigger value="all">{t("filterAll")}</TabsTrigger>
          <TabsTrigger value="unread" className="gap-1.5">
            {t("filterUnread")}
            {unreadCount > 0 && (
              <Badge variant="destructive" className="h-5 px-1.5 text-xs">
                {unreadCount > 99 ? "99+" : unreadCount}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>
      </Tabs>

      {isLoading ? (
        <div className="flex h-40 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : notifications.length === 0 ? (
        <Card className="flex flex-col items-center gap-2 py-16 text-center text-muted-foreground">
          <Inbox className="size-8" />
          <p>{t("empty")}</p>
        </Card>
      ) : (
        <div className="space-y-2">
          {notifications.map((n) => (
            <button
              key={n.id}
              onClick={() => handleClick(n)}
              className="flex w-full items-start gap-3 rounded-lg border p-4 text-start transition-colors hover:bg-muted/50"
            >
              <div
                className={`mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-full ${
                  n.status === "unread" ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground"
                }`}
              >
                {n.type === "reminder" ? <Bell className="size-4" /> : <Info className="size-4" />}
              </div>
              <div className="min-w-0 flex-1 space-y-0.5">
                <div className="flex items-center gap-2">
                  <p className="truncate font-medium">{n.title}</p>
                  {n.status === "unread" && (
                    <span className="size-2 shrink-0 rounded-full bg-primary" aria-hidden />
                  )}
                </div>
                {n.body && <p className="text-sm text-muted-foreground">{n.body}</p>}
                <p className="text-xs text-muted-foreground">
                  {n.created_at
                    ? formatDistanceToNow(new Date(n.created_at), { addSuffix: true })
                    : t("justNow")}
                </p>
              </div>
            </button>
          ))}

          {hasNextPage && (
            <div className="flex justify-center pt-2">
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
