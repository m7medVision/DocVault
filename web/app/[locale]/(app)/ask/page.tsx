"use client";

import { useTranslations } from "next-intl";
import { ChatPanel } from "@/components/ChatPanel";

export default function AskPage() {
  const t = useTranslations("ask");

  return (
    <div className="mx-auto flex h-[calc(100vh-10rem)] max-w-3xl flex-col">
      <div className="mb-4 space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{t("title")}</h1>
        <p className="text-sm text-muted-foreground">{t("description")}</p>
      </div>
      <div className="flex-1 overflow-hidden rounded-lg border bg-card">
        <ChatPanel />
      </div>
    </div>
  );
}
