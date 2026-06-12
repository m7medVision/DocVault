"use client";

import { useTranslations } from "next-intl";
import { FileText, Clock, CheckCircle2, HardDrive, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useDocumentStats } from "@/features/documents/useDocumentStats";
import { useStatusBreakdown } from "@/features/documents/useStatusBreakdown";

function formatBytes(bytes: number): string {
  if (!bytes) return "0 B";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

function BreakdownBars({ data }: { data: Record<string, number> }) {
  const entries = Object.entries(data).filter(([, v]) => v > 0);
  const max = Math.max(1, ...entries.map(([, v]) => v));
  if (entries.length === 0) {
    return null;
  }
  return (
    <div className="space-y-3">
      {entries.map(([key, value]) => (
        <div key={key} className="space-y-1">
          <div className="flex items-center justify-between text-sm">
            <span className="capitalize text-muted-foreground">{key}</span>
            <span className="font-medium tabular-nums">{value}</span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-primary"
              style={{ width: `${(value / max) * 100}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

export default function AnalyticsPage() {
  const t = useTranslations("analytics");
  const { data: stats, isLoading } = useDocumentStats();
  const { data: breakdown, isLoading: breakdownLoading } = useStatusBreakdown();

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const cards = [
    {
      label: t("totalDocuments"),
      value: stats?.total_documents ?? 0,
      icon: FileText,
    },
    {
      label: t("pending"),
      value: stats?.pending_documents ?? 0,
      icon: Clock,
    },
    {
      label: t("completedThisWeek"),
      value: stats?.completed_this_week ?? 0,
      icon: CheckCircle2,
    },
    {
      label: t("storageUsed"),
      value: formatBytes(stats?.storage_used_bytes ?? 0),
      icon: HardDrive,
    },
  ];

  const hasBreakdown =
    breakdown &&
    (Object.values(breakdown.byStatus).some((v) => v > 0) ||
      Object.values(breakdown.byType).some((v) => v > 0));

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-3xl font-bold">{t("title")}</h1>
        <p className="text-muted-foreground">{t("subtitle")}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map((card) => (
          <Card key={card.label}>
            <CardContent className="flex items-center gap-4 p-6">
              <div className="flex size-11 items-center justify-center rounded-lg bg-primary/10">
                <card.icon className="size-5 text-primary" />
              </div>
              <div className="min-w-0">
                <p className="truncate text-sm text-muted-foreground">{card.label}</p>
                <p className="text-2xl font-bold tabular-nums">{card.value}</p>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("byStatus")}</CardTitle>
          </CardHeader>
          <CardContent>
            {breakdownLoading ? (
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            ) : breakdown && Object.values(breakdown.byStatus).some((v) => v > 0) ? (
              <BreakdownBars data={breakdown.byStatus} />
            ) : (
              <p className="text-sm text-muted-foreground">{t("noData")}</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("byType")}</CardTitle>
          </CardHeader>
          <CardContent>
            {breakdownLoading ? (
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            ) : hasBreakdown ? (
              <BreakdownBars data={breakdown!.byType} />
            ) : (
              <p className="text-sm text-muted-foreground">{t("noData")}</p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
