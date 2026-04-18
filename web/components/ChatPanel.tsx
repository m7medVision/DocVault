"use client";

import { useState, type FormEvent } from "react";
import { useChat, fetchServerSentEvents } from "@tanstack/ai-react";
import { useTranslations } from "next-intl";
import { Send, Loader2, Bot, User } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface ChatPanelProps {
  documentId: string;
}

export function ChatPanel({ documentId }: ChatPanelProps) {
  const t = useTranslations("chat");
  const [input, setInput] = useState("");

  const { messages, sendMessage, isLoading, error } = useChat({
    connection: fetchServerSentEvents(`/api/chat?documentId=${documentId}`),
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

        {messages.map((message) => (
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
              {message.parts?.map((part, i) => {
                if (part.type === "text") return <span key={i}>{part.content}</span>;
                return null;
              })}
            </div>
            {message.role === "user" && (
              <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary">
                <User className="size-4 text-primary-foreground" />
              </div>
            )}
          </div>
        ))}

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
