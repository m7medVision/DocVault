import { NextRequest } from "next/server";
import { getAccessTokenFromRequest } from "@/lib/server-auth";
import { SERVER_API_BASE_URL } from "@/lib/auth/constants";

/**
 * Rewrite the backend SSE stream so the SOURCES event is surfaced to
 * @tanstack/ai-react. The library's stream processor silently ignores unknown
 * AG-UI event types, and only invokes `onCustomEvent` for events shaped as
 * `{ type: "CUSTOM", name, value }`. The backend emits a raw
 * `{ type: "SOURCES", messageId, sources, timestamp }` frame, so we transcode
 * it into a CUSTOM event (name "sources") while passing every other frame
 * through unchanged. The original messageId is carried inside `value` so the
 * client can attach the citations to the matching assistant message.
 */
function transformSourcesStream(
  upstream: ReadableStream<Uint8Array>
): ReadableStream<Uint8Array> {
  const decoder = new TextDecoder();
  const encoder = new TextEncoder();
  let buffer = "";

  const rewriteFrame = (frame: string): string => {
    // An SSE frame may contain comment lines and one or more `data:` lines.
    const lines = frame.split("\n");
    const dataLines = lines.filter((l) => l.startsWith("data:"));
    if (dataLines.length === 0) return frame;

    const payload = dataLines
      .map((l) => l.slice(l.startsWith("data: ") ? 6 : 5))
      .join("\n");

    try {
      const event = JSON.parse(payload);
      if (event && event.type === "SOURCES") {
        const custom = {
          type: "CUSTOM",
          name: "sources",
          value: {
            messageId: event.messageId,
            sources: event.sources,
            timestamp: event.timestamp,
          },
        };
        return `data: ${JSON.stringify(custom)}`;
      }
    } catch {
      // Not JSON (e.g. `[DONE]` or a comment) — leave untouched.
    }
    return frame;
  };

  return new ReadableStream<Uint8Array>({
    async start(controller) {
      const reader = upstream.getReader();
      try {
        for (;;) {
          const { done, value } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value, { stream: true });

          // SSE events are separated by a blank line ("\n\n").
          let sep: number;
          while ((sep = buffer.indexOf("\n\n")) !== -1) {
            const frame = buffer.slice(0, sep);
            buffer = buffer.slice(sep + 2);
            controller.enqueue(encoder.encode(rewriteFrame(frame) + "\n\n"));
          }
        }
        // Flush any trailing partial frame.
        buffer += decoder.decode();
        if (buffer.length > 0) {
          controller.enqueue(encoder.encode(rewriteFrame(buffer)));
        }
      } catch (err) {
        controller.error(err);
        return;
      } finally {
        reader.releaseLock();
      }
      controller.close();
    },
  });
}

export async function POST(req: NextRequest) {
  const accessToken = getAccessTokenFromRequest(req);
  if (!accessToken) {
    return new Response("Unauthorized", { status: 401 });
  }

  const { searchParams } = new URL(req.url);
  const documentId = searchParams.get("documentId");

  if (!documentId) {
    return new Response("Missing documentId", { status: 400 });
  }

  const body = await req.json();
  const { messages } = body as {
    messages: Array<{
      id?: string;
      role: string;
      parts?: Array<{ type: string; content?: string }>;
      content?: string;
    }>;
  };

  if (!messages || !Array.isArray(messages)) {
    return new Response("Missing messages", { status: 400 });
  }

  const normalizedMessages = messages.map((msg) => {
    if (Array.isArray(msg.parts)) {
      const text = msg.parts
        .filter((p) => p.type === "text" && typeof p.content === "string")
        .map((p) => p.content)
        .join("");
      return { role: msg.role, content: text };
    }
    return { role: msg.role, content: msg.content ?? "" };
  });

  const backendRes = await fetch(`${SERVER_API_BASE_URL}/documents/${documentId}/chat`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ messages: normalizedMessages }),
  });

  if (!backendRes.ok || !backendRes.body) {
    let errorMessage = "Chat request failed";
    try {
      const errBody = await backendRes.json();
      errorMessage = errBody.error || errorMessage;
    } catch {
      // non-JSON error
    }
    return new Response(JSON.stringify({ error: errorMessage }), {
      status: backendRes.ok ? 502 : backendRes.status,
      headers: { "Content-Type": "application/json" },
    });
  }

  return new Response(transformSourcesStream(backendRes.body), {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
