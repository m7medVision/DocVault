"use client";

import { useTranslations } from "next-intl";
import { History, FileText } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import type { DocumentVersion } from "@/lib/api/types";

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function VersionTimeline({ versions }: { versions: DocumentVersion[] }) {
  const t = useTranslations("versions");

  if (!versions || versions.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center text-sm text-muted-foreground">
          {t("noVersions")}
        </CardContent>
      </Card>
    );
  }

  // Highest version_number is the current one.
  const sorted = [...versions].sort((a, b) => b.version_number - a.version_number);
  const currentNumber = sorted[0]?.version_number;

  return (
    <div className="space-y-3">
      {sorted.map((version) => (
        <div
          key={version.id}
          className="flex items-start gap-4 rounded-lg border p-4"
        >
          <div className="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-full bg-muted">
            <History className="size-4 text-muted-foreground" />
          </div>
          <div className="min-w-0 flex-1 space-y-1">
            <div className="flex items-center gap-2">
              <span className="font-medium">
                {t("version", { number: version.version_number })}
              </span>
              {version.version_number === currentNumber && (
                <Badge variant="default">{t("current")}</Badge>
              )}
            </div>
            <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
              <span className="inline-flex items-center gap-1">
                <FileText className="size-3.5" />
                {version.mime_type}
              </span>
              <span>
                {t("size")}: {formatBytes(version.file_size)}
              </span>
              <span>{new Date(version.created_at).toLocaleString()}</span>
              {version.uploaded_by && (
                <span>
                  {t("uploadedBy")}: {version.uploaded_by}
                </span>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
