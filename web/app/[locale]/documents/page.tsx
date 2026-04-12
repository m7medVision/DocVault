"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import {
  FileText,
  FileCheck,
  Receipt,
  ShieldCheck,
  Fingerprint,
  FolderOpen,
  ChevronRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Loader2 } from "lucide-react";
import { useDocumentList, type DocumentFilters } from "@/features/documents/useDocumentList";

const DOC_TYPE_ICONS: Record<string, typeof FileText> = {
  contract: FileCheck,
  invoice: Receipt,
  warranty: ShieldCheck,
  identity: Fingerprint,
};

function DocTypeIcon({ docType }: { docType: string }) {
  const Icon = DOC_TYPE_ICONS[docType] ?? FileText;
  return <Icon className="size-6" />;
}

export default function DocumentsPage() {
  const t = useTranslations("documents");
  const [filter, setFilter] = useState<DocumentFilters>({ type: "", status: "" });

  const { documents, loading, error } = useDocumentList(filter);

  const handleFilterChange = (key: string, value: string) => {
    setFilter((prev) => ({ ...prev, [key]: value }));
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

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
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
    <div className="mx-auto max-w-5xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">{t("title")}</h1>
        <Button asChild>
          <Link href="/documents/upload">{t("upload")}</Link>
        </Button>
      </div>

      <div className="flex gap-4">
        <Select
          value={filter.type}
          onValueChange={(value) => handleFilterChange("type", value)}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t("docType")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="contract">{t("contract")}</SelectItem>
            <SelectItem value="invoice">{t("invoice")}</SelectItem>
            <SelectItem value="warranty">{t("warranty")}</SelectItem>
            <SelectItem value="identity">{t("identity")}</SelectItem>
          </SelectContent>
        </Select>

        <Select
          value={filter.status}
          onValueChange={(value) => handleFilterChange("status", value)}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t("status")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="pending">{t("pending")}</SelectItem>
            <SelectItem value="processed">{t("processed")}</SelectItem>
            <SelectItem value="failed">{t("failed")}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {documents.length === 0 ? (
        <Card className="flex flex-col items-center justify-center gap-3 p-12">
          <FolderOpen className="size-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">{t("noDocuments")}</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {documents.map((doc) => (
            <Link key={doc.id} href={`/documents/${doc.id}`} className="group">
              <Card className="flex h-full cursor-pointer flex-col overflow-hidden transition-shadow hover:shadow-md">
                <div className="flex items-center justify-center bg-muted p-6">
                  <DocTypeIcon docType={doc.doc_type} />
                </div>

                <div className="flex flex-1 flex-col gap-1 p-4">
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="line-clamp-1 font-medium leading-tight">
                      {doc.title}
                    </h3>
                    <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
                  </div>

                  <p className="text-sm capitalize text-muted-foreground">
                    {doc.doc_type} &middot;{" "}
                    {new Date(doc.created_at).toLocaleDateString()}
                  </p>

                  <div className="mt-auto flex items-center justify-between pt-2">
                    {doc.language && (
                      <span className="text-xs uppercase tracking-wide text-muted-foreground">
                        {doc.language}
                      </span>
                    )}
                    <Badge
                      variant={getStatusVariant(doc.status)}
                      className="ml-auto"
                    >
                      {doc.status}
                    </Badge>
                  </div>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
