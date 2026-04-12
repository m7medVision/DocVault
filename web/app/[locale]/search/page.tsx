"use client";

import { Suspense } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { useSearch } from "@/features/search/useSearch";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Loader2 } from "lucide-react";

function SearchResultsContent() {
  const t = useTranslations("search");
  const tCommon = useTranslations("common");
  const {
    results,
    loading,
    error,
    query,
    setQuery,
    filters,
    setFilter,
    performSearch,
  } = useSearch();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    performSearch(query, filters);
  };

  const getScoreColor = (score: number) => {
    if (score >= 0.9) return "#10B981";
    if (score >= 0.7) return "#F59E0B";
    return "#EF4444";
  };

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <form onSubmit={handleSearch} className="flex gap-3">
        <Input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={t("placeholder")}
          className="flex-1"
        />
        <Button type="submit">{t("search")}</Button>
      </form>

      <div className="flex gap-4">
        <Select
          value={filters.doc_type}
          onValueChange={(value) => setFilter("doc_type", value)}
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
          value={filters.language}
          onValueChange={(value) => setFilter("language", value)}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t("language")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="en">{t("english")}</SelectItem>
            <SelectItem value="ar">{t("arabic")}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {loading && (
        <div className="flex h-64 items-center justify-center gap-3">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          <span>{tCommon("searching")}</span>
        </div>
      )}

      {error && (
        <div className="rounded-lg bg-destructive/10 p-4 text-destructive">
          {error}
        </div>
      )}

      {!loading && !error && (!results || results.length === 0) && query && (
        <Card className="p-8 text-center">
          <p className="text-muted-foreground">{t("noResults")}</p>
        </Card>
      )}

      {!loading && results && results.length > 0 && (
        <p className="text-sm text-muted-foreground">
          {t("resultsCount", { count: results.length })}
        </p>
      )}

      <div className="space-y-4">
        {results.map((result, index) => (
          <Card key={`${result.document_id}-${index}`}>
            <CardContent className="space-y-4 p-4">
              <div className="flex items-start justify-between">
                <div
                  className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full text-white"
                  style={{ backgroundColor: getScoreColor(result.max_score) }}
                >
                  {(result.max_score * 100).toFixed(0)}%
                </div>
                <div className="flex flex-1 items-center px-4">
                  <span className="text-lg font-medium">{result.file}</span>
                </div>
              </div>

              <div className="flex justify-end">
                <Button variant="outline" size="sm" asChild>
                  <Link href={`/documents/${result.document_id}`}>
                    {tCommon("viewDocument")}
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="flex h-64 items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <SearchResultsContent />
    </Suspense>
  );
}
