"use client";

import { useTranslations } from "next-intl";
import { Bell, BellOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useReminders } from "@/features/reminders/useReminders";

function getUrgencyColor(daysUntil: number) {
  if (daysUntil < 0) return "bg-destructive/10 text-destructive";
  if (daysUntil <= 7) return "bg-orange-100 text-orange-700";
  if (daysUntil <= 30) return "bg-blue-100 text-blue-700";
  return "bg-green-100 text-green-700";
}

function getStatusVariant(status: string) {
  switch (status) {
    case "sent":
      return "default" as const;
    case "pending":
      return "secondary" as const;
    case "failed":
      return "destructive" as const;
    default:
      return "outline" as const;
  }
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr);
  return date.toLocaleDateString();
}

export default function RemindersPage() {
  const t = useTranslations("notifications");
  const tCommon = useTranslations("common");
  const {
    reminders,
    loading,
    error,
    filter,
    setFilter,
    handleDismiss,
    handleSnooze,
  } = useReminders();

  const filteredReminders = reminders.filter((r) => {
    if (filter === "all") return true;
    return r.status === filter;
  });

  const getDaysLabel = (daysUntil: number) => {
    if (daysUntil < 0) {
      return t("daysOverdue", { count: Math.abs(daysUntil) });
    }
    return t("daysUntil", { count: daysUntil });
  };

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="text-muted-foreground">{tCommon("loading")}</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-destructive/10 p-4 text-center text-destructive">
        {error}
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">{t("title")}</h1>
        <Bell className="h-6 w-6 text-muted-foreground" />
      </div>

      <Tabs value={filter} onValueChange={(v) => setFilter(v as typeof filter)}>
        <TabsList>
          <TabsTrigger value="all">{t("all")}</TabsTrigger>
          <TabsTrigger value="pending">{t("pending")}</TabsTrigger>
          <TabsTrigger value="sent">{t("sent")}</TabsTrigger>
        </TabsList>

        <TabsContent value={filter} className="mt-4">
          {filteredReminders.length === 0 ? (
            <Card className="p-8 text-center">
              <div className="flex flex-col items-center gap-3">
                <BellOff className="h-10 w-10 text-muted-foreground" />
                <p className="text-muted-foreground">{t("noNotifications")}</p>
              </div>
            </Card>
          ) : (
            <div className="space-y-3">
              {filteredReminders.map((reminder) => (
                <Card key={reminder.id} className="p-4">
                  <div className="flex items-center gap-4">
                    <div
                      className={`flex h-14 w-14 shrink-0 items-center justify-center rounded-full text-sm font-semibold ${getUrgencyColor(
                        reminder.days_until
                      )}`}
                    >
                      {getDaysLabel(reminder.days_until)}
                    </div>

                    <div className="flex-1 space-y-1">
                      <h3 className="font-medium">{reminder.document_title}</h3>
                      <p className="text-sm text-muted-foreground">
                        {reminder.rule_type.charAt(0).toUpperCase() +
                          reminder.rule_type.slice(1)}{" "}
                        - {formatDate(reminder.trigger_date)}
                      </p>
                    </div>

                    <Badge variant={getStatusVariant(reminder.status)}>
                      {t(reminder.status)}
                    </Badge>

                    {reminder.status === "pending" && (
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleSnooze(reminder.id)}
                        >
                          {tCommon("snooze")}
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => void handleDismiss(reminder.id)}
                        >
                          {tCommon("dismiss")}
                        </Button>
                      </div>
                    )}
                  </div>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
