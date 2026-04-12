"use client";

import dynamic from "next/dynamic";
import { Download, FileIcon, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const PdfViewer = dynamic(
  () => import("@/components/PdfViewer").then((mod) => mod.PdfViewer),
  { ssr: false }
);

type PreviewMode = "pdf" | "image" | "fallback";

function getPreviewMode(mimeType: string): PreviewMode {
  if (mimeType === "application/pdf") return "pdf";
  if (mimeType.startsWith("image/jpeg") || mimeType.startsWith("image/png")) return "image";
  return "fallback";
}

function getFallbackType(mimeType: string): string | null {
  if (mimeType === "image/tiff") return "tiff";
  if (
    mimeType === "application/msword" ||
    mimeType === "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
  ) {
    return "word";
  }
  return "unknown";
}

function ImageViewer({ url }: { url: string }) {
  return (
    <div
      className={cn(
        "mx-auto w-full max-w-3xl overflow-hidden rounded-lg border bg-white shadow-md",
        "flex items-center justify-center min-h-[600px] p-4"
      )}
    >
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={url}
        alt="Document preview"
        className="max-h-[800px] max-w-full rounded object-contain"
      />
    </div>
  );
}

function FallbackCard({
  message,
  downloadLabel,
  url,
  fileSize,
}: {
  message: string;
  downloadLabel: string;
  url: string;
  fileSize?: string;
}) {
  return (
    <div
      className={cn(
        "mx-auto flex w-full max-w-3xl flex-col items-center justify-center gap-4",
        "rounded-lg border bg-white shadow-md min-h-[400px] p-12"
      )}
    >
      <FileIcon className="size-12 text-muted-foreground" />
      <p className="text-center text-sm text-muted-foreground">{message}</p>
      {fileSize && (
        <span className="text-xs text-muted-foreground">{fileSize}</span>
      )}
      <Button asChild className="gap-2">
        <a href={url} download>
          <Download className="size-4" />
          {downloadLabel}
        </a>
      </Button>
    </div>
  );
}

interface FilePreviewProps {
  url: string | null;
  mimeType: string;
  fileSize?: number;
  t: (key: string) => string;
}

export function FilePreview({ url, mimeType, fileSize, t }: FilePreviewProps) {
  if (!url) {
    return (
      <div className="flex items-center justify-center py-20 text-sm text-muted-foreground">
        <Loader2 className="mr-2 size-4 animate-spin" />
        {t("loadingPreview")}
      </div>
    );
  }

  const mode = getPreviewMode(mimeType);
  const readableSize = fileSize ? `${(fileSize / 1024).toFixed(1)} KB` : undefined;

  if (mode === "pdf") {
    return (
      <PdfViewer
        url={url}
        failedMessage={t("failedToLoadPreview")}
        loadingMessage={t("loadingPreview")}
      />
    );
  }

  if (mode === "image") {
    return <ImageViewer url={url} />;
  }

  const fallbackType = getFallbackType(mimeType);
  const messageKey =
    fallbackType === "tiff"
      ? "tiffPreviewUnavailable"
      : fallbackType === "word"
        ? "wordPreviewUnavailable"
        : "previewUnavailable";

  return (
    <FallbackCard
      message={t(messageKey)}
      downloadLabel={t("downloadFile")}
      url={url}
      fileSize={readableSize}
    />
  );
}
