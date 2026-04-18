"use client";

import { useEffect, useRef, useState } from "react";
import { ChevronLeft, ChevronRight, FileIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const PDFJS_CDN = "https://cdn.jsdelivr.net/npm/pdfjs-dist@5.4.296/build";

interface PdfViewerProps {
  url: string;
  failedMessage: string;
  loadingMessage: string;
}

interface PDFPage {
  getViewport: (opts: { scale: number }) => {
    width: number;
    height: number;
    getViewport: (scale: number) => PDFPage["getViewport"] extends (opts: infer O) => infer R ? (scale: number) => R : never;
  };
  getTextContent: () => Promise<{ items: Array<{ str: string }> }>;
  render: (opts: {
    canvasContext: CanvasRenderingContext2D;
    viewport: ReturnType<PDFPage["getViewport"]>;
  }) => { promise: Promise<void> };
}

export function PdfViewer({ url, failedMessage, loadingMessage }: PdfViewerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const textLayerRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const [numPages, setNumPages] = useState(0);
  const [pageNumber, setPageNumber] = useState(1);
  const [pdfError, setPdfError] = useState(false);
  const [pdfLoading, setPdfLoading] = useState(true);
  const [containerWidth, setContainerWidth] = useState(800);
  const [pdfReady, setPdfReady] = useState(false);

  const pdfjsLibRef = useRef<typeof import("pdfjs-dist") | null>(null);
  const docRef = useRef<unknown>(null);
  const numPagesRef = useRef(0);

  // Init: load CDN script + resize observer
  useEffect(() => {
    const script = document.createElement("script");
    script.src = `${PDFJS_CDN}/pdf.min.mjs`;
    script.type = "module";
    script.onload = () => {
      const pdfjs = (globalThis as unknown as Record<string, unknown>).pdfjsLib as typeof import("pdfjs-dist");
      if (pdfjs) {
        pdfjs.GlobalWorkerOptions.workerSrc = `${PDFJS_CDN}/pdf.worker.min.mjs`;
        pdfjsLibRef.current = pdfjs;
        setPdfReady(true);
      }
    };
    document.head.appendChild(script);

    if (!containerRef.current) return;
    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerWidth(entry.contentRect.width);
      }
    });
    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, []);

  // Load PDF document when pdfjs is ready and url is available
  useEffect(() => {
    if (!pdfReady || !url) return;
    setPdfLoading(true);
    setPdfError(false);
    setPageNumber(1);

    const load = async () => {
      try {
        const pdf = await pdfjsLibRef.current!.getDocument(url).promise;
        docRef.current = pdf;
        numPagesRef.current = pdf.numPages;
        setNumPages(pdf.numPages);
        setPdfLoading(false);
      } catch {
        setPdfError(true);
        setPdfLoading(false);
      }
    };
    load();
  }, [pdfReady, url]);

  // Render current page when pageNumber or containerWidth changes
  useEffect(() => {
    const renderPage = async () => {
      if (!pdfjsLibRef.current || !canvasRef.current || !docRef.current) return;
      try {
        const page = await (docRef.current as { getPage: (n: number) => Promise<PDFPage> }).getPage(pageNumber);
        const viewport = page.getViewport({ scale: 1.2 }) as unknown as { width: number; height: number };
        const computedScale = (containerWidth - 4) / viewport.width;
        const scaledViewport = page.getViewport({ scale: computedScale }) as unknown as { width: number; height: number };

        const canvas = canvasRef.current;
        canvas.width = scaledViewport.width;
        canvas.height = scaledViewport.height;

        await (page as unknown as { render: (opts: { canvasContext: CanvasRenderingContext2D; viewport: unknown }) => { promise: Promise<void> } }).render({ canvasContext: canvas.getContext("2d")!, viewport: scaledViewport }).promise;

        if (textLayerRef.current) {
          textLayerRef.current.innerHTML = "";
          const textContent = await page.getTextContent();
          const textDiv = document.createElement("div");
          textDiv.style.position = "absolute";
          textDiv.style.top = "0";
          textDiv.style.left = "0";
          textDiv.style.width = `${scaledViewport.width}px`;
          textDiv.style.height = `${scaledViewport.height}px`;
          textDiv.style.fontSize = "10px";
          textDiv.style.lineHeight = "1.4";
          textDiv.style.transformOrigin = "top left";
          for (const item of textContent.items) {
            if ("str" in item) {
              const tx = document.createElement("span");
              tx.textContent = item.str as string;
              textDiv.appendChild(tx);
            }
          }
          textLayerRef.current.appendChild(textDiv);
        }
      } catch {
        setPdfError(true);
      }
    };
    if (!pdfLoading) renderPage();
  }, [pageNumber, containerWidth, pdfLoading]);

  // Keyboard navigation
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
      if (e.key === "ArrowLeft") {
        e.preventDefault();
        setPageNumber((p) => Math.max(1, p - 1));
      } else if (e.key === "ArrowRight") {
        e.preventDefault();
        setPageNumber((p) => Math.min(numPagesRef.current, p + 1));
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  if (pdfError) {
    return (
      <div className="mx-auto flex min-h-[400px] max-w-3xl flex-col items-center justify-center gap-4 rounded-lg border bg-white p-12 shadow-md">
        <FileIcon className="size-12 text-muted-foreground" />
        <p className="text-center text-sm text-muted-foreground">{failedMessage}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="flex w-full max-w-3xl items-center justify-between rounded-lg border bg-card px-4 py-2">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={() => setPageNumber((p) => Math.max(1, p - 1))}
            disabled={pageNumber <= 1}
            aria-label="Previous page"
          >
            <ChevronLeft className="size-4" />
          </Button>
          <span className="min-w-[100px] text-center text-sm tabular-nums text-muted-foreground">
            {pdfLoading ? "—" : `${pageNumber} / ${numPages}`}
          </span>
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            onClick={() => setPageNumber((p) => Math.min(numPages, p + 1))}
            disabled={pageNumber >= numPages}
            aria-label="Next page"
          >
            <ChevronRight className="size-4" />
          </Button>
        </div>
      </div>

      <div
        ref={containerRef}
        className={cn(
          "relative flex min-h-[600px] w-full max-w-3xl items-center justify-center overflow-hidden rounded-lg border bg-white shadow-md"
        )}
      >
        {pdfLoading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/80">
            <span className="text-sm text-muted-foreground">{loadingMessage}</span>
          </div>
        )}
        <canvas ref={canvasRef} className="block" />
        <div
          ref={textLayerRef}
          className="absolute left-0 top-0 overflow-hidden pointer-events-none"
          style={{ color: "black" }}
        />
      </div>
    </div>
  );
}
