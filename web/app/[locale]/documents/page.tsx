"use client";

import { useState, useEffect, useMemo, useCallback } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { format } from "date-fns";
import {
  FileText,
  FileCheck,
  Receipt,
  ShieldCheck,
  Fingerprint,
  Search,
  FolderOpen,
  Clock,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { FolderExplorer } from "@/components/FolderExplorer";
import { listDocuments, type Document } from "@/features/documents/api";
import { listAllFolders } from "@/lib/api/folders";
import type { Folder } from "@/lib/api/types";
import { useAuth } from "@/lib/useAuth";
import { cn } from "@/lib/utils";

const DOC_TYPE_ICONS: Record<string, typeof FileText> = {
  contract: FileCheck,
  invoice: Receipt,
  warranty: ShieldCheck,
  identity: Fingerprint,
};

const STATUS_COLORS: Record<string, string> = {
  processed: "bg-emerald-600 hover:bg-emerald-700",
  processing: "bg-amber-500 hover:bg-amber-600",
  pending: "bg-amber-500 hover:bg-amber-600",
  failed: "bg-red-600 hover:bg-red-700",
};

function DocTypeIcon({ docType, className }: { docType: string; className?: string }) {
  const Icon = DOC_TYPE_ICONS[docType] ?? FileText;
  return <Icon className={className ?? "size-4 shrink-0 text-muted-foreground"} />;
}

function StatusBadge({ status }: { status: string }) {
  return (
    <Badge className={`text-white ${STATUS_COLORS[status] ?? ""}`}>
      {status}
    </Badge>
  );
}

function RecentCard({
  doc,
  link,
}: {
  doc: Document;
  link: string;
}) {
  return (
    <Link href={link}>
      <Card className="flex items-start gap-3 overflow-hidden p-3 transition-shadow hover:shadow-sm">
        <DocTypeIcon docType={doc.doc_type} className="size-5 shrink-0 mt-0.5" />
        <div className="min-w-0 flex-1 overflow-hidden">
          <p className="truncate text-sm font-medium leading-tight">
            {doc.title}
          </p>
          <p className="text-xs text-muted-foreground capitalize mt-0.5">
            {doc.doc_type}
          </p>
          <p className="text-xs text-muted-foreground">
            {format(new Date(doc.created_at), "MMM dd yyyy")}
          </p>
        </div>
      </Card>
    </Link>
  );
}

function PageSkeleton() {
  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div className="flex items-center justify-between">
        <Skeleton className="h-9 w-40" />
        <Skeleton className="h-9 w-36" />
      </div>
      <div className="flex gap-6">
        <div className="w-[260px] shrink-0 space-y-4">
          <Skeleton className="h-9 w-full" />
          <Skeleton className="h-9 w-full" />
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-8 w-3/4" />
          ))}
        </div>
        <div className="flex-1 space-y-6">
          <div className="grid grid-cols-3 gap-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-20 w-full" />
            ))}
          </div>
          <div className="space-y-2 rounded-md border p-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function DocumentsPage() {
  const t = useTranslations("documents");
  const { isLoading: authLoading, isAuthenticated } = useAuth();

  const [documents, setDocuments] = useState<Document[]>([]);
  const [folders, setFolders] = useState<Folder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [searchQuery, setSearchQuery] = useState("");
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      setError(null);
      const [docsRes, foldersRes] = await Promise.all([
        listDocuments({ limit: 1000 }),
        listAllFolders(),
      ]);
      setDocuments(docsRes.documents ?? []);
      setFolders(foldersRes.folders ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load");
    }
  }, []);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      setLoading(false);
      return;
    }

    let cancelled = false;

    async function init() {
      setLoading(true);
      await loadData();
      if (!cancelled) setLoading(false);
    }

    init();

    return () => {
      cancelled = true;
    };
  }, [authLoading, isAuthenticated, loadData]);

  const foldersMap = useMemo(() => {
    const map: Record<string, string> = {};
    for (const f of folders) {
      map[f.id] = f.name;
    }
    return map;
  }, [folders]);

  const recentDocs = useMemo(
    () =>
      [...documents]
        .sort(
          (a, b) =>
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        )
        .slice(0, 3),
    [documents],
  );

  const filteredDocs = useMemo(() => {
    let list =
      selectedFolderId === null
        ? documents
        : documents.filter((d) => d.folder_id === selectedFolderId);

    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      list = list.filter((d) => d.title.toLowerCase().includes(q));
    }

    return list;
  }, [documents, selectedFolderId, searchQuery]);

  if (loading) {
    return <PageSkeleton />;
  }

  if (error) {
    return (
      <div className="mx-auto max-w-6xl">
        <div className="rounded-lg bg-destructive/10 p-4 text-center text-destructive">
          {error}
        </div>
      </div>
    );
  }

  const isAllDocs = selectedFolderId === null;
  const selectedFolderName =
    selectedFolderId && foldersMap[selectedFolderId]
      ? foldersMap[selectedFolderId]
      : null;

  return (
    <div className="mx-auto max-w-6xl">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">{t("title")}</h1>
        <Button asChild>
          <Link href="/documents/upload">{t("upload")}</Link>
        </Button>
      </div>

      <div className="flex gap-6">
        <div className="w-[260px] shrink-0 space-y-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="ps-9"
            />
          </div>

          <div className="space-y-1">
            <button
              type="button"
              onClick={() => setSelectedFolderId(null)}
              className={cn(
                "flex w-full items-center gap-2 rounded-md px-3 py-2 text-start text-sm transition-colors",
                isAllDocs
                  ? "bg-primary/10 font-medium text-primary"
                  : "hover:bg-accent",
              )}
            >
              <FileText className="size-4 shrink-0" />
              <span>{t("allDocuments")}</span>
            </button>
          </div>

          <FolderExplorer
            selectedFolderId={selectedFolderId ?? undefined}
            onFolderSelect={(id) => setSelectedFolderId(id)}
            documents={[]}
            onRefresh={loadData}
          />
        </div>

        <div className="min-w-0 flex-1">
          {isAllDocs && recentDocs.length > 0 && (
            <div className="mb-6">
              <div className="mb-3 flex items-center gap-2">
                <Clock className="size-4 text-muted-foreground" />
                <h2 className="text-sm font-medium uppercase tracking-wide text-muted-foreground">
                  {t("recent")}
                </h2>
              </div>
              <div className="grid grid-cols-3 gap-3">
                {recentDocs.map((doc) => (
                  <RecentCard
                    key={doc.id}
                    doc={doc}
                    link={`/documents/${doc.id}`}
                  />
                ))}
              </div>
            </div>
          )}

          {selectedFolderName && (
            <h2 className="mb-3 text-lg font-semibold">{selectedFolderName}</h2>
          )}

          {filteredDocs.length === 0 ? (
            <Card className="flex flex-col items-center justify-center gap-3 p-12">
              <FolderOpen className="size-10 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">
                {searchQuery ? t("noMatchingDocuments") : t("noDocuments")}
              </p>
            </Card>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-10" />
                    <TableHead>{t("titleColumn")}</TableHead>
                    <TableHead>{t("folder")}</TableHead>
                    <TableHead>{t("typeColumn")}</TableHead>
                    <TableHead>{t("dateColumn")}</TableHead>
                    <TableHead>{t("statusColumn")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredDocs.map((doc) => (
                    <TableRow key={doc.id}>
                      <TableCell>
                        <DocTypeIcon docType={doc.doc_type} />
                      </TableCell>
                  <TableCell className="max-w-[300px]">
                    <Link
                      href={`/documents/${doc.id}`}
                      className="block truncate font-medium hover:underline"
                    >
                      {doc.title}
                    </Link>
                  </TableCell>
                      <TableCell className="text-muted-foreground">
                        {doc.folder_id
                          ? (foldersMap[doc.folder_id] ?? t("unknown"))
                          : "—"}
                      </TableCell>
                      <TableCell className="capitalize">{doc.doc_type}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {format(new Date(doc.created_at), "MMM dd yyyy")}
                      </TableCell>
                      <TableCell>
                        <StatusBadge status={doc.status} />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
