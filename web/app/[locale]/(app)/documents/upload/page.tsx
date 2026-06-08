'use client';

import { useTranslations } from "next-intl";
import { Upload, X, FileText, CheckCircle2, AlertCircle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useUploadWithProgress, ALLOWED_FILE_TYPES, MAX_FILE_SIZE, type UploadStatus } from "@/features/documents/useUploadWithProgress";
import { toast } from "sonner";

export default function UploadPage() {
  const t = useTranslations("documents");
  const tCommon = useTranslations("common");

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
  } = useUploadWithProgress();

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
            <h3 className="mb-3 font-medium">{t("processingStatus")}</h3>
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
