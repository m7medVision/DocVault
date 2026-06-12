"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import { Tags as TagsIcon, Search, Loader2, FileText } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useTags } from "@/features/tags/useTags";
import { searchDocuments } from "@/lib/api/search";
import type { Tag } from "@/lib/api/tags";

export default function TagsPage() {
  const t = useTranslations("tags");
  const pathname = usePathname();
  const localePrefix = pathname.split("/")[1] || "en";

  const [query, setQuery] = useState("");
  const [selectedTag, setSelectedTag] = useState<Tag | null>(null);

  const { data: tagsData, isLoading } = useTags(query);
  const tags = tagsData?.tags ?? [];

  const { data: results, isLoading: searching } = useQuery({
    queryKey: ["tagSearch", selectedTag?.id],
    queryFn: () =>
      searchDocuments({ query: selectedTag!.name, tags: [selectedTag!.id], limit: 50 }),
    enabled: !!selectedTag,
  });

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-3xl font-bold">
          <TagsIcon className="size-7" />
          {t("title")}
        </h1>
        <p className="text-muted-foreground">{t("subtitle")}</p>
      </div>

      <div className="relative max-w-md">
        <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={t("searchPlaceholder")}
          className="ps-9"
        />
      </div>

      {isLoading ? (
        <div className="flex h-24 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : tags.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noTags")}</p>
      ) : (
        <div className="flex flex-wrap gap-2">
          {tags.map((tag) => (
            <button key={tag.id} onClick={() => setSelectedTag(tag)}>
              <Badge
                variant={selectedTag?.id === tag.id ? "default" : "secondary"}
                className="cursor-pointer px-3 py-1 text-sm"
              >
                {tag.name}
              </Badge>
            </button>
          ))}
        </div>
      )}

      {selectedTag && (
        <div className="space-y-3 pt-2">
          <h2 className="text-lg font-semibold">
            {t("documentsWithTag", { tag: selectedTag.name })}
          </h2>
          {searching ? (
            <div className="flex h-24 items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : !results || results.results.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("noDocuments")}</p>
          ) : (
            <div className="space-y-2">
              {results.results.map((r) => (
                <Link key={r.document_id} href={`/${localePrefix}/documents/${r.document_id}`}>
                  <Card className="transition-colors hover:bg-muted/50">
                    <CardContent className="flex items-center gap-3 p-4">
                      <FileText className="size-5 text-muted-foreground" />
                      <span className="truncate font-medium">{r.file}</span>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
