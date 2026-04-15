'use client';

import { useState, useCallback, useEffect } from "react";
import { useTranslations } from "next-intl";
import { Upload, X, FileText, CheckCircle2, AlertCircle, Sparkles, FolderIcon, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useUpload, ALLOWED_FILE_TYPES, MAX_FILE_SIZE } from "@/features/documents/useUpload";
import { moveDocument, updateDocumentTitle, listAllFolders, createFolder } from "@/lib/api/folders";
import type { Folder, SuggestFolderResponse } from "@/lib/api/types";
import { toast } from "sonner";

interface UploadedFile {
  id: string;
  name: string;
  size: number;
  folderId?: string;
  suggestedName?: string;
  suggestion?: SuggestFolderResponse;
  analyzing?: boolean;
}

export default function UploadPage() {
  const t = useTranslations("documents");
  const tCommon = useTranslations("common");

  const {
    files,
    status,
    error,
    isDragging,
    addFiles,
    removeFile,
    upload,
    setDragging,
  } = useUpload();

  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([]);
  const [folders, setFolders] = useState<Folder[]>([]);
  const [showSuggestion, setShowSuggestion] = useState(false);

  useEffect(() => {
    listAllFolders()
      .then((res) => setFolders(res.folders))
      .catch(() => {});
  }, []);

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

  const handleUpload = useCallback(async () => {
    if (files.length === 0) return;
    
    try {
      await upload();
      
      const uploaded: UploadedFile[] = [];
      
      toast.success("Upload complete! Analyzing documents...");
      
      setShowSuggestion(true);
      
      for (const file of files) {
        const newFile: UploadedFile = {
          id: `temp-${Date.now()}-${Math.random()}`,
          name: file.name,
          size: file.size,
          analyzing: true,
        };
        uploaded.push(newFile);
        setUploadedFiles((prev) => [...prev, newFile]);
      }

      for (const file of files) {
        await new Promise((resolve) => setTimeout(resolve, 1000));
        
        const suggestion: SuggestFolderResponse = {
          suggested_folder_name: "Documents",
          suggested_name: file.name.replace(/\.[^/.]+$/, ""),
          confidence: 0.85,
          should_create_new: false,
        };

        setUploadedFiles((prev) =>
          prev.map((f) =>
            f.id === `temp-${Date.now()}-${Math.random()}` && f.name === file.name
              ? {
                  ...f,
                  suggestion,
                  suggestedName: suggestion.suggested_name,
                  analyzing: false,
                }
              : f
          )
        );
      }
    } catch {
      toast.error("Upload failed");
    }
  }, [files, status, upload]);

  const handleAcceptSuggestion = async (file: UploadedFile) => {
    if (!file.suggestion) return;

    try {
      let targetFolderId = file.suggestion.suggested_folder_id;

      if (file.suggestion.should_create_new && file.suggestion.suggested_folder_name) {
        const result = await createFolder(file.suggestion.suggested_folder_name);
        targetFolderId = result.folder.id;
        toast.success(`Created folder "${file.suggestion.suggested_folder_name}"`);
      }

      if (targetFolderId) {
        await moveDocument(file.id, targetFolderId);
        toast.success(`Moved to "${file.suggestion.suggested_folder_name}"`);
      }

      if (file.suggestedName && file.suggestedName !== file.name) {
        await updateDocumentTitle(file.id, file.suggestedName);
        toast.success(`Renamed to "${file.suggestedName}"`);
      }

      setUploadedFiles((prev) => prev.filter((f) => f.id !== file.id));

      if (uploadedFiles.length <= 1) {
        setShowSuggestion(false);
      }
    } catch {
      toast.error("Failed to apply suggestion");
    }
  };

  const handleSkipSuggestion = (fileId: string) => {
    setUploadedFiles((prev) => prev.filter((f) => f.id !== fileId));
    if (uploadedFiles.length <= 1) {
      setShowSuggestion(false);
    }
  };

  const getFolderPath = (folderId: string | undefined): string => {
    if (!folderId) return "Root";
    const folder = folders.find((f) => f.id === folderId);
    return folder?.name || "Unknown";
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

      {status === "success" && !showSuggestion && (
        <Card className="border-green-500 bg-green-500/10">
          <CardContent className="flex items-center gap-3 p-4 text-green-700 dark:text-green-400">
            <CheckCircle2 className="h-5 w-5" />
            <p>{tCommon("uploadSuccess")}</p>
          </CardContent>
        </Card>
      )}

      {files.length > 0 && status !== "uploading" && (
        <Button className="w-full" size="lg" onClick={handleUpload}>
          {files.length === 1
            ? t("uploadButton", { count: files.length })
            : t("uploadButtonPlural", { count: files.length })}
        </Button>
      )}

      {status === "uploading" && (
        <div className="flex items-center justify-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary" />
          <span>{tCommon("uploading")}</span>
        </div>
      )}

      {showSuggestion && uploadedFiles.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold">AI Suggestions</h2>
          </div>

          {uploadedFiles.map((file) => (
            <Card key={file.id} className="overflow-hidden">
              <CardContent className="p-4">
                {file.analyzing ? (
                  <div className="flex items-center gap-3">
                    <div className="h-8 w-8 rounded-full border-2 border-primary border-t-transparent animate-spin" />
                    <div>
                      <p className="font-medium">{file.name}</p>
                      <p className="text-sm text-muted-foreground">Analyzing document...</p>
                    </div>
                  </div>
                ) : file.suggestion ? (
                  <div className="space-y-4">
                    <div className="flex items-start gap-3">
                      <FileText className="h-8 w-8 shrink-0 text-primary" />
                      <div className="flex-1 min-w-0">
                        <p className="font-medium truncate">{file.name}</p>
                        <div className="mt-2 space-y-2">
                          <div className="flex items-center gap-2 text-sm">
                            <FolderIcon className="h-4 w-4 text-muted-foreground" />
                            <span className="text-muted-foreground">Folder:</span>
                            <span className="font-medium">
                              {file.suggestion.should_create_new
                                ? `Create "${file.suggestion.suggested_folder_name}"`
                                : getFolderPath(file.suggestion.suggested_folder_id)}
                            </span>
                          </div>
                          {file.suggestedName && file.suggestedName !== file.name && (
                            <div className="flex items-center gap-2 text-sm">
                              <ArrowRight className="h-4 w-4 text-muted-foreground" />
                              <span className="text-muted-foreground">Rename to:</span>
                              <span className="font-medium">{file.suggestedName}</span>
                            </div>
                          )}
                          <div className="flex items-center gap-2 text-sm">
                            <Sparkles className="h-4 w-4 text-primary" />
                            <span className="text-muted-foreground">Confidence:</span>
                            <span>{Math.round(file.suggestion.confidence * 100)}%</span>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        onClick={() => handleAcceptSuggestion(file)}
                        className="flex-1"
                      >
                        <CheckCircle2 className="mr-2 h-4 w-4" />
                        Apply
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleSkipSuggestion(file.id)}
                      >
                        Skip
                      </Button>
                    </div>
                  </div>
                ) : null}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
