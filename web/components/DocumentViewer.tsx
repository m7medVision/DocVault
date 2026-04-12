"use client";

import { useEffect, useRef } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { DocumentPage } from "@/lib/api";

interface DocumentViewerProps {
  pages: DocumentPage[];
  selectedPage: number;
  onPageChange: (index: number) => void;
  toolbarActions?: React.ReactNode;
  noOcrText?: string;
  confidenceLabel?: string;
}

export function DocumentViewer({
  pages,
  selectedPage,
  onPageChange,
  toolbarActions,
  noOcrText = "No OCR text available for this page.",
  confidenceLabel = "Confidence",
}: DocumentViewerProps) {
  const totalPages = pages.length;
  const currentPage = pages[selectedPage];
  const onPageChangeRef = useRef(onPageChange);
  onPageChangeRef.current = onPageChange;

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
      if (e.key === "ArrowLeft") {
        e.preventDefault();
        onPageChangeRef.current(Math.max(0, selectedPage - 1));
      } else if (e.key === "ArrowRight") {
        e.preventDefault();
        onPageChangeRef.current(Math.min(totalPages - 1, selectedPage + 1));
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [selectedPage, totalPages]);

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="flex w-full max-w-3xl items-center justify-between rounded-lg border bg-card px-4 py-2">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={() => onPageChange(Math.max(0, selectedPage - 1))}
            disabled={selectedPage === 0}
            aria-label="Previous page"
          >
            <ChevronLeft className="size-4" />
          </Button>
          <span className="min-w-[100px] text-center text-sm tabular-nums text-muted-foreground">
            {selectedPage + 1} / {totalPages}
          </span>
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={() => onPageChange(Math.min(totalPages - 1, selectedPage + 1))}
            disabled={selectedPage === totalPages - 1}
            aria-label="Next page"
          >
            <ChevronRight className="size-4" />
          </Button>
        </div>
        <div className="flex items-center gap-3">
          {currentPage?.confidence != null && (
            <span className="text-xs text-muted-foreground">
              {(currentPage.confidence * 100).toFixed(0)}% {confidenceLabel}
            </span>
          )}
          {toolbarActions}
        </div>
      </div>

      <div
        className={cn(
          "w-full max-w-3xl overflow-hidden rounded-lg border bg-white shadow-md",
          "min-h-[600px]"
        )}
      >
        <div className="mx-auto max-w-none px-10 py-12 [color:oklch(0.15_0.02_250)]">
          <div className="prose prose-sm max-w-none prose-headings:mb-3 prose-headings:mt-6 prose-p:leading-relaxed prose-table:text-sm">
            {currentPage?.ocr_text ? (
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {currentPage.ocr_text}
              </ReactMarkdown>
            ) : (
              <p className="italic text-muted-foreground">{noOcrText}</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
