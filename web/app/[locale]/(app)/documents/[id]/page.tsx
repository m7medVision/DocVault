"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Loader2,
  Download,
  Eye,
  FileText,
  Languages,
  PanelRight,
  Calendar,
  Bell,
  MessageSquare,
  History,
} from "lucide-react";
import { DocumentViewer } from "@/components/DocumentViewer";
import { VersionTimeline } from "@/components/VersionTimeline";
import { FilePreview } from "@/components/FilePreview";
import { useDocumentDetail } from "@/features/documents/useDocumentDetail";
import { useDocumentReminders } from "@/features/documents/useDocumentReminders";
import { ChatPanel } from "@/components/ChatPanel";

export default function DocumentDetailPage() {
  const params = useParams();
  const documentId = params.id as string;
  const t = useTranslations("documents");
  const tCommon = useTranslations("common");
  const tVersions = useTranslations("versions");

  const { document, loading, error, presignedUrl, handleDownload, handleUpdateMetadata, handleRenameTitle } =
    useDocumentDetail(documentId);

  const [selectedPage, setSelectedPage] = useState(0);
  const [editingMetadata, setEditingMetadata] = useState<string | null>(null);
  const [correctedValue, setCorrectedValue] = useState("");
  const [metadataOpen, setMetadataOpen] = useState(false);
  const [remindersOpen, setRemindersOpen] = useState(false);
  const [chatOpen, setChatOpen] = useState(false);
  const [editingTitle, setEditingTitle] = useState(false);
  const [titleValue, setTitleValue] = useState("");

  const { reminders } = useDocumentReminders(documentId);

  const extractedDates = document?.metadata?.filter(
    (m) =>
      m.key.includes("date") ||
      m.key.includes("expiry") ||
      m.key.includes("deadline") ||
      m.key.includes("due")
  );

  const handleEditMetadata = (key: string, currentValue?: string) => {
    setEditingMetadata(key);
    setCorrectedValue(currentValue || "");
  };

  const handleSaveMetadata = async () => {
    if (!editingMetadata) return;
    await handleUpdateMetadata(editingMetadata, correctedValue);
    setEditingMetadata(null);
    setCorrectedValue("");
  };

  const handleEditTitle = () => {
    setTitleValue(globalThis.document?.title ?? "");
    setEditingTitle(true);
  };

  const handleSaveTitle = async () => {
    const trimmed = titleValue.trim();
    if (!trimmed) return;
    const ok = await handleRenameTitle(trimmed);
    if (ok) setEditingTitle(false);
  };

  const handleCancelTitle = () => {
    setEditingTitle(false);
    setTitleValue("");
  };

  const getStatusVariant = (status: string) => {
    switch (status) {
      case "processed":
        return "default" as const;
      case "pending":
        return "secondary" as const;
      case "failed":
        return "destructive" as const;
      default:
        return "outline" as const;
    }
  };

  const hasTranslation = document?.language && document.language !== "en";
  const latestVersion = document?.versions?.[document.versions.length - 1];

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !document) {
    return (
      <div className="rounded-lg bg-destructive/10 p-4 text-center text-destructive">
        {error || "Document not found"}
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div className="space-y-2">
        <div className="flex items-center gap-3">
          {editingTitle ? (
            <div className="flex items-center gap-2">
              <Input
                value={titleValue}
                onChange={(e) => setTitleValue(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleSaveTitle();
                  if (e.key === 'Escape') handleCancelTitle();
                }}
                className="text-2xl font-bold h-auto py-1"
                autoFocus
              />
              <Button size="sm" onClick={handleSaveTitle}>{tCommon("save")}</Button>
              <Button size="sm" variant="outline" onClick={handleCancelTitle}>{tCommon("cancel")}</Button>
            </div>
          ) : (
            <h1 className="text-3xl font-bold cursor-pointer hover:text-primary transition-colors" onClick={handleEditTitle}>{document.title}</h1>
          )}
          <Badge variant={getStatusVariant(document.status)}>
            {document.status}
          </Badge>
        </div>
        <div className="flex gap-4 text-sm text-muted-foreground">
          <span className="capitalize">{document.doc_type}</span>
          <span className="uppercase">{document.language || "-"}</span>
        </div>
      </div>

      <Tabs defaultValue="document" className="space-y-4">
        <TabsList>
          <TabsTrigger value="document" className="gap-1.5">
            <FileText className="size-4" />
            {t("documentTab")}
          </TabsTrigger>
          <TabsTrigger value="ocr" className="gap-1.5">
            <Eye className="size-4" />
            {t("ocrVersion")}
          </TabsTrigger>
          {hasTranslation && (
            <TabsTrigger value="translated" className="gap-1.5">
              <Languages className="size-4" />
              {t("translatedVersion")}
            </TabsTrigger>
          )}
          <TabsTrigger value="versions" className="gap-1.5">
            <History className="size-4" />
            {tVersions("title")}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="document">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              {latestVersion && (
                <div className="flex gap-4 text-sm text-muted-foreground">
                  <span>{latestVersion.mime_type}</span>
                  <span>{(latestVersion.file_size / 1024).toFixed(1)} KB</span>
                  <span>{new Date(latestVersion.created_at).toLocaleDateString()}</span>
                </div>
              )}
              <div className="ms-auto flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  className="gap-1.5"
                  onClick={() => setChatOpen(true)}
                >
                  <MessageSquare className="size-4" />
                  {t("chatWithDoc")}
                </Button>
                {(reminders.length > 0 || (extractedDates && extractedDates.length > 0)) && (
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-1.5"
                    onClick={() => setRemindersOpen(true)}
                  >
                    <Bell className="size-4" />
                    {t("viewDates")}
                  </Button>
                )}
                {document.metadata && document.metadata.length > 0 && (
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-1.5"
                    onClick={() => setMetadataOpen(true)}
                  >
                    <PanelRight className="size-4" />
                    {t("metadata")}
                  </Button>
                )}
                <Button
                  variant="outline"
                  size="sm"
                  className="gap-1.5"
                  onClick={handleDownload}
                >
                  <Download className="size-4" />
                  {t("downloadOriginal")}
                </Button>
              </div>
            </div>

            <FilePreview
              url={presignedUrl}
              mimeType={latestVersion?.mime_type || ""}
              fileSize={latestVersion?.file_size}
              t={t}
            />
          </div>
        </TabsContent>

        <TabsContent value="ocr">
          {document.pages && document.pages.length > 0 ? (
            <DocumentViewer
              pages={document.pages}
              selectedPage={selectedPage}
              onPageChange={setSelectedPage}
              noOcrText={tCommon("noOcrText")}
              confidenceLabel={tCommon("confidence")}
              toolbarActions={
                document.metadata && document.metadata.length > 0 ? (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="gap-1.5"
                    onClick={() => setMetadataOpen(true)}
                  >
                    <PanelRight className="size-4" />
                    {t("metadata")}
                  </Button>
                ) : undefined
              }
            />
          ) : (
            <Card>
              <CardContent className="py-12 text-center text-sm text-muted-foreground">
                {tCommon("noOcrText")}
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {hasTranslation && (
          <TabsContent value="translated">
            <div className="flex flex-col items-center gap-6">
              {document.pages && document.pages.length > 0 ? (
                document.pages.map((page) => (
                  <div key={page.id} className="w-full max-w-3xl space-y-2">
                    <div className="flex items-center gap-2 px-1">
                      <span className="text-sm font-medium text-muted-foreground">
                        {tCommon("page")} {page.page_number}
                      </span>
                    </div>
                    <div className="overflow-hidden rounded-lg border bg-white shadow-md min-h-[300px]">
                      <div className="mx-auto max-w-none px-10 py-10 [color:oklch(0.15_0.02_250)]">
                        <div className="prose prose-sm max-w-none prose-headings:mb-3 prose-headings:mt-6 prose-p:leading-relaxed prose-table:text-sm">
                          {page.translated_text ? (
                            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                              {page.translated_text}
                            </ReactMarkdown>
                          ) : (
                            <p className="italic text-muted-foreground">
                              {t("noTranslation")}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <Card>
                  <CardContent className="py-12 text-center text-sm text-muted-foreground">
                    {t("noTranslation")}
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>
        )}

        <TabsContent value="versions">
          <VersionTimeline versions={document.versions ?? []} />
        </TabsContent>
      </Tabs>

      <Sheet open={metadataOpen} onOpenChange={setMetadataOpen}>
        <SheetContent side="right" className="w-[400px] sm:max-w-[400px] overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{t("metadata")}</SheetTitle>
            <SheetDescription>
              {document.title}
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-4 px-4 pb-8">
            {document.metadata && document.metadata.length > 0 ? (
              document.metadata.map((item) => (
                <div key={item.key} className="space-y-2">
                  <label className="text-xs uppercase text-muted-foreground">
                    {item.key.replace(/_/g, " ")}
                  </label>
                  {editingMetadata === item.key ? (
                    <div className="space-y-2">
                      <Input
                        value={correctedValue}
                        onChange={(e) => setCorrectedValue(e.target.value)}
                      />
                      <div className="flex gap-2">
                        <Button size="sm" onClick={handleSaveMetadata}>
                          {tCommon("save")}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setEditingMetadata(null)}
                        >
                          {tCommon("cancel")}
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-center justify-between">
                      <span className="text-sm">
                        {item.corrected_value || item.extracted_value || "-"}
                      </span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          handleEditMetadata(
                            item.key,
                            item.corrected_value || item.extracted_value
                          )
                        }
                      >
                        {tCommon("edit")}
                      </Button>
                    </div>
                  )}
                  <Separator />
                </div>
              ))
            ) : (
              <p className="text-sm text-muted-foreground">
                {tCommon("noMetadata")}
              </p>
            )}
          </div>
        </SheetContent>
      </Sheet>

      <Sheet open={remindersOpen} onOpenChange={setRemindersOpen}>
        <SheetContent side="bottom" className="max-h-[60vh] overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="flex items-center gap-2">
              <Calendar className="size-5" />
              {t("datesAndReminders")}
            </SheetTitle>
            <SheetDescription>
              {document.title}
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-4 px-4 pt-4">
            {reminders.length > 0 && (
              <div className="space-y-3">
                <h3 className="text-sm font-medium flex items-center gap-2">
                  <Bell className="size-4" />
                  {t("reminders")}
                </h3>
                {reminders.map((reminder) => (
                  <div key={reminder.id} className="flex items-center justify-between rounded-lg border p-3">
                    <div>
                      <p className="text-sm font-medium">
                        {new Date(reminder.trigger_date).toLocaleDateString()}
                      </p>
                      <p className="text-xs text-muted-foreground capitalize">
                        {reminder.rule_type.replace(/_/g, " ")}
                      </p>
                    </div>
                    <Badge variant={reminder.status === "pending" ? "default" : "secondary"}>
                      {reminder.days_until > 0
                        ? t("inDays", { count: reminder.days_until })
                        : reminder.days_until === 0
                        ? t("today")
                        : t("overdue")}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
            {extractedDates && extractedDates.length > 0 && (
              <div className="space-y-3">
                <h3 className="text-sm font-medium flex items-center gap-2">
                  <Calendar className="size-4" />
                  {t("extractedDates")}
                </h3>
                {extractedDates.map((item) => (
                  <div key={item.key} className="flex items-center justify-between rounded-lg border p-3">
                    <div>
                      <p className="text-xs uppercase text-muted-foreground">
                        {item.key.replace(/_/g, " ")}
                      </p>
                      <p className="text-sm font-medium">
                        {item.corrected_value || item.extracted_value}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            )}
            {reminders.length === 0 && (!extractedDates || extractedDates.length === 0) && (
              <p className="text-sm text-muted-foreground text-center py-8">
                {t("noDatesOrReminders")}
              </p>
            )}
          </div>
        </SheetContent>
      </Sheet>

      <Sheet open={chatOpen} onOpenChange={setChatOpen}>
        <SheetContent side="right" className="w-[500px] sm:max-w-[500px] p-0 flex flex-col">
          <SheetHeader className="px-4 pt-4 pb-2">
            <SheetTitle className="flex items-center gap-2">
              <MessageSquare className="size-5" />
              {t("chatWithDoc")}
            </SheetTitle>
            <SheetDescription>
              {document.title}
            </SheetDescription>
          </SheetHeader>
          <div className="flex-1 min-h-0">
            <ChatPanel documentId={documentId} />
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
