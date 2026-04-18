import { NextRequest } from "next/server";
import { getAccessTokenFromRequest } from "@/lib/server-auth";
import { SERVER_API_BASE_URL } from "@/lib/auth/constants";

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
    messages: { role: string; content: string }[];
  };

  if (!messages || !Array.isArray(messages)) {
    return new Response("Missing messages", { status: 400 });
  }

  const backendRes = await fetch(
    `${SERVER_API_BASE_URL}/documents/${documentId}/chat`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${accessToken}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ messages }),
    }
  );

  if (!backendRes.ok) {
    let errorMessage = "Chat request failed";
    try {
      const errBody = await backendRes.json();
      errorMessage = errBody.error || errorMessage;
    } catch {
      // non-JSON error
    }
    return new Response(
      JSON.stringify({ error: errorMessage }),
      {
        status: backendRes.status,
        headers: { "Content-Type": "application/json" },
      }
    );
  }

  return new Response(backendRes.body, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
