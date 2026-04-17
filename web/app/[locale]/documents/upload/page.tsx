'use client';

import { useState, useCallback } from "react";
import { useTranslations } from "next-intl";
import { Upload, X, FileText, CheckCircle2, AlertCircle, Loader2, FolderOpen, FilePen } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { useUploadWithProgress, ALLOWED_FILE_TYPES, MAX_FILE_SIZE, type UploadStatus } from "@/features/documents/useUploadWithProgress";
import { getDocument } from "@/features/documents/api";
import type { DocumentDetailResponse } from "@/lib/api/types";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

interface CompletedUpload {
  id: string;
  name: string;
  suggestedFolderName?: string;
  suggestedFilename?: string;
  suggestionConfidence?: number;
  suggestionCreateNew?: boolean;
}

export default function UploadPage() {
  const t = useTranslations("documents");
  const tCommon = useTranslations("common");
  const router = useRouter();

  const [completedUploads, setCompletedUploads] = useState<Map<number, CompletedUpload>>(new Map());

  const handleComplete = useCallback(async (documentId: string, fileName: string, index: number) => {
    try {
      const token = localStorage.getItem("access_token");
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/documents/${documentId}`, {
        headers: token ? { 'Authorization': `Bearer ${token}` } : {},
      });
      if (!res.ok) return;
      const data: DocumentDetailResponse = await res.json();

      setCompletedUploads((prev) => {
        const next = new Map(prev);
        next.set(index, {
          id: documentId,
          name: fileName,
          suggestedFolderName: data.document.suggested_folder_name ?? undefined,
          suggestedFilename: data.document.suggested_filename ?? undefined,
          suggestionConfidence: data.document.suggestion_confidence ?? undefined,
          suggestionCreateNew: data.document.suggestion_create_new ?? undefined,
        });
        return next;
      });
    } catch {
      // non-fatal — suggestion not critical
    }
  }, []);

  const {
    files,
    uploads,
    status,
    error,
    isDragging,
    addFiles,
    removeFile,
    upload,
    setDragging,
  } = useUploadWithProgress({ onComplete: handleComplete });

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragging(false);

    const droppedFiles = Array.from(e.dataTransfer.files);
    processFiles(droppedFiles);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(e.target.files || []);
    processFiles(selectedFiles);
  };

  const processFiles = (newFiles: File[]) => {
    const validFiles: File[] = [];
    const errors: string[] = [];

    for (const file of newFiles) {
      if (!ALLOWED_FILE_TYPES.includes(file.type as typeof ALLOWED_FILE_TYPES[number])) {
        errors.push(`${file.name}: ${t("fileTypeNotSupported", { type: file.type })}`);
      } else if (file.size > MAX_FILE_SIZE) {
        errors.push(`${file.name}: ${t("fileTooLarge", { size: MAX_FILE_SIZE / 1024 / 1024 })}`);
      } else {
        validFiles.push(file);
      }
    }

    if (errors.length > 0) {
      toast.error(errors.join("\n"));
    }

    addFiles(validFiles);
  };

  const uploadIcon = (s: UploadStatus) => {
    if (s === 'completed') return <CheckCircle2 className="h-5 w-5 text-green-500" />;
    if (s === 'error') return <AlertCircle className="h-5 w-5 text-destructive" />;
    if (s === 'processing' || s === 'uploading') return <Loader2 className="h-5 w-5 animate-spin text-primary" />;
    return <FileText className="h-5 w-5 text-muted-foreground" />;
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-3xl font-bold">{t("uploadTitle")}</h1>

      <div
        className={`relative rounded-xl border-2 border-dashed p-8 text-center transition-colors ${
          isDragging
            ? "border-primary bg-primary/5"
            : "border-border hover:border-primary/50"
        }`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <input
          type="file"
          id="file-input"
          multiple
          onChange={handleFileSelect}
          accept={ALLOWED_FILE_TYPES.join(",")}
          className="absolute inset-0 cursor-pointer opacity-0"
        />
        <div className="flex flex-col items-center gap-3 pointer-events-none">
          <Upload className="h-10 w-10 text-muted-foreground" />
          <div className="space-y-1">
            <p className="text-sm font-medium">{t("dragDrop")}</p>
            <p className="text-xs text-muted-foreground">{t("fileTypes")}</p>
          </div>
        </div>
      </div>

      {error && (
        <Card className="border-destructive bg-destructive/10">
          <CardContent className="flex items-start gap-3 p-4">
            <AlertCircle className="h-5 w-5 text-destructive" />
            <pre className="flex-1 whitespace-pre-wrap text-sm text-destructive">
              {error}
            </pre>
          </CardContent>
        </Card>
      )}

      {files.length > 0 && (
        <Card>
          <CardContent className="p-4">
            <h3 className="mb-3 font-medium">
              {t("selectedFiles")} ({files.length})
            </h3>
            <div className="space-y-2">
              {files.map((file, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between rounded-lg bg-muted p-3"
                >
                  <div className="flex items-center gap-3">
                    <FileText className="h-5 w-5 text-muted-foreground" />
                    <div className="space-y-0.5">
                      <p className="text-sm font-medium">{file.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {(file.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => removeFile(index)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {uploads.length > 0 && (
        <Card>
          <CardContent className="p-4">
            <h3 className="mb-3 font-medium">{t("processingStatus", "Processing Status")}</h3>
            <div className="space-y-4">
              {uploads.map((upload, index) => (
                <div key={index} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {uploadIcon(upload.status)}
                      <span className="text-sm font-medium truncate max-w-[200px]">
                        {upload.name}
                      </span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {upload.message}
                    </span>
                  </div>
                  <Progress value={upload.progress} className="h-2" />
                  {upload.status === 'completed' && completedUploads.has(index) && (
                    <SuggestionCard
                      index={index}
                      completed={completedUploads.get(index)!}
                      t={t}
                      onDismiss={() => {
                        setCompletedUploads((prev) => {
                          const next = new Map(prev);
                          next.delete(index);
                          return next;
                        });
                      }}
                    />
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {status === "completed" && uploads.length === 0 && (
        <Card className="border-green-500 bg-green-500/10">
          <CardContent className="flex items-center gap-3 p-4 text-green-700 dark:text-green-400">
            <CheckCircle2 className="h-5 w-5" />
            <p>{tCommon("uploadSuccess")}</p>
          </CardContent>
        </Card>
      )}

      {files.length > 0 && status !== "uploading" && status !== "processing" && (
        <Button className="w-full" size="lg" onClick={upload}>
          {files.length === 1
            ? t("uploadButton", { count: files.length })
            : t("uploadButtonPlural", { count: files.length })}
        </Button>
      )}

      {(status === "uploading" || status === "processing") && (
        <div className="flex items-center justify-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary" />
          <span>{status === "uploading" ? tCommon("uploading") : "Processing..."}</span>
        </div>
      )}
    </div>
  );
}

function SuggestionCard({
  index,
  completed,
  t,
  onDismiss,
}: {
  index: number;
  completed: CompletedUpload;
  t: ReturnType<typeof useTranslations>;
  onDismiss: () => void;
}) {
  const [loading, setLoading] = useState(false);

  const handleAccept = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem("access_token");
      await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/documents/${completed.id}/move`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ folder_id: null }),
      });
      toast.success(t("suggestionAccepted", "Suggestion accepted"));
      onDismiss();
    } catch {
      toast.error(t("suggestionFailed", "Failed to apply suggestion"));
    } finally {
      setLoading(false);
    }
  };

  if (!completed.suggestedFolderName && !completed.suggestedFilename) {
    return null;
  }

  return (
    <div className="rounded-lg border bg-muted/50 p-3 space-y-2">
      <p className="text-xs font-medium text-muted-foreground">{t("suggestionLabel", "Suggestion")}</p>
      {completed.suggestedFolderName && (
        <div className="flex items-center gap-2 text-sm">
          <FolderOpen className="h-4 w-4 text-muted-foreground" />
          <span>{t("suggestedFolder", "Suggested folder")}: </span>
          <Badge variant="outline">{completed.suggestedFolderName}</Badge>
        </div>
      )}
      {completed.suggestedFilename && (
        <div className="flex items-center gap-2 text-sm">
          <FilePen className="h-4 w-4 text-muted-foreground" />
          <span>{t("suggestedName", "Suggested name")}: </span>
          <Badge variant="outline">{completed.suggestedFilename}</Badge>
        </div>
      )}
      {completed.suggestionConfidence != null && (
        <p className="text-xs text-muted-foreground">
          {t("confidence", "Confidence")}: {(completed.suggestionConfidence * 100).toFixed(0)}%
        </p>
      )}
      <div className="flex gap-2 pt-1">
        <Button size="sm" variant="default" onClick={handleAccept} disabled={loading}>
          {t("acceptSuggestion", "Accept")}
        </Button>
        <Button size="sm" variant="ghost" onClick={onDismiss}>
          {tCommon("cancel")}
        </Button>
      </div>
    </div>
  );
}
