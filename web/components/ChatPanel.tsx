"use client";

import { useCallback, useState, type FormEvent } from "react";
import { useChat, fetchServerSentEvents } from "@tanstack/ai-react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { usePathname } from "next/navigation";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Send, Loader2, Bot, User, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface ChatSource {
  n: number;
  documentId: string;
  title: string;
  page: number;
  score: number;
}

interface ChatPanelProps {
  /**
   * When provided, the chat is scoped to a single document. When omitted, the
   * chat spans the caller's whole organization (global "Ask" surface).
   */
  documentId?: string;
}

export function ChatPanel({ documentId }: ChatPanelProps) {
  const t = useTranslations("chat");
  const pathname = usePathname();
  const [input, setInput] = useState("");
  // Citations keyed by the assistant messageId they belong to. The backend
  // emits a SOURCES event (re-emitted by /api/chat as a CUSTOM "sources" event)
  // after the assistant text for a turn; absence of the event means no sources.
  const [sourcesByMessage, setSourcesByMessage] = useState<
    Record<string, ChatSource[]>
  >({});

  const localePrefix = (() => {
    const segment = pathname.split("/")[1];
    return segment === "ar" ? "/ar" : "";
  })();

  const onCustomEvent = useCallback(
    (eventType: string, data: unknown) => {
      if (eventType !== "sources" || !data || typeof data !== "object") return;
      const { messageId, sources } = data as {
        messageId?: string;
        sources?: ChatSource[];
      };
      if (!messageId || !Array.isArray(sources) || sources.length === 0) return;
      setSourcesByMessage((prev) => ({ ...prev, [messageId]: sources }));
    },
    []
  );

  const chatUrl = documentId
    ? `/api/chat?documentId=${documentId}`
    : "/api/chat";

  const { messages, sendMessage, isLoading, error } = useChat({
    connection: fetchServerSentEvents(chatUrl),
    onCustomEvent,
  });

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    const text = input.trim();
    if (!text || isLoading) return;
    setInput("");
    sendMessage(text);
  };

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4">
        {messages.length === 0 && (
          <div className="flex h-full items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("emptyState")}</p>
          </div>
        )}

        {messages.map((message) => {
          const textContent =
            message.parts
              ?.map((part) => (part.type === "text" ? part.content : ""))
              .join("") ?? "";

          const sources =
            message.role === "assistant"
              ? sortSourcesByNumber(sourcesByMessage[message.id])
              : [];

          return (
            <div
              key={message.id}
              className={`flex gap-3 ${
                message.role === "user" ? "justify-end" : "justify-start"
              }`}
            >
              {message.role === "assistant" && (
                <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10">
                  <Bot className="size-4 text-primary" />
                </div>
              )}
              <div
                className={`max-w-[80%] rounded-lg px-3 py-2 text-sm ${
                  message.role === "user"
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted"
                }`}
              >
                {message.role === "assistant" ? (
                  <>
                    <div className="prose prose-sm max-w-none break-words dark:prose-invert prose-p:my-0 prose-headings:mb-2 prose-headings:mt-4 prose-ul:my-2 prose-ol:my-2 prose-li:my-0 prose-pre:my-2 prose-code:before:content-none prose-code:after:content-none">
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {textContent}
                      </ReactMarkdown>
                    </div>
                    {sources.length > 0 && (
                      <div className="mt-2 border-t border-border/60 pt-2">
                        <p className="mb-1.5 text-xs font-medium text-muted-foreground">
                          {t("sourcesTitle")}
                        </p>
                        <div className="flex flex-wrap gap-1.5">
                          {sources.map((source) => (
                            <Link
                              key={source.n}
                              href={`${localePrefix}/documents/${source.documentId}?page=${source.page}`}
                              className="inline-flex items-center gap-1 rounded-full border bg-background px-2 py-0.5 text-xs text-foreground transition-colors hover:bg-accent"
                            >
                              <span className="font-medium text-muted-foreground tabular-nums">
                                [{source.n}]
                              </span>
                              <FileText className="size-3 shrink-0 text-muted-foreground" />
                              <span className="max-w-[14rem] truncate">
                                {source.title || t("untitledSource")}
                              </span>
                              {source.page > 0 && (
                                <span className="text-muted-foreground">
                                  {t("pageShort", { page: source.page })}
                                </span>
                              )}
                            </Link>
                          ))}
                        </div>
                      </div>
                    )}
                  </>
                ) : (
                  <span className="whitespace-pre-wrap">{textContent}</span>
                )}
              </div>
              {message.role === "user" && (
                <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary">
                  <User className="size-4 text-primary-foreground" />
                </div>
              )}
            </div>
          );
        })}

        {isLoading && (
          <div className="flex gap-3 justify-start">
            <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10">
              <Bot className="size-4 text-primary" />
            </div>
            <div className="flex items-center gap-1 rounded-lg bg-muted px-3 py-2">
              <Loader2 className="size-3 animate-spin" />
              <span className="text-xs text-muted-foreground">
                {t("thinking")}
              </span>
            </div>
          </div>
        )}
      </div>

      {error && (
        <div className="border-t px-4 py-2">
          <p className="text-xs text-destructive">
            {error.message || t("error")}
          </p>
        </div>
      )}

      <div className="border-t px-4 py-3">
        <form onSubmit={handleSubmit} className="flex gap-2">
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder={t("placeholder")}
            disabled={isLoading}
            className="flex-1"
          />
          <Button type="submit" size="icon" disabled={isLoading || !input.trim()}>
            <Send className="size-4" />
          </Button>
        </form>
      </div>
    </div>
  );
}

/**
 * Order citations by their passage number so the [n] chips line up with the
 * inline [n] markers in the assistant's answer. Every cited passage is kept
 * (numbers are unique per turn), so no inline marker is left without a chip.
 */
function sortSourcesByNumber(sources: ChatSource[] | undefined): ChatSource[] {
  if (!sources || sources.length === 0) return [];
  return [...sources].sort((a, b) => a.n - b.n);
}
